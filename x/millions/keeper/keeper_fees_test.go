package keeper_test

import (
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/lum-network/chain/x/millions/types"
)

func (suite *KeeperTestSuite) TestFees_FeeCollector() {
	app := suite.app
	ctx := suite.ctx

	// Fees should start at 0 and have the pool denom
	denom := app.StakingKeeper.BondDenom(ctx)
	feeAddr := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName).GetAddress()
	fc := app.MillionsKeeper.NewFeeCollector(ctx, types.Pool{Denom: denom, LocalAddress: suite.addrs[0].String()})
	suite.Require().Equal(math.ZeroInt(), fc.CollectedAmount().Amount)
	suite.Require().Equal(denom, fc.CollectedAmount().Denom)

	// 0 fees should do nothing
	params := app.MillionsKeeper.GetParams(ctx)
	params.FeesStakers = sdk.ZeroDec()
	app.MillionsKeeper.SetParams(ctx, params)

	fc = app.MillionsKeeper.NewFeeCollector(ctx, types.Pool{Denom: denom, LocalAddress: suite.addrs[0].String()})
	prize := &types.Prize{Amount: sdk.NewCoin(denom, math.ZeroInt())}
	a, f := fc.CollectPrizeFees(ctx, prize)
	suite.Require().Equal(math.ZeroInt(), a)
	suite.Require().Equal(math.ZeroInt(), f)

	prize.Amount.Amount = math.NewInt(123)
	a, f = fc.CollectPrizeFees(ctx, prize)
	suite.Require().Equal(math.NewInt(123), a)
	suite.Require().Equal(math.ZeroInt(), f)

	// 10% fees should store collected fees (if possible) and update prize amount
	params.FeesStakers = floatToDec(0.1)
	app.MillionsKeeper.SetParams(ctx, params)
	fc = app.MillionsKeeper.NewFeeCollector(ctx, types.Pool{Denom: denom, LocalAddress: suite.addrs[0].String()})

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
	collectedAmount := sdk.NewInt(123_456)
	fc.CollectPrizeFees(ctx, &types.Prize{Amount: sdk.NewCoin(denom, collectedAmount.MulRaw(10).SubRaw(1290))})
	suite.Require().Equal(denom, fc.CollectedAmount().Denom)
	suite.Require().Equal(collectedAmount, fc.CollectedAmount().Amount)
	err := fc.SendCollectedFees(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(0), fc.CollectedAmount().Amount)
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
	app.DistrKeeper.AllocateTokens(ctx, 100, 100, consAddr0, votes)
	comTax := app.DistrKeeper.GetCommunityTax(ctx)
	suite.Require().Equal(sdk.NewDec(1).MulInt(collectedAmount).Sub(comTax.MulInt(collectedAmount)), app.DistrKeeper.GetTotalRewards(ctx).AmountOf(denom))
	suite.Require().Equal(comTax.MulInt(collectedAmount), app.DistrKeeper.GetFeePoolCommunityCoins(ctx).AmountOf(denom))
}
