package keeper

import (
	"context"
	"strconv"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

// Deposit creates a deposit from the transaction message
func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the pool to validate id and denom
	pool, err := k.GetPool(ctx, msg.GetPoolId())
	if err != nil {
		return nil, types.ErrPoolNotFound
	}

	if pool.State != types.PoolState_Ready && pool.State != types.PoolState_Paused {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidPoolState, "cannot deposit in pool during state %s", pool.State.String(),
		)
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
	poolRunner, err := k.GetPoolRunner(pool.PoolType)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidPoolType, err.Error())
	}
	if err := poolRunner.SendDepositToPool(ctx, pool, deposit); err != nil {
		return nil, err
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
			sdk.NewAttribute(types.AttributeKeySponsor, strconv.FormatBool(msg.IsSponsor)),
		),
	})

	if err := k.TransferDepositToRemoteZone(ctx, deposit.GetPoolId(), deposit.GetDepositId()); err != nil {
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

	newState := types.DepositState_Unspecified
	if deposit.State == types.DepositState_IbcTransfer && ctx.BlockTime().After(deposit.UpdatedAt.Add(types.IBCForceRetryNanos*time.Nanosecond)) {
		// Handle IBC stucked operation for more than 3 days (no timeout received)
		newState = types.DepositState_IbcTransfer
		if err := k.TransferDepositToRemoteZone(ctx, deposit.PoolId, deposit.DepositId); err != nil {
			return nil, err
		}
	} else if deposit.ErrorState == types.DepositState_IbcTransfer {
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
			"state is %s instead of %s",
			deposit.State.String(), types.DepositState_Failure.String(),
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
