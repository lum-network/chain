package dfract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/dfract/keeper"
	"github.com/lum-network/chain/x/dfract/types"
)

func handleMsgDeposit(ctx sdk.Context, k keeper.Keeper, msg *types.MsgDeposit) (*sdk.Result, error) {
	err := k.CreateDeposit(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
