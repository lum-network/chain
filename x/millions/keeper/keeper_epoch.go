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

	// If it's the first epoch unbonding, create it with the first withdrawal
	if !k.hasEpochPoolUnbonding(ctx, epochTracker.EpochNumber, withdrawal.PoolId) {
		epochUnbonding := types.EpochUnbonding{
			EpochNumber:        epochTracker.EpochNumber,
			EpochIdentifier:    epochstypes.DAY_EPOCH,
			PoolId:             withdrawal.PoolId,
			WithdrawalIds:      []uint64{withdrawal.WithdrawalId},
			WithdrawalIdsCount: uint64(1),
			TotalAmount:        withdrawal.Amount,
		}
		k.setEpochPoolUnbonding(ctx, epochUnbonding)

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
				sdk.NewAttribute(sdk.AttributeKeyAmount, withdrawal.Amount.String()),
			),
		})

		return nil
	}

	// Get epoch unbonding entity for new withdrawals within the created epoch and poolID
	epochPoolUnbonding, err := k.GetEpochPoolUnbonding(ctx, epochTracker.EpochNumber, withdrawal.PoolId)
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

	nextEpoch := epochTracker.EpochNumber + 1

	for {
		epochPoolUnbonding, err := k.GetEpochPoolUnbonding(ctx, nextEpoch, withdrawal.PoolId)
		if err != nil {
			if errors.Is(err, types.ErrInvalidEpochUnbonding) {
				// Create a new epoch unbonding since it doesn't exist for the next epoch
				epochPoolUnbonding = types.EpochUnbonding{
					EpochNumber:        nextEpoch,
					EpochIdentifier:    epochstypes.DAY_EPOCH,
					PoolId:             withdrawal.PoolId,
					WithdrawalIds:      []uint64{withdrawal.WithdrawalId},
					WithdrawalIdsCount: uint64(1),
					TotalAmount:        withdrawal.Amount,
					CreatedAtHeight:    ctx.BlockHeight(),
					CreatedAt:          ctx.BlockTime(),
				}
				k.setEpochPoolUnbonding(ctx, epochPoolUnbonding)
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

		nextEpoch++
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

	k.setEpochPoolUnbonding(ctx, epochPoolUnbonding)

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
	withdrawDepositStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(withdrawDepositStore, types.GetWithdrawalsKey())
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

// setEpochPoolUnbonding sets an epoch unbonding by its epochNumber and poolID
func (k Keeper) setEpochPoolUnbonding(ctx sdk.Context, epochUnbonding types.EpochUnbonding) {
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
	k.setEpochTracker(ctx, eTracker, types.WithdrawalTrackerType)

	return eTracker, nil
}

// SetEpochTracker sets the internal epoch tracker
func (k Keeper) setEpochTracker(ctx sdk.Context, epochTracker types.EpochTracker, trackerType string) {
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
	kvStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(kvStore, types.GetEpochPoolUnbondingKey(epochID, poolID))

	defer iterator.Close()
	return iterator.Valid()
}

// GetEpochUnbondings lists the epoch unbondings by epochID
func (k Keeper) GetEpochUnbondings(ctx sdk.Context, epochID uint64) (epochUnbondings []types.EpochUnbonding) {
	epochStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(epochStore, types.GetEpochUnbondingsKey(epochID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var epochUnbonding types.EpochUnbonding
		k.cdc.MustUnmarshal(iterator.Value(), &epochUnbonding)
		epochUnbondings = append(epochUnbondings, epochUnbonding)
	}
	return
}
