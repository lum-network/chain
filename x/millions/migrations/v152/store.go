package v152

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
)

func MigrateFailedIcaUndelegationsToEpochUnbonding(ctx sdk.Context, k millionskeeper.Keeper) error {
	// Batch failed ica undelegations
	if err := k.AddFailedIcaUndelegationsToEpochUnbonding(ctx); err != nil {
		return err
	}

	ctx.Logger().Info("Batch Failed ICA Undelegations")

	return nil
}
