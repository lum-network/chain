package v162

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	beamkeeper "github.com/lum-network/chain/x/beam/keeper"
	beamtypes "github.com/lum-network/chain/x/beam/types"
)

func DeleteBeamsData(ctx sdk.Context, bk beamkeeper.Keeper) error {
	bk.IterateOpenBeamsQueue(ctx, func(beam beamtypes.Beam) bool {
		bk.RemoveFromOpenBeamQueue(ctx, beam.GetId())
		bk.Logger(ctx).Info(fmt.Sprintf("Removed beam %s from open beam queue", beam.GetId()))
		return false
	})
	bk.IterateClosedBeamsQueue(ctx, func(beam beamtypes.Beam) bool {
		bk.RemoveFromClosedBeamQueue(ctx, beam.GetId())
		bk.Logger(ctx).Info(fmt.Sprintf("Removed beam %s from open beam queue", beam.GetId()))
		return false
	})
	bk.IterateOpenBeamsByBlockQueue(ctx, func(beam beamtypes.Beam) bool {
		bk.RemoveFromOpenBeamByBlockQueue(ctx, int(beam.GetClosesAtBlock()), beam.GetId())
		bk.Logger(ctx).Info(fmt.Sprintf("Removed beam %s from open beam by block queue", beam.GetId()))
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
		bk.Logger(ctx).Info(fmt.Sprintf("Deleted beam entity with ID %s", beam.GetId()))
		return false
	})
	return nil
}
