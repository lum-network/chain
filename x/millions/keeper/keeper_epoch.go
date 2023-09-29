package keeper

import (
	"errors"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	epochstypes "github.com/lum-network/chain/x/epochs/types"
	"github.com/lum-network/chain/x/millions/types"
)

// AddEpochUnbonding adds an epoch unbonding record that serves to undelegate withdrawals via batches
func (k Keeper) AddEpochUnbonding(ctx sdk.Context, withdrawal types.Withdrawal, isUndelegateRetry bool) error {
	var epochInfo epochstypes.EpochInfo

	pool, err := k.GetPool(ctx, withdrawal.PoolId)
	if err != nil {
		return err
	}

	withdrawal, err = k.GetPoolWithdrawal(ctx, pool.PoolId, withdrawal.WithdrawalId)
	if err != nil {
		return err
	}

	if isUndelegateRetry {
		// Check retry state
		if withdrawal.ErrorState != types.WithdrawalState_IcaUndelegate {
			return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.WithdrawalState_IcaUndelegate.String(), withdrawal.ErrorState.String())
		}
		// Update withdrawal with state pending for the failed ones to retry
		k.UpdateWithdrawalStatus(ctx, withdrawal.PoolId, withdrawal.WithdrawalId, types.WithdrawalState_Pending, nil, false)
	} else {
		// check state of new withdrawals
		if withdrawal.State != types.WithdrawalState_Pending {
			return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.WithdrawalState_Pending.String(), withdrawal.State.String())
		}
	}

	// Get the millions internal module tracker
	epochTracker, err := k.GetEpochTracker(ctx, epochstypes.DAY_EPOCH, types.WithdrawalTrackerType)
	if err != nil {
		if errors.Is(err, types.ErrInvalidEpochTracker) {
			// Force update the epoch tracker if it doesn't exist
			epochInfo.Identifier = epochstypes.DAY_EPOCH
			epochInfo.StartTime = epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
			_, err = k.UpdateEpochTracker(ctx, epochInfo, types.WithdrawalTrackerType)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Get the next epoch that will be executed
	nextEpochUnbonding := pool.GetNextEpochUnbonding(epochTracker)

	// If it's the first epoch unbonding, create it with the first withdrawal
	if !k.hasEpochPoolUnbonding(ctx, nextEpochUnbonding, withdrawal.PoolId) {
		epochPoolUnbonding := types.EpochUnbonding{
			EpochNumber:        nextEpochUnbonding,
			EpochIdentifier:    epochstypes.DAY_EPOCH,
			PoolId:             withdrawal.PoolId,
			WithdrawalIds:      []uint64{withdrawal.WithdrawalId},
			WithdrawalIdsCount: uint64(1),
			TotalAmount:        withdrawal.Amount,
		}
		k.SetEpochPoolUnbonding(ctx, epochPoolUnbonding)

		// Emit event
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			),
			sdk.NewEvent(
				types.EventTypeEpochUnbonding,
				sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(withdrawal.PoolId, 10)),
				sdk.NewAttribute(types.AttributeKeyEpochID, strconv.FormatUint(epochTracker.EpochNumber, 10)),
				sdk.NewAttribute(sdk.AttributeKeyAmount, epochPoolUnbonding.TotalAmount.String()),
			),
		})

		return nil
	}

	// Get epoch unbonding entity for new withdrawals within the unbonding epoch and poolID
	epochPoolUnbonding, err := k.GetEpochPoolUnbonding(ctx, nextEpochUnbonding, withdrawal.PoolId)
	if err != nil {
		return err
	}

	// Prevent adding the same withdrawalID several times per epoch
	if epochPoolUnbonding.WithdrawalIDExists(withdrawal.WithdrawalId) {
		return errorsmod.Wrapf(types.ErrEntityOverride, "ID %d", withdrawal.GetWithdrawalId())
	}

	// Check the Hard cap limit on withdrawals per epoch unbonding
	if epochPoolUnbonding.WithdrawalIDsLimitReached(epochPoolUnbonding.WithdrawalIdsCount) {
		// Add to next available epoch if reached
		if err := k.AddWithdrawalToNextAvailableEpoch(ctx, withdrawal); err != nil {
			return err
		}
		return nil
	}

	// Update current epoch unbondings
	if err := k.UpdateEpochUnbonding(ctx, epochPoolUnbonding, withdrawal); err != nil {
		return err
	}

	return nil
}

// AddWithdrawalToNextAvailableEpoch adds a withdrawal to the next available epoch unbonding
func (k Keeper) AddWithdrawalToNextAvailableEpoch(ctx sdk.Context, withdrawal types.Withdrawal) error {
	pool, err := k.GetPool(ctx, withdrawal.PoolId)
	if err != nil {
		return err
	}

	withdrawal, err = k.GetPoolWithdrawal(ctx, pool.PoolId, withdrawal.WithdrawalId)
	if err != nil {
		return err
	}

	epochTracker, err := k.GetEpochTracker(ctx, epochstypes.DAY_EPOCH, types.WithdrawalTrackerType)
	if err != nil {
		return err
	}

	// Get the next epoch unbonding considering the unbonding frequency
	nextEpochUnbonding := pool.GetNextEpochUnbonding(epochTracker) + pool.GetUnbondingFrequency().Uint64()

	for {
		epochPoolUnbonding, err := k.GetEpochPoolUnbonding(ctx, nextEpochUnbonding, withdrawal.PoolId)
		if err != nil {
			if errors.Is(err, types.ErrInvalidEpochUnbonding) {
				// Create a new epoch unbonding since it doesn't exist for the next epoch
				epochPoolUnbonding = types.EpochUnbonding{
					EpochNumber:        nextEpochUnbonding,
					EpochIdentifier:    epochstypes.DAY_EPOCH,
					PoolId:             withdrawal.PoolId,
					WithdrawalIds:      []uint64{withdrawal.WithdrawalId},
					WithdrawalIdsCount: uint64(1),
					TotalAmount:        withdrawal.Amount,
					CreatedAtHeight:    ctx.BlockHeight(),
					CreatedAt:          ctx.BlockTime(),
				}
				k.SetEpochPoolUnbonding(ctx, epochPoolUnbonding)

				// Emit event
				ctx.EventManager().EmitEvents(sdk.Events{
					sdk.NewEvent(
						sdk.EventTypeMessage,
						sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					),
					sdk.NewEvent(
						types.EventTypeEpochUnbonding,
						sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(withdrawal.PoolId, 10)),
						sdk.NewAttribute(types.AttributeKeyEpochID, strconv.FormatUint(epochTracker.EpochNumber, 10)),
						sdk.NewAttribute(sdk.AttributeKeyAmount, epochPoolUnbonding.TotalAmount.String()),
					),
				})

				return nil
			}
			return err
		}

		if !epochPoolUnbonding.WithdrawalIDsLimitReached(epochPoolUnbonding.WithdrawalIdsCount) {
			// Add the withdrawal to the found epoch unbonding
			if err := k.UpdateEpochUnbonding(ctx, epochPoolUnbonding, withdrawal); err != nil {
				return err
			}
			return nil
		}

		nextEpochUnbonding += pool.GetUnbondingFrequency().Uint64()
	}
}

// RemoveEpochUnbonding removes an epoch unbonding after each epoch
func (k Keeper) RemoveEpochUnbonding(ctx sdk.Context, epochUnbonding types.EpochUnbonding) error {
	// Ensure payload is valid
	params := types.DefaultParams()
	if err := epochUnbonding.ValidateBasic(params); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetEpochPoolUnbondingKey(epochUnbonding.EpochNumber, epochUnbonding.PoolId))

	return nil
}

// UpdateEpochUnbonding allows essentially to sum-up the withdrawals as well as metadata
func (k Keeper) UpdateEpochUnbonding(ctx sdk.Context, epochPoolUnbonding types.EpochUnbonding, withdrawal types.Withdrawal) error {
	pool, err := k.GetPool(ctx, withdrawal.PoolId)
	if err != nil {
		return err
	}

	withdrawal, err = k.GetPoolWithdrawal(ctx, pool.PoolId, withdrawal.WithdrawalId)
	if err != nil {
		return err
	}

	epochPoolUnbonding, err = k.GetEpochPoolUnbonding(ctx, epochPoolUnbonding.EpochNumber, withdrawal.PoolId)
	if err != nil {
		return err
	}

	epochPoolUnbonding.PoolId = withdrawal.PoolId
	epochPoolUnbonding.TotalAmount = epochPoolUnbonding.TotalAmount.Add(withdrawal.Amount)
	epochPoolUnbonding.WithdrawalIds = append(epochPoolUnbonding.WithdrawalIds, withdrawal.WithdrawalId)
	epochPoolUnbonding.WithdrawalIdsCount++
	epochPoolUnbonding.UpdatedAtHeight = ctx.BlockHeight()
	epochPoolUnbonding.UpdatedAt = ctx.BlockTime()

	k.SetEpochPoolUnbonding(ctx, epochPoolUnbonding)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeEpochUnbonding,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(withdrawal.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyEpochID, strconv.FormatUint(epochPoolUnbonding.EpochNumber, 10)),
			sdk.NewAttribute(sdk.AttributeKeyAmount, epochPoolUnbonding.TotalAmount.String()),
		),
	})

	return nil
}

// AddFailedIcaUndelegationsToEpochUnbonding add failed ica withdrawals to current epoch unbonding record
func (k Keeper) AddFailedIcaUndelegationsToEpochUnbonding(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetWithdrawalsKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var withdrawal types.Withdrawal
		k.cdc.MustUnmarshal(iterator.Value(), &withdrawal)

		if withdrawal.ErrorState == types.WithdrawalState_IcaUndelegate {
			err := k.AddEpochUnbonding(ctx, withdrawal, true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UnsafeAddPendingWithdrawalsToEpochUnbonding raw updates pending withdrawal's epochUnbonding
// Unsafe and should only be used for store migration
func (k Keeper) UnsafeAddPendingWithdrawalsToNewEpochUnbonding(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetWithdrawalsKey())
	defer iterator.Close()

	epochTracker, err := k.GetEpochTracker(ctx, epochstypes.DAY_EPOCH, types.WithdrawalTrackerType)
	if err != nil {
		return err
	}

	for ; iterator.Valid(); iterator.Next() {
		var withdrawal types.Withdrawal
		k.cdc.MustUnmarshal(iterator.Value(), &withdrawal)

		if withdrawal.State == types.WithdrawalState_Pending {
			// Iterate through past epochs to find EpochUnbondings
			for i := 1; i <= int(epochTracker.EpochNumber); i++ {
				epochUnbonding, err := k.GetEpochPoolUnbonding(ctx, uint64(i), withdrawal.PoolId)
				if err != nil {
					// If no EpochUnbonding object is found for this epoch, skip it
					continue
				}

				// If there are withdrawals contained in past epochUnbondings remove them
				if len(epochUnbonding.WithdrawalIds) > 0 {
					if err := k.RemoveEpochUnbonding(ctx, epochUnbonding); err != nil {
						return err
					}
				}
			}

			// Add the withdrawal to the next appropriate epochUnbonding
			err := k.AddEpochUnbonding(ctx, withdrawal, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// SetEpochPoolUnbonding sets an epoch unbonding by its epochNumber and poolID
func (k Keeper) SetEpochPoolUnbonding(ctx sdk.Context, epochUnbonding types.EpochUnbonding) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEpochPoolUnbondingKey(epochUnbonding.EpochNumber, epochUnbonding.PoolId)
	encodedEpochUnbonding := k.cdc.MustMarshal(&epochUnbonding)
	store.Set(key, encodedEpochUnbonding)
}

// GetEpochPoolUnbonding gets an epoch unbonding by its epochNumber and poolID
func (k Keeper) GetEpochPoolUnbonding(ctx sdk.Context, epochID uint64, poolID uint64) (types.EpochUnbonding, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetEpochPoolUnbondingKey(epochID, poolID))
	if bz == nil {
		return types.EpochUnbonding{}, types.ErrInvalidEpochUnbonding
	}

	var eUnbonding types.EpochUnbonding
	if err := k.cdc.Unmarshal(bz, &eUnbonding); err != nil {
		return types.EpochUnbonding{}, err
	}

	return eUnbonding, nil
}

// UpdateEpochTracker updates the internal epoch tracker
func (k Keeper) UpdateEpochTracker(ctx sdk.Context, epochInfo epochstypes.EpochInfo, trackerType string) (eTracker types.EpochTracker, err error) {
	epochNumber, err := cast.ToUint64E(epochInfo.CurrentEpoch)
	if err != nil {
		return eTracker, err
	}

	eTracker = types.EpochTracker{
		EpochTrackerType:    trackerType,
		EpochIdentifier:     epochInfo.Identifier,
		EpochNumber:         epochNumber,
		PreviousEpochNumber: epochNumber - 1,
		NextEpochNumber:     epochNumber + 1,
		NextEpochStartTime:  epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration),
	}
	k.SetEpochTracker(ctx, eTracker, types.WithdrawalTrackerType)

	return eTracker, nil
}

// SetEpochTracker sets the internal epoch tracker
func (k Keeper) SetEpochTracker(ctx sdk.Context, epochTracker types.EpochTracker, trackerType string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEpochTrackerKey(epochTracker.EpochIdentifier, trackerType)
	encodedEpochTracker := k.cdc.MustMarshal(&epochTracker)
	store.Set(key, encodedEpochTracker)
}

// GetEpochTracker gets the internal epoch tracker
func (k Keeper) GetEpochTracker(ctx sdk.Context, epochIdentifier string, trackerType string) (types.EpochTracker, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetEpochTrackerKey(epochIdentifier, trackerType))
	if bz == nil {
		return types.EpochTracker{}, types.ErrInvalidEpochTracker
	}

	var eTracker types.EpochTracker
	if err := k.cdc.Unmarshal(bz, &eTracker); err != nil {
		return types.EpochTracker{}, err
	}

	return eTracker, nil
}

// hasEpochPoolUnbonding allows to check if an epoch unbonding needs to be freshly created or updated
func (k Keeper) hasEpochPoolUnbonding(ctx sdk.Context, epochID uint64, poolID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetEpochPoolUnbondingKey(epochID, poolID))

	defer iterator.Close()
	return iterator.Valid()
}

// GetEpochUnbondings lists the epoch unbondings by epochID
func (k Keeper) GetEpochUnbondings(ctx sdk.Context, epochID uint64) (epochUnbondings []types.EpochUnbonding) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetEpochUnbondingsKey(epochID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var epochUnbonding types.EpochUnbonding
		k.cdc.MustUnmarshal(iterator.Value(), &epochUnbonding)
		epochUnbondings = append(epochUnbondings, epochUnbonding)
	}
	return
}

func (k Keeper) ListEpochTrackers(ctx sdk.Context) (list []types.EpochTracker) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.EpochTrackerPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.EpochTracker
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
