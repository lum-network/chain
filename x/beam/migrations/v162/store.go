package v162

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	beamkeeper "github.com/lum-network/chain/x/beam/keeper"
	beamtypes "github.com/lum-network/chain/x/beam/types"
)

func DeleteBeamsData(ctx sdk.Context, bk beamkeeper.Keeper) error {
	bk.IterateOpenBeamsQueue(ctx, func(beam beamtypes.Beam) bool {
		bk.RemoveFromOpenBeamQueue(ctx, beam.GetId())
		return false
	})
	bk.IterateClosedBeamsQueue(ctx, func(beam beamtypes.Beam) bool {
		bk.RemoveFromClosedBeamQueue(ctx, beam.GetId())
		return false
	})
	bk.IterateOpenBeamsByBlockQueue(ctx, func(beam beamtypes.Beam) bool {
		bk.RemoveFromOpenBeamByBlockQueue(ctx, int(beam.GetClosesAtBlock()), beam.GetId())
		return false
	})
	bk.IterateBeams(ctx, func(beam beamtypes.Beam) bool {
		// Cancel the remaining beams if there is any
		if beam.GetStatus() == beamtypes.BeamState_StateOpen {
			if err := bk.UpdateBeamStatus(ctx, beam.GetId(), beamtypes.BeamState_StateCanceled); err != nil {
				panic(err)
			}
		}

		// Delete the beam entity itself
		if err := bk.DeleteBeam(ctx, beam.GetId()); err != nil {
			panic(err)
		}
		return false
	})
	return nil
}
