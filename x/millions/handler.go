package millions

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/lum-network/chain/x/millions/keeper"
	"github.com/lum-network/chain/x/millions/types"
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
		case *types.MsgDepositRetry:
			res, err := msgServer.DepositRetry(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDepositEdit:
			res, err := msgServer.DepositEdit(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgClaimPrize:
			res, err := msgServer.ClaimPrize(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDrawRetry:
			res, err := msgServer.DrawRetry(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgWithdrawDeposit:
			res, err := msgServer.WithdrawDeposit(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgWithdrawDepositRetry:
			res, err := msgServer.WithdrawDepositRetry(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRestoreInterchainAccounts:
			res, err := msgServer.RestoreInterchainAccounts(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
