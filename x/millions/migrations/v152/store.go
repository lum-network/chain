package v152

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
)

func MigrateFailedIcaUndelegationsToEpochUnbonding(ctx sdk.Context, k millionskeeper.Keeper) error {
	// Batch failed ica undelegations
	ctx.Logger().Info("Processing failed ICA undelegations and add them to pending queue...")
	if err := k.AddFailedIcaUndelegationsToEpochUnbonding(ctx); err != nil {
		return err
	}

	// Patch broken deposit entities to allow for retry
	// We acquire both entities that are broken. If there is any error, it shouldn't be returned to allow for testnet execution
	// We require to ensure it here, cuz on a sub level, it would panic out in case entity does not exist
	// Both entities are flagged on error state IcaDelegate to allow them to be retried out by their owner
	poolId := uint64(2)
	ctx.Logger().Info("Patching broken deposit entities...")
	_, err := k.GetPoolDeposit(ctx, poolId, 429)
	if err == nil {
		k.UpdateDepositStatus(ctx, poolId, 429, types.DepositState_IcaDelegate, true)
		ctx.Logger().Info("Deposit 429 patched")
	}
	_, err = k.GetPoolDeposit(ctx, poolId, 450)
	if err == nil {
		k.UpdateDepositStatus(ctx, poolId, 450, types.DepositState_IcaDelegate, true)
		ctx.Logger().Info("Deposit 450 patched")
	}

	return nil
}
