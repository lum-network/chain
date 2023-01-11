package beam

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/beam/keeper"
	"github.com/lum-network/chain/x/beam/types"
)

// EndBlocker Called every block, process the beam expiration and auto close
func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	// Notify the telemetry module
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Acquire the list of beams
	ids := keeper.GetBeamIDsFromBlockQueue(ctx, int(ctx.BlockHeight()))

	// Process beams expirations
	for _, value := range ids {
		err := keeper.UpdateBeamStatus(ctx, value, types.BeamState_StateCanceled)
		if err != nil {
			panic(err)
		}
		keeper.Logger(ctx).Info(fmt.Sprintf("Canceling beam #%s due to crossed auto close thresold", value), "height", ctx.BlockHeight())
	}
}
