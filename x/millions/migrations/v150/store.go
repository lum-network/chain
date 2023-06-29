package v150

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
	"strings"
)

func MigratePoolPortIdsToPortOwnerName(ctx sdk.Context, k millionskeeper.Keeper) error {
	k.IteratePools(ctx, func(pool millionstypes.Pool) bool {
		// Migrate the deposit port id only if required
		var updatedDepositPortId = pool.GetIcaDepositPortId()
		if strings.HasPrefix(pool.GetIcaDepositPortId(), icatypes.ControllerPortPrefix) {
			updatedDepositPortId = strings.TrimPrefix(pool.GetIcaDepositPortId(), icatypes.ControllerPortPrefix)
		}

		// Migrate the prizepool port id only if required
		var updatedPrizePoolPortId = pool.GetIcaPrizepoolPortId()
		if strings.HasPrefix(pool.GetIcaPrizepoolPortId(), icatypes.ControllerPortPrefix) {
			updatedPrizePoolPortId = strings.TrimPrefix(pool.GetIcaPrizepoolPortId(), icatypes.ControllerPortPrefix)
		}

		// Patch our entity
		if _, err := k.UnsafeUpdatePoolPortIds(ctx, pool.GetPoolId(), updatedDepositPortId, updatedPrizePoolPortId); err != nil {
			panic(err)
		}
		ctx.Logger().Info(fmt.Sprintf("Migrated pool %d port ids | old %s / %s | new %s / %s", pool.GetPoolId(), pool.GetIcaDepositPortId(), pool.GetIcaPrizepoolPortId(), updatedDepositPortId, updatedPrizePoolPortId))
		return false
	})
	return nil
}
