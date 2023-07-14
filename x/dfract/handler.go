package dfract

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/lum-network/chain/x/dfract/keeper"
	"github.com/lum-network/chain/x/dfract/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		goCtx := sdk.WrapSDKContext(ctx)

		switch msg := msg.(type) {
		case *types.MsgDeposit:
			res, err := msgServer.Deposit(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgWithdrawAndMint:
			res, err := msgServer.WithdrawAndMint(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
