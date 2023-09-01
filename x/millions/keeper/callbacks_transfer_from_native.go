package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"

	"github.com/lum-network/chain/x/millions/types"
)

// MarshalTransferFromNativeCallbackArgs Marshal TransferFromNativeCallback arguments
func (k Keeper) MarshalTransferFromNativeCallbackArgs(ctx sdk.Context, transferCallback types.TransferFromNativeCallback) ([]byte, error) {
	out, err := k.cdc.Marshal(&transferCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalTransferFromNativeCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// UnmarshalTransferFromNativeCallbackArgs Marshal callback arguments into a TransferFromNativeCallback struct
func (k Keeper) UnmarshalTransferFromNativeCallbackArgs(ctx sdk.Context, transferCallback []byte) (*types.TransferFromNativeCallback, error) {
	unmarshalledTransferCallback := types.TransferFromNativeCallback{}
	if err := k.cdc.Unmarshal(transferCallback, &unmarshalledTransferCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalTransferFromNativeCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledTransferCallback, nil
}

func TransferFromNativeCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Deserialize the callback args
	transferCallback, err := k.UnmarshalTransferFromNativeCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal transfer from native callback args: %s", err.Error()))
	}

	// Acquire the pool instance from the callback
	_, err = k.GetPool(ctx, transferCallback.GetPoolId())
	if err != nil {
		return err
	}

	// If the response status is a timeout, that's not an "error" since the relayer will retry then fail or succeed.
	// We just log it out and return no error
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Debug("Received timeout for a transfer from native packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Debug("Received failure for a transfer from native packet")
		if transferCallback.Type == types.TransferType_Claim {
			_, err := k.OnTransferFreshPrizePoolCoinsToLocalZoneCompleted(ctx, transferCallback.GetPoolId(), transferCallback.GetDrawId(), true)
			return err
		} else if transferCallback.Type == types.TransferType_Withdraw {
			return k.OnTransferWithdrawalToRecipientCompleted(ctx, transferCallback.GetPoolId(), transferCallback.GetWithdrawalId(), true)
		}
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_SUCCESS {
		k.Logger(ctx).Debug("Received success for a transfer from native packet")
		if transferCallback.Type == types.TransferType_Claim {
			_, err := k.OnTransferFreshPrizePoolCoinsToLocalZoneCompleted(ctx, transferCallback.GetPoolId(), transferCallback.GetDrawId(), false)
			return err
		} else if transferCallback.Type == types.TransferType_Withdraw {
			return k.OnTransferWithdrawalToRecipientCompleted(ctx, transferCallback.GetPoolId(), transferCallback.GetWithdrawalId(), false)
		}
	}
	return nil
}
