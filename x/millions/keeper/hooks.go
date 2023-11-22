package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/lum-network/chain/x/epochs/types"
	"github.com/lum-network/chain/x/millions/types"
)

// BeforeEpochStart is a hook triggered every defined epoch
// Currently triggers the undelegation of epochUnbonding withdrawals
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	logger := k.Logger(ctx).With("ctx", "epoch_unbonding")

	// Proceed daily based epoch
	if epochInfo.Identifier == epochstypes.DAY_EPOCH {
		k.processEpochUnbondings(ctx, epochInfo, logger)
	}
}

func (k Keeper) processEpochUnbondings(ctx sdk.Context, epochInfo epochstypes.EpochInfo, logger log.Logger) (successCount, errorCount, skippedCount int) {
	// Create temporary context
	cacheCtx, writeCache := ctx.CacheContext()

	// Update the epoch tracker
	epochTracker, err := k.UpdateEpochTracker(ctx, epochInfo, types.WithdrawalTrackerType)
	if err != nil {
		logger.Error(
			fmt.Sprintf("Unable to update epoch tracker, err: %v", err),
			"epoch_number", epochInfo.CurrentEpoch,
		)
		return
	}

	// Get epoch unbondings
	epochUnbondings := k.GetEpochUnbondings(cacheCtx, epochTracker.EpochNumber)

	// For each unbonding, try to proceed the undelegation otherwise, rollback the entire operation
	for _, epochUnbonding := range epochUnbondings {
		success, err := k.processEpochUnbonding(cacheCtx, epochUnbonding, epochTracker, logger)
		if err != nil {
			logger.Error(
				fmt.Sprintf("Error processing epoch unbonding: %v", err),
				"pool_id", epochUnbonding.GetPoolId(),
				"epoch_number", epochUnbonding.GetEpochNumber(),
			)
			errorCount++
			break
		} else if success {
			successCount++
		} else {
			skippedCount++
		}
	}

	// If there was an error, we are supposed to cancel the entire operation
	if errorCount > 0 {
		// Log out the critical error
		logger.Error(
			"epoch unbonding undelegate processed with errors, cache not written",
			"nbr_success", successCount,
			"nbr_error", errorCount,
			"nbr_skipped", skippedCount,
		)

		// Allocate new cache context
		rollbackTmpCacheCtx, writeRollbackCache := ctx.CacheContext()

		// Get epoch unbondings
		epochUnbondings = k.GetEpochUnbondings(rollbackTmpCacheCtx, epochTracker.EpochNumber)

		// Rollback the operations and put everything back up in the next epoch
		// We explicitly don't use the cache context here, as we want to proceed
		for _, epochUnbonding := range epochUnbondings {
			for _, wid := range epochUnbonding.WithdrawalIds {
				withdrawal, err := k.GetPoolWithdrawal(rollbackTmpCacheCtx, epochUnbonding.PoolId, wid)
				if err != nil {
					logger.Error(fmt.Sprintf("Failure in getting the pool withdrawals for pool %d, err: %v", epochUnbonding.PoolId, err))
					return 0, 1, 0
				}
				if err := k.AddEpochUnbonding(rollbackTmpCacheCtx, withdrawal, false); err != nil {
					logger.Error(fmt.Sprintf("Failure in adding the withdrawal to epoch unbonding for pool %d, err: %v", epochUnbonding.PoolId, err))
					return 0, 1, 0
				}
			}

			if err := k.RemoveEpochUnbonding(rollbackTmpCacheCtx, epochUnbonding); err != nil {
				logger.Error(fmt.Sprintf("Failure in removing the epoch unbonding for pool %d, err: %v", epochUnbonding.PoolId, err))
				return 0, 1, 0
			}
		}

		// Write the cache
		writeRollbackCache()
	} else {
		// Otherwise we can just commit the cache, and return
		writeCache()
		logger.Info(
			"epoch unbonding undelegate processed",
			"nbr_success", successCount,
			"nbr_error", errorCount,
			"nbr_skipped", skippedCount,
		)
	}
	return successCount, errorCount, skippedCount
}

func (k Keeper) processEpochUnbonding(ctx sdk.Context, epochUnbonding types.EpochUnbonding, epochTracker types.EpochTracker, logger log.Logger) (bool, error) {
	pool, err := k.GetPool(ctx, epochUnbonding.GetPoolId())
	if err != nil {
		return false, err
	}

	// Epoch unbonding is supposed to happen every X where X is (unbonding_frequency/7)+1
	// Modulo operation allows to ensure that this one can run
	if pool.IsInvalidEpochUnbonding(epochTracker) {
		logger.Info(
			"Unbonding isn't supposed to trigger at this epoch",
			"pool_id", epochUnbonding.PoolId,
			"epoch_number", epochUnbonding.GetEpochNumber(),
		)

		// Fail safe - each withdrawal is added back to the next executable epoch unbonding
		// Then we just remove the actual epoch entity
		for _, wid := range epochUnbonding.WithdrawalIds {
			withdrawal, err := k.GetPoolWithdrawal(ctx, epochUnbonding.PoolId, wid)
			if err != nil {
				return false, err
			}
			if err := k.AddEpochUnbonding(ctx, withdrawal, false); err != nil {
				return false, err
			}
		}

		if err := k.RemoveEpochUnbonding(ctx, epochUnbonding); err != nil {
			return false, err
		}

		return false, nil
	}

	// In case everything worked fine, we just proceed as expected
	// Undelegate the batch withdrawals to remote zone, then remove the actual epoch entity
	if err := k.UndelegateWithdrawalsOnRemoteZone(ctx, epochUnbonding); err != nil {
		return false, err
	}
	if err := k.RemoveEpochUnbonding(ctx, epochUnbonding); err != nil {
		return false, err
	}

	return true, nil
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.AfterEpochEnd(ctx, epochInfo)
}
