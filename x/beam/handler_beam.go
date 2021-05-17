package beam

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/beam/keeper"
	"github.com/lum-network/chain/x/beam/types"
)

func handleMsgOpenBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgOpenBeam) (*sdk.Result, error) {
	err := k.OpenBeam(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgUpdateBeam(ctx sdk.Context, k keeper.Keeper, msg *types.MsgUpdateBeam) (*sdk.Result, error) {
	err := k.UpdateBeam(ctx, *msg)
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
