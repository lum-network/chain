package v164

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
)

func SaveEntitiesOnAccountLevel(ctx sdk.Context, k millionskeeper.Keeper) error {
	ctx.Logger().Info("Saving entities on account level")

	updatedDeposits := k.UnsafeSetUnpersistedDeposits(ctx)
	updatedWithdrawals := k.UnsafeSetUnpersistedWithdrawals(ctx)

	ctx.Logger().Info("Saving entities on account level: deposits", "count", updatedDeposits)
	ctx.Logger().Info("Saving entities on account level: withdrawals", "count", updatedWithdrawals)
	return nil
}
