package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"

	"github.com/lum-network/chain/x/millions/types"
)

// MarshalSetWithdrawAddressCallbackArgs Marshal delegate RedelegateCallback arguments
func (k Keeper) MarshalSetWithdrawAddressCallbackArgs(ctx sdk.Context, setWithdrawAddrCallback types.SetWithdrawAddressCallback) ([]byte, error) {
	out, err := k.cdc.Marshal(&setWithdrawAddrCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalSetWithdrawAddressCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// UnmarshalSetWithdrawAddressCallbackArgs Marshal delegate callback arguments into a RedelegateCallback struct
func (k Keeper) UnmarshalSetWithdrawAddressCallbackArgs(ctx sdk.Context, setWithdrawAddrCallback []byte) (*types.SetWithdrawAddressCallback, error) {
	unmarshalledWithdrawAddrCallback := types.SetWithdrawAddressCallback{}
	if err := k.cdc.Unmarshal(setWithdrawAddrCallback, &unmarshalledWithdrawAddrCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalSetWithdrawAddressCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledWithdrawAddrCallback, nil
}

func SetWithdrawAddressCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Deserialize the callback args
	setWithdrawAddressCallback, err := k.UnmarshalSetWithdrawAddressCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal set withdraw address callback args: %s", err.Error()))
	}

	// Acquire the pool instance from the callback
	pool, err := k.GetPool(ctx, setWithdrawAddressCallback.GetPoolId())
	if err != nil {
		return err
	}

	// Scenarios:
	// This operation is done right after the ICA channel open procedure
	// - Timeout: Does nothing - fix may be needed manually
	// - Error: Does nothing - fix may be needed manually
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Debug("Received timeout for a set withdraw address packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Debug("Received failure for a set withdraw address packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_SUCCESS {
		k.Logger(ctx).Debug("Received success for a set withdraw address packet")
		_, err := k.OnSetupPoolWithdrawalAddressCompleted(ctx, pool.PoolId)
		return err
	}
	return nil
}
