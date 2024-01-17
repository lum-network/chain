package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v164 "github.com/lum-network/chain/x/millions/migrations/v164"

	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	v150 "github.com/lum-network/chain/x/millions/migrations/v150"
	v152 "github.com/lum-network/chain/x/millions/migrations/v152"
	v161 "github.com/lum-network/chain/x/millions/migrations/v161"
	v162 "github.com/lum-network/chain/x/millions/migrations/v162"
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

// Migrate2To3 migrates from version 2 to 3
func (m Migrator) Migrate2To3(ctx sdk.Context) error {
	return v152.MigrateFailedIcaUndelegationsToEpochUnbonding(ctx, m.keeper)
}

// Migrate3To4 migrates from version 3 to 4
func (m Migrator) Migrate3To4(ctx sdk.Context) error {
	return v161.MigratePoolTypeAndUnbondingFrequency(ctx, m.keeper)
}

// Migrate4To5 migrates from version 4 to 5
func (m Migrator) Migrate4To5(ctx sdk.Context) error {
	return v162.MigratePendingWithdrawalsToNewEpochUnbonding(ctx, m.keeper)
}

// Migrate5To6 migrates from version 5 to 6
func (m Migrator) Migrate5To6(ctx sdk.Context) error {
	if err := v164.SaveEntitiesOnAccountLevel(ctx, m.keeper); err != nil {
		return err
	}

	if err := v164.InitializePoolFeeTakers(ctx, m.keeper); err != nil {
		return err
	}

	return nil
}
