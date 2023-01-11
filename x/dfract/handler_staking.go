package dfract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/dfract/keeper"
	"github.com/lum-network/chain/x/dfract/types"
)

func handleMsgBond(ctx sdk.Context, k keeper.Keeper, msg *types.MsgBond) (*sdk.Result, error) {
	err := k.Bond(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgUnbond(ctx sdk.Context, k keeper.Keeper, msg *types.MsgUnbond) (*sdk.Result, error) {
	err := k.Unbond(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
