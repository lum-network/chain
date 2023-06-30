package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distribtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	apptesting "github.com/lum-network/chain/app/testing"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

// TestDraw_NoPrizeToWin tests cases where no prize can be computed and therefore no winner
func (suite *KeeperTestSuite) TestDraw_NoPrizeToWin() {
	app := suite.app
	ctx := suite.ctx

	seed := int64(42)

	// Empty prize pool should return 0 prizes
	prizePool := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.ZeroInt())
	result, err := app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 20, Quantity: 1, DrawProbability: floatToDec(0.01)},
				{PoolPercent: 30, Quantity: 1, DrawProbability: floatToDec(0.11)},
				{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
			},
		},
		[]millionskeeper.DepositTWB{},
		seed,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroInt(), result.TotalWinAmount)
	suite.Require().Equal(uint64(0), result.TotalWinCount)
	suite.Require().Len(result.PrizeDraws, 0)

	result, err = app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 20, Quantity: 1, DrawProbability: floatToDec(0.01)},
				{PoolPercent: 30, Quantity: 1, DrawProbability: floatToDec(0.11)},
				{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
			},
		},
		[]millionskeeper.DepositTWB{
			{Address: suite.addrs[0].String(), Amount: sdk.NewInt(1_000_000)},
			{Address: suite.addrs[1].String(), Amount: sdk.NewInt(1_000_000)},
			{Address: suite.addrs[2].String(), Amount: sdk.NewInt(1_000_000)},
		},
		seed,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroInt(), result.TotalWinAmount)
	suite.Require().Equal(uint64(0), result.TotalWinCount)
	suite.Require().Len(result.PrizeDraws, 0)

	// No depositors should return prizes without winners
	prizePool = sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000))
	result, err = app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 20, Quantity: 1, DrawProbability: floatToDec(0.01)},
				{PoolPercent: 30, Quantity: 1, DrawProbability: floatToDec(0.11)},
				{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
			},
		},
		[]millionskeeper.DepositTWB{},
		seed,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroInt(), result.TotalWinAmount)
	suite.Require().Equal(uint64(0), result.TotalWinCount)
	suite.Require().Len(result.PrizeDraws, 3)
}

// TestDraw_NoPrizesWon tests the outcome of winnable prizes without winners
func (suite *KeeperTestSuite) TestDraw_NoPrizesWon() {
	app := suite.app
	ctx := suite.ctx

	seed := int64(84)

	// No depositors should return prizes without winners
	prizePool := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000))
	result, err := app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 20, Quantity: 1, DrawProbability: floatToDec(0.01)},
				{PoolPercent: 30, Quantity: 1, DrawProbability: floatToDec(0.11)},
				{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
			},
		},
		[]millionskeeper.DepositTWB{},
		seed,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroInt(), result.TotalWinAmount)
	suite.Require().Equal(uint64(0), result.TotalWinCount)
	suite.Require().Len(result.PrizeDraws, 3)
	suite.Require().Nil(result.PrizeDraws[0].Winner)
	suite.Require().Nil(result.PrizeDraws[1].Winner)
	suite.Require().Nil(result.PrizeDraws[2].Winner)

	// Unlikely odds to win despite a large number of draws (deterministic due to the seed used)
	result, err = app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 20, Quantity: 1000, DrawProbability: floatToDec(0.00001)},
				{PoolPercent: 30, Quantity: 100, DrawProbability: floatToDec(0.00001)},
				{PoolPercent: 50, Quantity: 10, DrawProbability: floatToDec(0.00001)},
			},
		},
		[]millionskeeper.DepositTWB{
			{Address: suite.addrs[0].String(), Amount: sdk.NewInt(1_000_000)},
			{Address: suite.addrs[1].String(), Amount: sdk.NewInt(1_000_000)},
			{Address: suite.addrs[2].String(), Amount: sdk.NewInt(1_000_000)},
		},
		seed,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroInt(), result.TotalWinAmount)
	suite.Require().Equal(uint64(0), result.TotalWinCount)
	suite.Require().Len(result.PrizeDraws, 1110)
	for i, pd := range result.PrizeDraws {
		suite.Require().Nil(pd.Winner, "prize draw %d unexpected winner", i)
	}
}

// TestDraw_PrizeComputation tests rounding approximation on prizes computations
func (suite *KeeperTestSuite) TestDraw_PrizeComputation() {
	app := suite.app
	ctx := suite.ctx
	params := app.MillionsKeeper.GetParams(ctx)

	denom := app.StakingKeeper.BondDenom(ctx)

	pb1 := millionstypes.PrizeBatch{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(1.0)}
	suite.Require().NoError(pb1.Validate(params))
	suite.Require().Equal(int64(0), pb1.GetTotalPrizesAmount(sdk.NewCoin(denom, sdk.NewInt(0))).Int64())
	suite.Require().Equal(int64(100), pb1.GetTotalPrizesAmount(sdk.NewCoin(denom, sdk.NewInt(100))).Int64())
	suite.Require().Equal(int64(0), pb1.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(0))).Int64())
	suite.Require().Equal(int64(100), pb1.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(100))).Int64())

	pb2 := millionstypes.PrizeBatch{PoolPercent: 100, Quantity: 100, DrawProbability: floatToDec(1.0)}
	suite.Require().NoError(pb1.Validate(params))
	suite.Require().Equal(int64(0), pb2.GetTotalPrizesAmount(sdk.NewCoin(denom, sdk.NewInt(0))).Int64())
	suite.Require().Equal(int64(100), pb2.GetTotalPrizesAmount(sdk.NewCoin(denom, sdk.NewInt(100))).Int64())
	suite.Require().Equal(int64(0), pb2.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(0))).Int64())
	suite.Require().Equal(int64(1), pb2.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(100))).Int64())
	suite.Require().Equal(int64(0), pb2.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(99))).Int64())
	suite.Require().Equal(int64(1), pb2.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(101))).Int64())
	suite.Require().Equal(int64(9), pb2.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(990))).Int64())

	pb3 := millionstypes.PrizeBatch{PoolPercent: 1, Quantity: 1, DrawProbability: floatToDec(1.0)}
	suite.Require().NoError(pb1.Validate(params))
	suite.Require().Equal(int64(0), pb3.GetTotalPrizesAmount(sdk.NewCoin(denom, sdk.NewInt(0))).Int64())
	suite.Require().Equal(int64(1), pb3.GetTotalPrizesAmount(sdk.NewCoin(denom, sdk.NewInt(100))).Int64())
	suite.Require().Equal(int64(0), pb3.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(0))).Int64())
	suite.Require().Equal(int64(1), pb3.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(100))).Int64())
	suite.Require().Equal(int64(0), pb3.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(99))).Int64())
	suite.Require().Equal(int64(1), pb3.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(101))).Int64())
	suite.Require().Equal(int64(9), pb3.GetPrizeAmount(sdk.NewCoin(denom, sdk.NewInt(990))).Int64())

	// We expect the prizes to be rounded down as a safety measure
	// The rounding happens at total prize batch computation and at prize unit value computation
	// Meaning that:
	// - pb1 -> 990 * 98 / 100 = 970 -> 970 / 100 = 9
	// - pb2 -> 990 * 1 / 100 = 9 -> 9 / 10 = 0
	// - pb3 -> 990 * 1 / 100 = 9 -> 9 / 1 = 9
	// Totals:
	// - pb1 -> 100 items for 900 total value
	// - pb2 -> 0 items for 0 total value
	// - pb3 -> 1 item for 9 total value
	ps1 := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 98, Quantity: 100, DrawProbability: floatToDec(1.0)},
			{PoolPercent: 1, Quantity: 10, DrawProbability: floatToDec(1.0)},
			{PoolPercent: 1, Quantity: 1, DrawProbability: floatToDec(1.0)},
		},
	}
	suite.Require().NoError(ps1.Validate(params))
	ps1PrizesProbs, usedAmount, remainingAmount, err := ps1.ComputePrizesProbs(sdk.NewCoin(denom, sdk.NewInt(990)))
	ps1PrizesSum := int64(0)
	for _, p := range ps1PrizesProbs {
		ps1PrizesSum += p.Amount.Int64()
	}
	suite.Require().NoError(err)
	suite.Require().Len(ps1PrizesProbs, 101)
	suite.Require().Equal(usedAmount.Int64(), ps1PrizesSum)
	suite.Require().Equal(int64(909), usedAmount.Int64())
	suite.Require().Equal(int64(81), remainingAmount.Int64())

	// Same configuration with large numbers should not result in rounding approximation
	ps2 := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 98, Quantity: 100, DrawProbability: floatToDec(1.0)},
			{PoolPercent: 1, Quantity: 10, DrawProbability: floatToDec(1.0)},
			{PoolPercent: 1, Quantity: 1, DrawProbability: floatToDec(1.0)},
		},
	}
	suite.Require().NoError(ps1.Validate(params))
	ps2PrizesProbs, usedAmount, remainingAmount, err := ps2.ComputePrizesProbs(sdk.NewCoin(denom, sdk.NewInt(990000)))
	ps2PrizesSum := int64(0)
	for _, p := range ps2PrizesProbs {
		ps2PrizesSum += p.Amount.Int64()
	}
	suite.Require().NoError(err)
	suite.Require().Len(ps2PrizesProbs, 111)
	suite.Require().Equal(usedAmount.Int64(), ps2PrizesSum)
	suite.Require().Equal(int64(990000), usedAmount.Int64())
	suite.Require().Equal(int64(0), remainingAmount.Int64())
}

// TestDraw_UnusualPrizeCases tests if unusual cases lead to predictable behaviours
func (suite *KeeperTestSuite) TestDraw_UnusualPrizeCases() {
	app := suite.app
	ctx := suite.ctx

	seed := int64(336)

	// No depositors post TWB should result in no winner
	prizePool := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000))
	result, err := app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 100, DrawProbability: floatToDec(1.0)},
			},
		},
		[]millionskeeper.DepositTWB{
			{Address: suite.addrs[0].String(), Amount: sdk.NewInt(0)},
			{Address: suite.addrs[0].String(), Amount: sdk.NewInt(0)},
			{Address: suite.addrs[1].String(), Amount: sdk.NewInt(0)},
		},
		seed,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), result.TotalWinCount)
	suite.Require().Len(result.PrizeDraws, 100)
	for _, pd := range result.PrizeDraws {
		suite.Require().Nil(pd.Winner)
	}

	// Lone depositors should win with certainty
	prizePool = sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000))
	result, err = app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 100, DrawProbability: floatToDec(1.0)},
			},
		},
		[]millionskeeper.DepositTWB{
			{Address: suite.addrs[0].String(), Amount: sdk.NewInt(1_000_000)},
		},
		seed,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(100), result.TotalWinCount)
	suite.Require().Len(result.PrizeDraws, 100)
	for _, pd := range result.PrizeDraws {
		suite.Require().NotNil(pd.Winner)
		suite.Require().Equal(suite.addrs[0].String(), pd.Winner.Address)
	}

	// Small deposits and prize pool should be possible even though they are not recommended
	prizePool = sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1))
	result, err = app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 1, Quantity: 1, DrawProbability: floatToDec(1.0)},
				{PoolPercent: 1, Quantity: 1, DrawProbability: floatToDec(1.0)},
				{PoolPercent: 98, Quantity: 100, DrawProbability: floatToDec(1.0)},
			},
		},
		[]millionskeeper.DepositTWB{
			{Address: suite.addrs[0].String(), Amount: sdk.NewInt(1)},
			{Address: suite.addrs[1].String(), Amount: sdk.NewInt(1)},
			{Address: suite.addrs[2].String(), Amount: sdk.NewInt(1)},
		},
		seed,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), result.TotalWinCount)
	suite.Require().Equal(int64(0), result.TotalWinAmount.Int64())

	prizePool = sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(100))
	result, err = app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 1, Quantity: 1, DrawProbability: floatToDec(1.0)},
				{PoolPercent: 1, Quantity: 1, DrawProbability: floatToDec(1.0)},
				{PoolPercent: 98, Quantity: 100, DrawProbability: floatToDec(1.0)},
			},
		},
		[]millionskeeper.DepositTWB{
			{Address: suite.addrs[0].String(), Amount: sdk.NewInt(1)},
			{Address: suite.addrs[1].String(), Amount: sdk.NewInt(1)},
			{Address: suite.addrs[2].String(), Amount: sdk.NewInt(1)},
		},
		seed,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(2), result.TotalWinCount)
	suite.Require().Equal(int64(2), result.TotalWinAmount.Int64())

	// Large deposits and prize pool should work as expected
	prizePool = sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(int64(^uint64(0)>>1)))
	result, err = app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 2, Quantity: 1, DrawProbability: floatToDec(1.0)},
				{PoolPercent: 3, Quantity: 1, DrawProbability: floatToDec(1.0)},
				{PoolPercent: 95, Quantity: 1_000_000, DrawProbability: floatToDec(0.0)},
			},
		},
		[]millionskeeper.DepositTWB{
			{Address: suite.addrs[0].String(), Amount: prizePool.Amount.QuoRaw(3)},
			{Address: suite.addrs[1].String(), Amount: prizePool.Amount.QuoRaw(3)},
			{Address: suite.addrs[2].String(), Amount: prizePool.Amount.QuoRaw(3)},
		},
		seed,
	)
	suite.Require().Equal(int64(8762203435012), millionstypes.PrizeBatch{PoolPercent: 95, Quantity: 1_000_000, DrawProbability: floatToDec(0.0)}.GetPrizeAmount(prizePool).Int64())
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(2), result.TotalWinCount)
	suite.Require().Equal(prizePool.Amount.QuoRaw(100).MulRaw(5).Int64(), result.TotalWinAmount.Int64())
}

// TestDraw_PrizesOrdering tests prizes descending ordering by amount
func (suite *KeeperTestSuite) TestDraw_PrizesOrdering() {
	app := suite.app
	ctx := suite.ctx

	seed := int64(168)

	// Prizes should always be order by value descending
	prizePool := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000))
	result, err := app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 20, Quantity: 1000, DrawProbability: floatToDec(0.60)},
				{PoolPercent: 30, Quantity: 1000, DrawProbability: floatToDec(0.60)},
				{PoolPercent: 50, Quantity: 1000, DrawProbability: floatToDec(0.60)},
			},
		},
		[]millionskeeper.DepositTWB{
			{Address: suite.addrs[0].String(), Amount: sdk.NewInt(1_000_000)},
			{Address: suite.addrs[1].String(), Amount: sdk.NewInt(1_000_000)},
			{Address: suite.addrs[2].String(), Amount: sdk.NewInt(1_000_000)},
		},
		seed,
	)
	suite.Require().NoError(err)
	// wins should likely to be >= 55% <= 65% since all draws have 60% win chance
	// not a very strong test condition but still helpful to debug early faulty implementations
	suite.Require().True(result.TotalWinAmount.GTE(prizePool.Amount.MulRaw(55).QuoRaw(100)))
	suite.Require().True(result.TotalWinAmount.LTE(prizePool.Amount.MulRaw(65).QuoRaw(100)))
	suite.Require().True(result.TotalWinCount >= 3000*55/100)
	suite.Require().True(result.TotalWinCount <= 3000*65/100)
	suite.Require().Len(result.PrizeDraws, 3000)
	for i := 1; i < len(result.PrizeDraws); i++ {
		suite.Require().True(result.PrizeDraws[i].Amount.LTE(result.PrizeDraws[i-1].Amount))
	}
}

// TestDraw_PrizesDrawDeterminism tests prizes draw determinism using the same input parameters
// Each draw using the same parameters should be identical (deterministic)
func (suite *KeeperTestSuite) TestDraw_PrizesDrawDeterminism() {
	app := suite.app
	ctx := suite.ctx

	prizePool := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(42_000_000_000))
	depositors := []millionskeeper.DepositTWB{}
	for i := 0; i < 100; i++ {
		uid := fmt.Sprintf("small_%d", i)
		depositors = append(depositors, millionskeeper.DepositTWB{Address: uid, Amount: sdk.NewInt(1_000_000)})
	}

	// Init stable strategy for all draws
	strat := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 100, Quantity: 100, DrawProbability: floatToDec(1.0)},
		},
	}

	seed := int64(42)
	firstDraw, err := app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		strat,
		depositors,
		seed,
	)
	suite.Require().NoError(err)

	for i := 0; i < 100; i++ {
		draw, err := app.MillionsKeeper.RunDrawPrizes(ctx,
			prizePool,
			strat,
			depositors,
			seed,
		)
		suite.Require().NoError(err)
		suite.Require().Equal(firstDraw.TotalWinAmount.Int64(), draw.TotalWinAmount.Int64())
		suite.Require().Equal(firstDraw.TotalWinCount, draw.TotalWinCount)
		suite.Require().Equal(len(firstDraw.PrizeDraws), len(draw.PrizeDraws))
		for j, pd := range draw.PrizeDraws {
			suite.Require().Equal(firstDraw.PrizeDraws[j].Amount.Int64(), pd.Amount.Int64())
			if firstDraw.PrizeDraws[j].Winner == nil {
				suite.Require().Nil(pd.Winner)
			} else {
				suite.Require().Equal(firstDraw.PrizeDraws[j].Winner.Address, pd.Winner.Address)
				suite.Require().Equal(firstDraw.PrizeDraws[j].Winner.Amount.Int64(), pd.Winner.Amount.Int64())
			}
		}
	}
}

// TestDraw_PrizePoolPersistence tests propagation of prize pool sources (fresh, clawback, remains) for subsequent draws
func (suite *KeeperTestSuite) TestDraw_PrizePoolPersistence() {
	app := suite.app
	ctx := suite.ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)

	// Force save test pools
	p := newValidPool(suite, millionstypes.Pool{
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 100, DrawProbability: floatToDec(1.00)},
			},
		},
	})
	newID, err := app.MillionsKeeper.RegisterPool(
		ctx,
		millionstypes.PoolType_Staking,
		p.Denom, p.NativeDenom, p.ChainId, p.ConnectionId, p.TransferChannelId,
		[]string{suite.valAddrs[0].String()},
		p.Bech32PrefixAccAddr, p.Bech32PrefixValAddr,
		p.MinDepositAmount,
		p.DrawSchedule,
		p.PrizeStrategy,
	)
	suite.Require().NoError(err)
	pool1, err := app.MillionsKeeper.GetPool(ctx, newID)
	suite.Require().NoError(err)

	p = newValidPool(suite, millionstypes.Pool{
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 50, Quantity: 100, DrawProbability: floatToDec(0.00)},
				{PoolPercent: 50, Quantity: 200, DrawProbability: floatToDec(1.00)},
			},
		},
	})
	newID, err = app.MillionsKeeper.RegisterPool(
		ctx,
		millionstypes.PoolType_Staking,
		p.Denom, p.NativeDenom, p.ChainId, p.ConnectionId, p.TransferChannelId,
		[]string{suite.valAddrs[0].String()},
		p.Bech32PrefixAccAddr, p.Bech32PrefixValAddr,
		p.MinDepositAmount,
		p.DrawSchedule,
		p.PrizeStrategy,
	)
	suite.Require().NoError(err)
	pool2, err := app.MillionsKeeper.GetPool(ctx, newID)
	suite.Require().NoError(err)

	// Pools should have 0 draw
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 0)

	// Run 2 draws for pool #2
	// - Create one deposit for pool #2 which will allow to draw prizes (no prize pool = no prizes)
	// - Run reward distribution
	// - Run draw #1 which should attribute 50% of the rewards
	// - Run draw #2 which has no reward but should get the 50% rewards remaining from draw #1
	resp, err := msgServer.Deposit(goCtx, millionstypes.NewMsgDeposit(
		suite.addrs[0].String(),
		sdk.NewCoin(pool2.Denom, sdk.NewInt(1_000_000)),
		pool2.PoolId,
	))
	suite.Require().NoError(err)
	deposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, pool2.PoolId, resp.DepositId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DepositState_Success, deposit.State)
	pool2Delegations := app.StakingKeeper.GetDelegatorDelegations(ctx, sdk.MustAccAddressFromBech32(pool2.IcaDepositAddress), 100)
	suite.Require().Len(pool2Delegations, 1)

	// Move to next block and next draw
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(pool2.DrawSchedule.DrawDelta)).WithBlockHeight(ctx.BlockHeight() + 1)

	// Manually trigger reward distribution phases
	for _, val := range pool2.Validators {
		validator, found := app.StakingKeeper.GetValidator(ctx, val.MustValAddressFromBech32())
		suite.Require().True(found)

		// Attribute delegation
		staking.EndBlocker(ctx, app.StakingKeeper)
		err = app.DistrKeeper.Hooks().AfterDelegationModified(ctx, sdk.MustAccAddressFromBech32(pool2.IcaDepositAddress), val.MustValAddressFromBech32())
		suite.Require().NoError(err)

		// Simulate inflation minted
		err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(pool2.Denom, sdk.NewInt(1_000_000))))
		suite.Require().NoError(err)
		err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, distribtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(pool2.Denom, sdk.NewInt(1_000_000))))
		suite.Require().NoError(err)
		app.DistrKeeper.AllocateTokensToValidator(ctx, validator, sdk.NewDecCoins(sdk.NewDecCoin(pool2.Denom, sdk.NewInt(1_000_000))))
	}

	// Move to next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// Run draw #1
	p2Draw1, err := app.MillionsKeeper.LaunchNewDraw(ctx, pool2.PoolId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_Success, p2Draw1.State)
	suite.Require().Equal(pool2.PoolId, p2Draw1.PoolId)
	suite.Require().Len(p2Draw1.PrizesRefs, 200)
	suite.Require().Equal(200, int(p2Draw1.TotalWinCount))
	suite.Require().Equal(p2Draw1.PrizePool.Amount.Int64()/2, p2Draw1.TotalWinAmount.Int64())

	// Run draw #2
	p2Draw2, err := app.MillionsKeeper.LaunchNewDraw(ctx, pool2.PoolId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_Success, p2Draw2.State)
	suite.Require().Equal(pool2.PoolId, p2Draw2.PoolId)
	suite.Require().Len(p2Draw2.PrizesRefs, 200)
	suite.Require().Equal(200, int(p2Draw2.TotalWinCount))
	suite.Require().Equal(p2Draw1.PrizePool.Amount.Int64()/2, p2Draw2.PrizePoolRemainsAmount.Int64())
	suite.Require().Equal(p2Draw2.PrizePool.Amount.Int64()/2, p2Draw2.TotalWinAmount.Int64())
	suite.Require().Equal(p2Draw1.PrizePool.Amount.Int64()/4, p2Draw2.TotalWinAmount.Int64())

	_, err = app.MillionsKeeper.GetPoolDraw(ctx, pool2.PoolId, 3)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolDrawNotFound)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 2)
	suite.Require().Len(app.MillionsKeeper.ListPoolDraws(ctx, pool2.PoolId), 2)

	// Pool #1 should have 0 draw
	suite.Require().Len(app.MillionsKeeper.ListPoolDraws(ctx, pool1.PoolId), 0)

	// Run 3 draws for pool #1 which should have no effect since their is no depositor
	_, err = msgServer.Deposit(goCtx, millionstypes.NewMsgDeposit(
		suite.addrs[0].String(),
		sdk.NewCoin(pool1.Denom, sdk.NewInt(1_000_000)),
		pool1.PoolId,
	))
	suite.Require().NoError(err)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1_000)
	p1Draw1, err := app.MillionsKeeper.LaunchNewDraw(ctx, pool1.PoolId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_Success, p1Draw1.State)

	_, err = msgServer.Deposit(goCtx, millionstypes.NewMsgDeposit(
		suite.addrs[0].String(),
		sdk.NewCoin(pool1.Denom, sdk.NewInt(1_000_000)),
		pool1.PoolId,
	))
	suite.Require().NoError(err)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1_000)
	p1Draw2, err := app.MillionsKeeper.LaunchNewDraw(ctx, pool1.PoolId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_Success, p1Draw2.State)

	_, err = msgServer.Deposit(goCtx, millionstypes.NewMsgDeposit(
		suite.addrs[0].String(),
		sdk.NewCoin(pool1.Denom, sdk.NewInt(1_000_000)),
		pool1.PoolId,
	))
	suite.Require().NoError(err)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1_000)
	p1Draw3, err := app.MillionsKeeper.LaunchNewDraw(ctx, pool1.PoolId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_Success, p1Draw3.State)

	for i := 1; i <= 3; i++ {
		draw, err := app.MillionsKeeper.GetPoolDraw(ctx, pool1.PoolId, uint64(i))
		suite.Require().NoError(err)
		suite.Require().Equal(pool1.PoolId, draw.PoolId)
		suite.Require().Equal(uint64(i), draw.DrawId)
		suite.Require().NotEqual(int64(0), draw.RandSeed)
		suite.Require().Len(draw.PrizesRefs, 0)
		suite.Require().Equal(uint64(0), draw.TotalWinCount)
		suite.Require().Equal(sdk.ZeroInt(), draw.TotalWinAmount)
	}
	_, err = app.MillionsKeeper.GetPoolDraw(ctx, pool1.PoolId, 4)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolDrawNotFound)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 5)
	suite.Require().Len(app.MillionsKeeper.ListPoolDraws(ctx, pool1.PoolId), 3)

	// Pool #2 should still have 2 draws
	for i := 1; i <= 2; i++ {
		draw, err := app.MillionsKeeper.GetPoolDraw(ctx, pool2.PoolId, uint64(i))
		suite.Require().NoError(err)
		suite.Require().Equal(pool2.PoolId, draw.PoolId)
		suite.Require().Equal(uint64(i), draw.DrawId)
		suite.Require().NotEqual(int64(0), draw.RandSeed)
	}
	_, err = app.MillionsKeeper.GetPoolDraw(ctx, pool2.PoolId, 3)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolDrawNotFound)
	suite.Require().Len(app.MillionsKeeper.ListPoolDraws(ctx, pool2.PoolId), 2)

	// Draws should be ordered by alphabetical order (poolID then drawID)
	allDraws := app.MillionsKeeper.ListDraws(ctx)
	suite.Require().Equal(pool1.PoolId, allDraws[0].PoolId)
	suite.Require().Equal(uint64(1), allDraws[0].DrawId)
	suite.Require().Equal(pool1.PoolId, allDraws[1].PoolId)
	suite.Require().Equal(uint64(2), allDraws[1].DrawId)
	suite.Require().Equal(pool1.PoolId, allDraws[2].PoolId)
	suite.Require().Equal(uint64(3), allDraws[2].DrawId)
	suite.Require().Equal(pool2.PoolId, allDraws[3].PoolId)
	suite.Require().Equal(uint64(1), allDraws[3].DrawId)
	suite.Require().Equal(pool2.PoolId, allDraws[4].PoolId)
	suite.Require().Equal(uint64(2), allDraws[4].DrawId)

	draws1 := app.MillionsKeeper.ListPoolDraws(ctx, pool1.PoolId)
	suite.Require().Equal(pool1.PoolId, draws1[0].PoolId)
	suite.Require().Equal(uint64(1), draws1[0].DrawId)
	suite.Require().Equal(pool1.PoolId, draws1[1].PoolId)
	suite.Require().Equal(uint64(2), draws1[1].DrawId)
	suite.Require().Equal(pool1.PoolId, draws1[2].PoolId)
	suite.Require().Equal(uint64(3), draws1[2].DrawId)

	draws2 := app.MillionsKeeper.ListPoolDraws(ctx, pool2.PoolId)
	suite.Require().Equal(pool2.PoolId, draws2[0].PoolId)
	suite.Require().Equal(uint64(1), draws2[0].DrawId)
	suite.Require().Equal(pool2.PoolId, draws2[1].PoolId)
	suite.Require().Equal(uint64(2), draws2[1].DrawId)
}

// TestDraw_TimeWeightedBalance tests TWB enforcement at draw level
func (suite *KeeperTestSuite) TestDraw_TimeWeightedBalance() {
	app := suite.app
	ctx := suite.ctx

	// Init fake pool
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	pool := newValidPool(suite, millionstypes.Pool{
		PoolId: poolID,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 50, Quantity: 100, DrawProbability: floatToDec(0.50)},
				{PoolPercent: 50, Quantity: 200, DrawProbability: floatToDec(0.70)},
			},
		},
	})
	app.MillionsKeeper.AddPool(ctx, pool)

	// Simulate deposits at various time for first draw
	t0 := time.Now().UTC()
	denom := app.StakingKeeper.BondDenom(ctx)
	dState := millionstypes.DepositState_Success
	depositsTWB := app.MillionsKeeper.ComputeDepositsTWB(
		ctx,
		t0.Add(-1_000_000*time.Second),
		t0,
		[]millionstypes.Deposit{
			{State: dState, CreatedAt: t0.Add(-2_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(-1_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(-500_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(-250_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(0 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(1_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, IsSponsor: true, WinnerAddress: "6", DepositorAddress: "6", CreatedAt: t0.Add(-2_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, WinnerAddress: "8", DepositorAddress: "7", CreatedAt: t0.Add(-2_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: millionstypes.DepositState_Failure, CreatedAt: t0.Add(-2_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
		},
	)
	suite.Require().Equal(int64(1_000_000), depositsTWB[0].Amount.Int64())
	suite.Require().Equal(int64(1_000_000), depositsTWB[1].Amount.Int64())
	suite.Require().Equal(int64(500_000), depositsTWB[2].Amount.Int64())
	suite.Require().Equal(int64(250_000), depositsTWB[3].Amount.Int64())
	suite.Require().Equal(int64(0), depositsTWB[4].Amount.Int64())
	suite.Require().Equal(int64(0), depositsTWB[5].Amount.Int64())
	// TWB should not return sponsors since they waive their drawing chances
	suite.Require().Len(depositsTWB, 7)
	suite.Require().Equal("8", depositsTWB[6].Address)

	// Simulate deposits at various time for a subsequent draw
	params := app.MillionsKeeper.GetParams(ctx)
	lastDrawAt := t0.Add(-500_000 * time.Second)
	depositsTWB = app.MillionsKeeper.ComputeDepositsTWB(
		ctx,
		lastDrawAt,
		t0,
		[]millionstypes.Deposit{
			{State: dState, CreatedAt: t0.Add(-2_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(-1_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(-500_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(-250_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(-params.MinDepositDrawDelta), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(-params.MinDepositDrawDelta + 1*time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(0 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, CreatedAt: t0.Add(1_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, IsSponsor: true, WinnerAddress: "6", DepositorAddress: "6", CreatedAt: t0.Add(-2_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: dState, WinnerAddress: "8", DepositorAddress: "7", CreatedAt: t0.Add(-2_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
			{State: millionstypes.DepositState_IbcTransfer, CreatedAt: t0.Add(-2_000_000 * time.Second), Amount: sdk.NewCoin(denom, sdk.NewInt(1_000_000))},
		},
	)
	suite.Require().Equal(int64(1_000_000), depositsTWB[0].Amount.Int64())
	suite.Require().Equal(int64(1_000_000), depositsTWB[1].Amount.Int64())
	suite.Require().Equal(int64(1_000_000), depositsTWB[2].Amount.Int64())
	suite.Require().Equal(int64(500_000), depositsTWB[3].Amount.Int64())
	suite.Require().Greater(depositsTWB[4].Amount.Int64(), int64(0))
	suite.Require().Equal(int64(0), depositsTWB[5].Amount.Int64())
	suite.Require().Equal(int64(0), depositsTWB[6].Amount.Int64())
	suite.Require().Equal(int64(0), depositsTWB[7].Amount.Int64())
	// TWB should not return sponsors since they waive their drawing chances
	suite.Require().Len(depositsTWB, 9)
	suite.Require().Equal("8", depositsTWB[8].Address)
}

// TestDraw_PrizeDistribution tests draw prize distribution phase
func (suite *KeeperTestSuite) TestDraw_PrizeDistribution() {
	app := suite.app
	ctx := suite.ctx

	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)

	seed := int64(84)

	prizePool := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000))
	drawRes, err := app.MillionsKeeper.RunDrawPrizes(ctx,
		prizePool,
		millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 20, Quantity: 1, DrawProbability: floatToDec(0.01)},
				{PoolPercent: 30, Quantity: 1, DrawProbability: floatToDec(0.11)},
				{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
			},
		},
		[]millionskeeper.DepositTWB{
			{Address: suite.addrs[0].String(), Amount: sdk.NewInt(1_000_000)},
			{Address: suite.addrs[1].String(), Amount: sdk.NewInt(1_000_000)},
			{Address: suite.addrs[2].String(), Amount: sdk.NewInt(1_000_000)},
		},
		seed,
	)
	suite.Require().NoError(err)

	pool := newValidPool(suite, millionstypes.Pool{
		PoolId: poolID,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 50, Quantity: 100, DrawProbability: floatToDec(0.5)},
				{PoolPercent: 50, Quantity: 200, DrawProbability: floatToDec(0.7)},
			},
		},
	})
	app.MillionsKeeper.AddPool(ctx, pool)

	app.MillionsKeeper.SetPoolDraw(ctx, millionstypes.Draw{
		PoolId:          poolID,
		DrawId:          pool.GetNextDrawId(),
		State:           millionstypes.DrawState_Drawing,
		ErrorState:      millionstypes.DrawState_Unspecified,
		PrizePool:       sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000)),
		UpdatedAtHeight: ctx.BlockHeight(),
		UpdatedAt:       ctx.BlockTime(),
	})

	draw, err := app.MillionsKeeper.GetPoolDraw(ctx, poolID, pool.GetNextDrawId())
	suite.Require().NoError(err)

	err = app.MillionsKeeper.DistributePrizes(ctx, app.MillionsKeeper.NewFeeCollector(ctx, *pool), drawRes, draw)
	suite.Require().NoError(err)

	for _, pd := range drawRes.PrizeDraws {
		if pd.Winner != nil {
			winnerAddress, err := sdk.AccAddressFromBech32(pd.Winner.Address)
			suite.Require().NoError(err)
			accountPrizes := app.MillionsKeeper.ListAccountPoolPrizes(ctx, winnerAddress, poolID)
			suite.Require().NoError(err)
			suite.Require().GreaterOrEqual(len(accountPrizes), 1)
			draw, err := app.MillionsKeeper.GetPoolDraw(ctx, poolID, pool.GetNextDrawId())
			suite.Require().NoError(err)
			suite.Require().GreaterOrEqual(len(draw.PrizesRefs), 1)
		}
	}

}

// TestDraw_TriggerAndPrizePool tests draw triggered by block updates and prize pool computation
func (suite *KeeperTestSuite) TestDraw_TriggerAndPrizePool() {
	app := suite.app
	ctx := suite.ctx

	// Create pool with a fake amount already available
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: poolID,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(24 * time.Hour),
			DrawDelta:     24 * time.Hour,
		},
		AvailablePrizePool: sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), math.NewInt(1000)),
	}))

	// First draw should pick up the available prize pool amount but return it to the pool (no winner)
	app.MillionsKeeper.BlockPoolUpdates(ctx.WithBlockTime(ctx.BlockTime().Add(25 * time.Hour)))
	pool, err := app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(1000), pool.AvailablePrizePool.Amount)
	draw, err := app.MillionsKeeper.GetPoolDraw(ctx, pool.PoolId, pool.NextDrawId-1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_Success, draw.State)
	suite.Require().Equal(math.NewInt(1000), draw.PrizePoolRemainsAmount)
	suite.Require().Equal(sdk.ZeroInt(), draw.PrizePoolFreshAmount)
	suite.Require().Equal(math.NewInt(1000), draw.PrizePool.Amount)

	// Simulate a prize clawback
	app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
		PoolId:        poolID,
		DrawId:        pool.NextDrawId,
		PrizeId:       1,
		Amount:        sdk.NewCoin(pool.Denom, math.NewInt(200)),
		WinnerAddress: suite.addrs[1].String(),
		CreatedAt:     ctx.BlockTime(),
		ExpiresAt:     ctx.BlockTime().Add(1 * time.Second),
		State:         millionstypes.PrizeState_Pending,
	})
	app.MillionsKeeper.BlockPrizeUpdates(ctx.WithBlockTime(ctx.BlockTime().Add(50 * time.Hour)))

	// Pool should have clawed back the amount
	pool, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(1200), pool.AvailablePrizePool.Amount)

	// Next draw should pick the new available amount but return it to the pool (still no winner)
	app.MillionsKeeper.BlockPoolUpdates(ctx.WithBlockTime(ctx.BlockTime().Add(50 * time.Hour)))
	pool, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().Equal(math.NewInt(1200), pool.AvailablePrizePool.Amount)
	suite.Require().NoError(err)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pool.PoolId, pool.NextDrawId-1)
	suite.Require().NoError(err)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(1200), draw.PrizePoolRemainsAmount)
	suite.Require().Equal(math.NewInt(0), draw.PrizePoolFreshAmount)
	suite.Require().Equal(math.NewInt(1200), draw.PrizePool.Amount)
}

// TestDraw_EvenPrizeDistributionLLN tests if the prize distribution abides by the law of large numbers
// A large number of draws should converge to the expected drawing probabilities
// In this case an even distribution for depositors with similar amounts
func (suite *KeeperTestSuite) TestDraw_EvenPrizeDistributionLLN() {
	app := suite.app
	ctx := suite.ctx

	type balanceControl struct {
		address        string
		deposit        math.Int
		inflationGains math.Int
		millionsGains  math.Int
		millionsDrawn  int
	}

	prizePool := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(0))
	// Prepare expected inflation
	yearlyInflation := 0.34
	weeklyInflation := math.LegacyNewDecWithPrec(int64(yearlyInflation/52.0*1_000_000), 6)
	// Prepare depositors
	depositors := []millionskeeper.DepositTWB{}
	balancesControls := map[string]*balanceControl{}
	for i := 0; i < 100; i++ {
		uid := fmt.Sprintf("small_%d", i)
		deposit := sdk.NewCoin(prizePool.Denom, sdk.NewInt(1_000_000))
		depositors = append(depositors, millionskeeper.DepositTWB{Address: uid, Amount: deposit.Amount})
		balancesControls[uid] = &balanceControl{address: uid, deposit: deposit.Amount, inflationGains: math.ZeroInt(), millionsGains: math.ZeroInt()}
	}

	// Init stable strategy for all draws
	strat := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 100, Quantity: 100, DrawProbability: floatToDec(1.0)},
		},
	}

	// Simulate 52,000 weekly draws (1000 years)
	totalInf := math.ZeroInt()
	totalWinAmount := math.ZeroInt()
	totalWinners := 0
	remainsFromLastDraw := math.ZeroInt()
	drawCount := 52_000
	for i := 0; i < drawCount; i++ {
		seed := int64(42 * (i + 1))

		// Attribute inflation gains and compute prizePool
		prizePool := math.ZeroInt().Add(remainsFromLastDraw)
		for i := range balancesControls {
			inf := weeklyInflation.MulInt(balancesControls[i].deposit).RoundInt()
			balancesControls[i].inflationGains = balancesControls[i].inflationGains.Add(inf)
			prizePool = prizePool.Add(inf)
			totalInf = totalInf.Add(inf)
		}

		// DrawPrizes and save results
		results, err := app.MillionsKeeper.RunDrawPrizes(ctx,
			sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), prizePool),
			strat,
			depositors,
			seed,
		)
		suite.Require().NoError(err, "error at draw %d: %v", i, err)
		totalWinAmount = totalWinAmount.Add(results.TotalWinAmount)
		totalWinners += int(results.TotalWinCount)
		remainsFromLastDraw = prizePool.Sub(results.TotalWinAmount)
		for _, d := range results.PrizeDraws {
			if d.Winner != nil {
				balancesControls[d.Winner.Address].millionsGains = balancesControls[d.Winner.Address].millionsGains.Add(d.Amount)
				balancesControls[d.Winner.Address].millionsDrawn++
			}
		}
	}

	// Number of wins should be close to probabilities (less than 0.01% difference)
	probaTotalWinners := float64(0)
	for _, pb := range strat.PrizeBatches {
		drawProb, err := pb.DrawProbability.Float64()
		suite.Require().NoError(err)
		probaTotalWinners += float64(pb.Quantity) * drawProb * float64(drawCount)
	}

	suite.Require().Greater(float64(totalWinners)/float64(probaTotalWinners), 0.999)
	suite.Require().Less(float64(totalWinners)/float64(probaTotalWinners), 1.001)
	// Draws outcome + remains should be equal to inflation
	suite.Require().Equal(totalInf.Int64(), totalWinAmount.Add(remainsFromLastDraw).Int64())
	// Draws outcome should be close to inflation (less than 0.01% difference)
	suite.Require().Greater(float64(totalWinAmount.Int64())/float64(totalInf.Int64()), 0.999)
	suite.Require().Less(float64(totalWinAmount.Int64())/float64(totalInf.Int64()), 1.001)

	// Each depositor gains should be close to inflation (less than 0.1% difference) since they are all identical
	for _, b := range balancesControls {
		suite.Require().Greater(float64(b.millionsGains.Int64())/float64(b.inflationGains.Int64()), float64(0.99))
		suite.Require().Less(float64(b.millionsGains.Int64())/float64(b.inflationGains.Int64()), float64(1.01))
	}
}

// TestDraw_WeightedPrizeDistributionLLN tests if the prize distribution abides by the law of large numbers
// A large number of draws should converge to the expected drawing probabilities
// In this case a weighted distribution based on each depositor stakes
func (suite *KeeperTestSuite) TestDraw_WeightedPrizeDistributionLLN() {
	app := suite.app
	ctx := suite.ctx

	type balanceControl struct {
		address        string
		deposit        math.Int
		inflationGains math.Int
		millionsGains  math.Int
		millionsDrawn  int
	}

	prizePool := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(0))
	// Prepare expected inflation
	yearlyInflation := 0.34
	weeklyInflation := math.LegacyNewDecWithPrec(int64(yearlyInflation/52.0*1_000_000), 6)
	// Prepare depositors
	depositors := []millionskeeper.DepositTWB{}
	balancesControls := map[string]*balanceControl{}
	// large depositors
	for i := 0; i < 10; i++ {
		uid := fmt.Sprintf("large_%d", i)
		deposit := sdk.NewCoin(prizePool.Denom, sdk.NewInt(100_000_000))
		depositors = append(depositors, millionskeeper.DepositTWB{Address: uid, Amount: deposit.Amount})
		balancesControls[uid] = &balanceControl{address: uid, deposit: deposit.Amount, inflationGains: math.ZeroInt(), millionsGains: math.ZeroInt()}
	}
	// medium depositors
	for i := 0; i < 20; i++ {
		uid := fmt.Sprintf("medium_%d", i)
		deposit := sdk.NewCoin(prizePool.Denom, sdk.NewInt(10_000_000))
		depositors = append(depositors, millionskeeper.DepositTWB{Address: uid, Amount: deposit.Amount})
		balancesControls[uid] = &balanceControl{address: uid, deposit: deposit.Amount, inflationGains: math.ZeroInt(), millionsGains: math.ZeroInt()}
	}
	// small depositors
	for i := 0; i < 40; i++ {
		uid := fmt.Sprintf("small_%d", i)
		deposit := sdk.NewCoin(prizePool.Denom, sdk.NewInt(1_000_000))
		depositors = append(depositors, millionskeeper.DepositTWB{Address: uid, Amount: deposit.Amount})
		balancesControls[uid] = &balanceControl{address: uid, deposit: deposit.Amount, inflationGains: math.ZeroInt(), millionsGains: math.ZeroInt()}
	}

	// Init stable strategy for all draws
	strat := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 100, Quantity: 200, DrawProbability: floatToDec(0.80)},
		},
	}

	// Simulate 52,000 weekly draws (1000 years)
	totalInf := math.ZeroInt()
	totalWinAmount := math.ZeroInt()
	totalWinners := 0
	remainsFromLastDraw := math.ZeroInt()
	drawCount := 52_000
	for i := 0; i < drawCount; i++ {
		seed := int64(84 * (i + 1))

		// Attribute inflation gains and compute prizePool
		prizePool := math.ZeroInt().Add(remainsFromLastDraw)
		for i := range balancesControls {
			inf := weeklyInflation.MulInt(balancesControls[i].deposit).RoundInt()
			balancesControls[i].inflationGains = balancesControls[i].inflationGains.Add(inf)
			prizePool = prizePool.Add(inf)
			totalInf = totalInf.Add(inf)
		}

		// DrawPrizes and save results
		results, err := app.MillionsKeeper.RunDrawPrizes(ctx,
			sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), prizePool),
			strat,
			depositors,
			seed,
		)
		suite.Require().NoError(err, "error at draw %d: %v", i, err)
		totalWinAmount = totalWinAmount.Add(results.TotalWinAmount)
		totalWinners += int(results.TotalWinCount)
		remainsFromLastDraw = prizePool.Sub(results.TotalWinAmount)
		for _, d := range results.PrizeDraws {
			if d.Winner != nil {
				balancesControls[d.Winner.Address].millionsGains = balancesControls[d.Winner.Address].millionsGains.Add(d.Amount)
				balancesControls[d.Winner.Address].millionsDrawn++
			}
		}
	}

	// Number of wins should be close to probabilities (less than 0.01% difference)
	probaTotalWinners := float64(0)
	for _, pb := range strat.PrizeBatches {
		drawProb, err := pb.DrawProbability.Float64()
		suite.Require().NoError(err)
		probaTotalWinners += float64(pb.Quantity) * drawProb * float64(drawCount)
	}

	suite.Require().Greater(float64(totalWinners)/float64(probaTotalWinners), 0.999)
	suite.Require().Less(float64(totalWinners)/float64(probaTotalWinners), 1.001)
	// Draws outcome + remains should be equal to inflation
	suite.Require().Equal(totalInf.Int64(), totalWinAmount.Add(remainsFromLastDraw).Int64())
	// Draws outcome should be close to inflation (less than 0.01% difference)
	suite.Require().Greater(float64(totalWinAmount.Int64())/float64(totalInf.Int64()), 0.999)
	suite.Require().Less(float64(totalWinAmount.Int64())/float64(totalInf.Int64()), 1.00)

	// Each depositor gains should be close to inflation (less than 5% difference)
	// A higher difference from inflation is understandable here especially for small depositors for which a win
	// is less likely but has a higher impact on their final balance
	for _, b := range balancesControls {
		suite.Require().Greater(float64(b.millionsGains.Int64())/float64(b.inflationGains.Int64()), float64(0.95))
		suite.Require().Less(float64(b.millionsGains.Int64())/float64(b.inflationGains.Int64()), float64(1.05))
	}
}

// TestDraw_IDsGeneration test that drawID is always incremented for a new draw
func (suite *KeeperTestSuite) TestDraw_IDsGeneration() {
	app := suite.app
	ctx := suite.ctx

	pool := newValidPool(suite, millionstypes.Pool{
		PoolId: uint64(1),
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 50, Quantity: 10, DrawProbability: floatToDec(1.00)},
			},
		},
		AvailablePrizePool: sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), math.NewInt(1_000_000)),
	})
	app.MillionsKeeper.AddPool(ctx, pool)
	// Launch 10 new draws
	for i := 0; i < 10; i++ {
		poolDraw, err := app.MillionsKeeper.LaunchNewDraw(ctx, pool.PoolId)
		suite.Require().NoError(err)
		// drawID should increment by 1
		suite.Require().Equal(uint64(i+1), poolDraw.DrawId)
	}
}

// TestDraw_SetPoolDraw tests that a draw result is set in the KVStore for a given poolID and drawID
func (suite *KeeperTestSuite) TestDraw_SetPoolDraw() {
	app := suite.app
	ctx := suite.ctx

	pool := newValidPool(suite, millionstypes.Pool{
		PoolId: uint64(1),
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 50, Quantity: 10, DrawProbability: floatToDec(1.00)},
			},
		},
		AvailablePrizePool: sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), math.NewInt(1_000_000)),
	})
	draw1 := millionstypes.Draw{
		PoolId:          pool.PoolId,
		DrawId:          uint64(1),
		State:           millionstypes.DrawState_Drawing,
		ErrorState:      millionstypes.DrawState_Unspecified,
		PrizePool:       sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000)),
		UpdatedAtHeight: ctx.BlockHeight(),
		UpdatedAt:       ctx.BlockTime(),
	}
	// List draws
	app.MillionsKeeper.SetPoolDraw(ctx, draw1)
	draws := app.MillionsKeeper.ListDraws(ctx)
	suite.Require().Len(draws, 1)
	draws = app.MillionsKeeper.ListPoolDraws(ctx, pool.PoolId)
	suite.Require().Len(draws, 1)

	draw, err := app.MillionsKeeper.GetPoolDraw(ctx, pool.PoolId, draw1.DrawId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_Drawing, draw.State)
	// Update the same draw
	draw1.State = millionstypes.DrawState_Success
	app.MillionsKeeper.SetPoolDraw(ctx, draw1)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pool.PoolId, draw1.DrawId)
	suite.Require().NoError(err)

	// Verify the state was updated
	suite.Require().Equal(millionstypes.DrawState_Success, draw.State)

	// Create second draw
	draw2 := millionstypes.Draw{
		PoolId:          pool.PoolId,
		DrawId:          uint64(2),
		State:           millionstypes.DrawState_Success,
		ErrorState:      millionstypes.DrawState_Unspecified,
		PrizePool:       sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000)),
		UpdatedAtHeight: ctx.BlockHeight(),
		UpdatedAt:       ctx.BlockTime(),
	}

	app.MillionsKeeper.SetPoolDraw(ctx, draw2)
	draws = app.MillionsKeeper.ListDraws(ctx)
	suite.Require().Len(draws, 2)
	draws = app.MillionsKeeper.ListPoolDraws(ctx, pool.PoolId)
	suite.Require().Len(draws, 2)

	// Create third draw
	draw3 := millionstypes.Draw{
		PoolId:          pool.PoolId,
		DrawId:          uint64(3),
		State:           millionstypes.DrawState_Success,
		ErrorState:      millionstypes.DrawState_Unspecified,
		PrizePool:       sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000)),
		UpdatedAtHeight: ctx.BlockHeight(),
		UpdatedAt:       ctx.BlockTime(),
	}

	app.MillionsKeeper.SetPoolDraw(ctx, draw3)
	draws = app.MillionsKeeper.ListDraws(ctx)
	suite.Require().Len(draws, 3)
	draws = app.MillionsKeeper.ListPoolDraws(ctx, pool.PoolId)
	suite.Require().Len(draws, 3)
}

// TestDraw_ClaimYieldOnRemoteZone tests claim of staking rewards from the native chain validators
func (suite *KeeperTestSuite) TestDraw_ClaimYieldOnRemoteZone() {
	app := suite.app
	ctx := suite.ctx
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)
	goCtx := sdk.WrapSDKContext(ctx)

	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta1 := 1 * time.Hour
	uatomAddress7 := apptesting.AddTestAddrsWithDenom(app, ctx, 7, sdk.NewInt(1_000_0000_000), remotePoolDenom)
	suite.addrs = append(suite.addrs, uatomAddress7...)

	// Remote pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:              poolID,
		Bech32PrefixValAddr: remoteBech32PrefixValAddr,
		ChainId:             remoteChainId,
		Denom:               remotePoolDenom,
		NativeDenom:         remotePoolDenom,
		ConnectionId:        remoteConnectionId,
		TransferChannelId:   remoteTransferChannelId,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: cosmosPoolValidator,
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		IcaDepositAddress:   cosmosIcaDepositAddress,
		IcaPrizepoolAddress: cosmosIcaPrizePoolAddress,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(remotePoolDenom, math.NewInt(1000)),
	}))

	pools := app.MillionsKeeper.ListPools(ctx)

	draw1 := millionstypes.Draw{
		PoolId:          pools[0].PoolId,
		DrawId:          pools[0].GetNextDrawId(),
		State:           millionstypes.DrawState_IcaWithdrawRewards,
		ErrorState:      millionstypes.DrawState_Unspecified,
		PrizePool:       sdk.NewCoin(remotePoolDenom, sdk.NewInt(1_000_000)),
		UpdatedAtHeight: ctx.BlockHeight(),
		UpdatedAt:       ctx.BlockTime(),
	}

	app.MillionsKeeper.SetPoolDraw(ctx, draw1)
	// Test to acquire a pool with wrong poolID
	_, err := app.MillionsKeeper.ClaimYieldOnRemoteZone(ctx, uint64(0), pools[0].NextDrawId)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolNotFound)
	// Test to acquire a draw with wrong drawID
	_, err = app.MillionsKeeper.ClaimYieldOnRemoteZone(ctx, pools[0].PoolId, uint64(0))
	suite.Require().ErrorIs(err, millionstypes.ErrPoolDrawNotFound)
	draw, err := app.MillionsKeeper.GetPoolDraw(ctx, pools[0].PoolId, pools[0].NextDrawId)
	suite.Require().NoError(err)
	// Test state before callBack
	suite.Require().Equal(millionstypes.DrawState_IcaWithdrawRewards, draw.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw.ErrorState)
	// Simulate failed callBack
	_, err = app.MillionsKeeper.OnClaimYieldOnRemoteZoneCompleted(ctx, draw.PoolId, draw.DrawId, true)
	suite.Require().NoError(err)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pools[0].PoolId, pools[0].NextDrawId)
	suite.Require().NoError(err)
	// Test state after callBack
	suite.Require().Equal(millionstypes.DrawState_Failure, draw.State)
	suite.Require().Equal(millionstypes.DrawState_IcaWithdrawRewards, draw.ErrorState)
	// Update draw state to do a retrial
	draw.State = millionstypes.DrawState_IcaWithdrawRewards
	draw.ErrorState = millionstypes.DrawState_Unspecified
	app.MillionsKeeper.SetPoolDraw(ctx, draw)
	// Simulate successful ICA callback
	// Move to ICQ phase (query available prize pool)
	_, err = app.MillionsKeeper.OnClaimYieldOnRemoteZoneCompleted(ctx, draw.PoolId, draw.DrawId, false)
	suite.Require().NoError(err)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pools[0].PoolId, pools[0].NextDrawId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_IcqBalance, draw.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw.ErrorState)
	// Simulate succesful ICQ callback
	// Should complete the Draw since no coins was found by the simulated callback (skip transfer phase)
	_, err = app.MillionsKeeper.OnQueryFreshPrizePoolCoinsOnRemoteZoneCompleted(ctx, draw.PoolId, draw.DrawId, sdk.NewCoins(), false)
	suite.Require().NoError(err)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pools[0].PoolId, pools[0].NextDrawId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_Success, draw.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw.ErrorState)

	// Test claimrewards for local pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      app.MillionsKeeper.GetNextPoolID(ctx),
		Denom:       localPoolDenom,
		NativeDenom: localPoolDenom,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: suite.valAddrs[0].String(),
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(1.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	}))
	// List pools
	pools = app.MillionsKeeper.ListPools(ctx)
	_, err = msgServer.Deposit(goCtx, millionstypes.NewMsgDeposit(
		suite.addrs[0].String(),
		sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		pools[1].PoolId,
	))
	suite.Require().NoError(err)

	draw2 := millionstypes.Draw{
		PoolId:          pools[1].PoolId,
		DrawId:          pools[1].GetNextDrawId(),
		State:           millionstypes.DrawState_IcaWithdrawRewards,
		ErrorState:      millionstypes.DrawState_Unspecified,
		PrizePool:       sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		UpdatedAtHeight: ctx.BlockHeight(),
		UpdatedAt:       ctx.BlockTime(),
	}

	app.MillionsKeeper.SetPoolDraw(ctx, draw2)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pools[1].PoolId, pools[1].NextDrawId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_IcaWithdrawRewards, draw.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw.ErrorState)

	// Trigger ClaimYieldOnRemoteZone
	_, err = app.MillionsKeeper.ClaimYieldOnRemoteZone(ctx, draw.PoolId, draw.DrawId)
	suite.Require().NoError(err)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pools[1].PoolId, pools[1].NextDrawId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_Success, draw.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw.ErrorState)
}

// TestDraw_TransferFreshPrizePoolCoinsToLocalZone tests the transfer of the claimed rewards to the local chain
func (suite *KeeperTestSuite) TestDraw_TransferFreshPrizePoolCoinsToLocalZone() {
	app := suite.app
	ctx := suite.ctx
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)
	goCtx := sdk.WrapSDKContext(ctx)

	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta1 := 1 * time.Hour
	uatomAddress7 := apptesting.AddTestAddrsWithDenom(app, ctx, 7, sdk.NewInt(1_000_0000_000), remotePoolDenom)
	suite.addrs = append(suite.addrs, uatomAddress7...)

	// Remote pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:              poolID,
		Bech32PrefixValAddr: remoteBech32PrefixValAddr,
		ChainId:             remoteChainId,
		Denom:               remotePoolDenom,
		NativeDenom:         remotePoolDenom,
		ConnectionId:        remoteConnectionId,
		TransferChannelId:   remoteTransferChannelId,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: cosmosPoolValidator,
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		IcaDepositAddress:   cosmosIcaDepositAddress,
		IcaPrizepoolAddress: cosmosIcaPrizePoolAddress,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(remotePoolDenom, math.NewInt(1000)),
	}))

	pools := app.MillionsKeeper.ListPools(ctx)

	draw1 := millionstypes.Draw{
		PoolId:          pools[0].PoolId,
		DrawId:          pools[0].GetNextDrawId(),
		State:           millionstypes.DrawState_IbcTransfer,
		ErrorState:      millionstypes.DrawState_Unspecified,
		PrizePool:       sdk.NewCoin(remotePoolDenom, sdk.NewInt(1_000_000)),
		UpdatedAtHeight: ctx.BlockHeight(),
		UpdatedAt:       ctx.BlockTime(),
	}

	app.MillionsKeeper.SetPoolDraw(ctx, draw1)
	// Test to acquire a pool with wrong poolID
	_, err := app.MillionsKeeper.TransferFreshPrizePoolCoinsToLocalZone(ctx, uint64(0), pools[0].NextDrawId)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolNotFound)
	// Test to acquire a draw with wrong drawID
	_, err = app.MillionsKeeper.TransferFreshPrizePoolCoinsToLocalZone(ctx, pools[0].PoolId, uint64(0))
	suite.Require().ErrorIs(err, millionstypes.ErrPoolDrawNotFound)
	draw, err := app.MillionsKeeper.GetPoolDraw(ctx, pools[0].PoolId, pools[0].NextDrawId)
	suite.Require().NoError(err)
	// Test state before failed callBack
	suite.Require().Equal(millionstypes.DrawState_IbcTransfer, draw.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw.ErrorState)
	// Simulate failed callBack
	_, err = app.MillionsKeeper.OnTransferFreshPrizePoolCoinsToLocalZoneCompleted(ctx, draw.PoolId, draw.DrawId, true)
	suite.Require().NoError(err)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pools[0].PoolId, pools[0].NextDrawId)
	suite.Require().NoError(err)
	// Test state after callBack
	suite.Require().Equal(millionstypes.DrawState_Failure, draw.State)
	suite.Require().Equal(millionstypes.DrawState_IbcTransfer, draw.ErrorState)
	// Update draw state to do a retrial
	draw.State = millionstypes.DrawState_IbcTransfer
	draw.ErrorState = millionstypes.DrawState_Unspecified
	app.MillionsKeeper.SetPoolDraw(ctx, draw)
	// Simulate successful callBack
	_, err = app.MillionsKeeper.OnTransferFreshPrizePoolCoinsToLocalZoneCompleted(ctx, draw.PoolId, draw.DrawId, false)
	suite.Require().NoError(err)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pools[0].PoolId, pools[0].NextDrawId)
	suite.Require().NoError(err)
	// Test state after successful callback
	suite.Require().Equal(millionstypes.DrawState_Success, draw.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw.ErrorState)

	// Test transferRewards for local pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      app.MillionsKeeper.GetNextPoolID(ctx),
		Denom:       localPoolDenom,
		NativeDenom: localPoolDenom,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: suite.valAddrs[0].String(),
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(1.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	}))
	// List pools
	pools = app.MillionsKeeper.ListPools(ctx)
	_, err = msgServer.Deposit(goCtx, millionstypes.NewMsgDeposit(
		suite.addrs[0].String(),
		sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		pools[1].PoolId,
	))
	suite.Require().NoError(err)

	draw2 := millionstypes.Draw{
		PoolId:          pools[1].PoolId,
		DrawId:          pools[1].GetNextDrawId(),
		State:           millionstypes.DrawState_IbcTransfer,
		ErrorState:      millionstypes.DrawState_Unspecified,
		PrizePool:       sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		UpdatedAtHeight: ctx.BlockHeight(),
		UpdatedAt:       ctx.BlockTime(),
	}

	app.MillionsKeeper.SetPoolDraw(ctx, draw2)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pools[1].PoolId, pools[1].NextDrawId)
	suite.Require().NoError(err)
	// Test state before TransferFreshPrizePoolCoinsToLocalZone
	suite.Require().Equal(millionstypes.DrawState_IbcTransfer, draw.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw.ErrorState)

	// Trigger TransferFreshPrizePoolCoinsToLocalZone
	_, err = app.MillionsKeeper.TransferFreshPrizePoolCoinsToLocalZone(ctx, draw.PoolId, draw.DrawId)
	suite.Require().NoError(err)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pools[1].PoolId, pools[1].NextDrawId)
	suite.Require().NoError(err)
	// Test state after TransferFreshPrizePoolCoinsToLocalZone
	suite.Require().Equal(millionstypes.DrawState_Success, draw.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw.ErrorState)
}

// TestDraw_ExecuteDraw test the last draw phases by drawing prizes
func (suite *KeeperTestSuite) TestDraw_ExecuteDraw() {
	app := suite.app
	ctx := suite.ctx
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)
	goCtx := sdk.WrapSDKContext(ctx)

	pool := newValidPool(suite, millionstypes.Pool{
		PoolId: uint64(1),
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 2, DrawProbability: floatToDec(1.00)},
			},
		},
	})
	app.MillionsKeeper.AddPool(ctx, pool)
	pools := app.MillionsKeeper.ListPools(ctx)

	draw1 := millionstypes.Draw{
		PoolId:          pools[0].PoolId,
		DrawId:          pools[0].GetNextDrawId(),
		State:           millionstypes.DrawState_IbcTransfer,
		ErrorState:      millionstypes.DrawState_Unspecified,
		PrizePool:       sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		UpdatedAtHeight: ctx.BlockHeight(),
		UpdatedAt:       ctx.BlockTime(),
	}

	app.MillionsKeeper.SetPoolDraw(ctx, draw1)
	// Test to acquire a pool with wrong poolID
	_, err := app.MillionsKeeper.ExecuteDraw(ctx, uint64(0), pools[0].NextDrawId)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolNotFound)
	// Test to acquire a draw with wrong drawID
	_, err = app.MillionsKeeper.ExecuteDraw(ctx, pools[0].PoolId, uint64(0))
	suite.Require().ErrorIs(err, millionstypes.ErrPoolDrawNotFound)
	// Test to acquire that the draw has a valid state
	_, err = app.MillionsKeeper.ExecuteDraw(ctx, pools[0].PoolId, pools[0].NextDrawId)
	suite.Require().ErrorIs(err, millionstypes.ErrIllegalStateOperation)

	pool = newValidPool(suite, millionstypes.Pool{
		PoolId: uint64(2),
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 50, Quantity: 100, DrawProbability: floatToDec(0.00)},
				{PoolPercent: 50, Quantity: 200, DrawProbability: floatToDec(1.00)},
			},
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1_000_000)),
	})
	app.MillionsKeeper.AddPool(ctx, pool)
	pools = app.MillionsKeeper.ListPools(ctx)

	_, err = msgServer.Deposit(goCtx, millionstypes.NewMsgDeposit(
		suite.addrs[0].String(),
		sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		pools[1].PoolId,
	))
	suite.Require().NoError(err)

	// Create new draw
	draw2 := millionstypes.Draw{
		PoolId:          pools[1].PoolId,
		DrawId:          pools[1].GetNextDrawId(),
		State:           millionstypes.DrawState_Drawing,
		ErrorState:      millionstypes.DrawState_Unspecified,
		PrizePool:       sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		UpdatedAtHeight: ctx.BlockHeight(),
		UpdatedAt:       ctx.BlockTime(),
	}

	app.MillionsKeeper.SetPoolDraw(ctx, draw2)

	// Test draw before draw execution
	draw, err := app.MillionsKeeper.GetPoolDraw(ctx, pools[1].PoolId, pools[1].NextDrawId)
	suite.Require().NoError(err)
	suite.Require().Zero(draw.RandSeed)
	suite.Require().Equal(int64(0), draw.PrizePoolRemainsAmount.Int64())
	suite.Require().Equal(int64(1_000_000), draw.PrizePool.Amount.Int64())

	_, err = app.MillionsKeeper.ExecuteDraw(ctx, pools[1].PoolId, pools[1].NextDrawId)
	suite.Require().NoError(err)
	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, pools[1].PoolId, pools[1].NextDrawId)
	suite.Require().NoError(err)
	// Test properties after successful draw execution
	suite.Require().Equal(millionstypes.DrawState_Success, draw.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw.ErrorState)
	suite.Require().NotZero(draw.RandSeed)
	suite.Require().Equal(int64(1_000_000), draw.PrizePoolRemainsAmount.Int64())
	suite.Require().Equal(int64(2_000_000), draw.PrizePool.Amount.Int64())
}
