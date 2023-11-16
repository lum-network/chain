package keeper

import (
	"context"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

// WithdrawDeposit withdraw a deposit from the transaction message
func (k msgServer) WithdrawDeposit(goCtx context.Context, msg *types.MsgWithdrawDeposit) (*types.MsgWithdrawDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pool, err := k.GetPool(ctx, msg.PoolId)
	if err != nil {
		return nil, types.ErrPoolNotFound
	}

	// Get the pool deposit
	deposit, err := k.GetPoolDeposit(ctx, msg.PoolId, msg.DepositId)
	if err != nil {
		return nil, types.ErrDepositNotFound
	}

	// Verify addresses
	depositorAddr, err := sdk.AccAddressFromBech32(msg.DepositorAddress)
	if err != nil {
		return nil, types.ErrInvalidDepositorAddress
	}

	_, toAddr, err := pool.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidDestinationAddress, err.Error())
	}

	// Verify that the incoming message withdrawAddr matches the deposit depositor_address
	if depositorAddr.String() != deposit.DepositorAddress {
		return nil, types.ErrInvalidWithdrawalDepositorAddress
	}

	// Verify that the deposit can be withdrawn
	if deposit.State != types.DepositState_Success {
		return nil, types.ErrInvalidDepositState
	}

	// Prepare the withdrawal instance
	withdrawal := types.Withdrawal{
		PoolId:           msg.PoolId,
		DepositId:        msg.DepositId,
		WithdrawalId:     k.GetNextWithdrawalIdAndIncrement(ctx),
		State:            types.WithdrawalState_Pending,
		DepositorAddress: deposit.DepositorAddress,
		ToAddress:        toAddr.String(),
		Amount:           deposit.Amount,
		CreatedAtHeight:  ctx.BlockHeight(),
		UpdatedAtHeight:  ctx.BlockHeight(),
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
	}

	// Adds the withdrawal and remove the deposit
	k.AddWithdrawal(ctx, withdrawal)
	k.RemoveDeposit(ctx, &deposit)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeWithdrawDeposit,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(withdrawal.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyWithdrawalID, strconv.FormatUint(withdrawal.WithdrawalId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositID, strconv.FormatUint(deposit.DepositId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositor, withdrawal.DepositorAddress),
			sdk.NewAttribute(types.AttributeKeyRecipient, withdrawal.ToAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, withdrawal.Amount.String()),
		),
	})

	// Add epoch unbonding
	if err := k.AddEpochUnbonding(ctx, withdrawal, false); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawDepositResponse{WithdrawalId: withdrawal.WithdrawalId}, err
}

// WithdrawDepositRetry allows the user to manually trigger the withdrawal of their deposit in case of failure
func (k msgServer) WithdrawDepositRetry(goCtx context.Context, msg *types.MsgWithdrawDepositRetry) (*types.MsgWithdrawDepositRetryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Ensure msg received is valid
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	withdrawal, err := k.GetPoolWithdrawal(ctx, msg.PoolId, msg.WithdrawalId)
	if err != nil {
		return nil, err
	}

	depositorAddr, err := sdk.AccAddressFromBech32(msg.DepositorAddress)
	if err != nil {
		return nil, types.ErrInvalidDepositorAddress
	}

	// Verify that the depositorAddress msg is the same as the entity pool withdrawal depositor
	if depositorAddr.String() != withdrawal.DepositorAddress {
		return nil, types.ErrInvalidDepositorAddress
	}

	// State should be set to failure in order to retry something
	if withdrawal.State != types.WithdrawalState_Failure {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidWithdrawalState,
			"state is %s instead of %s",
			withdrawal.State.String(), types.WithdrawalState_Failure.String(),
		)
	}

	newState := types.WithdrawalState_Unspecified
	if withdrawal.ErrorState == types.WithdrawalState_IbcTransfer {
		newState = types.WithdrawalState_IbcTransfer
		k.UpdateWithdrawalStatus(ctx, withdrawal.PoolId, withdrawal.WithdrawalId, newState, withdrawal.UnbondingEndsAt, false)
		if err := k.TransferWithdrawalToRecipient(ctx, withdrawal.PoolId, withdrawal.WithdrawalId); err != nil {
			return nil, err
		}
	} else if withdrawal.ErrorState == types.WithdrawalState_IcaUndelegate {
		// Only possible following a restore account which put entities in this error state
		// otherwise the retry is automatically triggered by the ICA callback
		err := k.AddEpochUnbonding(ctx, withdrawal, true)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidWithdrawalState,
			"error_state is %s instead of %s or %s",
			withdrawal.ErrorState.String(), types.WithdrawalState_IcaUndelegate.String(), types.WithdrawalState_IbcTransfer.String(),
		)
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeWithdrawDepositRetry,
			sdk.NewAttribute(types.AttributeKeyState, newState.String()),
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(withdrawal.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyWithdrawalID, strconv.FormatUint(withdrawal.WithdrawalId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositor, withdrawal.DepositorAddress),
			sdk.NewAttribute(types.AttributeKeyRecipient, withdrawal.ToAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, withdrawal.Amount.String()),
		),
	})

	return &types.MsgWithdrawDepositRetryResponse{}, nil
}
