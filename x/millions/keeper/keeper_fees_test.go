package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func (suite *KeeperTestSuite) TestFees_FeeCollector() {
	app := suite.app
	ctx := suite.ctx

	// Fees should start at 0 and have the pool denom
	denom := app.StakingKeeper.BondDenom(ctx)
	feeAddr := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName).GetAddress()
	fc := app.MillionsKeeper.NewFeeManager(ctx, millionstypes.Pool{Denom: denom, LocalAddress: suite.addrs[0].String()})
	suite.Require().Equal(math.ZeroInt(), fc.CollectedAmount().Amount)
	suite.Require().Equal(denom, fc.CollectedAmount().Denom)

	// 0 fees should do nothing
	fc = app.MillionsKeeper.NewFeeManager(ctx, millionstypes.Pool{Denom: denom, LocalAddress: suite.addrs[0].String()})
	prize := &millionstypes.Prize{Amount: sdk.NewCoin(denom, math.ZeroInt())}
	a, f := fc.CollectPrizeFees(ctx, prize)
	suite.Require().Equal(math.ZeroInt(), a)
	suite.Require().Equal(math.ZeroInt(), f)

	prize.Amount.Amount = math.NewInt(123)
	a, f = fc.CollectPrizeFees(ctx, prize)
	suite.Require().Equal(math.NewInt(123), a)
	suite.Require().Equal(math.ZeroInt(), f)

	// 10% fees should store collected fees (if possible) and update prize amount
	fc = app.MillionsKeeper.NewFeeManager(ctx, millionstypes.Pool{Denom: denom, LocalAddress: suite.addrs[0].String(), FeeTakers: []millionstypes.FeeTaker{
		{Destination: authtypes.FeeCollectorName, Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeesStakers, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
	}})

	prize.Amount.Amount = math.NewInt(0)
	a, f = fc.CollectPrizeFees(ctx, prize)
	suite.Require().Equal(math.NewInt(0), f)
	suite.Require().Equal(math.NewInt(0), a)
	suite.Require().Equal(a, prize.Amount.Amount)

	prize.Amount.Amount = math.NewInt(100)
	a, f = fc.CollectPrizeFees(ctx, prize)
	suite.Require().Equal(math.NewInt(10), f)
	suite.Require().Equal(math.NewInt(90), a)
	suite.Require().Equal(a, prize.Amount.Amount)

	prize.Amount.Amount = math.NewInt(1000)
	a, f = fc.CollectPrizeFees(ctx, prize)
	suite.Require().Equal(math.NewInt(100), f)
	suite.Require().Equal(math.NewInt(900), a)
	suite.Require().Equal(a, prize.Amount.Amount)

	// Fees should add up in the collected but pending send value
	suite.Require().Equal(math.NewInt(110), fc.CollectedAmount().Amount)

	// Fees should use a closest round up or down logic in case of floating approximation
	prize.Amount.Amount = math.NewInt(99)
	a, f = fc.CollectPrizeFees(ctx, prize)
	suite.Require().Equal(math.NewInt(10), f)
	suite.Require().Equal(math.NewInt(89), a)
	suite.Require().Equal(a, prize.Amount.Amount)

	prize.Amount.Amount = math.NewInt(91)
	a, f = fc.CollectPrizeFees(ctx, prize)
	suite.Require().Equal(math.NewInt(9), f)
	suite.Require().Equal(math.NewInt(82), a)
	suite.Require().Equal(a, prize.Amount.Amount)

	// Fees should add up in the collected but pending send value
	suite.Require().Equal(math.NewInt(129), fc.CollectedAmount().Amount)

	// Succeeding at sending collected fees should send and clear the collected amount
	poolBalanceBefore := app.BankKeeper.GetBalance(ctx, suite.addrs[0], denom)
	feeCollectorBalanceBefore := app.BankKeeper.GetBalance(ctx, feeAddr, denom)

	collectedAmount := sdk.NewInt(123_456) // Manually set collected amount

	fc.CollectPrizeFees(ctx, &millionstypes.Prize{Amount: sdk.NewCoin(denom, collectedAmount.MulRaw(10).SubRaw(1290))})
	suite.Require().Equal(denom, fc.CollectedAmount().Denom)
	suite.Require().Equal(collectedAmount, fc.CollectedAmount().Amount)

	err := fc.SendCollectedFees(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(0).Int64(), fc.CollectedAmount().Amount.Int64())
	suite.Require().Equal(poolBalanceBefore.SubAmount(collectedAmount), app.BankKeeper.GetBalance(ctx, suite.addrs[0], denom))
	suite.Require().Equal(feeCollectorBalanceBefore.AddAmount(collectedAmount), app.BankKeeper.GetBalance(ctx, feeAddr, denom))

	// Trying to send collected fees again should result in a no-op
	err = fc.SendCollectedFees(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(0), fc.CollectedAmount().Amount)
	suite.Require().Equal(poolBalanceBefore.SubAmount(collectedAmount), app.BankKeeper.GetBalance(ctx, suite.addrs[0], denom))
	suite.Require().Equal(feeCollectorBalanceBefore.AddAmount(collectedAmount), app.BankKeeper.GetBalance(ctx, feeAddr, denom))

	// Stakers should receive their share of the collected fees upon send success
	// Community tax should apply as well
	vals := app.StakingKeeper.GetAllValidators(ctx)
	consAddr0, err := vals[0].GetConsAddr()
	suite.Require().NoError(err)

	suite.Require().Equal(sdk.ZeroDec(), app.DistrKeeper.GetTotalRewards(ctx).AmountOf(denom))
	suite.Require().Equal(sdk.ZeroDec(), app.DistrKeeper.GetFeePoolCommunityCoins(ctx).AmountOf(denom))
	votes := []abci.VoteInfo{
		{
			Validator:       abci.Validator{Address: consAddr0.Bytes(), Power: 100},
			SignedLastBlock: true,
		},
	}
	app.DistrKeeper.AllocateTokens(ctx, 100, votes)
	comTax := app.DistrKeeper.GetCommunityTax(ctx)
	suite.Require().Equal(sdk.NewDec(1).MulInt(collectedAmount).Sub(comTax.MulInt(collectedAmount)), app.DistrKeeper.GetTotalRewards(ctx).AmountOf(denom))
	suite.Require().Equal(comTax.MulInt(collectedAmount), app.DistrKeeper.GetFeePoolCommunityCoins(ctx).AmountOf(denom))
}

func (suite *KeeperTestSuite) TestFees_DrawPrizesFees() {
	app := suite.app
	ctx := suite.ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)

	// init pool
	p := *newValidPool(suite, millionstypes.Pool{
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 100, DrawProbability: floatToDec(1.00)},
			},
		},
	})
	poolID, err := app.MillionsKeeper.RegisterPool(
		ctx,
		millionstypes.PoolType_Staking,
		p.Denom, p.NativeDenom, p.ChainId, p.ConnectionId, p.TransferChannelId,
		[]string{suite.valAddrs[0].String()},
		p.Bech32PrefixAccAddr, p.Bech32PrefixValAddr,
		p.MinDepositAmount,
		p.UnbondingDuration,
		p.MaxUnbondingEntries,
		p.DrawSchedule,
		p.PrizeStrategy,
		p.FeeTakers,
	)
	suite.Require().NoError(err)
	p, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)

	// make a deposit for the draw to draw prizes
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(24 * time.Hour))
	_, err = msgServer.Deposit(goCtx, millionstypes.NewMsgDeposit(
		suite.addrs[0].String(),
		sdk.NewCoin(p.Denom, sdk.NewInt(1_000_000)),
		p.PoolId,
	))
	suite.Require().NoError(err)

	// send amount to have a prize pool
	var prizePoolAmount int64 = 6_700_800_000
	err = app.BankKeeper.SendCoins(
		ctx,
		suite.addrs[0],
		sdk.MustAccAddressFromBech32(p.IcaPrizepoolAddress),
		sdk.NewCoins(sdk.NewCoin(p.Denom, sdk.NewInt(prizePoolAmount))),
	)
	suite.Require().NoError(err)

	// force run draw and fee collection
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(24 * time.Hour))
	draw, err := app.MillionsKeeper.LaunchNewDraw(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(prizePoolAmount, draw.PrizePoolFreshAmount.Int64())
	suite.Require().Equal(prizePoolAmount, draw.TotalWinAmount.Int64())
	// prize ref should reflect the total win amount
	suite.Require().Len(draw.PrizesRefs, 100)
	suite.Require().Equal(prizePoolAmount/100, draw.PrizesRefs[0].Amount.Int64())
	// prize entity should have fees subtracted
	prizes := app.MillionsKeeper.ListPrizes(ctx)
	suite.Require().Len(prizes, 100)
	suite.Require().Equal(prizePoolAmount/100-p.FeeTakers[0].Amount.MulInt64(prizePoolAmount/100).RoundInt64(), prizes[0].Amount.Amount.Int64())

	// Stakers should receive their share of the collected fees upon send success
	// Community tax should apply as well
	collectedAmount := p.FeeTakers[0].Amount.MulInt64(prizePoolAmount).RoundInt64()
	vals := app.StakingKeeper.GetAllValidators(ctx)
	consAddr0, err := vals[0].GetConsAddr()
	suite.Require().NoError(err)

	suite.Require().Equal(sdk.ZeroDec(), app.DistrKeeper.GetTotalRewards(ctx).AmountOf(p.Denom))
	suite.Require().Equal(sdk.ZeroDec(), app.DistrKeeper.GetFeePoolCommunityCoins(ctx).AmountOf(p.Denom))
	votes := []abci.VoteInfo{
		{
			Validator:       abci.Validator{Address: consAddr0.Bytes(), Power: 100},
			SignedLastBlock: true,
		},
	}
	app.DistrKeeper.AllocateTokens(ctx, 100, votes)
	comTax := app.DistrKeeper.GetCommunityTax(ctx)
	suite.Require().Equal(sdk.NewDec(1).MulInt64(collectedAmount).Sub(comTax.MulInt64(collectedAmount)), app.DistrKeeper.GetTotalRewards(ctx).AmountOf(p.Denom))
	suite.Require().Equal(comTax.MulInt64(collectedAmount), app.DistrKeeper.GetFeePoolCommunityCoins(ctx).AmountOf(p.Denom))
}
