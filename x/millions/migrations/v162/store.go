package v162

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
)

func MigratePendingWithdrawalsToNewEpochUnbonding(ctx sdk.Context, k millionskeeper.Keeper) error {
	ctx.Logger().Info("Processing pending withdrawals and add them to the new EpochUnbonding queue...")
	if err := k.UnsafeAddPendingWithdrawalsToNewEpochUnbonding(ctx); err != nil {
		return err
	}

	ctx.Logger().Info("Migration of pending withdrawals to new EpochUnbonding completed successfully")
	return nil
}
