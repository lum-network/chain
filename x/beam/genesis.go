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

	// Persist the beams to raw store
	for _, beam := range genState.Beams {
		k.SetBeam(ctx, beam.GetId(), beam)
	}

	// Persist the closed queue
	for _, beam := range genState.BeamClosedQueue {
		k.InsertClosedBeamQueue(ctx, beam.GetId())
	}

	// Persist the old open queue
	for _, beam := range genState.BeamOpenOldQueue {
		k.InsertOpenBeamQueue(ctx, beam.GetId())
	}

	// Persist the open beam by block queue
	for _, beam := range genState.BeamOpenQueue {
		k.InsertOpenBeamByBlockQueue(ctx, int(beam.GetClosesAtBlock()), beam.GetId())
	}

	return nil
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	beams := k.ListBeams(ctx)

	beamsFromOldOpenQueue := k.ListBeamsFromOldOpenQueue(ctx)
	beamsFromClosedQueue := k.ListBeamsFromClosedQueue(ctx)
	beamsFromOpenQueue := k.ListBeamsFromOpenQueue(ctx)

	return &types.GenesisState{
		Beams:                beams,
		BeamOpenOldQueue:     beamsFromOldOpenQueue,
		BeamClosedQueue:      beamsFromClosedQueue,
		BeamOpenQueue:        beamsFromOpenQueue,
		ModuleAccountBalance: k.GetBeamAccountBalance(ctx),
	}
}
