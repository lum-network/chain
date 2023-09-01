package v153

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func MigratePoolType(ctx sdk.Context, k millionskeeper.Keeper) error {
	k.IteratePools(ctx, func(pool millionstypes.Pool) bool {
		// Patch our entity with the new poolType
		if _, err := k.UnsafeUpdatePoolType(ctx, pool.GetPoolId(), millionstypes.PoolType_Staking); err != nil {
			panic(err)
		}
		ctx.Logger().Info(fmt.Sprintf("Migrated pool %d pool type %s", pool.GetPoolId(), millionstypes.PoolType_Staking))
		return false
	})
	return nil
}
