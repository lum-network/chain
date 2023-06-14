package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	"github.com/lum-network/chain/x/millions/types"
)

// MarshalRedelegateCallbackArgs Marshal delegate RedelegateCallback arguments.
func (k Keeper) MarshalRedelegateCallbackArgs(ctx sdk.Context, redelegateCallback types.RedelegateCallback) ([]byte, error) {
	out, err := k.cdc.Marshal(&redelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalRedelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// UnmarshalRedelegateCallbackArgs Marshal delegate callback arguments into a RedelegateCallback struct.
func (k Keeper) UnmarshalRedelegateCallbackArgs(ctx sdk.Context, redelegateCallback []byte) (*types.RedelegateCallback, error) {
	unmarshalledRedelegateCallback := types.RedelegateCallback{}
	if err := k.cdc.Unmarshal(redelegateCallback, &unmarshalledRedelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalRedelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledRedelegateCallback, nil
}

func RedelegateCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Deserialize the callback args
	redelegateCallback, err := k.UnmarshalRedelegateCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal redelegate callback args: %s", err.Error()))
	}

	// Acquire the pool instance from the callback
	_, err = k.GetPool(ctx, redelegateCallback.GetPoolId())
	if err != nil {
		return err
	}

	// If the response status is a timeout, that's not an "error" since the relayer will retry then fail or succeed.
	// We just log it out and return no error
	if ackResponse.Status == icacallbackstypes.AckResponseStatusTimeout {
		k.Logger(ctx).Debug("Received timeout for a redelegate packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatusFailure {
		k.Logger(ctx).Debug("Received failure for a redelegate packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatusSuccess {
		k.Logger(ctx).Debug("Received success for a redelegate packet")
	}
	return nil
}
