package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

// BlockPoolUpdates runs all pool updates and draws
// Called in each BeginBlock.
func (k Keeper) BlockPoolUpdates(ctx sdk.Context) (successCount, errorCount int) {
	logger := k.Logger(ctx).With("ctx", "blocker_pools")

	// Launch all mature draws
	poolsToDraw := k.ListPoolsToDraw(ctx)
	for _, pool := range poolsToDraw {
		// Initiate a dedicated cache context to avoid creating draws if they fail at their very creation
		cacheCtx, writeCache := ctx.CacheContext()
		draw, err := k.LaunchNewDraw(cacheCtx, pool.GetPoolId())
		if err != nil {
			// Do not commit changes in case of failure
			logger.Error(
				fmt.Sprintf("failed to launch draw: %v", err),
				"pool_id", pool.PoolId,
				"draw_id", pool.GetNextDrawId(),
			)
			errorCount++
		} else {
			// Commit successful changes (draw and pool updates)
			logger.Info(
				"draw launched with success",
				"pool_id", pool.PoolId,
				"draw_id", pool.GetNextDrawId(),
			)
			writeCache()
			successCount++

			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					sdk.EventTypeMessage,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				),
				sdk.NewEvent(
					types.EventTypeNewDraw,
					sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(draw.PoolId, 10)),
					sdk.NewAttribute(types.AttributeKeyDrawID, strconv.FormatUint(draw.DrawId, 10)),
				),
			})
		}
	}
	if successCount+errorCount > 0 {
		logger.Info(
			"draws launched",
			"nbr_success", successCount,
			"nbr_error", errorCount,
		)
	}
	return
}

// BlockPrizeUpdates runs all prize updates (clawback)
// Called in each EndBlock.
func (k Keeper) BlockPrizeUpdates(ctx sdk.Context) (successCount, errorCount int) {
	logger := k.Logger(ctx).With("ctx", "blocker_prizes")

	// Expired unclaimed prizes
	prizesIDsToClawback := k.DequeueEPCBQueue(ctx, ctx.BlockTime())
	for _, pid := range prizesIDsToClawback {
		prize, err := k.GetPoolDrawPrize(ctx, pid.PoolId, pid.DrawId, pid.PrizeId)
		if err == nil {
			// Fetch master entity and ignore not found issues since we cannot recover it
			if err := k.ClawBackPrize(ctx, prize.PoolId, prize.DrawId, prize.PrizeId); err != nil {
				logger.Error(
					fmt.Sprintf("failed to clawback draw: %v", err),
					"pool_id", prize.PoolId,
					"draw_id", prize.DrawId,
					"prize_id", prize.PrizeId,
				)
				errorCount++
			} else {
				successCount++
				ctx.EventManager().EmitEvents(sdk.Events{
					sdk.NewEvent(
						sdk.EventTypeMessage,
						sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					),
					sdk.NewEvent(
						types.EventTypeClawbackPrize,
						sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(prize.PoolId, 10)),
						sdk.NewAttribute(types.AttributeKeyDrawID, strconv.FormatUint(prize.DrawId, 10)),
						sdk.NewAttribute(types.AttributeKeyPrizeID, strconv.FormatUint(prize.PrizeId, 10)),
						sdk.NewAttribute(types.AttributeKeyWinner, prize.WinnerAddress),
						sdk.NewAttribute(sdk.AttributeKeyAmount, prize.Amount.String()),
					),
				})
			}
		} else {
			logger.Error(
				fmt.Sprintf("failed to acquire prize: %v", err),
				"pool_id", pid.PoolId,
				"draw_id", pid.DrawId,
				"prize_id", pid.PrizeId,
			)
			errorCount++
		}
	}
	if successCount+errorCount > 0 {
		logger.Info(
			"prizes clawed back",
			"nbr_success", successCount,
			"nbr_error", errorCount,
		)
	}
	return
}

// BlockWithdrawalUpdates runs all matured withdrawals updates (transfer post unbonding)
// Called in each EndBlock.
func (k Keeper) BlockWithdrawalUpdates(ctx sdk.Context) (successCount, errorCount int) {
	logger := k.Logger(ctx).With("ctx", "blocker_withdrawals")

	// Matured withdrawals
	withdrawalsIDs := k.DequeueMaturedWithdrawalQueue(ctx, ctx.BlockTime())
	for _, wid := range withdrawalsIDs {
		withdrawal, err := k.GetPoolWithdrawal(ctx, wid.PoolId, wid.WithdrawalId)
		// Fetch master entity and ignore not found issues since we cannot recover it
		if err == nil && withdrawal.State == types.WithdrawalState_IcaUnbonding {
			// Initiate a dedicated cache context to avoid partial withdrawal commits
			cacheCtx, writeCache := ctx.CacheContext()
			k.UpdateWithdrawalStatus(cacheCtx, withdrawal.PoolId, withdrawal.WithdrawalId, types.WithdrawalState_IbcTransfer, nil, false)
			if err := k.TransferWithdrawalToLocalChain(cacheCtx, withdrawal.PoolId, withdrawal.WithdrawalId); err != nil {
				logger.Error(
					fmt.Sprintf("failed to launch transfer withdrawal amount to local zone: %v", err),
					"pool_id", withdrawal.PoolId,
					"withdrawal_id", withdrawal.WithdrawalId,
				)
				// Commit failing state using base context to make retries possible
				k.UpdateWithdrawalStatus(ctx, withdrawal.PoolId, withdrawal.WithdrawalId, types.WithdrawalState_IbcTransfer, nil, true)
				errorCount++
			} else {
				// Commit successful changes
				writeCache()
				successCount++
			}
		} else {
			logger.Error(
				fmt.Sprintf("failed to acquire withdrawal: %v", err),
				"pool_id", wid.PoolId,
				"withdrawal_id", wid.WithdrawalId,
			)
			errorCount++
		}
	}
	if successCount+errorCount > 0 {
		logger.Info(
			"withdrawals transfers started",
			"nbr_success", successCount,
			"nbr_error", errorCount,
		)
	}
	return
}
