package v110

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	beamkeeper "github.com/lum-network/chain/x/beam/keeper"
	"github.com/lum-network/chain/x/beam/types"
)

func MigrateBeamQueues(ctx sdk.Context, bk beamkeeper.Keeper) error {
	// Iterate over the whole beams list
	bk.IterateOpenBeamsQueue(ctx, func(beam types.Beam) bool {
		// In case the beam is not intended to auto close, we just pop it out
		if beam.GetClosesAtBlock() <= 0 {
			bk.RemoveFromOpenBeamQueue(ctx, beam.GetId())
			bk.Logger(ctx).Info(fmt.Sprintf("Remove beam #%s from open queue", beam.GetId()))
			return false
		}

		// Append the beam ID to the height entry
		bk.InsertOpenBeamByBlockQueue(ctx, int(beam.GetClosesAtBlock()), beam.GetId())

		// Remove the beam ID from the open beams queue
		bk.RemoveFromOpenBeamQueue(ctx, beam.GetId())

		bk.Logger(ctx).Info(fmt.Sprintf("Migrated beam #%s from open to open by height queues", beam.GetId()))
		return false
	})
	return nil
}
