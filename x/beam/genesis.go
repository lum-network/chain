package beam

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/beam/keeper"
	"github.com/lum-network/chain/x/beam/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// InitGenesis initializes the capability module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) (res []abci.ValidatorUpdate) {
	k.CreateBeamModuleAccount(ctx, genState.ModuleAccountBalance)

	for _, beam := range genState.Beams {
		k.SetBeam(ctx, beam.GetId(), beam)
	}
	return nil
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	beams := k.ListBeams(ctx)

	return &types.GenesisState{
		Beams:                beams,
		ModuleAccountBalance: k.GetBeamAccountBalance(ctx),
	}
}
