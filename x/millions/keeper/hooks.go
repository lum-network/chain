package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/lum-network/chain/x/epochs/types"
	"github.com/lum-network/chain/x/millions/types"
)

// BeforeEpochStart is a hook triggered every defined epoch
// Currently triggers the undelegation of epochUnbonding withdrawals
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) (successCount, errorCount int) {
	logger := k.Logger(ctx).With("ctx", "epoch_unbonding")

	// Get the correct identifier
	if epochInfo.Identifier == epochstypes.DAY_EPOCH {
		epochTracker, err := k.UpdateEpochTracker(ctx, epochInfo, types.WithdrawalTrackerType)
		if err == nil {
			// List the unbondings from the previous epoch
			// The epoch unbondings are aggregated by poolIDEpochNumber
			epochUnbondings := k.GetEpochUnbondings(ctx, epochTracker.PreviousEpochNumber)
			for _, epochUnbonding := range epochUnbondings {
				if err := k.UndelegateWithdrawalsOnRemoteZone(ctx, epochUnbonding); err != nil {
					logger.Error(
						fmt.Sprintf("failed to launch undelegation for epoch unbonding: %v", err),
						"pool_id", epochUnbonding.PoolId,
						"epoch_number", epochUnbonding.EpochNumber,
					)
					errorCount++
				} else {
					successCount++
				}
				// Remove the epoch unbonding each time to start with a fresh one
				if err := k.RemoveEpochUnbonding(ctx, epochUnbonding); err != nil {
					logger.Error(
						fmt.Sprintf("failed to remove record for epoch unbonding: %v", err),
						"pool_id", epochUnbonding.PoolId,
						"epoch_number", epochUnbonding.EpochNumber,
					)
					errorCount++
				}
			}
		} else {
			logger.Error(
				fmt.Sprintf("Unable to update epoch tracker, err: %v", err),
				"epoch_number", epochInfo.CurrentEpoch,
			)
			errorCount++
		}

		if successCount+errorCount > 0 {
			logger.Info(
				"epoch unbonding transfers started",
				"nbr_success", successCount,
				"nbr_error", errorCount,
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
