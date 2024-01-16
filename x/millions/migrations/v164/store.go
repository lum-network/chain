package v164

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func InitializePoolFeeTakers(ctx sdk.Context, k millionskeeper.Keeper) error {
	ctx.Logger().Info("Initializing pool fee takers...")

	defaultFeeTakers := []millionstypes.FeeTaker{
		{Destination: authtypes.FeeCollectorName, Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
	}

	k.IteratePools(ctx, func(pool millionstypes.Pool) bool {
		if _, err := k.UnsafeUpdatePoolFeeTakers(ctx, pool.GetPoolId(), defaultFeeTakers); err != nil {
			panic(err)
		}
		return false
	})

	ctx.Logger().Info("Migration of pool fee takers completed successfully")
	return nil
}
