package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	v150 "github.com/lum-network/chain/x/millions/migrations/v150"
)

type Migrator struct {
	keeper millionskeeper.Keeper
}

func NewMigrator(keeper millionskeeper.Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1To2 migrates from version 1 to 2
func (m Migrator) Migrate1To2(ctx sdk.Context) error {
	return v150.MigratePoolPortIdsToPortOwnerName(ctx, m.keeper)
}
