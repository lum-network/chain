package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"

	"github.com/lum-network/chain/x/millions/types"
)

// MarshalDelegateCallbackArgs Marshal delegate DelegateCallback arguments.
func (k Keeper) MarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback types.DelegateCallback) ([]byte, error) {
	out, err := k.cdc.Marshal(&delegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalDelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// UnmarshalDelegateCallbackArgs Marshal delegate callback arguments into a DelegateCallback struct.
func (k Keeper) UnmarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback []byte) (*types.DelegateCallback, error) {
	unmarshalledDelegateCallback := types.DelegateCallback{}
	if err := k.cdc.Unmarshal(delegateCallback, &unmarshalledDelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalDelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledDelegateCallback, nil
}

func DelegateCallback(k Keeper, ctx sdk.Context, _ channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Deserialize the callback args
	delegateCallback, err := k.UnmarshalDelegateCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal delegate callback args: %s", err.Error()))
	}

	// If the response status is a timeout, that's not an "error" since the relayer will retry then fail or succeed.
	// We just log it out and return no error
	if ackResponse.Status == icacallbackstypes.AckResponseStatusTimeout {
		k.Logger(ctx).Debug("Received timeout for a delegate packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatusFailure {
		k.Logger(ctx).Debug("Received failure for a delegate packet")
		return k.OnDelegateDepositOnNativeChainCompleted(ctx, delegateCallback.GetPoolId(), delegateCallback.GetDepositId(), delegateCallback.GetSplitDelegations(), true)
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatusSuccess {
		k.Logger(ctx).Debug("Received success for a delegate packet")
		return k.OnDelegateDepositOnNativeChainCompleted(ctx, delegateCallback.GetPoolId(), delegateCallback.GetDepositId(), delegateCallback.GetSplitDelegations(), false)
	}
	return nil
}
