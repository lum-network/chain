package beam

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/lum-network/chain/x/beam/keeper"
	"github.com/lum-network/chain/x/beam/types"
)

// NewHandler new handler for the beam module
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		goCtx := sdk.WrapSDKContext(ctx)

		switch msg := msg.(type) {
		case *types.MsgOpenBeam:
			res, err := msgServer.OpenBeam(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUpdateBeam:
			res, err := msgServer.UpdateBeam(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgClaimBeam:
			res, err := msgServer.ClaimBeam(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
