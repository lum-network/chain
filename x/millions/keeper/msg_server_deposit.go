package keeper

import (
	"context"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/millions/types"
)

// Deposit creates a deposit from the transaction message.
func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the pool to validate id and denom
	pool, err := k.GetPool(ctx, msg.GetPoolId())
	if err != nil {
		return nil, types.ErrPoolNotFound
	}
	if pool.State != types.PoolState_Ready {
		return nil, types.ErrPoolNotReady
	}

	depositorAddr, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress())
	if err != nil {
		return nil, types.ErrInvalidDepositorAddress
	}

	if msg.Amount.Denom != pool.Denom {
		return nil, types.ErrInvalidDepositDenom
	}

	// Make sure the deposit is sufficient
	if msg.GetAmount().Amount.LT(pool.MinDepositAmount) {
		return nil, types.ErrInsufficientDepositAmount
	}

	winnerAddress := depositorAddr
	if strings.TrimSpace(msg.GetWinnerAddress()) != "" {
		winnerAddress, err = sdk.AccAddressFromBech32(msg.GetWinnerAddress())
		if err != nil {
			return nil, types.ErrInvalidWinnerAddress
		}
	}

	if !depositorAddr.Equals(winnerAddress) && msg.GetIsSponsor() {
		return nil, types.ErrInvalidSponsorWinnerCombo
	}

	// New deposit instance
	deposit := types.Deposit{
		PoolId:           pool.PoolId,
		State:            types.DepositState_IbcTransfer,
		DepositorAddress: depositorAddr.String(),
		Amount:           msg.Amount,
		WinnerAddress:    winnerAddress.String(),
		IsSponsor:        msg.GetIsSponsor(),
		CreatedAtHeight:  ctx.BlockHeight(),
		UpdatedAtHeight:  ctx.BlockHeight(),
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
	}

	// Move funds
	if pool.IsLocalZone(ctx) {
		// Directly send funds to the local deposit address in case of local pool
		if err := k.BankKeeper.SendCoins(
			ctx,
			depositorAddr,
			sdk.MustAccAddressFromBech32(pool.GetIcaDepositAddress()),
			sdk.NewCoins(msg.Amount),
		); err != nil {
			return nil, err
		}
	} else {
		if err := k.BankKeeper.SendCoins(
			ctx,
			depositorAddr,
			sdk.MustAccAddressFromBech32(pool.GetLocalAddress()),
			sdk.NewCoins(msg.Amount),
		); err != nil {
			return nil, err
		}
	}

	// Store deposit
	k.AddDeposit(ctx, &deposit)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeDeposit,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(deposit.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositID, strconv.FormatUint(deposit.DepositId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositor, deposit.DepositorAddress),
			sdk.NewAttribute(types.AttributeKeyWinner, deposit.WinnerAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, deposit.Amount.String()),
		),
	})

	if err := k.TransferDepositToNativeChain(ctx, deposit.GetPoolId(), deposit.GetDepositId()); err != nil {
		return nil, err
	}
	return &types.MsgDepositResponse{DepositId: deposit.DepositId}, nil
}

func (k msgServer) DepositRetry(goCtx context.Context, msg *types.MsgDepositRetry) (*types.MsgDepositRetryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Ensure msg received is valid
	if err := msg.ValidateDepositRetryBasic(); err != nil {
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
		if err := k.TransferDepositToNativeChain(ctx, deposit.PoolId, deposit.DepositId); err != nil {
			return nil, err
		}
	} else if deposit.ErrorState == types.DepositState_IcaDelegate {
		newState = types.DepositState_IcaDelegate
		k.UpdateDepositStatus(ctx, deposit.PoolId, deposit.DepositId, newState, false)
		if err := k.DelegateDepositOnNativeChain(ctx, deposit.PoolId, deposit.DepositId); err != nil {
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
