package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/lum-network/chain/x/epochs/types"
	"github.com/lum-network/chain/x/millions/types"
)

// BeforeEpochStart is a hook triggered every defined epoch
// Currently triggers the undelegation of epochUnbonding withdrawals
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) (successCount, errorCount, skippedCount int) {
	logger := k.Logger(ctx).With("ctx", "epoch_unbonding")

	// Get the correct identifier
	if epochInfo.Identifier == epochstypes.DAY_EPOCH {
		// Update the epoch with the actual information
		epochTracker, err := k.UpdateEpochTracker(ctx, epochInfo, types.WithdrawalTrackerType)
		if err != nil {
			logger.Error(
				fmt.Sprintf("Unable to update epoch tracker, err: %v", err),
				"epoch_number", epochInfo.CurrentEpoch,
			)
			skippedCount++
			return
		}

		// List the unbondings from the previous epoch
		epochUnbondings := k.GetEpochUnbondings(ctx, epochTracker.PreviousEpochNumber)
		for _, epochUnbonding := range epochUnbondings {
			// Confirm the unbonding is supposed to get triggered at this epoch time
			pool, _ := k.GetPool(ctx, epochUnbonding.GetPoolId())
			if epochTracker.EpochNumber%pool.UnbondingFrequency.Uint64() != 0 {
				logger.Info(
					fmt.Sprintf("Unbonding isn't supposed to trigger at this epoch"),
					"pool_id", epochUnbonding.PoolId,
					"epoch_number", epochUnbonding.GetEpochNumber(),
				)
				errorCount++
				continue
			}

			if err := k.UndelegateWithdrawalsOnRemoteZone(ctx, epochUnbonding); err != nil {
				logger.Error(
					fmt.Sprintf("failed to launch undelegation for epoch unbonding: %v", err),
					"pool_id", epochUnbonding.GetPoolId(),
					"epoch_number", epochUnbonding.GetEpochNumber(),
				)
				errorCount++
			} else {
				successCount++
			}

			// Remove the epoch unbonding record to start the next one with a fresh one
			if err := k.RemoveEpochUnbonding(ctx, epochUnbonding); err != nil {
				logger.Error(
					fmt.Sprintf("failed to remove record for epoch unbonding: %v", err),
					"pool_id", epochUnbonding.GetPoolId(),
					"epoch_number", epochUnbonding.GetEpochNumber(),
				)
				errorCount++
			}
		}

		if successCount+errorCount > 0 {
			logger.Info(
				"epoch unbonding undelegate started",
				"nbr_success", successCount,
				"nbr_error", errorCount,
				"nbr_skipped", skippedCount,
			)
		}
	}

	return
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
