package keeper

import (
	"context"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

// Deposit creates a deposit from the transaction message
func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check internal validation before creating a deposit
	deposit, err := k.CreateDeposit(ctx, msg, types.DepositOrigin_Direct)
	if err != nil {
		return nil, err
	}

	return &types.MsgDepositResponse{DepositId: deposit.DepositId}, nil
}

func (k msgServer) DepositRetry(goCtx context.Context, msg *types.MsgDepositRetry) (*types.MsgDepositRetryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Ensure msg received is valid
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	deposit, err := k.GetPoolDeposit(ctx, msg.PoolId, msg.DepositId)
	if err != nil {
		return nil, err
	}

	depositorAddr, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress())
	if err != nil {
		return nil, types.ErrInvalidDepositorAddress
	}

	// Verify that the depositorAddr msg is the same as the entity depositor
	if depositorAddr.String() != deposit.DepositorAddress {
		return nil, types.ErrInvalidDepositorAddress
	}

	// State should be set to failure in order to retry something
	if deposit.State != types.DepositState_Failure {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidDepositState,
			"state is %s instead of %s",
			deposit.State.String(), types.DepositState_Failure.String(),
		)
	}

	newState := types.DepositState_Unspecified
	if deposit.ErrorState == types.DepositState_IbcTransfer {
		newState = types.DepositState_IbcTransfer
		k.UpdateDepositStatus(ctx, deposit.PoolId, deposit.DepositId, newState, false)
		if err := k.TransferDepositToRemoteZone(ctx, deposit.PoolId, deposit.DepositId); err != nil {
			return nil, err
		}
	} else if deposit.ErrorState == types.DepositState_IcaDelegate {
		newState = types.DepositState_IcaDelegate
		k.UpdateDepositStatus(ctx, deposit.PoolId, deposit.DepositId, newState, false)
		if err := k.DelegateDepositOnRemoteZone(ctx, deposit.PoolId, deposit.DepositId); err != nil {
			return nil, err
		}
	} else {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidDepositState,
			"error_state is %s instead of %s or %s",
			deposit.ErrorState.String(), types.DepositState_IbcTransfer.String(), types.DepositState_IcaDelegate.String(),
		)
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeDepositRetry,
			sdk.NewAttribute(types.AttributeKeyState, newState.String()),
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(deposit.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositID, strconv.FormatUint(deposit.DepositId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositor, deposit.DepositorAddress),
			sdk.NewAttribute(types.AttributeKeyWinner, deposit.WinnerAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, deposit.Amount.String()),
		),
	})

	return &types.MsgDepositRetryResponse{}, nil
}

func (k msgServer) DepositEdit(goCtx context.Context, msg *types.MsgDepositEdit) (*types.MsgDepositEditResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var isSponsor bool

	// Ensure msg received is valid
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	deposit, err := k.GetPoolDeposit(ctx, msg.PoolId, msg.DepositId)
	if err != nil {
		return nil, err
	}

	depositorAddr, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress())
	if err != nil {
		return nil, types.ErrInvalidDepositorAddress
	}

	// Verify that the depositorAddr msg is the same as the entity depositor
	if depositorAddr.String() != deposit.DepositorAddress {
		return nil, types.ErrInvalidDepositorAddress
	}

	winnerAddress, err := sdk.AccAddressFromBech32(deposit.WinnerAddress)
	if err != nil {
		return nil, types.ErrInvalidWinnerAddress
	}

	if strings.TrimSpace(msg.GetWinnerAddress()) != "" {
		winnerAddress, err = sdk.AccAddressFromBech32(msg.GetWinnerAddress())
		if err != nil {
			return nil, types.ErrInvalidWinnerAddress
		}
	}

	if msg.IsSponsor != nil {
		isSponsor = msg.IsSponsor.Value
	} else {
		isSponsor = deposit.IsSponsor
	}

	if !depositorAddr.Equals(winnerAddress) && isSponsor {
		return nil, types.ErrInvalidSponsorWinnerCombo
	}

	// Edit deposit
	if err := k.EditDeposit(ctx, msg.PoolId, deposit.DepositId, winnerAddress, isSponsor); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeDepositEdit,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(deposit.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositID, strconv.FormatUint(deposit.DepositId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyWinner, winnerAddress.String()),
			sdk.NewAttribute(types.AttributeKeySponsor, strconv.FormatBool(isSponsor)),
		),
	})

	return &types.MsgDepositEditResponse{}, nil
}
