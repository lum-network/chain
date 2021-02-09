package chain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sandblockio/chain/x/chain/keeper"
	"github.com/sandblockio/chain/x/chain/types"
)

func handleMsgOpenBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgOpenBeam) (*sdk.Result, error) {
	//TODO: handle errors
	k.OpenBeam(ctx, *msg)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgIncreaseBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgIncreaseBeam) (*sdk.Result, error) {
	//TODO: handle errors
	k.IncreaseBeam(ctx, *msg)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgCloseBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgCloseBeam) (*sdk.Result, error) {
	//TODO: handle errors
	k.CloseBeam(ctx, *msg)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgCancelBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgCancelBeam) (*sdk.Result, error) {
	//TODO: handle errors
	k.CancelBeam(ctx, *msg)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgClaimBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgClaimBeam) (*sdk.Result, error) {
	//TODO: handle errors
	k.ClaimBeam(ctx, *msg)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
