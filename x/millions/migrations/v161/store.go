package v161

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func MigrateAutoCompoundDeposits(ctx sdk.Context, k millionskeeper.Keeper) error {
	ctx.Logger().Info("Processing autocompound deposits")

	deposits := k.ListDeposits(ctx)
	updatedEntityCount := 0
	for _, d := range deposits {
		if _, err := k.UnsafeUpdateAutoCompoundDeposits(ctx, d.PoolId, d.DepositId, millionstypes.DepositOrigin_Direct); err != nil {
			panic(err)
		}
		updatedEntityCount++
	}

	ctx.Logger().Info(fmt.Sprintf("Successfully updated %d deposits with %s state", updatedEntityCount, millionstypes.DepositOrigin_Direct.String()))
	return nil
}
