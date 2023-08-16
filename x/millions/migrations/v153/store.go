package v153

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func MigratePoolDefaultUnbondingFrequency(ctx sdk.Context, k millionskeeper.Keeper) error {
	ctx.Logger().Info("Processing unbonding frequency migration on pools")
	k.IteratePools(ctx, func(pool millionstypes.Pool) bool {
		// Unbonding frequency here is Cosmos Hub (unbonding time/7)+1
		if _, err := k.UnsafeUpdatePoolUnbondingFrequency(ctx, pool.GetPoolId(), 4); err != nil {
			panic(err)
		}
		return false
	})
	ctx.Logger().Info("Unbonding frequencies patched")
	return nil
}
