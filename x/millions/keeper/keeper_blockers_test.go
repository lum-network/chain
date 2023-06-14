package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionstypes "github.com/lum-network/chain/x/millions/types"
)

// TestBlockers_PoolUpdates tests block based pool updates (draw launches).
func (suite *KeeperTestSuite) TestBlockers_PoolUpdates() {
	app := suite.app
	ctx := suite.ctx.WithBlockHeight(0).WithBlockTime(time.Now().UTC())

	// Pool1 should always be able to Draw
	poolID1 := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta1 := 1 * time.Hour
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: poolID1,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), math.NewInt(1000)),
	}))

	// Pool2 should not be able to create any draw (invalid hardset validators configuration)
	poolID2 := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta2 := 2 * time.Hour
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: poolID2,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta2),
			DrawDelta:     drawDelta2,
		},
		AvailablePrizePool: sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), math.NewInt(1000)),
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: "lumvaloper1qx2dts3tglxcu0jh47k7ghstsn4nactufgmmlk",
			IsEnabled:       true,
			BondedAmount:    math.NewInt(1_000_000),
		}},
	}))

	// Pool3 should be able to draw (until we hardset faulty draw config)
	poolID3 := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta3 := 3 * time.Hour
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: poolID3,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta3),
			DrawDelta:     drawDelta3,
		},
		AvailablePrizePool: sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), math.NewInt(1000)),
	}))

	// We should have 3 pools and 0 draws
	suite.Require().Len(app.MillionsKeeper.ListPools(ctx), 3)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 0)

	// Block pool updates should have no effect until a pool is ready to draw
	sc, ec := app.MillionsKeeper.BlockPoolUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 0)
	ctx = ctx.WithBlockHeight(1).WithBlockTime(ctx.BlockTime().Add(60 * time.Second))
	sc, ec = app.MillionsKeeper.BlockPoolUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 0)

	// Shoud launch pool1 draw1 with success
	ctx = ctx.WithBlockHeight(2).WithBlockTime(ctx.BlockTime().Add(drawDelta1))
	sc, ec = app.MillionsKeeper.BlockPoolUpdates(ctx)
	suite.Require().Equal(1, sc)
	suite.Require().Equal(0, ec)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 1)
	p1, err := app.MillionsKeeper.GetPool(ctx, poolID1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.UnknownID+2, p1.NextDrawId)
	suite.Require().NotNil(p1.LastDrawCreatedAt)
	suite.Require().Equal(ctx.BlockTime(), *p1.LastDrawCreatedAt)
	suite.Require().Equal(millionstypes.DrawState_Success, p1.LastDrawState)
	p2, err := app.MillionsKeeper.GetPool(ctx, poolID2)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.UnknownID+1, p2.NextDrawId)
	suite.Require().Nil(p2.LastDrawCreatedAt)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, p2.LastDrawState)
	p3, err := app.MillionsKeeper.GetPool(ctx, poolID3)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.UnknownID+1, p3.NextDrawId)
	suite.Require().Nil(p3.LastDrawCreatedAt)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, p3.LastDrawState)

	// Should do nothing
	sc, ec = app.MillionsKeeper.BlockPoolUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 1)
	ctx = ctx.WithBlockHeight(3).WithBlockTime(ctx.BlockTime().Add(60 * time.Second))
	sc, ec = app.MillionsKeeper.BlockPoolUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 1)

	// Should trigger p1d2, p2d1 (error without draw creation)
	ctx = ctx.WithBlockHeight(4).WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockPoolUpdates(ctx)
	suite.Require().Equal(1, sc)
	suite.Require().Equal(1, ec)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 2)
	p1, err = app.MillionsKeeper.GetPool(ctx, poolID1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.UnknownID+3, p1.NextDrawId)
	suite.Require().NotNil(p1.LastDrawCreatedAt)
	suite.Require().Equal(ctx.BlockTime(), *p1.LastDrawCreatedAt)
	suite.Require().Equal(millionstypes.DrawState_Success, p1.LastDrawState)
	p2, err = app.MillionsKeeper.GetPool(ctx, poolID2)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.UnknownID+1, p2.NextDrawId)
	suite.Require().Nil(p2.LastDrawCreatedAt)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, p2.LastDrawState)
	p3, err = app.MillionsKeeper.GetPool(ctx, poolID3)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.UnknownID+1, p3.NextDrawId)
	suite.Require().Nil(p3.LastDrawCreatedAt)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, p3.LastDrawState)

	// Should trigger p2d1 (error without draw creation)
	ctx = ctx.WithBlockHeight(5).WithBlockTime(ctx.BlockTime().Add(60 * time.Second))
	sc, ec = app.MillionsKeeper.BlockPoolUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(1, ec)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 2)

	// Should trigger p1d3, p2d1 (error without draw creation) and p3d1
	ctx = ctx.WithBlockHeight(6).WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockPoolUpdates(ctx)
	suite.Require().Equal(2, sc)
	suite.Require().Equal(1, ec)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 4)
	p1, err = app.MillionsKeeper.GetPool(ctx, poolID1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.UnknownID+4, p1.NextDrawId)
	suite.Require().NotNil(p1.LastDrawCreatedAt)
	suite.Require().Equal(ctx.BlockTime(), *p1.LastDrawCreatedAt)
	suite.Require().Equal(millionstypes.DrawState_Success, p1.LastDrawState)
	p2, err = app.MillionsKeeper.GetPool(ctx, poolID2)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.UnknownID+1, p2.NextDrawId)
	suite.Require().Nil(p2.LastDrawCreatedAt)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, p2.LastDrawState)
	p3, err = app.MillionsKeeper.GetPool(ctx, poolID3)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.UnknownID+2, p3.NextDrawId)
	suite.Require().NotNil(p3.LastDrawCreatedAt)
	suite.Require().Equal(ctx.BlockTime(), *p3.LastDrawCreatedAt)
	suite.Require().Equal(millionstypes.DrawState_Success, p3.LastDrawState)

	// Should trigger p2d1 (error without draw creation)
	ctx = ctx.WithBlockHeight(7).WithBlockTime(ctx.BlockTime().Add(60 * time.Second))
	sc, ec = app.MillionsKeeper.BlockPoolUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(1, ec)
	suite.Require().Len(app.MillionsKeeper.ListDraws(ctx), 4)
}

// TestBlockers_PrizeUpdates tests block based prizes updates (clawbacks).
func (suite *KeeperTestSuite) TestBlockers_PrizeUpdates() {
	app := suite.app
	ctx := suite.ctx.WithBlockHeight(0).WithBlockTime(time.Now().UTC())

	// Add pool with ID 1 to make its prizes work
	poolID1 := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: poolID1,
	}))
	app.MillionsKeeper.SetPoolDraw(ctx, millionstypes.Draw{PoolId: poolID1, DrawId: millionstypes.UnknownID + 1})

	// Add expirable prizes with various expiration times
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
				PoolId:        uint64(i + 1),
				DrawId:        millionstypes.UnknownID + 1,
				PrizeId:       suite.app.MillionsKeeper.GetNextPrizeIdAndIncrement(ctx),
				State:         millionstypes.PrizeState_Pending,
				WinnerAddress: suite.addrs[0].String(),
				Amount:        sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), math.NewInt(1)),
				ExpiresAt:     ctx.BlockTime().Add(time.Duration(i) * time.Hour),
			})
		}
	}

	// Should clawback the first 10 prizes with success (poolID1 has been created)
	ctx = ctx.WithBlockHeight(1).WithBlockTime(ctx.BlockTime().Add(1 * time.Second))
	sc, ec := app.MillionsKeeper.BlockPrizeUpdates(ctx)
	suite.Require().Equal(10, sc)
	suite.Require().Equal(0, ec)

	// Should do nothing
	sc, ec = app.MillionsKeeper.BlockPrizeUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)

	// Should clawback the next 10 prizes (should silently ignore the errors since pool will not be found)
	ctx = ctx.WithBlockHeight(2).WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockPrizeUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(10, ec)

	// Should do nothing
	sc, ec = app.MillionsKeeper.BlockPrizeUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)

	// Should clawback the next 50 prizes
	ctx = ctx.WithBlockHeight(3).WithBlockTime(ctx.BlockTime().Add(5 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockPrizeUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(50, ec)

	// Should do nothing
	sc, ec = app.MillionsKeeper.BlockPrizeUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)

	// Should clawback all remaining prizes
	ctx = ctx.WithBlockHeight(4).WithBlockTime(ctx.BlockTime().Add(5 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockPrizeUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(30, ec)

	// Should do nothing
	sc, ec = app.MillionsKeeper.BlockPrizeUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)
	ctx = ctx.WithBlockHeight(5).WithBlockTime(ctx.BlockTime().Add(10 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockPrizeUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)
	ctx = ctx.WithBlockHeight(6).WithBlockTime(ctx.BlockTime().Add(-100 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockPrizeUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)
}

// TestBlockers_WithdrawalUpdates tests block based withdrawal updates (unbonding completed).
func (suite *KeeperTestSuite) TestBlockers_WithdrawalUpdates() {
	app := suite.app
	ctx := suite.ctx.WithBlockHeight(0).WithBlockTime(time.Now().UTC())

	// Add pool with ID 1 to make its withdrawals work
	p1 := newValidPool(suite, millionstypes.Pool{
		PoolId: app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx),
	})
	app.MillionsKeeper.AddPool(ctx, p1)

	// Add pool with ID 2 to make withdrawal work on pool level but fail at transfer level
	p2 := newValidPool(suite, millionstypes.Pool{
		PoolId:            app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx),
		IcaDepositAddress: suite.moduleAddrs[4].String(),
	})
	app.MillionsKeeper.AddPool(ctx, p2)

	// Add withdrawals with various unbonding times
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			t := ctx.BlockTime().Add(time.Duration(i) * time.Hour)
			app.MillionsKeeper.AddWithdrawal(ctx, millionstypes.Withdrawal{
				PoolId:           uint64(i + 1),
				DepositId:        millionstypes.UnknownID + 1,
				WithdrawalId:     app.MillionsKeeper.GetNextWithdrawalIdAndIncrement(ctx),
				State:            millionstypes.WithdrawalState_IcaUnbonding,
				DepositorAddress: suite.addrs[0].String(),
				ToAddress:        suite.addrs[0].String(),
				Amount:           sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), math.NewInt(1)),
				UnbondingEndsAt:  &t,
			})
		}
	}
	err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(p1.IcaDepositAddress), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), math.NewInt(1_000))))
	suite.Require().NoError(err)

	// Should complete the first 10 withdrawals with success (poolID1 has been created and funded)
	ctx = ctx.WithBlockHeight(1).WithBlockTime(ctx.BlockTime().Add(1 * time.Second))
	sc, ec := app.MillionsKeeper.BlockWithdrawalUpdates(ctx)
	suite.Require().Equal(10, sc)
	suite.Require().Equal(0, ec)
	// State updates should have been committed (verify first entity of the batch)
	// After successful completion we delete the withdrawal
	_, err = app.MillionsKeeper.GetPoolWithdrawal(ctx, p1.PoolId, millionstypes.UnknownID+1)
	suite.Require().ErrorIs(err, millionstypes.ErrWithdrawalNotFound)

	// Should do nothing
	sc, ec = app.MillionsKeeper.BlockWithdrawalUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)

	// Should complete the next 10 withdrawals (should silently ignore the errors since pool module account is not funded for the transfers)
	ctx = ctx.WithBlockHeight(2).WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockWithdrawalUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(10, ec)
	// State updates should have been committed (verify first entity of the batch)
	w, err := app.MillionsKeeper.GetPoolWithdrawal(ctx, p2.PoolId, millionstypes.UnknownID+11)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.WithdrawalState_Failure, w.State)
	suite.Require().Equal(millionstypes.WithdrawalState_IbcTransfer, w.ErrorState)

	// Should do nothing
	sc, ec = app.MillionsKeeper.BlockWithdrawalUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)

	// Should complete the next 50 withdrawals (should silently ignore the errors since pool will not be found)
	ctx = ctx.WithBlockHeight(3).WithBlockTime(ctx.BlockTime().Add(5 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockWithdrawalUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(50, ec)

	// Should do nothing
	sc, ec = app.MillionsKeeper.BlockWithdrawalUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)

	// Should complete all remaining withdrawals
	ctx = ctx.WithBlockHeight(4).WithBlockTime(ctx.BlockTime().Add(5 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockWithdrawalUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(30, ec)

	// Should do nothing
	sc, ec = app.MillionsKeeper.BlockWithdrawalUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)
	ctx = ctx.WithBlockHeight(5).WithBlockTime(ctx.BlockTime().Add(10 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockWithdrawalUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)
	ctx = ctx.WithBlockHeight(6).WithBlockTime(ctx.BlockTime().Add(-100 * time.Hour))
	sc, ec = app.MillionsKeeper.BlockWithdrawalUpdates(ctx)
	suite.Require().Equal(0, sc)
	suite.Require().Equal(0, ec)
}
