package keeper

import (
	"context"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/lum-network/chain/x/beam/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// OpenBeam Create a new beam instance
func (k Keeper) OpenBeam(goCtx context.Context, msg *types.MsgOpenBeam) (*types.MsgOpenBeamResponse, error) {
	return &types.MsgOpenBeamResponse{}, sdkerrors.ErrNotSupported
}

// UpdateBeam Update a beam instance and proceeds any require state machine update
func (k Keeper) UpdateBeam(goCtx context.Context, msg *types.MsgUpdateBeam) (*types.MsgUpdateBeamResponse, error) {
	return &types.MsgUpdateBeamResponse{}, sdkerrors.ErrNotSupported
}

// ClaimBeam Final user endpoint to claim and acquire the money
func (k Keeper) ClaimBeam(goCtx context.Context, msg *types.MsgClaimBeam) (*types.MsgClaimBeamResponse, error) {
	return &types.MsgClaimBeamResponse{}, sdkerrors.ErrNotSupported

}
