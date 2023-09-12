package v160

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func MigratePoolTypeAndUnbondingFrequency(ctx sdk.Context, k millionskeeper.Keeper) error {
	ctx.Logger().Info("Processing unbonding frequency migration on pools")
	k.IteratePools(ctx, func(pool millionstypes.Pool) bool {
		// Unbonding frequency here is Cosmos Hub (unbonding time/7)+1
		// Pool type is staking
		// We have to do this in the same operation to avoid chicken-egg problem when it comes to ValidateBasic
		if _, err := k.UnsafeUpdatePoolUnbondingFrequencyAndType(ctx, pool.GetPoolId(), millionstypes.DefaultUnbondingDuration, sdk.NewInt(millionstypes.DefaultMaxUnbondingEntries), millionstypes.PoolType_Staking); err != nil {
			panic(err)
		}
		return false
	})
	ctx.Logger().Info("Pool type and unbonding frequency migration completed successfully")
	return nil
}
