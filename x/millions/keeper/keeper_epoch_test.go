package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/lum-network/chain/x/epochs/types"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

// TestEpoch_BeforeEpochStartHook tests the full epoch unbonding flow for withdrawals
func (suite *KeeperTestSuite) TestEpoch_BeforeEpochStartHook() {
	app := suite.app
	ctx := suite.ctx
	drawDelta1 := 1 * time.Hour
	var epochTracker millionstypes.EpochTracker

	// Assuming first epoch is 1, and nextEpochUnbonding is the 4th one
	for epoch := int64(1); epoch <= 4; epoch++ {
		epochInfo, err := TriggerEpochUpdate(suite)
		suite.Require().NoError(err)
		suite.Require().Equal(epoch, epochInfo.CurrentEpoch)

		epochTracker, err = TriggerEpochTrackerUpdate(suite, epochInfo)
		suite.Require().NoError(err)
	}

	// Test epochs with local pool
	// Unbonding frequency of 4 days
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      1,
		Denom:       localPoolDenom,
		NativeDenom: localPoolDenom,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: suite.valAddrs[0].String(),
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool:  sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
		UnbondingDuration:   time.Duration(millionstypes.DefaultUnbondingDuration),
		MaxUnbondingEntries: sdk.NewInt(millionstypes.DefaultMaxUnbondingEntries),
	}))
	// List pools
	pools := app.MillionsKeeper.ListPools(ctx)

	// - Test if the validator is enabled and bondedamount
	suite.Require().Equal(true, pools[0].Validators[0].IsEnabled)
	suite.Require().Equal(sdk.NewInt(1_000_000), pools[0].Validators[0].BondedAmount)

	for i := 0; i < 5; i++ {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           pools[0].PoolId,
			DepositorAddress: suite.addrs[i].String(),
			WinnerAddress:    suite.addrs[i].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_0)),
		})
	}

	err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)

	deposits := app.MillionsKeeper.ListDeposits(ctx)
	suite.Require().Len(deposits, 5)

	// Simulate the transfer to native chain
	for i := 0; i < 5; i++ {
		err = app.MillionsKeeper.TransferDepositToRemoteZone(ctx, deposits[i].PoolId, deposits[i].DepositId)
		suite.Require().NoError(err)
	}

	// Add 2 withdrawals for the first epoch
	for i := 0; i < 5; i++ {
		app.MillionsKeeper.AddWithdrawal(ctx, millionstypes.Withdrawal{
			PoolId:           deposits[i].PoolId,
			DepositId:        deposits[i].DepositId,
			DepositorAddress: suite.addrs[i].String(),
			ToAddress:        suite.addrs[i].String(),
			State:            millionstypes.WithdrawalState_Pending,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_0)),
		})
	}
	// There should be 2 withdrawals
	withdrawals := app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 5)

	for _, w := range withdrawals {
		err = app.MillionsKeeper.AddEpochUnbonding(ctx, w, false)
		suite.Require().NoError(err)
	}

	// Get the millions internal module tracker
	epochTracker, err = app.MillionsKeeper.GetEpochTracker(ctx, epochstypes.DAY_EPOCH, millionstypes.WithdrawalTrackerType)
	suite.Require().NoError(err)

	// Get epoch unbonding
	currentEpochUnbonding := app.MillionsKeeper.GetEpochUnbondings(ctx, epochTracker.EpochNumber+4)
	suite.Require().NoError(err)
	suite.Require().Len(currentEpochUnbonding, 1)
	suite.Require().Len(currentEpochUnbonding[0].WithdrawalIds, 5)
	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)

	for _, w := range withdrawals {
		suite.Require().Equal(millionstypes.WithdrawalState_Pending, w.State)
		suite.Require().Equal(millionstypes.WithdrawalState_Unspecified, w.ErrorState)
	}

	epochInfo, err := GetEpochInfo(suite)
	suite.Require().NoError(err)
	epochInfo = epochstypes.EpochInfo{
		Identifier:              epochInfo.Identifier,
		StartTime:               epochInfo.StartTime,
		Duration:                epochInfo.Duration,
		CurrentEpoch:            epochInfo.CurrentEpoch + 4,
		CurrentEpochStartTime:   epochInfo.CurrentEpochStartTime,
		EpochCountingStarted:    true,
		CurrentEpochStartHeight: 1,
	}
	app.MillionsKeeper.Hooks().BeforeEpochStart(ctx, epochInfo)

	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)

	for _, w := range withdrawals {
		suite.Require().Equal(millionstypes.WithdrawalState_IcaUnbonding, w.State)
		suite.Require().Equal(millionstypes.WithdrawalState_Unspecified, w.ErrorState)
		suite.Require().NotNil(w.UnbondingEndsAt)
	}
}

// TestEpoch_AddEpochUnbonding test the epoch unbonding for different pools
func (suite *KeeperTestSuite) TestEpoch_AddEpochUnbonding() {
	app := suite.app
	ctx := suite.ctx
	drawDelta1 := 1 * time.Hour
	var epochTracker millionstypes.EpochTracker
	frequency := uint64(4)

	// Assuming first epoch is 1, and nextEpochUnbonding is the 4th one
	for epoch := int64(1); epoch <= 4; epoch++ {
		epochInfo, err := TriggerEpochUpdate(suite)
		suite.Require().NoError(err)
		suite.Require().Equal(epoch, epochInfo.CurrentEpoch)

		epochTracker, err = TriggerEpochTrackerUpdate(suite, epochInfo)
		suite.Require().NoError(err)
	}

	// Test epochs with local pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      1,
		Denom:       localPoolDenom,
		NativeDenom: localPoolDenom,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: suite.valAddrs[0].String(),
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	}))

	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:              2,
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
	// List pools
	pools := app.MillionsKeeper.ListPools(ctx)

	for i := 0; i < 2; i++ {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           pools[0].PoolId,
			DepositorAddress: suite.addrs[i].String(),
			WinnerAddress:    suite.addrs[i].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_0)),
		})
	}

	err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)

	deposits := app.MillionsKeeper.ListDeposits(ctx)
	// Simulate the transfer to native chain
	for i := 0; i < 2; i++ {
		err = app.MillionsKeeper.TransferDepositToRemoteZone(ctx, deposits[i].PoolId, deposits[i].DepositId)
		suite.Require().NoError(err)
	}

	// Add 2 withdrawals for the first epoch
	for i := 0; i < 2; i++ {
		app.MillionsKeeper.AddWithdrawal(ctx, millionstypes.Withdrawal{
			PoolId:           deposits[i].PoolId,
			DepositId:        deposits[i].DepositId,
			DepositorAddress: suite.addrs[i].String(),
			ToAddress:        suite.addrs[i].String(),
			State:            millionstypes.WithdrawalState_Pending,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_0)),
		})
	}
	// There should be 2 withdrawals
	withdrawals := app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 2)

	for _, w := range withdrawals {
		err = app.MillionsKeeper.AddEpochUnbonding(ctx, w, false)
		suite.Require().NoError(err)
	}

	// Error should be triggered as no
	epochUnbondingPool, err := app.MillionsKeeper.GetEpochPoolUnbonding(ctx, epochTracker.EpochNumber+frequency, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(2_000_0), epochUnbondingPool.TotalAmount.Amount)
	suite.Require().Equal(uint64(1), epochUnbondingPool.PoolId)
	suite.Require().Equal([]uint64{1, 2}, epochUnbondingPool.WithdrawalIds)
	suite.Require().Equal(uint64(2), epochUnbondingPool.WithdrawalIdsCount)

	// Same withdrawalID to the epoch unbonding should faild
	err = app.MillionsKeeper.AddEpochUnbonding(ctx, withdrawals[0], false)
	suite.Require().ErrorIs(millionstypes.ErrEntityOverride, err)

	// Trigger new deposits for new pool
	for i := 0; i < 3; i++ {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           pools[1].PoolId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_0)),
		})
	}

	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)

	deposits = app.MillionsKeeper.ListDeposits(ctx)
	// Simulate the transfer to native chain
	for i := 2; i < 5; i++ {
		err = app.MillionsKeeper.TransferDepositToRemoteZone(ctx, deposits[i].PoolId, deposits[i].DepositId)
		suite.Require().Error(err)
	}

	// Add 2 withdrawals for the first epoch
	for i := 2; i < 5; i++ {
		app.MillionsKeeper.AddWithdrawal(ctx, millionstypes.Withdrawal{
			PoolId:           2,
			DepositId:        deposits[i].DepositId,
			DepositorAddress: suite.addrs[i].String(),
			ToAddress:        suite.addrs[i].String(),
			State:            millionstypes.WithdrawalState_Pending,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_0)),
		})
	}
	// There should be 5 withdrawals
	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 5)

	for _, w := range withdrawals[2:5] {
		err = app.MillionsKeeper.AddEpochUnbonding(ctx, w, false)
		suite.Require().NoError(err)
	}
	epochUnbondingPool2, err := app.MillionsKeeper.GetEpochPoolUnbonding(ctx, epochTracker.EpochNumber+frequency, 2)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(3_000_0), epochUnbondingPool2.TotalAmount.Amount)
	suite.Require().Equal(uint64(2), epochUnbondingPool2.PoolId)
	suite.Require().Equal([]uint64{3, 4, 5}, epochUnbondingPool2.WithdrawalIds)
	suite.Require().Equal(uint64(3), epochUnbondingPool2.WithdrawalIdsCount)
}

// TestEpoch_AddWithdrawalsToNextAvailableEpoch pushes the withdrawal to the next epoch if already full
func (suite *KeeperTestSuite) TestEpoch_AddWithdrawalsToNextAvailableEpoch() {
	app := suite.app
	ctx := suite.ctx
	drawDelta1 := 1 * time.Hour
	// Assuming first epoch is 1, and nextEpochUnbonding is the 4th one
	for epoch := int64(1); epoch <= 4; epoch++ {
		epochInfo, err := TriggerEpochUpdate(suite)
		suite.Require().NoError(err)
		suite.Require().Equal(epoch, epochInfo.CurrentEpoch)

		_, err = TriggerEpochTrackerUpdate(suite, epochInfo)
		suite.Require().NoError(err)
	}

	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      1,
		Denom:       localPoolDenom,
		NativeDenom: localPoolDenom,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: suite.valAddrs[0].String(),
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	}))

	pools := app.MillionsKeeper.ListPools(ctx)

	for i := 0; i < 305; i++ {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           1,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000)),
		})
	}

	err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(30_000_000))})
	suite.Require().NoError(err)

	deposits := app.MillionsKeeper.ListDeposits(ctx)
	// Add 30k+ withdrawals for the first epoch
	for i := 0; i < 305; i++ {
		app.MillionsKeeper.AddWithdrawal(ctx, millionstypes.Withdrawal{
			PoolId:           1,
			DepositId:        deposits[i].DepositId,
			DepositorAddress: suite.addrs[0].String(),
			ToAddress:        suite.addrs[0].String(),
			State:            millionstypes.WithdrawalState_Pending,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000)),
		})
	}
	withdrawals := app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 305)

	for _, w := range withdrawals {
		err = app.MillionsKeeper.AddEpochUnbonding(ctx, w, false)
		suite.Require().NoError(err)
	}

	// First 100 goes to the 8th epoch
	epochUnbonding, err := app.MillionsKeeper.GetEpochPoolUnbonding(ctx, 8, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(100), uint64(epochUnbonding.WithdrawalIdsCount))

	// Second 100 goes to the 12th epoch
	epochUnbonding, err = app.MillionsKeeper.GetEpochPoolUnbonding(ctx, 12, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(100), uint64(epochUnbonding.WithdrawalIdsCount))

	// Third 100 goes to the 16th epoch
	epochUnbonding, err = app.MillionsKeeper.GetEpochPoolUnbonding(ctx, 16, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(100), uint64(epochUnbonding.WithdrawalIdsCount))
	epochUnbonding, _ = app.MillionsKeeper.GetEpochPoolUnbonding(ctx, 9, 1)

	// Remaing 5 goes to the 20th epoch
	epochUnbonding, err = app.MillionsKeeper.GetEpochPoolUnbonding(ctx, 20, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(5), uint64(epochUnbonding.WithdrawalIdsCount))

	// Create second pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      2,
		Denom:       localPoolDenom,
		NativeDenom: localPoolDenom,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: suite.valAddrs[0].String(),
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	}))

	pools = app.MillionsKeeper.ListPools(ctx)

	// Add 30k+ withdrawals for the second pool
	for i := 0; i < 305; i++ {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           2,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000)),
		})
	}

	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(30_000_000))})
	suite.Require().NoError(err)

	deposits = app.MillionsKeeper.ListDeposits(ctx)
	// Add 30k+ withdrawals for the second pool
	for i := 305; i < 610; i++ {
		app.MillionsKeeper.AddWithdrawal(ctx, millionstypes.Withdrawal{
			PoolId:           2,
			DepositId:        deposits[i].DepositId,
			DepositorAddress: suite.addrs[0].String(),
			ToAddress:        suite.addrs[0].String(),
			State:            millionstypes.WithdrawalState_Pending,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000)),
		})
	}
	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 610)

	for _, w := range withdrawals[305:610] {
		err = app.MillionsKeeper.AddEpochUnbonding(ctx, w, false)
		suite.Require().NoError(err)
	}

	// Same process but with different poolID
	epochUnbonding, err = app.MillionsKeeper.GetEpochPoolUnbonding(ctx, 8, 2)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(100), uint64(epochUnbonding.WithdrawalIdsCount))

	epochUnbonding, err = app.MillionsKeeper.GetEpochPoolUnbonding(ctx, 12, 2)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(100), uint64(epochUnbonding.WithdrawalIdsCount))

	epochUnbonding, err = app.MillionsKeeper.GetEpochPoolUnbonding(ctx, 16, 2)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(100), uint64(epochUnbonding.WithdrawalIdsCount))

	epochUnbonding, err = app.MillionsKeeper.GetEpochPoolUnbonding(ctx, 20, 2)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(5), uint64(epochUnbonding.WithdrawalIdsCount))
}

// TestEpoch_AddFailedIcaUndelegationsToEpochUnbonding adds all failed withdrawals to the unbonding epoch
// Simulate migration test
func (suite *KeeperTestSuite) TestEpoch_AddFailedIcaUndelegationsToEpochUnbonding() {
	app := suite.app
	ctx := suite.ctx
	drawDelta1 := 1 * time.Hour

	epochInfo, err := TriggerEpochUpdate(suite)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(1), epochInfo.CurrentEpoch)
	app.MillionsKeeper.Hooks().BeforeEpochStart(ctx, epochInfo)

	// Initit pool with unbonding frequency of 1 day
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      1,
		Denom:       localPoolDenom,
		NativeDenom: localPoolDenom,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: suite.valAddrs[0].String(),
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool:  sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
		UnbondingDuration:   time.Duration(millionstypes.DefaultUnbondingDuration),
		MaxUnbondingEntries: sdk.NewInt(millionstypes.DefaultMaxUnbondingEntries),
	}))

	pools := app.MillionsKeeper.ListPools(ctx)

	for i := 0; i < 50; i++ {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           1,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000)),
		})
	}
	deposits := app.MillionsKeeper.ListDeposits(ctx)

	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(30_000_000))})
	suite.Require().NoError(err)

	for i := 0; i < 50; i++ {
		err := app.MillionsKeeper.TransferDepositToRemoteZone(ctx, deposits[i].PoolId, deposits[i].DepositId)
		suite.Require().NoError(err)
	}

	deposits = app.MillionsKeeper.ListDeposits(ctx)
	for i := 0; i < 50; i++ {
		app.MillionsKeeper.AddWithdrawal(ctx, millionstypes.Withdrawal{
			PoolId:           1,
			DepositId:        deposits[i].DepositId,
			DepositorAddress: suite.addrs[0].String(),
			ToAddress:        suite.addrs[0].String(),
			State:            millionstypes.WithdrawalState_Failure,
			ErrorState:       millionstypes.WithdrawalState_IcaUndelegate,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000)),
		})
	}

	err = app.MillionsKeeper.AddFailedIcaUndelegationsToEpochUnbonding(ctx)
	suite.Require().NoError(err)

	// Get the millions internal module tracker
	epochTracker, err := app.MillionsKeeper.GetEpochTracker(ctx, epochstypes.DAY_EPOCH, millionstypes.WithdrawalTrackerType)
	suite.Require().NoError(err)

	// In the current epoch there is no epochUnbondings
	epochUnbondings := app.MillionsKeeper.GetEpochUnbondings(ctx, uint64(epochTracker.EpochNumber))
	suite.Require().Equal(0, len(epochUnbondings))

	// The withdrawals should have been pushed to the first executable epoch wich is the 4th
	epochUnbondings = app.MillionsKeeper.GetEpochUnbondings(ctx, uint64(4))
	suite.Require().Equal(1, len(epochUnbondings))
	suite.Require().Equal(50, len(epochUnbondings[0].WithdrawalIds))

	withdrawals := app.MillionsKeeper.ListWithdrawals(ctx)

	for _, w := range withdrawals {
		suite.Require().Equal(millionstypes.WithdrawalState_Pending, w.State)
		suite.Require().Equal(millionstypes.WithdrawalState_Unspecified, w.ErrorState)
	}
}

// TestEpoch_RemoveEpochUnbonding tests the removal of an epoch unbonding
func (suite *KeeperTestSuite) TestEpoch_RemoveEpochUnbonding() {
	app := suite.app
	ctx := suite.ctx
	drawDelta1 := 1 * time.Hour
	var epochTracker millionstypes.EpochTracker
	frequency := uint64(4)
	// Assuming first epoch is 1, and nextEpochUnbonding is the 4th one
	for epoch := int64(1); epoch <= 4; epoch++ {
		epochInfo, err := TriggerEpochUpdate(suite)
		suite.Require().NoError(err)
		suite.Require().Equal(epoch, epochInfo.CurrentEpoch)

		epochTracker, err = TriggerEpochTrackerUpdate(suite, epochInfo)
		suite.Require().NoError(err)
	}

	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:      1,
		Denom:       localPoolDenom,
		NativeDenom: localPoolDenom,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: suite.valAddrs[0].String(),
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	}))

	pools := app.MillionsKeeper.ListPools(ctx)

	for i := 0; i < 50; i++ {
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           1,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000)),
		})
	}
	deposits := app.MillionsKeeper.ListDeposits(ctx)

	for i := 0; i < 50; i++ {
		err := app.MillionsKeeper.TransferDepositToRemoteZone(ctx, deposits[i].PoolId, deposits[i].DepositId)
		suite.Require().NoError(err)
	}

	err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(30_000_000))})
	suite.Require().NoError(err)

	deposits = app.MillionsKeeper.ListDeposits(ctx)
	for i := 0; i < 50; i++ {
		app.MillionsKeeper.AddWithdrawal(ctx, millionstypes.Withdrawal{
			PoolId:           1,
			DepositId:        deposits[i].DepositId,
			DepositorAddress: suite.addrs[0].String(),
			ToAddress:        suite.addrs[0].String(),
			State:            millionstypes.WithdrawalState_Pending,
			ErrorState:       millionstypes.WithdrawalState_Unspecified,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000)),
		})
	}

	withdrawals := app.MillionsKeeper.ListWithdrawals(ctx)
	for _, w := range withdrawals {
		err = app.MillionsKeeper.AddEpochUnbonding(ctx, w, false)
		suite.Require().NoError(err)
	}

	epochUnbonding, err := app.MillionsKeeper.GetEpochPoolUnbonding(ctx, epochTracker.EpochNumber+frequency, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(50), uint64(epochUnbonding.WithdrawalIdsCount))

	err = app.MillionsKeeper.RemoveEpochUnbonding(ctx, epochUnbonding)
	suite.Require().NoError(err)

	epochUnbonding, err = app.MillionsKeeper.GetEpochPoolUnbonding(ctx, epochTracker.EpochNumber+frequency, 1)
	suite.Require().ErrorIs(millionstypes.ErrInvalidEpochUnbonding, err)
}
