package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v104 "github.com/lum-network/chain/x/airdrop/migrations/v104"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate2to3 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	err := v104.MigrateModuleBalance(ctx, m.keeper.AuthKeeper, m.keeper.BankKeeper, m.keeper.GetClaimRecords(ctx))
	return err
}
