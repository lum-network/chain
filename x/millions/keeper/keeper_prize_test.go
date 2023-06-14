package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionstypes "github.com/lum-network/chain/x/millions/types"
)

// TestPrize_IDsGeneration runs test related to prizeID generation.
func (suite *KeeperTestSuite) TestPrize_IDsGeneration() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx

	// Add 10 Prizes
	for i := 0; i < 10; i++ {
		poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
		prizeID := app.MillionsKeeper.GetNextPrizeID(ctx)
		app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{PoolId: poolID}))

		pool, err := app.MillionsKeeper.GetPool(ctx, poolID)
		suite.Require().NoError(err)

		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        poolID,
			DrawId:        pool.NextDrawId,
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		})
		// Test that prizeID is incremented
		suite.Require().Equal(uint64(i+1), prizeID)

		// Test that we never override an existing entity
		panicF := func() {
			app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
				PoolId:        poolID,
				DrawId:        pool.NextDrawId,
				PrizeId:       prizeID,
				State:         millionstypes.PrizeState_Pending,
				WinnerAddress: suite.addrs[0].String(),
				Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
			})
		}
		suite.Require().Panics(panicF)
	}
}

// TestPrize_AddPrize tests the logic of adding a prize to the store.
func (suite *KeeperTestSuite) TestPrize_AddPrize() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx

	prizesBefore := app.MillionsKeeper.ListPrizes(ctx)
	suite.Require().Len(prizesBefore, 0)
	// Initialize prizeBefore
	var prizesAccountPoolBefore []millionstypes.Prize
	var prizesAccountBefore []millionstypes.Prize
	var prizesPoolDrawBefore []millionstypes.Prize
	var prizesPoolBefore []millionstypes.Prize

	// Add 5 prizes
	for i := 0; i < 5; i++ {
		poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
		drawDelta1 := 1 * time.Hour
		app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
			PoolId: poolID,
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
		// Retrieve the pool from the state
		pool, err := app.MillionsKeeper.GetPool(ctx, poolID)
		suite.Require().NoError(err)

		prizesAccountPoolBefore = app.MillionsKeeper.ListAccountPrizes(ctx, suite.addrs[i])
		prizesAccountBefore = app.MillionsKeeper.ListAccountPoolPrizes(ctx, suite.addrs[i], pool.PoolId)
		prizesPoolDrawBefore = app.MillionsKeeper.ListPoolDrawPrizes(ctx, pool.PoolId, pool.NextDrawId)
		prizesPoolBefore = app.MillionsKeeper.ListPoolPrizes(ctx, pool.PoolId)

		// Create a new deposit and add it to the state
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        pool.NextDrawId,
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[i].String(),
			Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		})
	}

	prizes := app.MillionsKeeper.ListPrizes(ctx)
	suite.Require().Len(prizes, 5)
	pool, err := app.MillionsKeeper.GetPool(ctx, prizes[0].PoolId)
	suite.Require().NoError(err)
	// - Test validateBasics
	// -- Test that the prize validation panics with an invalid pool ID
	panicF := func() {
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        uint64(0),
			DrawId:        pool.NextDrawId,
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)
	// -- Test that the prize validation panics with an invalid draw ID
	panicF = func() {
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        uint64(0),
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)
	// -- Test that the prize validation panics with an invalid prize ID
	panicF = func() {
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        pool.NextDrawId,
			PrizeId:       uint64(1),
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		})
		prizes = app.MillionsKeeper.ListPrizes(ctx)
		prizes[0].PrizeId = 0
	}
	suite.Require().Panics(panicF)
	// -- Test that the prize validation panics with an invalid state
	panicF = func() {
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        pool.NextDrawId,
			State:         millionstypes.PrizeState_Unspecified,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)
	// -- Test that the prize validation panics with an invalid winnerAddress
	panicF = func() {
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        pool.NextDrawId,
			State:         millionstypes.PrizeState_Unspecified,
			WinnerAddress: "",
			Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)
	// -- Test that the prize validation panics with an invalid prize amount
	panicF = func() {
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        pool.NextDrawId,
			State:         millionstypes.PrizeState_Unspecified,
			WinnerAddress: "",
			Amount:        sdk.NewCoin(localPoolDenom, sdk.ZeroInt()),
		})
	}
	suite.Require().Panics(panicF)
	// - Test List Methods to verify AddPrize was correctly added
	prizes = app.MillionsKeeper.ListPrizes(ctx)

	for _, prize := range prizes {
		addr := sdk.MustAccAddressFromBech32(prize.WinnerAddress)
		// Test ListAccountPrizes
		prizesAccountPrize := app.MillionsKeeper.ListAccountPrizes(ctx, addr)
		suite.Require().Equal(len(prizesAccountBefore)+1, len(prizesAccountPrize))
		// Test ListAccountPoolPrizes
		prizesAccountPool := app.MillionsKeeper.ListAccountPoolPrizes(ctx, addr, prize.PoolId)
		suite.Require().Equal(len(prizesAccountPoolBefore)+1, len(prizesAccountPool))
		// Test ListPoolDrawPrizes
		prizesPoolDraw := app.MillionsKeeper.ListPoolDrawPrizes(ctx, prize.PoolId, prize.DrawId)
		suite.Require().Equal(len(prizesPoolDrawBefore)+1, len(prizesPoolDraw))
		// Test ListPoolPrizes
		prizesPool := app.MillionsKeeper.ListPoolPrizes(ctx, prize.PoolId)
		suite.Require().Equal(len(prizesPoolBefore)+1, len(prizesPool))

		// Test GetPoolDrawPrize
		poolDrawPrize, err := app.MillionsKeeper.GetPoolDrawPrize(ctx, prize.PoolId, prize.DrawId, prize.PrizeId)
		suite.Require().NoError(err)
		suite.Require().Equal(poolDrawPrize.PoolId, prize.PoolId)
	}
}

// TestPrize_RemovePrize tests the logic of removing a prize from the store.
func (suite *KeeperTestSuite) TestPrize_RemovePrize() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx

	prizesBefore := app.MillionsKeeper.ListPrizes(ctx)
	suite.Require().Len(prizesBefore, 0)
	// Initialize prize before
	var prizesAccountPoolBefore []millionstypes.Prize
	var prizesAccountBefore []millionstypes.Prize
	var prizesPoolDrawBefore []millionstypes.Prize
	var prizesPoolBefore []millionstypes.Prize

	// Add 5 prizes
	for i := 0; i < 5; i++ {
		poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
		drawDelta1 := 1 * time.Hour
		app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
			PoolId: poolID,
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
		// Retrieve the pool from the state
		pool, err := app.MillionsKeeper.GetPool(ctx, poolID)
		suite.Require().NoError(err)

		prizesAccountPoolBefore = app.MillionsKeeper.ListAccountPrizes(ctx, suite.addrs[i])
		prizesAccountBefore = app.MillionsKeeper.ListAccountPoolPrizes(ctx, suite.addrs[i], pool.PoolId)
		prizesPoolDrawBefore = app.MillionsKeeper.ListPoolDrawPrizes(ctx, pool.PoolId, pool.NextDrawId)
		prizesPoolBefore = app.MillionsKeeper.ListPoolPrizes(ctx, pool.PoolId)

		// Create a new deposit and add it to the state
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        pool.NextDrawId,
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[i].String(),
			Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		})
	}

	prizes := app.MillionsKeeper.ListPrizes(ctx)
	suite.Require().Len(prizes, 5)
	pool, err := app.MillionsKeeper.GetPool(ctx, prizes[0].PoolId)
	suite.Require().NoError(err)
	// - Test validateBasics
	// -- Test that the prize validation return an error with an invalid pool ID
	err = app.MillionsKeeper.RemovePrize(ctx, millionstypes.Prize{
		PoolId:        uint64(0),
		DrawId:        pool.NextDrawId,
		PrizeId:       prizes[0].PrizeId,
		State:         millionstypes.PrizeState_Pending,
		WinnerAddress: suite.addrs[0].String(),
		Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)
	// -- Test that the prize validation return an error with an invalid draw ID
	err = app.MillionsKeeper.RemovePrize(ctx, millionstypes.Prize{
		PoolId:        pool.PoolId,
		DrawId:        uint64(0),
		PrizeId:       prizes[0].PrizeId,
		State:         millionstypes.PrizeState_Pending,
		WinnerAddress: suite.addrs[0].String(),
		Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)
	// -- Test that the prize validation return an error with an invalid prize ID
	err = app.MillionsKeeper.RemovePrize(ctx, millionstypes.Prize{
		PoolId:        pool.PoolId,
		DrawId:        pool.NextDrawId,
		PrizeId:       uint64(0),
		State:         millionstypes.PrizeState_Pending,
		WinnerAddress: suite.addrs[0].String(),
		Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)
	// -- Test that the prize validation return an error with an invalid state
	err = app.MillionsKeeper.RemovePrize(ctx, millionstypes.Prize{
		PoolId:        pool.PoolId,
		DrawId:        pool.NextDrawId,
		PrizeId:       prizes[0].PrizeId,
		State:         millionstypes.PrizeState_Unspecified,
		WinnerAddress: suite.addrs[0].String(),
		Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidPrizeState)
	// -- Test that the prize validation panics with an invalid winnerAddress
	err = app.MillionsKeeper.RemovePrize(ctx, millionstypes.Prize{
		PoolId:        pool.PoolId,
		DrawId:        pool.NextDrawId,
		PrizeId:       prizes[0].PrizeId,
		State:         millionstypes.PrizeState_Pending,
		WinnerAddress: "",
		Amount:        sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidWinnerAddress)
	// -- Test that the prize validation return an error with an invalid prize amount
	err = app.MillionsKeeper.RemovePrize(ctx, millionstypes.Prize{
		PoolId:        pool.PoolId,
		DrawId:        pool.NextDrawId,
		PrizeId:       prizes[0].PrizeId,
		State:         millionstypes.PrizeState_Pending,
		WinnerAddress: suite.addrs[0].String(),
		Amount:        sdk.NewCoin(localPoolDenom, sdk.ZeroInt()),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidPrizeAmount)
	// - Test List Methods to verify AddPrize was correctly added
	prizes = app.MillionsKeeper.ListPrizes(ctx)

	for _, prize := range prizes {
		addr := sdk.MustAccAddressFromBech32(prize.WinnerAddress)
		prizesAccountPrize := app.MillionsKeeper.ListAccountPrizes(ctx, addr)
		suite.Require().Equal(len(prizesAccountBefore)+1, len(prizesAccountPrize))
		// Test ListAccountPoolPrizes before RemovePrize
		prizesAccountPool := app.MillionsKeeper.ListAccountPoolPrizes(ctx, addr, prize.PoolId)
		suite.Require().Equal(len(prizesAccountPoolBefore)+1, len(prizesAccountPool))
		// Test ListPoolDrawPrizes before RemovePrize
		prizesPoolDraw := app.MillionsKeeper.ListPoolDrawPrizes(ctx, prize.PoolId, prize.DrawId)
		suite.Require().Equal(len(prizesPoolDrawBefore)+1, len(prizesPoolDraw))
		// Test ListPoolPrizes before RemovePrize
		prizesPool := app.MillionsKeeper.ListPoolPrizes(ctx, prize.PoolId)
		suite.Require().Equal(len(prizesPoolBefore)+1, len(prizesPool))

		// Test GetPoolDrawPrize
		poolDrawPrize, err := app.MillionsKeeper.GetPoolDrawPrize(ctx, prize.PoolId, prize.DrawId, prize.PrizeId)
		suite.Require().NoError(err)
		suite.Require().Equal(poolDrawPrize.PoolId, prize.PoolId)

		// Remove prize
		err = app.MillionsKeeper.RemovePrize(ctx, prize)
		suite.Require().NoError(err)
		// Test ListAccountPrizes after RemovePrize
		prizesAccountPool = app.MillionsKeeper.ListAccountPrizes(ctx, addr)
		suite.Require().Equal(len(prizesAccountBefore), len(prizesAccountPool))
		// Test ListAccountPoolPrizes after RemovePrize
		prizesAccountPool = app.MillionsKeeper.ListAccountPoolPrizes(ctx, addr, prize.PoolId)
		suite.Require().Equal(len(prizesAccountPoolBefore), len(prizesAccountPool))
		// Test ListPoolDrawPrizes after RemovePrize
		prizesPoolDraw = app.MillionsKeeper.ListPoolDrawPrizes(ctx, prize.PoolId, prize.DrawId)
		suite.Require().Equal(len(prizesPoolDrawBefore), len(prizesPoolDraw))
		// Test ListPoolPrizes after RemovePrize
		prizesPool = app.MillionsKeeper.ListPoolPrizes(ctx, prize.PoolId)
		suite.Require().Equal(len(prizesPoolBefore), len(prizesPool))
	}
}

// TestPrize_ClawBackPrize test the logic of the prize clawback.
func (suite *KeeperTestSuite) TestPrize_ClawBackPrize() {
	app := suite.app
	ctx := suite.ctx

	pool := newValidPool(suite, millionstypes.Pool{
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 50, Quantity: 100, DrawProbability: floatToDec(0.5)},
				{PoolPercent: 50, Quantity: 200, DrawProbability: floatToDec(0.7)},
			},
		},
	})
	app.MillionsKeeper.AddPool(ctx, pool)

	// Create prizes with various expiration time
	drawID := pool.NextDrawId
	for i := 1; i <= 5; i++ {
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        drawID,
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
			ExpiresAt:     ctx.BlockTime().Add(time.Duration(i) * time.Second),
		})
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        drawID,
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
			ExpiresAt:     ctx.BlockTime().Add(time.Duration(i) * time.Second),
		})
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        drawID,
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
			ExpiresAt:     ctx.BlockTime().Add(time.Duration(i) * time.Second),
		})
	}

	prizes := app.MillionsKeeper.ListPrizes(ctx)
	suite.Require().Len(prizes, 15)

	// Test invalid prize with wrong PoolID
	err := app.MillionsKeeper.ClawBackPrize(ctx, uint64(0), prizes[0].DrawId, prizes[0].PrizeId)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolNotFound)

	// Test invalid prize with wrong drawID
	err = app.MillionsKeeper.ClawBackPrize(ctx, prizes[0].PoolId, uint64(0), prizes[0].PrizeId)
	suite.Require().ErrorIs(err, millionstypes.ErrPrizeNotFound)

	// Test invalid pool with wrong prizeID
	err = app.MillionsKeeper.ClawBackPrize(ctx, prizes[0].PoolId, prizes[0].DrawId, uint64(0))
	suite.Require().ErrorIs(err, millionstypes.ErrPrizeNotFound)

	for _, prize := range prizes {
		// Get pool before clawedBack
		pool, err := app.MillionsKeeper.GetPool(ctx, pool.PoolId)
		suite.Require().NoError(err)
		// Get prize before clawedBack
		prizes = app.MillionsKeeper.ListPrizes(ctx)
		// ClawBackPrize
		err = app.MillionsKeeper.ClawBackPrize(ctx, prize.PoolId, prize.DrawId, prize.PrizeId)
		suite.Require().NoError(err)
		// Get pool after clawedBack
		poolAftrCB, err := app.MillionsKeeper.GetPool(ctx, pool.PoolId)
		suite.Require().NoError(err)
		suite.Require().Equal(pool.AvailablePrizePool.Add(prize.Amount), poolAftrCB.AvailablePrizePool)
		// Get prize after clawedBack
		prizesAftrCB := app.MillionsKeeper.ListPrizes(ctx)
		suite.Require().NoError(err)
		suite.Require().Equal(len(prizes)-1, len(prizesAftrCB))
	}
	// List prize should be empty
	prizes = app.MillionsKeeper.ListPrizes(ctx)
	suite.Require().Len(prizes, 0)
}

// TestPrize_ClawbackQueueExpiration tests the Dequeueing of expired prizes.
func (suite *KeeperTestSuite) TestPrize_ClawbackQueueExpiration() {
	app := suite.app
	ctx := suite.ctx

	pool := newValidPool(suite, millionstypes.Pool{
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 50, Quantity: 100, DrawProbability: floatToDec(0.5)},
				{PoolPercent: 50, Quantity: 200, DrawProbability: floatToDec(0.7)},
			},
		},
	})
	app.MillionsKeeper.AddPool(ctx, pool)

	// Create prizes with various expiration time
	drawID := pool.NextDrawId
	for i := 1; i <= 5; i++ {
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        drawID,
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
			ExpiresAt:     ctx.BlockTime().Add(time.Duration(i) * time.Second),
		})
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        drawID,
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
			ExpiresAt:     ctx.BlockTime().Add(time.Duration(i) * time.Second),
		})
		app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
			PoolId:        pool.PoolId,
			DrawId:        drawID,
			State:         millionstypes.PrizeState_Pending,
			WinnerAddress: suite.addrs[0].String(),
			Amount:        sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
			ExpiresAt:     ctx.BlockTime().Add(time.Duration(i) * time.Second),
		})
	}

	// EPCB queue should contain the prizes at their expiration time
	curPrizeId := millionstypes.UnknownID + 1
	for i := 1; i <= 5; i++ {
		col := app.MillionsKeeper.GetPrizeIDsEPCBQueue(ctx, ctx.BlockTime().Add(time.Duration(i)*time.Second))
		suite.Require().Len(col.PrizesIds, 3)
		for _, pid := range col.PrizesIds {
			suite.Require().Equal(pool.PoolId, pid.PoolId)
			suite.Require().Equal(drawID, pid.DrawId)
			suite.Require().Equal(curPrizeId, pid.PrizeId)
			curPrizeId += 1
		}
	}

	// Dequeueing EPCBQueue at endTime should dequeue all items up to endTime
	curPrizeId = millionstypes.UnknownID + 1

	// Should do nothing - 15 remaining prizes
	prizesToDequeue := app.MillionsKeeper.DequeueEPCBQueue(ctx, ctx.BlockTime().Add(time.Duration(0)*time.Second))
	suite.Require().Len(prizesToDequeue, 0)

	// Should dequeue 6 prizes - 9 remaining prizes
	prizesToDequeue = app.MillionsKeeper.DequeueEPCBQueue(ctx, ctx.BlockTime().Add(time.Duration(2)*time.Second))
	suite.Require().Len(prizesToDequeue, 6)
	for _, prize := range prizesToDequeue {
		suite.Require().Equal(curPrizeId, prize.PrizeId)
		curPrizeId += 1
	}

	// Should dequeue 6 more prizes - 3 remaining prizes
	prizesToDequeue = app.MillionsKeeper.DequeueEPCBQueue(ctx, ctx.BlockTime().Add(time.Duration(4)*time.Second))
	suite.Require().Len(prizesToDequeue, 6)
	for _, prize := range prizesToDequeue {
		suite.Require().Equal(curPrizeId, prize.PrizeId)
		curPrizeId += 1
	}

	// Should dequeue all remaining prizes
	prizesToDequeue = app.MillionsKeeper.DequeueEPCBQueue(ctx, ctx.BlockTime().Add(1_000_000*time.Second))
	suite.Require().Len(prizesToDequeue, 3)
	for _, prize := range prizesToDequeue {
		suite.Require().Equal(curPrizeId, prize.PrizeId)
		curPrizeId += 1
	}
}
