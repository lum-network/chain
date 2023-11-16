package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"

	"github.com/lum-network/chain/x/millions/types"
)

// MarshalTransferToNativeCallbackArgs Marshal TransferToNativeCallback arguments
func (k Keeper) MarshalTransferToNativeCallbackArgs(ctx sdk.Context, transferCallback types.TransferToNativeCallback) ([]byte, error) {
	out, err := k.cdc.Marshal(&transferCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalTransferToNativeCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// UnmarshalTransferToNativeCallbackArgs Marshal callback arguments into a TransferToNativeCallback struct
func (k Keeper) UnmarshalTransferToNativeCallbackArgs(ctx sdk.Context, transferCallback []byte) (*types.TransferToNativeCallback, error) {
	unmarshalledTransferCallback := types.TransferToNativeCallback{}
	if err := k.cdc.Unmarshal(transferCallback, &unmarshalledTransferCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalTransferToNativeCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledTransferCallback, nil
}

func TransferToNativeCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Create a custom temporary cache
	cacheCtx, writeCache := ctx.CacheContext()

	// Deserialize the callback args
	transferCallback, err := k.UnmarshalTransferToNativeCallbackArgs(cacheCtx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal transfer to native callback args: %s", err.Error()))
	}

	// If the response status is a timeout, that's not an "error" since the relayer will retry then fail or succeed.
	// We just log it out and return no error
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(cacheCtx).Debug("Received timeout for a transfer to native packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(cacheCtx).Debug("Received failure for a transfer to native packet")
		if err := k.OnTransferDepositToRemoteZoneCompleted(ctx, transferCallback.GetPoolId(), transferCallback.GetDepositId(), true); err != nil {
			return err
		}

		// Commit the cache
		writeCache()
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_SUCCESS {
		k.Logger(cacheCtx).Debug("Received success for a transfer to native packet.")
		if err := k.OnTransferDepositToRemoteZoneCompleted(ctx, transferCallback.GetPoolId(), transferCallback.GetDepositId(), false); err != nil {
			return err
		}

		// Commit the cache
		writeCache()
	}
	return nil
}
