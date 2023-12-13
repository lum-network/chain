package beam

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/beam/keeper"
)

// EndBlocker Called every block, process the beam expiration and auto close
func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
}
