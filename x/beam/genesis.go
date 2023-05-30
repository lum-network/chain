package beam

import (
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/beam/keeper"
	"github.com/lum-network/chain/x/beam/types"
)

// InitGenesis initializes the capability module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) (res []abci.ValidatorUpdate) {
	k.CreateBeamModuleAccount(ctx, genState.ModuleAccountBalance)

	// Persist the beams to raw store
	for _, beam := range genState.Beams {
		k.SetBeam(ctx, beam.GetId(), beam)

		// Append to the correct queue from the beam state
		toQueue := false
		if beam.GetStatus() == types.BeamState_StateClosed || beam.GetStatus() == types.BeamState_StateCanceled {
			k.InsertClosedBeamQueue(ctx, beam.GetId())
			toQueue = true
		} else if beam.GetStatus() == types.BeamState_StateOpen {
			// Make sure we don't add a beam that is intended to be already closed at the current height
			if beam.GetClosesAtBlock() > 0 && int(beam.GetClosesAtBlock()) > int(ctx.BlockHeight()) {
				k.InsertOpenBeamByBlockQueue(ctx, int(beam.GetClosesAtBlock()), beam.GetId())
				toQueue = true
			}
		} else {
			ctx.Logger().Info(fmt.Sprintf("Not appending beam %s to any queue due to unhandled status", beam.GetId()), "height", ctx.BlockHeight(), "state", beam.GetStatus())
		}

		ctx.Logger().Info(fmt.Sprintf("Persisted beam %s from genesis file", beam.GetId()), "height", ctx.BlockHeight(), "state", beam.GetStatus(), "added_in_queue", toQueue)
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
