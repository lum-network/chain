package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	apptesting "github.com/lum-network/chain/app/testing"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

// TestDeposit_IDsGeneration tests that the depositID is properly incremented for each new deposit
func (suite *KeeperTestSuite) TestDeposit_IDsGeneration() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx

	// Set the denom
	denom := app.StakingKeeper.BondDenom(ctx)
	// Add 10 deposits
	for i := 0; i < 10; i++ {
		poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
		depositID := app.MillionsKeeper.GetNextDepositIdAndIncrement(ctx)
		app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{PoolId: poolID}))

		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           poolID,
			DepositId:        depositID,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
		// Test that depositID is incremented
		suite.Require().Equal(uint64(i+1), depositID)
		// Test that we never override an existing entity
		panicF := func() {
			app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
				PoolId:           poolID,
				DepositId:        depositID,
				DepositorAddress: suite.addrs[0].String(),
				WinnerAddress:    suite.addrs[0].String(),
				State:            millionstypes.DepositState_IbcTransfer,
				Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
			})
		}
		suite.Require().Panics(panicF)
	}
}

// TestDeposit_AddDeposit tests the logic of adding a deposit to the store
func (suite *KeeperTestSuite) TestDeposit_AddDeposit() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx

	// Set the denom
	denom := app.StakingKeeper.BondDenom(ctx)

	depositsBefore := app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Len(depositsBefore, 0)
	deposits := app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Len(deposits, 0)

	// Track poolTVLBefore
	var poolTvlBefore math.Int
	// Track poolDepositorsCount
	var poolDepositorCountBefore uint64
	// Add 5 deposits
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

		// Test poolTVLBefore
		poolTvlBefore = pool.TvlAmount
		suite.Require().Equal(int64(0), poolTvlBefore.Int64())

		// Test poolDepositorCount
		poolDepositorCountBefore = pool.DepositorsCount
		suite.Require().Equal(uint64(0), poolDepositorCountBefore)

		// Create a new deposit and add it to the state
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           pool.PoolId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}

	// - Test that deposits with unknown IDs have a new depositIDs reassigned
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	for i, deposit := range deposits {
		suite.Require().NotEqual(uint64(0), deposit.DepositId)
		suite.Require().Equal(uint64(i+1), deposit.DepositId)
	}

	// - Test the validate basics
	// -- Test one successful deposit with all correct validate basics
	panicF := func() {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           deposits[len(deposits)-1].PoolId,
			DepositorAddress: suite.addrs[1].String(),
			WinnerAddress:    suite.addrs[1].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().NotPanics(panicF)
	// -- Test that the deposit validation panics with an invalid pool ID
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           uint64(0),
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)
	// -- Test that the deposit validation panics with an invalid depositID
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		deposits[7].DepositId = 0
	}
	suite.Require().Panics(panicF)
	// -- Test that the deposit validation panics with an invalid state
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           deposits[len(deposits)-1].PoolId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_Unspecified,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)
	// -- Test that the deposit validation panics with an invalid DepositorAddress
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           deposits[len(deposits)-1].PoolId,
			DepositorAddress: "",
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)
	// -- Test that the deposit validation panics with an invalid WinnerAddress
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           deposits[len(deposits)-1].PoolId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    "",
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)

	// - Test that we never override an existing entity (double deposit)
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           deposits[len(deposits)-1].PoolId,
			DepositId:        deposits[len(deposits)-1].DepositId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)

	// - Test that the pool deposits and account deposits are correctly updated
	depositsAccount := app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Equal(len(depositsBefore)+5, len(depositsAccount))
	depositsAccount = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[1])
	suite.Require().Equal(len(depositsBefore)+1, len(depositsAccount))
	deposits = app.MillionsKeeper.ListDeposits(ctx)
	suite.Require().Equal(len(depositsBefore)+6, len(deposits))

	pools := app.MillionsKeeper.ListPools(ctx)

	for i, pool := range pools {
		// Test that the first 4 pools are updated
		if i <= 3 {
			depositsAccountPool := app.MillionsKeeper.ListAccountPoolDeposits(ctx, suite.addrs[0], pool.PoolId)
			suite.Require().Equal(len(depositsBefore)+1, len(depositsAccountPool))

			depositsPool := app.MillionsKeeper.ListPoolDeposits(ctx, pool.PoolId)
			suite.Require().Equal(len(depositsBefore)+1, len(depositsPool))
			// - Test that the TVL is added to the pool by the amount of the deposit
			suite.Require().Equal(poolTvlBefore.Int64()+1_000_000, pool.TvlAmount.Int64())
			// - Test that the deposit count is correctly incremented if it's a new depositor
			suite.Require().Equal(poolDepositorCountBefore+1, pool.DepositorsCount)

			// Test GetPoolDeposit
			poolDeposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, pool.PoolId, deposits[i].DepositId)
			suite.Require().NoError(err)
			suite.Require().Equal(deposits[i].Amount.Amount.Int64(), poolDeposit.Amount.Amount.Int64())
			suite.Require().Equal(deposits[i].PoolId, poolDeposit.PoolId)

			// Test that the last pool is updated accordingly
		} else if i == len(pools)-1 {
			depositsPool := app.MillionsKeeper.ListPoolDeposits(ctx, pool.PoolId)
			suite.Require().Equal(len(depositsBefore)+2, len(depositsPool))
			depositsAccountPool := app.MillionsKeeper.ListAccountPoolDeposits(ctx, suite.addrs[1], pool.PoolId)
			suite.Require().Equal(len(depositsBefore)+1, len(depositsAccountPool))
			// - Test that the TVL is added to the pool by the amount of the deposit for the last pool
			suite.Require().Equal(poolTvlBefore.Int64()+2_000_000, pool.TvlAmount.Int64())
			// - Test that the deposit count is correctly incremented if it's a new depositor for the last pool
			suite.Require().Equal(poolDepositorCountBefore+2, pool.DepositorsCount)

			// Test GetPoolDeposit
			poolDeposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, pool.PoolId, deposits[i].DepositId)
			suite.Require().NoError(err)
			suite.Require().Equal(deposits[i].Amount.Amount.Int64(), poolDeposit.Amount.Amount.Int64())
			suite.Require().Equal(deposits[i].PoolId, poolDeposit.PoolId)
		}
	}
}

// TestDeposit_RemoveDeposit tests the logic of removing a deposit from the store
func (suite *KeeperTestSuite) TestDeposit_RemoveDeposit() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx

	// Set the denom
	denom := app.StakingKeeper.BondDenom(ctx)

	depositsBefore := app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Len(depositsBefore, 0)
	deposits := app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Len(deposits, 0)

	// Track poolTVLBefore
	var poolTvlBefore math.Int
	// Track poolDepositorsCount
	var poolDepositorCountBefore uint64

	// Add 5 deposits
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

		// Test poolTVLBefore
		poolTvlBefore = pool.TvlAmount
		suite.Require().Equal(int64(0), poolTvlBefore.Int64())

		// Test poolDepositorCountBefore
		poolDepositorCountBefore = pool.DepositorsCount
		suite.Require().Equal(uint64(0), poolDepositorCountBefore)

		// Create a new deposit and add it to the state
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           pool.PoolId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}

	deposits = app.MillionsKeeper.ListDeposits(ctx)
	// - Test the validate basics
	// -- Test one successful deposit with all correct validate basics
	panicF := func() {
		app.MillionsKeeper.RemoveDeposit(ctx, &millionstypes.Deposit{
			PoolId:           deposits[len(deposits)-1].PoolId,
			DepositId:        deposits[len(deposits)-1].PoolId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().NotPanics(panicF)
	// -- Test that the deposit validation panics with an invalid pool ID
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		app.MillionsKeeper.RemoveDeposit(ctx, &millionstypes.Deposit{
			PoolId:           uint64(0),
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)
	// -- Test that the deposit validation panics with an invalid depositID
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		deposits[7].DepositId = 0
	}
	suite.Require().Panics(panicF)
	// -- Test that the deposit validation panics with an invalid state
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		app.MillionsKeeper.RemoveDeposit(ctx, &millionstypes.Deposit{
			PoolId:           deposits[len(deposits)-1].PoolId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_Unspecified,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)
	// -- Test that the deposit validation panics with an invalid DepositorAddress
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		app.MillionsKeeper.RemoveDeposit(ctx, &millionstypes.Deposit{
			PoolId:           deposits[len(deposits)-1].PoolId,
			DepositorAddress: "",
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)
	// -- Test that the deposit validation panics with an invalid WinnerAddress
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	panicF = func() {
		app.MillionsKeeper.RemoveDeposit(ctx, &millionstypes.Deposit{
			PoolId:           deposits[len(deposits)-1].PoolId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    "",
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}
	suite.Require().Panics(panicF)

	// - Test that the store is delete for pool deposits and account deposits
	// List all pools
	pools := app.MillionsKeeper.ListPools(ctx)

	for i, pool := range pools {
		if i <= 3 {
			depositsAccountPool := app.MillionsKeeper.ListAccountPoolDeposits(ctx, suite.addrs[0], pool.PoolId)
			suite.Require().Equal(len(depositsBefore)+1, len(depositsAccountPool))

			depositsPool := app.MillionsKeeper.ListPoolDeposits(ctx, pool.PoolId)
			suite.Require().Equal(len(depositsBefore)+1, len(depositsPool))
			// - Test that the TVL is added to the pool by the amount of the deposit
			suite.Require().Equal(poolTvlBefore.Int64()+1_000_000, pool.TvlAmount.Int64())
			// - Test that the deposit count is correctly incremented if it's a new depositor
			suite.Require().Equal(poolDepositorCountBefore+1, pool.DepositorsCount)
			// Check that deposit was removed from last pool
		} else if i == len(pools)-1 {
			depositsPool := app.MillionsKeeper.ListPoolDeposits(ctx, pool.PoolId)
			suite.Require().Equal(len(depositsBefore), len(depositsPool))
			depositsAccountPool := app.MillionsKeeper.ListAccountPoolDeposits(ctx, suite.addrs[0], pool.PoolId)
			suite.Require().Equal(len(depositsBefore), len(depositsAccountPool))
			// - Test that the TVL is added to the pool by the amount of the deposit for the last pool
			suite.Require().Equal(poolTvlBefore.Int64(), pool.TvlAmount.Int64())
			// - Test that the deposit count is correctly incremented if it's a new depositor for the last pool
			suite.Require().Equal(poolDepositorCountBefore, pool.DepositorsCount)

		}
	}

	for _, deposit := range deposits {
		// Get the pool before remove deposit
		poolBefore, err := app.MillionsKeeper.GetPool(ctx, deposit.PoolId)
		suite.Require().NoError(err)
		// Get the Account deposit before remove deposit
		addr := sdk.MustAccAddressFromBech32(deposit.DepositorAddress)
		deposits = app.MillionsKeeper.ListAccountDeposits(ctx, addr)
		// Get the Account pool deposit before remove deposit
		depositsAccountPoolBefore := app.MillionsKeeper.ListAccountPoolDeposits(ctx, addr, poolBefore.PoolId)
		// Get the pool deposit before remove deposit
		depositsPoolBefore := app.MillionsKeeper.ListPoolDeposits(ctx, poolBefore.PoolId)

		// Test GetPoolDeposit
		poolDeposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, deposit.PoolId, deposit.DepositId)
		suite.Require().NoError(err)
		suite.Require().Equal(deposit.Amount.Amount.Int64(), poolDeposit.Amount.Amount.Int64())
		suite.Require().Equal(deposit.PoolId, poolDeposit.PoolId)

		// Remove the deposit
		app.MillionsKeeper.RemoveDeposit(ctx, &deposit)

		// Decrement by 1 for each loop
		depositsAccount := app.MillionsKeeper.ListAccountDeposits(ctx, addr)
		suite.Require().Equal(len(deposits)-1, len(depositsAccount))

		// Decrement by 1 for each loop
		depositsAccountPool := app.MillionsKeeper.ListAccountPoolDeposits(ctx, addr, deposit.PoolId)
		suite.Require().Equal(len(depositsAccountPoolBefore)-1, len(depositsAccountPool))

		// Decrement by 1 for each loop
		depositsPool := app.MillionsKeeper.ListPoolDeposits(ctx, deposit.PoolId)
		suite.Require().Equal(len(depositsPoolBefore)-1, len(depositsPool))
		// Get the pool
		pool, err := app.MillionsKeeper.GetPool(ctx, deposit.PoolId)
		suite.Require().NoError(err)
		// Test that the pool tvl is deducted by the deposit amount
		suite.Require().Equal(poolBefore.TvlAmount.Int64()-deposit.Amount.Amount.Int64(), pool.TvlAmount.Int64())
		// - Test that the pool count is correctly decremented for each removed deposit
		poolBefore.DepositorsCount--
		suite.Require().Equal(poolBefore.DepositorsCount, pool.DepositorsCount)
	}

	depositsAll := app.MillionsKeeper.ListDeposits(ctx)
	suite.Require().Equal(len(depositsBefore), len(depositsAll))
}

// TestDeposit_UpdateDepositStatus tests the logic of updating a deposit from the store
func (suite *KeeperTestSuite) TestDeposit_UpdateDepositStatus() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx

	// Set the denom
	denom := app.StakingKeeper.BondDenom(ctx)

	// Add 5 deposits
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

		// Create a new deposit and add it to the state
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           pool.PoolId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
	}

	// - Test if the deposit exist
	// Retrived deposits
	deposits := app.MillionsKeeper.ListDeposits(ctx)
	// Retrieve the pool from the state
	pools := app.MillionsKeeper.ListPools(ctx)
	for i, pool := range pools {
		app.MillionsKeeper.UpdateDepositStatus(ctx, pool.PoolId, deposits[i].DepositId, millionstypes.DepositState_IcaDelegate, false)
		deposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, pool.PoolId, deposits[i].DepositId)
		suite.Require().NoError(err)
		suite.Require().NotNil(deposit, "Not nil")
		suite.Require().Equal(millionstypes.DepositState_IcaDelegate, deposit.State)
		suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)
	}
	deposits = app.MillionsKeeper.ListDeposits(ctx)
	// Trigger panic on wrong poolID
	panicF := func() {
		_, err := app.MillionsKeeper.GetPoolDeposit(ctx, deposits[10].PoolId, uint64(5))
		suite.Require().NoError(err)
	}
	suite.Require().Panics(panicF)
	// Trigger panic on wrong depositID
	panicF = func() {
		_, err := app.MillionsKeeper.GetPoolDeposit(ctx, uint64(5), deposits[10].PoolId)
		suite.Require().NoError(err)
	}
	suite.Require().Panics(panicF)

	// - Test Error Management
	deposits = app.MillionsKeeper.ListDeposits(ctx)
	// Retrieve the pool from the state
	pools = app.MillionsKeeper.ListPools(ctx)
	for i, pool := range pools {
		app.MillionsKeeper.UpdateDepositStatus(ctx, pool.PoolId, deposits[i].DepositId, millionstypes.DepositState_IcaDelegate, true)
		deposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, pool.PoolId, deposits[i].DepositId)
		suite.Require().NoError(err)
		suite.Require().NotNil(deposit, "Not nil")
		suite.Require().Equal(millionstypes.DepositState_Failure, deposit.State)
		suite.Require().Equal(millionstypes.DepositState_IcaDelegate, deposit.ErrorState)
	}

	// - Test that the account deposit was updated
	// - Test that the pool deposit was updated
	pools = app.MillionsKeeper.ListPools(ctx)
	for i, pool := range pools {
		depositsAccounts := app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
		suite.Require().Equal(ctx.BlockTime(), depositsAccounts[i].UpdatedAt)
		suite.Require().Equal(ctx.BlockHeight(), depositsAccounts[i].UpdatedAtHeight)

		depositsPoolAccount := app.MillionsKeeper.ListAccountPoolDeposits(ctx, suite.addrs[0], pool.PoolId)
		suite.Require().Equal(ctx.BlockTime(), depositsPoolAccount[0].UpdatedAt)
		suite.Require().Equal(ctx.BlockHeight(), depositsPoolAccount[0].UpdatedAtHeight)

		depositsPool := app.MillionsKeeper.ListPoolDeposits(ctx, pool.PoolId)
		suite.Require().Equal(ctx.BlockTime(), depositsPool[0].UpdatedAt)
		suite.Require().Equal(ctx.BlockHeight(), depositsPool[0].UpdatedAtHeight)
	}
}

// TestDeposit_TransferDeposit tests the full flow from the transfer till the delegation point
func (suite *KeeperTestSuite) TestDeposit_TransferDeposit() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta1 := 1 * time.Hour
	uatomAddresses := apptesting.AddTestAddrsWithDenom(app, ctx, 7, sdk.NewInt(1_000_0000_000), "uatom")

	// Remote pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:              poolID,
		Bech32PrefixValAddr: "cosmosvaloper",
		ChainId:             "cosmos",
		Denom:               "uatom",
		NativeDenom:         "uatom",
		ConnectionId:        "connection-id",
		TransferChannelId:   "transferChannel-id",
		Validators: map[string]*millionstypes.PoolValidator{
			cosmosPoolValidator: {
				OperatorAddress: cosmosPoolValidator,
				BondedAmount:    sdk.NewInt(1_000_000),
				IsEnabled:       true,
			},
		},
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
		AvailablePrizePool: sdk.NewCoin("uatom", math.NewInt(1000)),
	}))

	pools := app.MillionsKeeper.ListPools(ctx)
	suite.Require().Len(pools, 1)

	// - Test the full deposit process from the transfer till the delegation point
	// - Test if the validator is enabled and bonded amount
	suite.Require().Equal(true, pools[0].Validators[cosmosPoolValidator].IsEnabled)
	suite.Require().Equal(sdk.NewInt(1_000_000), pools[0].Validators[cosmosPoolValidator].BondedAmount)

	// Create a new deposit and add it to the state
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[0].PoolId,
		DepositorAddress: uatomAddresses[0].String(),
		WinnerAddress:    uatomAddresses[0].String(),
		State:            millionstypes.DepositState_IbcTransfer,
		Amount:           sdk.NewCoin("uatom", sdk.NewInt(1_000_000)),
	})
	err := app.BankKeeper.SendCoins(ctx, uatomAddresses[0], sdk.MustAccAddressFromBech32(pools[0].LocalAddress), sdk.Coins{sdk.NewCoin(pools[0].Denom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)
	deposits := app.MillionsKeeper.ListAccountDeposits(ctx, uatomAddresses[0])

	// Test that the pool is not local
	suite.Require().Equal(false, pools[0].IsLocalZone(ctx))
	deposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().NotNil(deposit, "Not nil")
	suite.Require().Equal(millionstypes.DepositState_IbcTransfer, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)

	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, uatomAddresses[0])
	suite.Require().Len(deposits, 1)

	// Simulate failed ackResponse AckResponseStatus_FAILURE
	err = app.MillionsKeeper.OnTransferDepositToRemoteZoneCompleted(ctx, deposits[0].GetPoolId(), deposits[0].GetDepositId(), true)
	suite.Require().NoError(err)

	// List deposits to get latest status
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, uatomAddresses[0])
	suite.Require().Len(deposits, 1)
	// Failed Transfer should set the errorState to DepositState_IbcTransfer and state to DepositState_Failure
	suite.Require().Equal(millionstypes.DepositState_Failure, deposits[0].State)
	suite.Require().Equal(millionstypes.DepositState_IbcTransfer, deposits[0].ErrorState)

	// Update status to simulate that the state is initially on DepositState_IbcTransfer
	app.MillionsKeeper.UpdateDepositStatus(ctx, deposits[0].PoolId, deposits[0].DepositId, millionstypes.DepositState_IbcTransfer, false)

	// Simulate succesful ackResponse AckResponseStatus_SUCCESS
	err = app.MillionsKeeper.OnTransferDepositToRemoteZoneCompleted(ctx, deposits[0].GetPoolId(), deposits[0].GetDepositId(), false)
	suite.Require().NoError(err)

	// List deposits to get latest state
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, uatomAddresses[0])
	suite.Require().Len(deposits, 1)
	// The deposit will reached the sequence, err := k.BroadcastICAMessages an fail on OnDelegateDepositOnRemoteZoneCompleted as we don't BroadcastICAMessages in our test
	suite.Require().Equal(millionstypes.DepositState_Failure, deposits[0].State)
	suite.Require().Equal(millionstypes.DepositState_IcaDelegate, deposits[0].ErrorState)

	// Test local pool
	// Set the denom
	denom := app.StakingKeeper.BondDenom(ctx)
	// local pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      app.MillionsKeeper.GetNextPoolID(ctx),
		Denom:       "ulum",
		NativeDenom: "ulum",
		Validators: map[string]*millionstypes.PoolValidator{
			suite.valAddrs[0].String(): {
				OperatorAddress: suite.valAddrs[0].String(),
				BondedAmount:    sdk.NewInt(1_000_000),
				IsEnabled:       true,
			},
		},
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
	pools = app.MillionsKeeper.ListPools(ctx)
	suite.Require().Len(pools, 2)

	// - Test if the validator is enabled and bonded amount
	suite.Require().Equal(true, pools[1].Validators[suite.valAddrs[0].String()].IsEnabled)
	suite.Require().Equal(sdk.NewInt(1_000_000), pools[1].Validators[suite.valAddrs[0].String()].BondedAmount)

	// Create a new deposit and add it to the state
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[1].PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_IbcTransfer,
		Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
	})
	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), sdk.Coins{sdk.NewCoin(pools[1].Denom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Len(deposits, 1)

	// Test that the pool is local
	suite.Require().Equal(true, pools[1].IsLocalZone(ctx))
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().NotNil(deposit, "Not nil")
	suite.Require().Equal(millionstypes.DepositState_IbcTransfer, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)

	// Trigger the Transfer deposit to native chain
	err = app.MillionsKeeper.TransferDepositToRemoteZone(ctx, deposit.PoolId, deposit.DepositId)
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DepositState_Success, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)
}

// TestDeposit_DelegateDeposit tests the delegate deposit from the transfer point till the final delegation
func (suite *KeeperTestSuite) TestDeposit_DelegateDeposit() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta1 := 1 * time.Hour
	uatomAddresses := apptesting.AddTestAddrsWithDenom(app, ctx, 7, sdk.NewInt(1_000_0000_000), "uatom")

	// Remote pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:              poolID,
		Bech32PrefixValAddr: "cosmosvaloper",
		ChainId:             "cosmos",
		Denom:               "uatom",
		NativeDenom:         "uatom",
		ConnectionId:        "connection-id",
		TransferChannelId:   "transferChannel-id",
		Validators: map[string]*millionstypes.PoolValidator{
			cosmosPoolValidator: {
				OperatorAddress: cosmosPoolValidator,
				BondedAmount:    sdk.NewInt(1_000_000),
				IsEnabled:       true,
			},
		},
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
		AvailablePrizePool: sdk.NewCoin("uatom", math.NewInt(1000)),
	}))

	pools := app.MillionsKeeper.ListPools(ctx)
	suite.Require().Len(pools, 1)

	// - Test if the validator is enabled and bonded amount
	suite.Require().Equal(true, pools[0].Validators[cosmosPoolValidator].IsEnabled)
	suite.Require().Equal(sdk.NewInt(1_000_000), pools[0].Validators[cosmosPoolValidator].BondedAmount)

	// Create a new deposit and add it to the state
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[0].PoolId,
		DepositorAddress: uatomAddresses[0].String(),
		WinnerAddress:    uatomAddresses[0].String(),
		State:            millionstypes.DepositState_IcaDelegate,
		Amount:           sdk.NewCoin("uatom", sdk.NewInt(1_000_000)),
	})
	err := app.BankKeeper.SendCoins(ctx, uatomAddresses[0], sdk.MustAccAddressFromBech32(pools[0].LocalAddress), sdk.Coins{sdk.NewCoin(pools[0].Denom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)
	deposits := app.MillionsKeeper.ListAccountDeposits(ctx, uatomAddresses[0])
	suite.Require().Len(deposits, 1)

	// Test that the pool is not local
	suite.Require().Equal(false, pools[0].IsLocalZone(ctx))
	deposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().NotNil(deposit, "Not nil")
	suite.Require().Equal(millionstypes.DepositState_IcaDelegate, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)

	// Simulate a failed DelegateDepositOnRemoteZone
	err = app.MillionsKeeper.DelegateDepositOnRemoteZone(ctx, deposit.PoolId, deposit.DepositId)
	suite.Require().NoError(err)

	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	// The deposit will reached the sequence, err := k.BroadcastICAMessages and fail on OnDelegateDepositOnRemoteZoneCompleted as we don't BroadcastICAMessages in our test
	suite.Require().Equal(millionstypes.DepositState_Failure, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_IcaDelegate, deposit.ErrorState)

	// Update status to simulate that the state is initially on DepositState_IcaDelegate
	app.MillionsKeeper.UpdateDepositStatus(ctx, deposit.PoolId, deposit.DepositId, millionstypes.DepositState_IcaDelegate, false)

	splits := []*millionstypes.SplitDelegation{{ValidatorAddress: cosmosPoolValidator, Amount: sdk.NewInt(int64(1_000_000))}}

	// Now trigger a successfull Delegation to simulate icacallbackstypes.AckResponseStatus_SUCCESS
	err = app.MillionsKeeper.OnDelegateDepositOnRemoteZoneCompleted(ctx, deposit.PoolId, deposit.DepositId, splits, false)
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposit.PoolId, deposit.DepositId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DepositState_Success, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)

	// Test local pool
	// Set the denom
	denom := app.StakingKeeper.BondDenom(ctx)
	// local pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      app.MillionsKeeper.GetNextPoolID(ctx),
		Denom:       "ulum",
		NativeDenom: "ulum",
		Validators: map[string]*millionstypes.PoolValidator{
			suite.valAddrs[0].String(): {
				OperatorAddress: suite.valAddrs[0].String(),
				BondedAmount:    sdk.NewInt(1_000_000),
				IsEnabled:       true,
			},
		},
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(denom, math.NewInt(1000)),
	}))
	pools = app.MillionsKeeper.ListPools(ctx)
	suite.Require().Len(pools, 2)

	// - Test if the validator is enabled and bonded amount
	suite.Require().Equal(true, pools[1].Validators[suite.valAddrs[0].String()].IsEnabled)
	suite.Require().Equal(sdk.NewInt(1_000_000), pools[1].Validators[suite.valAddrs[0].String()].BondedAmount)

	// Create a new deposit and add it to the state
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[1].PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_IcaDelegate,
		Amount:           sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_000)),
	})
	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), sdk.Coins{sdk.NewCoin(pools[1].Denom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Len(deposits, 1)

	pools = app.MillionsKeeper.ListPools(ctx)
	suite.Require().Len(pools, 2)

	// Test that the pool is local
	suite.Require().Equal(true, pools[1].IsLocalZone(ctx))
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().NotNil(deposit, "Not nil")
	suite.Require().Equal(millionstypes.DepositState_IcaDelegate, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)

	// As it's local pool trigger a DelegateDepositOnRemoteZone for the new deposit
	err = app.MillionsKeeper.DelegateDepositOnRemoteZone(ctx, deposit.PoolId, deposit.DepositId)
	suite.Require().NoError(err)

	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DepositState_Success, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)
}

// TestDeposit_BalanceDeposit tests the balance in case of success or failed deposit
func (suite *KeeperTestSuite) TestDeposit_BalanceDeposit() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx
	drawDelta1 := 1 * time.Hour
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	uatomAddresses := apptesting.AddTestAddrsWithDenom(app, ctx, 7, sdk.NewInt(1_000_0000_000), "uatom")

	// Remote pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:              poolID,
		Bech32PrefixValAddr: "cosmosvaloper",
		ChainId:             "cosmos",
		Denom:               "uatom",
		NativeDenom:         "uatom",
		ConnectionId:        "connection-id",
		TransferChannelId:   "transferChannel-id",
		Validators: map[string]*millionstypes.PoolValidator{
			cosmosPoolValidator: {
				OperatorAddress: cosmosPoolValidator,
				BondedAmount:    sdk.NewInt(1_000_000),
				IsEnabled:       true,
			},
		},
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
		AvailablePrizePool: sdk.NewCoin("uatom", math.NewInt(1000)),
	}))

	pools := app.MillionsKeeper.ListPools(ctx)
	suite.Require().Len(pools, 1)

	// Initialize balance before
	balanceBefore := app.BankKeeper.GetBalance(ctx, uatomAddresses[0], "uatom")
	suite.Require().Equal(int64(1_000_000_000_0), balanceBefore.Amount.Int64())
	// Initialize balance module account
	balanceModuleAccBefore := app.BankKeeper.GetBalance(ctx, suite.moduleAddrs[7], "uatom")
	suite.Require().Equal(int64(0), balanceModuleAccBefore.Amount.Int64())

	// - Test if the validator is enabled and bonded amount
	suite.Require().Equal(true, pools[0].Validators[cosmosPoolValidator].IsEnabled)
	suite.Require().Equal(sdk.NewInt(1_000_000), pools[0].Validators[cosmosPoolValidator].BondedAmount)

	// Create a new deposit and add it to the state
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[0].PoolId,
		DepositorAddress: uatomAddresses[0].String(),
		WinnerAddress:    uatomAddresses[0].String(),
		State:            millionstypes.DepositState_IbcTransfer,
		Amount:           sdk.NewCoin("uatom", sdk.NewInt(1_000_000)),
	})

	err := app.BankKeeper.SendCoins(ctx, uatomAddresses[0], sdk.MustAccAddressFromBech32(pools[0].LocalAddress), sdk.Coins{sdk.NewCoin(pools[0].Denom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)
	// Balance should have decreased by one deposit
	balance := app.BankKeeper.GetBalance(ctx, uatomAddresses[0], "uatom")
	suite.Require().Equal(balanceBefore.Amount.Int64()-1_000_000, balance.Amount.Int64())
	// Module Acc should have increased by one deposit
	balanceModuleAcc := app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(pools[0].LocalAddress), "uatom")
	suite.Require().Equal(balanceModuleAccBefore.Amount.Int64()+1_000_000, balanceModuleAcc.Amount.Int64())

	deposits := app.MillionsKeeper.ListAccountDeposits(ctx, uatomAddresses[0])
	suite.Require().Len(deposits, 1)

	app.MillionsKeeper.UpdateDepositStatus(ctx, deposits[0].PoolId, deposits[0].DepositId, millionstypes.DepositState_IcaDelegate, false)

	suite.Require().Equal(false, pools[0].IsLocalZone(ctx))
	deposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().NotNil(deposit, "Not nil")
	suite.Require().Equal(millionstypes.DepositState_IcaDelegate, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)

	splits := []*millionstypes.SplitDelegation{{ValidatorAddress: cosmosPoolValidator, Amount: sdk.NewInt(int64(1_000_000))}}

	// Now trigger a successfull Delegation to simulate icacallbackstypes.AckResponseStatus_SUCCESS
	err = app.MillionsKeeper.OnDelegateDepositOnRemoteZoneCompleted(ctx, deposit.PoolId, deposit.DepositId, splits, false)
	suite.Require().NoError(err)

	// For remote pools we don't reach the success as k.BroadcastICAMessages fails on OnDelegateDepositOnRemoteZoneCompleted

	// Balance should remain unchanged
	balance = app.BankKeeper.GetBalance(ctx, uatomAddresses[0], "uatom")
	suite.Require().Equal(balanceBefore.Amount.Int64()-1_000_000, balance.Amount.Int64())
	// Module Account should have the deposit amount
	balanceModuleAcc = app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(pools[0].LocalAddress), "uatom")
	suite.Require().Equal(balanceModuleAccBefore.Amount.Int64()+1_000_000, balanceModuleAcc.Amount.Int64())

	// There should no delegation shares
	delegationAmount := app.StakingKeeper.GetDelegatorDelegations(ctx, sdk.MustAccAddressFromBech32(pools[0].LocalAddress), 10)
	suite.Require().Len(delegationAmount, 0)

	// Test local pool
	// Set the denom
	denom := app.StakingKeeper.BondDenom(ctx)
	// local pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      app.MillionsKeeper.GetNextPoolID(ctx),
		Denom:       "ulum",
		NativeDenom: "ulum",
		Validators: map[string]*millionstypes.PoolValidator{
			suite.valAddrs[0].String(): {
				OperatorAddress: suite.valAddrs[0].String(),
				BondedAmount:    sdk.NewInt(1_000_000),
				IsEnabled:       true,
			},
		},
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
	pools = app.MillionsKeeper.ListPools(ctx)
	suite.Require().Len(pools, 2)

	// Initialize balance before
	balanceBefore = app.BankKeeper.GetBalance(ctx, suite.addrs[0], "ulum")
	suite.Require().Equal(int64(1_000_000_000_0), balanceBefore.Amount.Int64())
	// Initialize balance module account
	balanceModuleAccBefore = app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), "ulum")
	suite.Require().Equal(int64(0), balanceModuleAccBefore.Amount.Int64())

	// - Test if the validator is enabled and bonded amount
	suite.Require().Equal(true, pools[1].Validators[suite.valAddrs[0].String()].IsEnabled)
	suite.Require().Equal(sdk.NewInt(1_000_000), pools[1].Validators[suite.valAddrs[0].String()].BondedAmount)

	// Create a new deposit and add it to the state
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[1].PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_IbcTransfer,
		Amount:           sdk.NewCoin(denom, sdk.NewInt(2_000_000)),
	})
	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), sdk.Coins{sdk.NewCoin(pools[1].Denom, sdk.NewInt(2_000_000))})
	suite.Require().NoError(err)
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Len(deposits, 1)

	// Balance should have decreased by two deposits
	balance = app.BankKeeper.GetBalance(ctx, suite.addrs[0], "ulum")
	suite.Require().Equal(balanceBefore.Amount.Int64()-2_000_000, balance.Amount.Int64())
	// Module Acc should have increased by two deposits
	balanceModuleAcc = app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), "ulum")
	suite.Require().Equal(balanceModuleAccBefore.Amount.Int64()+2_000_000, balanceModuleAcc.Amount.Int64())

	// There should be no delegation shares yet
	delegationAmount = app.StakingKeeper.GetDelegatorDelegations(ctx, suite.moduleAddrs[0], 10)
	suite.Require().Len(delegationAmount, 0)

	suite.Require().Equal(true, pools[1].IsLocalZone(ctx))
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().NotNil(deposit, "Not nil")
	suite.Require().Equal(millionstypes.DepositState_IbcTransfer, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)

	err = app.MillionsKeeper.OnTransferDepositToRemoteZoneCompleted(ctx, deposit.PoolId, deposit.DepositId, true)
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DepositState_Failure, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_IbcTransfer, deposit.ErrorState)

	// Balance should be unchanged
	balance = app.BankKeeper.GetBalance(ctx, suite.addrs[0], "ulum")
	suite.Require().Equal(balanceBefore.Amount.Int64()-2_000_000, balance.Amount.Int64())
	// Module Acc should have still the deposited amount
	balanceModuleAcc = app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), "ulum")
	suite.Require().Equal(balanceModuleAccBefore.Amount.Int64()+2_000_000, balanceModuleAcc.Amount.Int64())

	// There should be no delegation shares yet
	delegationAmount = app.StakingKeeper.GetDelegatorDelegations(ctx, sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), 10)
	suite.Require().Len(delegationAmount, 0)

	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	// Update the status to trigger successful DepositToNativeChain
	app.MillionsKeeper.UpdateDepositStatus(ctx, deposit.PoolId, deposit.DepositId, millionstypes.DepositState_IbcTransfer, false)

	err = app.MillionsKeeper.TransferDepositToRemoteZone(ctx, deposit.PoolId, deposit.DepositId)
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DepositState_Success, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)

	// Balance should be unchanged
	balance = app.BankKeeper.GetBalance(ctx, suite.addrs[0], "ulum")
	suite.Require().Equal(balanceBefore.Amount.Int64()-2_000_000, balance.Amount.Int64())
	// Module Acc should have delegated the deposit
	balanceModuleAcc = app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), "ulum")
	suite.Require().Equal(balanceModuleAccBefore.Amount.Int64(), balanceModuleAcc.Amount.Int64())

	// The delegation amount should reflect the shares delegated - client precision
	delegationAmount = app.StakingKeeper.GetDelegatorDelegations(ctx, sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), 10)
	depositToShare := sdk.NewDecFromInt(sdk.NewInt(deposit.Amount.Amount.Int64() / 1_000_000))
	suite.Require().Equal(depositToShare, delegationAmount[0].Shares)
}

// TestDeposit_FullDepositProcess tests the complete logic for deposits for a local pool
func (suite *KeeperTestSuite) TestDeposit_FullDepositProcess() {
	// Set the app context
	app := suite.app
	ctx := suite.ctx
	drawDelta1 := 1 * time.Hour
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)

	// - Test the full deposit process for a local pool
	denom := app.StakingKeeper.BondDenom(ctx)
	// local pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      poolID,
		Denom:       "ulum",
		NativeDenom: "ulum",
		Validators: map[string]*millionstypes.PoolValidator{
			suite.valAddrs[0].String(): {
				OperatorAddress: suite.valAddrs[0].String(),
				BondedAmount:    sdk.NewInt(1_000_000),
				IsEnabled:       true,
			},
		},
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

	pools := app.MillionsKeeper.ListPools(ctx)
	suite.Require().Len(pools, 1)

	// Initialize balance before
	balanceBefore := app.BankKeeper.GetBalance(ctx, suite.addrs[0], "ulum")
	suite.Require().Equal(int64(1_000_000_000_0), balanceBefore.Amount.Int64())
	// Initialize balance module account
	balanceModuleAccBefore := app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), "ulum")
	suite.Require().Equal(int64(0), balanceModuleAccBefore.Amount.Int64())

	// - Test if the validator is enabled and bonded amount
	suite.Require().Equal(true, pools[0].Validators[suite.valAddrs[0].String()].IsEnabled)
	suite.Require().Equal(sdk.NewInt(1_000_000), pools[0].Validators[suite.valAddrs[0].String()].BondedAmount)

	// Create a new deposit and add it to the state
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[0].PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_IbcTransfer,
		Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
	})
	err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), sdk.Coins{sdk.NewCoin(pools[0].Denom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)
	deposits := app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Len(deposits, 1)

	// Balance should have decreased by one deposit
	balance := app.BankKeeper.GetBalance(ctx, suite.addrs[0], "ulum")
	suite.Require().Equal(balanceBefore.Amount.Int64()-1_000_000, balance.Amount.Int64())
	// Module Acc should have increased by one deposit
	balanceModuleAcc := app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), "ulum")
	suite.Require().Equal(balanceModuleAccBefore.Amount.Int64()+1_000_000, balanceModuleAcc.Amount.Int64())

	// There should no delegation shares
	delegationAmount := app.StakingKeeper.GetDelegatorDelegations(ctx, sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), 10)
	suite.Require().Len(delegationAmount, 0)

	// Test if it's a local pool
	suite.Require().Equal(true, pools[0].IsLocalZone(ctx))
	deposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().NotNil(deposit, "Not nil")
	suite.Require().Equal(millionstypes.DepositState_IbcTransfer, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)

	err = app.MillionsKeeper.TransferDepositToRemoteZone(ctx, deposit.PoolId, deposit.DepositId)
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DepositState_Success, deposit.State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposit.ErrorState)

	// Balance should be unchanged
	balance = app.BankKeeper.GetBalance(ctx, suite.addrs[0], "ulum")
	suite.Require().Equal(balanceBefore.Amount.Int64()-1_000_000, balance.Amount.Int64())
	// Module Acc should have delegated the deposit
	balanceModuleAcc = app.BankKeeper.GetBalance(ctx, suite.moduleAddrs[0], "ulum")
	suite.Require().Equal(balanceModuleAccBefore.Amount.Int64(), balanceModuleAcc.Amount.Int64())

	// The delegation amount should reflect the shares delegated - client precision
	delegationAmount = app.StakingKeeper.GetDelegatorDelegations(ctx, sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), 10)
	depositToShare := sdk.NewDecFromInt(sdk.NewInt(deposit.Amount.Amount.Int64() / 1_000_000))
	suite.Require().Equal(depositToShare, delegationAmount[0].Shares)
}
