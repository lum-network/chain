package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v162 "github.com/lum-network/chain/x/beam/migrations/v162"

	keeper2 "github.com/lum-network/chain/x/beam/keeper"
	v110 "github.com/lum-network/chain/x/beam/migrations/v110"
)

type Migrator struct {
	keeper keeper2.Keeper
}

func NewMigrator(keeper keeper2.Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1To2 migrates from version 1 to 2
func (m Migrator) Migrate1To2(ctx sdk.Context) error {
	return v110.MigrateBeamQueues(ctx, m.keeper)
}

func (m Migrator) Migrate2To3(ctx sdk.Context) error {
	return v162.DeleteBeamsData(ctx, m.keeper)
}
