package beam

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/beam/keeper"
	"github.com/lum-network/chain/x/beam/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) (res []abci.ValidatorUpdate) {
	// Acquire the module account, if it does not exist yet, it is created
	moduleAcc := k.GetBeamAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set as it should have been", types.ModuleName))
	}

	// If account has zero, it probably means it does not exist yet
	balance := k.BankKeeper.GetAllBalances(ctx, moduleAcc.GetAddress())
	if balance.IsZero() {
		k.AuthKeeper.SetModuleAccount(ctx, moduleAcc)
	}

	return nil
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return types.DefaultGenesis()
}
