package v161

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
)

func MigrateAutoCompoundDeposits(ctx sdk.Context, k millionskeeper.Keeper) error {
	ctx.Logger().Info("Processing autocompound deposits")

	deposits := k.ListDeposits(ctx)
	for _, d := range deposits {
		if _, err := k.UnsafeUpdateAutoCompoundDeposits(ctx, d); err != nil {
			panic(err)
		}
	}

	ctx.Logger().Info("Successfully updated deposits with autocompound state")
	return nil
}
