package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	"github.com/lum-network/chain/x/millions/types"
)

// MarshalSendBankTransferFromNativeCallbackArgs Marshal SendBankTransferFromNativeCallback arguments
func (k Keeper) MarshalBankSendCallbackArgs(ctx sdk.Context, bankSendCallback types.BankSendCallback) ([]byte, error) {
	out, err := k.cdc.Marshal(&bankSendCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalBankSendCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// UnmarshalSendBankTransferFromNativeCallbackArgs Unmarshal SendBankTransferFromNativeCallback arguments
func (k Keeper) UnmarshalBankSendCallbackArgs(ctx sdk.Context, bankSendCallback []byte) (*types.BankSendCallback, error) {
	unmarshalledBankSendCallback := types.BankSendCallback{}
	if err := k.cdc.Unmarshal(bankSendCallback, &unmarshalledBankSendCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalBankSendCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledBankSendCallback, nil
}

func BankSendCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Deserialize the callback args
	bankSendCallback, err := k.UnmarshalBankSendCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal bank send callback args: %s", err.Error()))
	}

	// Acquire the pool instance from the callback
	pool, err := k.GetPool(ctx, bankSendCallback.GetPoolId())
	if err != nil {
		return err
	}

	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Debug("Received timeout for a bank send to native packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Debug("Received failure for a bank send to native packet")
		return k.OnTransferWithdrawalToRecipientCompleted(ctx, pool.PoolId, bankSendCallback.GetWithdrawalId(), true)
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_SUCCESS {
		k.Logger(ctx).Debug("Received success for a bank send to native packet")
		return k.OnTransferWithdrawalToRecipientCompleted(ctx, pool.PoolId, bankSendCallback.GetWithdrawalId(), false)
	}
	return nil
}
