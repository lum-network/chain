package faucet

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sandblockio/chain/x/faucet/keeper"
	"github.com/sandblockio/chain/x/faucet/types"
)

func handleMsgMintAndSend(ctx sdk.Context, k keeper.Keeper, msg *types.MsgMintAndSend) (*sdk.Result, error) {
	err := k.MintAndSend(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
