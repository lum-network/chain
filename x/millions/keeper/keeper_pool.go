package keeper

import (
	"fmt"
	"strconv"
	"time"

	icqueriestypes "github.com/lum-network/chain/x/icqueries/types"

	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"

	"github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distribtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	"github.com/lum-network/chain/x/millions/types"
)

// SetupPoolICA registers the ICA account on the native chain
// - waits for the ICA callback to move to OnSetupPoolICACompleted
// - or go to OnSetupPoolICACompleted directly if local zone
func (k Keeper) SetupPoolICA(ctx sdk.Context, poolID uint64) (*types.Pool, error) {
	// Acquire and deserialize our pool entity
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	if pool.IsLocalZone(ctx) {
		// Skip ICA setup for local pools
		return k.SetupPoolWithdrawalAddress(ctx, poolID)
	}

	// Get the chain ID from the connection
	chainID, err := k.GetChainID(ctx, pool.GetConnectionId())
	if err != nil {
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, "unable to obtain chain id from connection %s, err: %v", pool.GetConnectionId(), err)
	}
	if chainID != pool.ChainId {
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, "provided chain id %s differs from the connection chain id %s", pool.ChainId, chainID)
	}

	// Compute the app version structure for ICA registration
	appVersion, err := k.getPoolAppVersion(ctx, pool)
	if err != nil {
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
	}

	// Register the accounts deposit account first
	// Wait for this account to be setup to register the prize pool account
	// This is done to avoid race conditions for the last setup step (SetWithdrawAddress)
	pool.IcaDepositPortId = string(types.NewPoolName(pool.GetPoolId(), types.ICATypeDeposit))
	if err != nil {
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, fmt.Sprintf("Unable to create deposit account port id, err: %s", err.Error()))
	}
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, pool.GetConnectionId(), pool.GetIcaDepositPortId(), appVersion); err != nil {
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, fmt.Sprintf("Unable to trigger deposit account registration, err: %s", err.Error()))
	}

	k.updatePool(ctx, &pool)
	return &pool, nil
}

// OnPoolICASetupCompleted Acknowledge the ICA account creation on the native chain
// then moves to SetupPoolWithdrawalAddress once all ICA accounts have been created
func (k Keeper) OnSetupPoolICACompleted(ctx sdk.Context, poolID uint64, icaType string, icaAddress string) (*types.Pool, error) {
	logger := k.Logger(ctx).With("ctx", "pool_on_setup_ica_completed")

	// Grab our local pool instance
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	// Make sure our pool is in state created, otherwise just continue without error
	if pool.GetState() != types.PoolState_Created {
		return &pool, nil
	}

	// If it's a local pool, not further processing required
	if pool.IsLocalZone(ctx) {
		// Ignore this step for local pools
		return k.SetupPoolWithdrawalAddress(ctx, poolID)
	}

	if pool.IcaDepositAddress == "" && icaType == types.ICATypeDeposit && len(icaAddress) > 0 {
		// Assign the ICA deposit address
		pool.IcaDepositAddress = icaAddress

		if pool.IcaPrizepoolPortId == "" {
			// First time registering ICA Deposit
			// Initialize ICA PrizePool
			appVersion, err := k.getPoolAppVersion(ctx, pool)
			if err != nil {
				return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
			}
			pool.IcaPrizepoolPortId = string(types.NewPoolName(pool.GetPoolId(), types.ICATypePrizePool))
			if err != nil {
				return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, fmt.Sprintf("Unable to create prizepool account port id, err: %s", err.Error()))
			}
			if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, pool.GetConnectionId(), pool.GetIcaPrizepoolPortId(), appVersion); err != nil {
				logger.Error("Unable to trigger prizepool account registration, err: %s", err.Error())
			}
		}

		// Save pool state
		k.updatePool(ctx, &pool)
	} else if pool.IcaPrizepoolAddress == "" && icaType == types.ICATypePrizePool && len(icaAddress) > 0 {
		// Assign the ICA prize pool address
		pool.IcaPrizepoolAddress = icaAddress
		k.updatePool(ctx, &pool)
	}

	if len(pool.IcaDepositAddress) > 0 && len(pool.IcaPrizepoolAddress) > 0 {
		// Move to next step since both accounts have been created
		return k.SetupPoolWithdrawalAddress(ctx, poolID)
	}

	return &pool, nil
}

// SetupPoolWithdrawalAddress sets the PrizePoolAddress as the Staking withdrawal address for the DepositAddress
// - waits for the ICA callback to move to OnSetupPoolWithdrawalAddressCompleted
// - or go to OnSetupPoolWithdrawalAddressCompleted directly upon setting up the withdrawal address if local zone
func (k Keeper) SetupPoolWithdrawalAddress(ctx sdk.Context, poolID uint64) (*types.Pool, error) {
	logger := k.Logger(ctx).With("ctx", "pool_setup_withdrawal")

	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	if pool.GetState() != types.PoolState_Created {
		return &pool, nil
	}

	if pool.IsLocalZone(ctx) {
		if err := k.DistributionKeeper.SetWithdrawAddr(
			ctx,
			sdk.MustAccAddressFromBech32(pool.IcaDepositAddress),
			sdk.MustAccAddressFromBech32(pool.IcaPrizepoolAddress),
		); err != nil {
			logger.Error(
				fmt.Sprintf("failed to dispatch set withdrawal address for local pool: %v", err),
				"pool_id", poolID,
				"chain_id", pool.GetChainId(),
			)
			return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
		}
		return k.OnSetupPoolWithdrawalAddressCompleted(ctx, poolID)
	}

	callbackData := types.SetWithdrawAddressCallback{
		PoolId: poolID,
	}
	marshalledCallbackData, err := k.MarshalSetWithdrawAddressCallbackArgs(ctx, callbackData)
	if err != nil {
		return &pool, err
	}
	msgs := []sdk.Msg{&distribtypes.MsgSetWithdrawAddress{
		DelegatorAddress: pool.GetIcaDepositAddress(),
		WithdrawAddress:  pool.GetIcaPrizepoolAddress(),
	}}
	sequence, err := k.BroadcastICAMessages(ctx, pool, types.ICATypeDeposit, msgs, types.IBCTimeoutNanos, ICACallbackID_SetWithdrawAddress, marshalledCallbackData)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to dispatch ICA set withdraw address: %v", err),
			"pool_id", poolID,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
	} else {
		logger.Debug(
			"ICA set withdraw address dispatched",
			"pool_id", poolID,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
	}

	return &pool, nil
}

// OnSetupPoolWithdrawalAddressCompleted Acknowledge the withdrawal address configuration on the native chain
// then sets the pool to status ready in case of success
func (k Keeper) OnSetupPoolWithdrawalAddressCompleted(ctx sdk.Context, poolID uint64) (*types.Pool, error) {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	if pool.GetState() != types.PoolState_Created {
		return &pool, nil
	}
	pool.State = types.PoolState_Ready
	k.updatePool(ctx, &pool)
	return &pool, nil
}

// GetNextPoolID Return the next pool id to be used
func (k Keeper) GetNextPoolID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextPoolId := gogotypes.UInt64Value{}

	b := store.Get(types.NextPoolIdPrefix)
	if b == nil {
		panic(fmt.Errorf("getting at key (%v) should not have been nil", types.NextPoolIdPrefix))
	}
	k.cdc.MustUnmarshal(b, &nextPoolId)
	return nextPoolId.GetValue()
}

// SetNextPoolID sets next pool ID
func (k Keeper) SetNextPoolID(ctx sdk.Context, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: poolId})
	store.Set(types.NextPoolIdPrefix, bz)
}

func (k Keeper) GetNextPoolIDAndIncrement(ctx sdk.Context) uint64 {
	nextPoolId := k.GetNextPoolID(ctx)
	k.SetNextPoolID(ctx, nextPoolId+1)
	return nextPoolId
}

// HasPool Returns a boolean that indicates if the given poolID exists in the KVStore or not
func (k Keeper) HasPool(ctx sdk.Context, poolID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetPoolKey(poolID))
}

// AddPool Set a pool structure in the KVStore for a given pool ID
func (k Keeper) AddPool(ctx sdk.Context, pool *types.Pool) {
	// Automatically affect ID if missing
	if pool.GetPoolId() == types.UnknownID {
		pool.PoolId = k.GetNextPoolIDAndIncrement(ctx)
	}
	// Ensure payload is valid
	if err := pool.ValidateBasic(k.GetParams(ctx)); err != nil {
		panic(err)
	}
	// Ensure runner is valid
	if _, err := k.GetPoolRunner(pool.PoolType); err != nil {
		panic(err)
	}
	// Ensure we never override an existing entity
	if k.HasPool(ctx, pool.GetPoolId()) {
		panic(errorsmod.Wrapf(types.ErrEntityOverride, "ID %d", pool.GetPoolId()))
	}
	store := ctx.KVStore(k.storeKey)
	encodedPool := k.cdc.MustMarshal(pool)
	store.Set(types.GetPoolKey(pool.GetPoolId()), encodedPool)
}

// RegisterPool Register a given pool from the transaction message
func (k Keeper) RegisterPool(
	ctx sdk.Context,
	poolType types.PoolType,
	denom, nativeDenom, chainId, connectionId, transferChannelId string,
	vals []string,
	bech32Acc, bech32Val string,
	minDepositAmount math.Int,
	UnbondingDuration time.Duration,
	maxUnbondingEntries math.Int,
	drawSchedule types.DrawSchedule,
	prizeStrategy types.PrizeStrategy,
	fees []types.FeeTaker,
) (uint64, error) {

	// Acquire new pool ID
	poolID := k.GetNextPoolIDAndIncrement(ctx)

	// Initialize validators
	var validators []types.PoolValidator
	for _, addr := range vals {
		validators = append(validators, types.PoolValidator{
			OperatorAddress: addr,
			IsEnabled:       true,
			BondedAmount:    sdk.ZeroInt(),
		})
	}

	// Initialize our local deposit address
	localAddress := types.NewPoolAddress(poolID, types.ICATypeDeposit)
	poolAccount := k.AccountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(authtypes.NewBaseAccountWithAddress(localAddress), localAddress.String()))
	k.AccountKeeper.SetAccount(ctx, poolAccount)

	// Prepare new pool
	var pool = types.Pool{
		PoolId:              poolID,
		PoolType:            poolType,
		Denom:               denom,
		NativeDenom:         nativeDenom,
		ChainId:             chainId,
		ConnectionId:        connectionId,
		Validators:          validators,
		Bech32PrefixAccAddr: bech32Acc,
		Bech32PrefixValAddr: bech32Val,
		MinDepositAmount:    minDepositAmount,
		UnbondingDuration:   UnbondingDuration,
		MaxUnbondingEntries: maxUnbondingEntries,
		DrawSchedule:        drawSchedule.Sanitized(),
		PrizeStrategy:       prizeStrategy,
		LocalAddress:        localAddress.String(),
		NextDrawId:          1,
		TvlAmount:           sdk.ZeroInt(),
		DepositorsCount:     0,
		SponsorshipAmount:   sdk.ZeroInt(),
		AvailablePrizePool:  sdk.NewCoin(denom, sdk.ZeroInt()),
		State:               types.PoolState_Created,
		FeeTakers:           fees,
		TransferChannelId:   transferChannelId,
		CreatedAtHeight:     ctx.BlockHeight(),
		UpdatedAtHeight:     ctx.BlockHeight(),
		CreatedAt:           ctx.BlockTime(),
		UpdatedAt:           ctx.BlockTime(),
	}

	// Validate pool configuration
	if err := pool.ValidateBasic(k.GetParams(ctx)); err != nil {
		return 0, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
	}
	// Validate pool runner exists
	if _, err := k.GetPoolRunner(poolType); err != nil {
		return 0, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
	}

	// If it's a local zone, we have more setup steps for module accounts
	if pool.IsLocalZone(ctx) {
		// Set the deposit address to the local module account address
		pool.IcaDepositAddress = localAddress.String()

		// Initialize our local prizepool address
		icaPrizePoolAddress := types.NewPoolAddress(poolID, types.ICATypePrizePool)
		icaPrizePoolAccount := k.AccountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(authtypes.NewBaseAccountWithAddress(icaPrizePoolAddress), icaPrizePoolAddress.String()))
		k.AccountKeeper.SetAccount(ctx, icaPrizePoolAccount)

		pool.IcaPrizepoolAddress = icaPrizePoolAddress.String()
	}

	// Commit the pool to the KVStore
	k.AddPool(ctx, &pool)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeRegisterPool,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(pool.PoolId, 10)),
		),
	})

	// Remote zone pool
	_, err := k.SetupPoolICA(ctx, poolID)
	if err != nil {
		return poolID, err
	}

	return poolID, nil
}

// UpdatePool Update the updatable properties of a pool from the transaction message
func (k Keeper) UpdatePool(
	ctx sdk.Context,
	poolID uint64,
	vals []string,
	minDepositAmount *math.Int,
	UnbondingDuration *time.Duration,
	maxUnbondingEntries *math.Int,
	drawSchedule *types.DrawSchedule,
	prizeStrategy *types.PrizeStrategy,
	state types.PoolState,
	feeTakers []types.FeeTaker,
) error {
	// Acquire and deserialize our pool entity
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	// Only a few properties can be updated
	if minDepositAmount != nil {
		pool.MinDepositAmount = *minDepositAmount
	}
	if UnbondingDuration != nil {
		pool.UnbondingDuration = *UnbondingDuration
	}
	if maxUnbondingEntries != nil {
		pool.MaxUnbondingEntries = *maxUnbondingEntries
	}
	if drawSchedule != nil {
		pool.DrawSchedule = *drawSchedule
		if pool.DrawSchedule.InitialDrawAt.After(ctx.BlockTime()) {
			// Specifying a new valid InitialDrawAt resets the Pool draw timing to this date
			// Also useful for governance to change the timing of Draws in case of time drift
			pool.LastDrawCreatedAt = nil
		}
	}
	if prizeStrategy != nil {
		pool.PrizeStrategy = *prizeStrategy
	}
	if feeTakers != nil {
		pool.FeeTakers = feeTakers
	}

	// Update pool state only if current pool state is in paused and incoming state ready
	// else if current pool state is in ready and incoming state paused
	if state == types.PoolState_Paused && pool.State == types.PoolState_Ready {
		pool.State = state
	} else if state == types.PoolState_Ready && pool.State == types.PoolState_Paused {
		pool.State = state
	} else if state != types.PoolState_Unspecified {
		return types.ErrPoolStateChangeNotAllowed
	}

	// Update enabled validators
	if pool.PoolType == types.PoolType_Staking && len(vals) > 0 {
		for i := range pool.Validators {
			pool.Validators[i].IsEnabled = false
		}
		valIdx := pool.GetValidatorsMapIndex()
		for _, addr := range vals {
			if _, exists := valIdx[addr]; exists {
				pool.Validators[valIdx[addr]].IsEnabled = true
			} else {
				pool.Validators = append(pool.Validators, types.PoolValidator{
					OperatorAddress: addr,
					IsEnabled:       true,
					BondedAmount:    sdk.ZeroInt(),
				})
				valIdx[addr] = len(pool.Validators) - 1
			}
		}
	}

	// Validate pool configuration
	if err := pool.ValidateBasic(k.GetParams(ctx)); err != nil {
		return errorsmod.Wrapf(types.ErrFailedToUpdatePool, err.Error())
	}
	// Validate pool runner exists
	poolRunner, err := k.GetPoolRunner(pool.PoolType)
	if err != nil {
		return errorsmod.Wrapf(types.ErrFailedToUpdatePool, err.Error())
	}
	// Commit the pool to the KVStore
	k.updatePool(ctx, &pool)

	// Trigger rebalance distribution
	if err := poolRunner.OnUpdatePool(ctx, pool); err != nil {
		return errorsmod.Wrapf(types.ErrFailedToUpdatePool, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeUpdatePool,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(pool.PoolId, 10)),
		),
	})

	return nil
}

// ClosePool It is a deterministic method that proceed with pool closure steps based on the current pool state
func (k Keeper) ClosePool(ctx sdk.Context, poolID uint64, finalStep bool) error {
	// Make sure we always have a valid pool entity
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	// Make sure the pool is not already closed
	if pool.State == types.PoolState_Closed {
		return types.ErrPoolClosed
	}

	// If the pool is in state "ready", it means we are in the very first step of the closure process
	if pool.State == types.PoolState_Ready {
		// Update the pool state to "Closing"
		pool.State = types.PoolState_Closing
		k.updatePool(ctx, &pool)
		k.Logger(ctx).Info(fmt.Sprintf("Updated pool %d to state closing", poolID))

		// Acquire all deposits
		deposits := k.ListPoolDeposits(ctx, poolID)

		// Update the prize strategy now to distribute 100% of the prize pool to the depositors
		// Do it now because once you create a withdrawal, you delete deposit
		pool.PrizeStrategy = types.PrizeStrategy{
			PrizeBatches: []types.PrizeBatch{
				{Quantity: uint64(len(deposits)), PoolPercent: 1, DrawProbability: sdk.NewDec(1), IsUnique: true},
			},
		}
		k.updatePool(ctx, &pool)
		k.Logger(ctx).Info(fmt.Sprintf("Updated pool %d prize strategy to distribute 100%% of the prize pool to the depositors", poolID))

		// Iterate over all deposits
		for _, deposit := range deposits {
			// Make sure the deposit can be withdrawn
			if deposit.State != types.DepositState_Success {
				continue
			}

			// Trigger the withdrawal
			withdrawal := types.Withdrawal{
				PoolId:           poolID,
				DepositId:        deposit.DepositId,
				WithdrawalId:     k.GetNextWithdrawalIdAndIncrement(ctx),
				State:            types.WithdrawalState_Pending,
				DepositorAddress: deposit.GetDepositorAddress(),
				ToAddress:        deposit.GetDepositorAddress(),
				Amount:           deposit.Amount,
				CreatedAtHeight:  ctx.BlockHeight(),
				UpdatedAtHeight:  ctx.BlockHeight(),
				CreatedAt:        ctx.BlockTime(),
				UpdatedAt:        ctx.BlockTime(),
			}

			// Adds the withdrawal and remove the deposit
			k.AddWithdrawal(ctx, withdrawal)
			k.RemoveDeposit(ctx, &deposit)
		}

		k.Logger(ctx).Info(fmt.Sprintf("Created %d withdrawals for pool %d", len(deposits), poolID))
	} else if pool.State == types.PoolState_Closing {
		if !finalStep {
			// Acquire the withdrawals and make sure we have none
			withdrawals := k.ListPoolWithdrawals(ctx, poolID)
			if len(withdrawals) > 0 {
				k.Logger(ctx).Debug(fmt.Sprintf("Pool %d has %d withdrawals pending", poolID, len(withdrawals)))
				return nil
			}

			// If the last draw state is success, it means we haven't launched the final draw yet
			if pool.LastDrawState == types.DrawState_Success {
				draw, err := k.LaunchNewDraw(ctx, poolID)
				if err != nil {
					return err
				}
				k.Logger(ctx).Info(fmt.Sprintf("Updated pool %d to launch draw with drawID %d", poolID, draw.DrawId))
			}
		} else {
			// Only proceed if the final draw succeeded
			if pool.LastDrawState == types.DrawState_Success {
				// Transfer the remaining dust on local module account to community pool
				// This is to avoid having dust on the module account
				// This is not a critical step, so we don't need to return an error if it fails
				currentLocalBalance := k.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(pool.LocalAddress), pool.Denom)
				if currentLocalBalance.IsPositive() {
					if err := k.BankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, distribtypes.ModuleName, sdk.NewCoins(currentLocalBalance)); err != nil {
						k.Logger(ctx).Error(fmt.Sprintf("Failed to transfer dust from pool %d local module account to community pool: %v", poolID, err))
					}
				}

				// Once everything is done, we can update the state to closed
				pool.State = types.PoolState_Closed
				k.updatePool(ctx, &pool)
				k.Logger(ctx).Info(fmt.Sprintf("Updated pool %d to state closed", poolID))
			} else {
				k.Logger(ctx).Error(fmt.Sprintf("Pool %d cannot be closed because the last draw state is not success", poolID))
			}
		}
	} else {
		return types.ErrPoolNotClosing
	}

	return nil
}

// RebalanceValidatorsBondings allows rebalancing of validators bonded assets
// Current implementation:
// - Initiate an even redelegate distribution from inactive bonded validators to active validators
func (k Keeper) RebalanceValidatorsBondings(ctx sdk.Context, poolID uint64) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	// Make sure pool is ready
	if pool.State == types.PoolState_Created || pool.State == types.PoolState_Unspecified {
		return types.ErrPoolNotReady
	}

	_, valsSrc := pool.BondedValidators()
	for _, valSrc := range valsSrc {
		// Double check that valSrc is inactive
		if !valSrc.IsEnabled {
			if err := k.RedelegateToActiveValidatorsOnRemoteZone(ctx, pool.PoolId, valSrc.GetOperatorAddress()); err != nil {
				return err
			}
		}
	}

	return nil
}

// RedelegateToActiveValidatorsOnRemoteZone redistribute evenly the bondedAmount from the bonded inactive to the active valitator set of the pool
func (k Keeper) RedelegateToActiveValidatorsOnRemoteZone(ctx sdk.Context, poolID uint64, valSrcAddr string) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	// Make sure pool is ready
	if pool.State == types.PoolState_Created || pool.State == types.PoolState_Unspecified {
		return types.ErrPoolNotReady
	}

	// Get the validator to redelegate
	valIdx := pool.GetValidatorsMapIndex()
	index, found := valIdx[valSrcAddr]
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "%s", valSrcAddr)
	}
	inactiveVal := pool.Validators[index]

	// Check that the validator is inactive
	if inactiveVal.IsEnabled {
		return errorsmod.Wrapf(
			types.ErrInvalidValidatorEnablementStatus,
			"status is %t instead of %t",
			inactiveVal.IsEnabled, !inactiveVal.IsEnabled,
		)
	}

	splits := pool.ComputeSplitDelegations(ctx, inactiveVal.BondedAmount)
	if len(splits) == 0 {
		return types.ErrPoolEmptySplitDelegations
	}

	// Validate pool runner
	poolRunner := k.MustGetPoolRunner(pool.PoolType)

	// ApplySplitRedelegate to pool validator set
	pool.ApplySplitRedelegate(ctx, inactiveVal.GetOperatorAddress(), splits)
	k.updatePool(ctx, &pool)

	if err := poolRunner.RedelegateToActiveValidatorsOnRemoteZone(ctx, pool, inactiveVal, splits); err != nil {
		return err
	}

	if pool.IsLocalZone(ctx) {
		return k.OnRedelegateToActiveValidatorsOnRemoteZoneCompleted(ctx, pool.PoolId, inactiveVal.GetOperatorAddress(), splits, false)
	}

	return nil
}

// OnRedelegateToActiveValidatorsOnRemoteZoneCompleted Acknowledged a redelegation of an inactive validator's bondedAmount
func (k Keeper) OnRedelegateToActiveValidatorsOnRemoteZoneCompleted(ctx sdk.Context, poolID uint64, valSrcAddr string, splits []*types.SplitDelegation, isError bool) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	// Make sure pool is ready
	if pool.State == types.PoolState_Created || pool.State == types.PoolState_Unspecified {
		return types.ErrPoolNotReady
	}

	// Get the validator
	valIdx := pool.GetValidatorsMapIndex()
	index, found := valIdx[valSrcAddr]
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "%s", valSrcAddr)
	}
	inactiveVal := pool.Validators[index]

	// RevertSplitRedelegate in case of failure
	if isError {
		pool.RevertSplitRedelegate(ctx, inactiveVal.GetOperatorAddress(), splits)
		k.updatePool(ctx, &pool)
		return nil
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeRedelegate,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(pool.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, inactiveVal.GetOperatorAddress()),
		),
	})

	return nil
}

// UnsafeUpdatePoolPortIds This method raw update the provided pool port ids.
// It's heavily unsafe and could break the ICA implementation. It should only be used by store migrations.
func (k Keeper) UnsafeUpdatePoolPortIds(ctx sdk.Context, poolID uint64, icaDepositPortId, icaPrizePoolPortId string) (types.Pool, error) {
	// Grab our pool instance
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return types.Pool{}, err
	}

	// Patch and update our pool entity
	pool.IcaDepositPortId = icaDepositPortId
	pool.IcaPrizepoolPortId = icaPrizePoolPortId
	k.updatePool(ctx, &pool)
	return pool, nil
}

// UnsafeUpdatePoolUnbondingFrequencyAndType raw updates the UnbondingDuration, mexUnbonding and pool type entries
// Unsafe and should only be used for store migration
func (k Keeper) UnsafeUpdatePoolUnbondingFrequencyAndType(ctx sdk.Context, poolID uint64, UnbondingDuration time.Duration, maxUnbondingEntries math.Int, poolType types.PoolType) (types.Pool, error) {
	// Grab our pool instance
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return types.Pool{}, err
	}

	// Patch and update our pool entity
	pool.UnbondingDuration = UnbondingDuration
	pool.MaxUnbondingEntries = maxUnbondingEntries
	pool.PoolType = poolType
	k.updatePool(ctx, &pool)

	return pool, nil
}

// UnsafeUpdatePoolFeeTakers raw updates the fee takers entries
// Unsafe and should only be used for store migration
func (k Keeper) UnsafeUpdatePoolFeeTakers(ctx sdk.Context, poolID uint64, feeTakers []types.FeeTaker) (types.Pool, error) {
	// Grab our pool instance
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return types.Pool{}, err
	}

	// Patch and update our pool entity
	pool.FeeTakers = feeTakers
	k.updatePool(ctx, &pool)

	return pool, nil
}

func (k Keeper) updatePool(ctx sdk.Context, pool *types.Pool) {
	pool.UpdatedAt = ctx.BlockTime()
	pool.UpdatedAtHeight = ctx.BlockHeight()
	// Ensure payload is valid
	if err := pool.ValidateBasic(k.GetParams(ctx)); err != nil {
		panic(err)
	}
	// Validate pool runner exists
	if _, err := k.GetPoolRunner(pool.PoolType); err != nil {
		panic(err)
	}
	store := ctx.KVStore(k.storeKey)
	encodedPool := k.cdc.MustMarshal(pool)
	store.Set(types.GetPoolKey(pool.GetPoolId()), encodedPool)
}

// GetPool Returns a pool instance for the given poolID
func (k Keeper) GetPool(ctx sdk.Context, poolID uint64) (types.Pool, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolKey(poolID))
	if bz == nil {
		return types.Pool{}, errorsmod.Wrapf(types.ErrPoolNotFound, "%d", poolID)
	}

	var pool types.Pool
	if err := k.cdc.Unmarshal(bz, &pool); err != nil {
		return types.Pool{}, err
	}

	return pool, nil
}

func (k Keeper) GetPoolForChainID(ctx sdk.Context, chainID string) (types.Pool, bool) {
	var pool = types.Pool{}
	found := false
	k.IteratePools(ctx, func(p types.Pool) bool {
		if p.GetChainId() == chainID {
			pool = p
			found = true
			return true
		}
		return false
	})

	return pool, found
}

func (k Keeper) GetPoolForConnectionID(ctx sdk.Context, connectionID string) (types.Pool, bool) {
	var pool = types.Pool{}
	found := false
	k.IteratePools(ctx, func(p types.Pool) bool {
		if p.GetConnectionId() == connectionID {
			pool = p
			found = true
			return true
		}
		return false
	})

	return pool, found
}

func (k Keeper) GetPoolForControllerPortID(ctx sdk.Context, controllerPortID string) (types.Pool, bool) {
	var pool = types.Pool{}
	found := false
	k.IteratePools(ctx, func(p types.Pool) bool {
		if p.GetIcaDepositPortIdWithPrefix() == controllerPortID || p.GetIcaPrizepoolPortIdWithPrefix() == controllerPortID {
			pool = p
			found = true
			return true
		}
		return false
	})

	return pool, found
}

// IteratePools Iterate over the pools, and for each entry call the callback
func (k Keeper) IteratePools(ctx sdk.Context, callback func(pool types.Pool) (stop bool)) {
	iterator := k.PoolsIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var pool types.Pool
		k.cdc.MustUnmarshal(iterator.Value(), &pool)
		if callback(pool) {
			break
		}
	}
}

// PoolsIterator Return a ready to use iterator for the pools store
func (k Keeper) PoolsIterator(ctx sdk.Context) sdk.Iterator {
	kvStore := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(kvStore, types.PoolPrefix)
}

// ListPools Return the pools
func (k Keeper) ListPools(ctx sdk.Context) (pools []types.Pool) {
	k.IteratePools(ctx, func(pool types.Pool) bool {
		pools = append(pools, pool)
		return false
	})
	return pools
}

// ListPoolsToDraw Returns the pools which should be launching a Draw
func (k Keeper) ListPoolsToDraw(ctx sdk.Context) (pools []types.Pool) {
	allPools := k.ListPools(ctx)
	for _, p := range allPools {
		if p.ShouldDraw(ctx) {
			pools = append(pools, p)
		}
	}
	return pools
}

// BroadcastIBCTransfer broadcasts and IBC transfer for the a given amount to the Pool remote zone ICA Deposit account from the Pool Local address
// Amount denom must be based on pool.Denom
// Also registers the requested callback
// This method should not and DOEST NOT update the pool state
func (k Keeper) BroadcastIBCTransfer(ctx sdk.Context, pool types.Pool, amount sdk.Coin, timeoutNanos uint64, callbackID string, callbackArgs []byte) (uint64, error) {
	logger := k.Logger(ctx).With("ctx", "pool_broadcast_ibc")

	// We always want our pool to have passed the created state and not be unspecified
	if pool.State == types.PoolState_Created || pool.State == types.PoolState_Unspecified {
		return 0, types.ErrPoolNotReady
	}

	// Timeout is now plus 30 minutes in nanoseconds
	// We use the standard transfer port ID and not the one opened for ICA
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + timeoutNanos
	// From Local to Remote - use transfer channel ID
	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		pool.GetTransferChannelId(),
		amount,
		pool.GetLocalAddress(),
		pool.GetIcaDepositAddress(),
		clienttypes.Height{},
		timeoutTimestamp,
		"Cosmos Millions",
	)

	// Broadcast the transfer
	res, err := k.IBCTransferKeeper.Transfer(ctx, msg)
	if err != nil {
		return 0, err
	}

	if callbackID != "" && callbackArgs != nil {
		callback := icacallbackstypes.CallbackData{
			CallbackKey:  icacallbackstypes.PacketID(msg.SourcePort, msg.SourceChannel, res.Sequence),
			PortId:       msg.SourcePort,
			ChannelId:    msg.SourceChannel,
			Sequence:     res.Sequence,
			CallbackId:   callbackID,
			CallbackArgs: callbackArgs,
		}
		k.ICACallbacksKeeper.SetCallbackData(ctx, callback)
	}

	logger.Debug(
		fmt.Sprintf("Broadcasted IBC transfer with sequence %d", res.Sequence),
		"pool_id", pool.GetPoolId(),
	)
	return res.Sequence, nil
}

// BroadcastICAMessages broadcasts ICA transaction containing the specified Messages
// Also registers the requested callback
// This method should not and DOEST NOT update the pool state
func (k Keeper) BroadcastICAMessages(ctx sdk.Context, pool types.Pool, accountType string, msgs []sdk.Msg, timeoutNanos uint64, callbackID string, callbackArgs []byte) (uint64, error) {
	logger := k.Logger(ctx).With("ctx", "pool_broadcast_ica")

	// Pool must be ready or created (creating) to process those kind of operations
	if pool.GetState() != types.PoolState_Ready && pool.GetState() != types.PoolState_Created {
		return 0, types.ErrPoolNotReady
	}

	// Acquire the ICA PortID
	// Compute the port ID
	var portOwner string
	var portID string
	if accountType == types.ICATypeDeposit {
		portOwner = pool.GetIcaDepositPortId()
		portID = pool.GetIcaDepositPortIdWithPrefix()
	} else if accountType == types.ICATypePrizePool {
		portOwner = pool.GetIcaPrizepoolPortId()
		portID = pool.GetIcaPrizepoolPortIdWithPrefix()
	} else {
		return 0, fmt.Errorf("unknown account type %q", accountType)
	}

	// Acquire the channel capacities
	channelID, found := k.ICAControllerKeeper.GetActiveChannelID(ctx, pool.GetConnectionId(), portID)
	if !found {
		return 0, errorsmod.Wrapf(icatypes.ErrActiveChannelNotFound, "Millions failed to retrieve open active channel for port %s (%s / %s) on connection %s", portID, pool.GetIcaDepositPortId(), pool.GetIcaPrizepoolPortId(), pool.GetConnectionId())
	}

	// Serialize the data and construct the packet to send
	var protoMsgs []proto.Message
	for _, msg := range msgs {
		protoMsgs = append(protoMsgs, msg)
	}
	data, err := icatypes.SerializeCosmosTx(k.cdc, protoMsgs)
	if err != nil {
		return 0, err
	}
	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: "Cosmos Millions ICA",
	}

	// Broadcast the messages
	msgServer := icacontrollerkeeper.NewMsgServerImpl(&k.ICAControllerKeeper)
	msgSendTx := icacontrollertypes.NewMsgSendTx(portOwner, pool.GetConnectionId(), timeoutNanos, packetData)
	res, err := msgServer.SendTx(ctx, msgSendTx)
	if err != nil {
		return 0, err
	}
	sequence := res.Sequence

	// Store the callback data
	if callbackID != "" && callbackArgs != nil {
		callback := icacallbackstypes.CallbackData{
			CallbackKey:  icacallbackstypes.PacketID(portID, channelID, sequence),
			PortId:       portID,
			ChannelId:    channelID,
			Sequence:     sequence,
			CallbackId:   callbackID,
			CallbackArgs: callbackArgs,
		}
		k.ICACallbacksKeeper.SetCallbackData(ctx, callback)
	}

	logger.Debug(
		fmt.Sprintf("Broadcasted ICA messages with sequence %d", sequence),
		"pool_id", pool.GetPoolId(),
	)
	return sequence, nil
}

// BroadcastICQuery broadcasts an ICQ query
// Also registeres the requested callback
// This method should not and DOEST NOT update the pool state
func (k Keeper) BroadcastICQuery(ctx sdk.Context, pool types.Pool, callbackID string, extraID string, queryType string, queryData []byte, timeoutPolicy icqueriestypes.TimeoutPolicy, timeoutTimestamp time.Duration) error {
	// Pool must be ready to process those kind of operations
	if pool.State == types.PoolState_Created || pool.State == types.PoolState_Unspecified {
		return types.ErrPoolNotReady
	}

	// Broadcast the query
	return k.ICQueriesKeeper.MakeRequest(
		ctx,
		types.ModuleName,
		callbackID,
		pool.GetChainId(),
		pool.GetConnectionId(),
		extraID,
		queryType,
		queryData,
		timeoutPolicy,
		timeoutTimestamp,
	)
}

// getPoolAppVersion returns the ICA app version for the pool connection
func (k Keeper) getPoolAppVersion(ctx sdk.Context, pool types.Pool) (string, error) {
	connectionEnd, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, pool.GetConnectionId())
	if !found {
		return "", fmt.Errorf("connection with id %s not found", pool.GetConnectionId())
	}
	return string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: pool.GetConnectionId(),
		HostConnectionId:       connectionEnd.Counterparty.GetConnectionID(),
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	})), nil
}
