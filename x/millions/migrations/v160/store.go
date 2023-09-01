package v160

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func MigratePoolTypeAndUnbondingFrequency(ctx sdk.Context, k millionskeeper.Keeper) error {
	ctx.Logger().Info("Processing unbonding frequency migration on pools")
	k.IteratePools(ctx, func(pool millionstypes.Pool) bool {
		// First update pool type as the RegisterPoolRunners relies on it
		if _, err := k.UnsafeUpdatePoolType(ctx, pool.GetPoolId(), millionstypes.PoolType_Staking); err != nil {
			panic(err)
		}

		// Unbonding frequency here is Cosmos Hub (unbonding time/7)+1
		if _, err := k.UnsafeUpdatePoolUnbondingFrequency(ctx, pool.GetPoolId(), millionstypes.DefaultUnbondingDuration, sdk.NewInt(millionstypes.DefaultMaxUnbondingEntries)); err != nil {
			panic(err)
		}
		return false
	})
	ctx.Logger().Info("Pool type and unbonding frequency migration completed successfully")
	return nil
}
