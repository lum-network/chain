package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/millions/types"
	"strconv"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// GenerateSeed generate a random seed using the same logic as draws, and return it (allows for off-chain drawing of data)
func (k msgServer) GenerateSeed(goCtx context.Context, msg *types.MsgGenerateSeed) (*types.MsgGenerateSeedResponse, error) {
	// Grab our application context
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Compute our random seed
	seed := k.ComputeRandomSeed(ctx)

	// Emit events and return the message response
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeGenerateSeed,
			sdk.NewAttribute(types.AttributeSeed, strconv.FormatInt(seed, 10)),
		),
	})
	return &types.MsgGenerateSeedResponse{Seed: seed}, nil
}
