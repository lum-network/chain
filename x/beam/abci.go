package beam

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/beam/keeper"
	"github.com/lum-network/chain/x/beam/types"
	"time"
)

// EndBlocker Called every block, process the beam expiration and auto close
func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	// Notify the telemetry module
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Process beams expiration
	keeper.IterateOpenBeamsQueue(ctx, func(beam types.Beam) bool {
		keeper.Logger(ctx).Debug(fmt.Sprintf("Processing beam %s", beam.GetId()))

		// Here, zero is considered as disabled auto close
		if beam.GetClosesAtBlock() == 0 {
			return false
		}

		// If beam hasn't passed thresold, continue
		if ctx.BlockHeight() != int64(beam.GetClosesAtBlock()) {
			return false
		}

		err := keeper.UpdateBeamStatus(ctx, beam.GetId(), types.BeamState_StateCanceled)
		if err != nil {
			panic(err)
		}

		keeper.Logger(ctx).Info(fmt.Sprintf("Canceling beam #%s due to crossed auto close thresold", beam.GetId()))
		return false
	})
}
