package millions

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/keeper"
	"github.com/lum-network/chain/x/millions/types"
)

func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	keeper.BlockPoolUpdates(ctx)
}

func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	keeper.BlockPrizeUpdates(ctx)
	keeper.BlockWithdrawalUpdates(ctx)
}
