package chain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sandblockio/chain/x/chain/keeper"
	"github.com/sandblockio/chain/x/chain/types"
)

func handleMsgOpenBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgOpenBeam) (*sdk.Result, error) {
	k.OpenBeam(ctx, *msg)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
