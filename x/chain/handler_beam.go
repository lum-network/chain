package chain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/chain/keeper"
	"github.com/lum-network/chain/x/chain/types"
)

func handleMsgOpenBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgOpenBeam) (*sdk.Result, error) {
	err := k.OpenBeam(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgIncreaseBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgIncreaseBeam) (*sdk.Result, error) {
	err := k.IncreaseBeam(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgCloseBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgCloseBeam) (*sdk.Result, error) {
	err := k.CloseBeam(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgCancelBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgCancelBeam) (*sdk.Result, error) {
	err := k.CancelBeam(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgClaimBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgClaimBeam) (*sdk.Result, error) {
	err := k.ClaimBeam(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
