package beam

import (
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

	logger := keeper.Logger(ctx)

	// Process beams expiration
	keeper.IterateOpenBeamsQueue(ctx, func(beam types.Beam) bool {
		// Here, zero is considered as disabled auto close
		if beam.GetClosesAtBlock() == 0 {
			return false
		}

		// If beam hasn't passed thresold, continue
		if ctx.BlockHeight() < int64(beam.GetClosesAtBlock()) {
			return false
		}

		err := keeper.UpdateBeamStatus(ctx, beam.GetId(), types.BeamState_StateCanceled)
		if err != nil {
			panic(err)
		}

		logger.Info("Expired beam #", beam.GetId(), " by crossing closes at block treshold")

		return false
	})
}
