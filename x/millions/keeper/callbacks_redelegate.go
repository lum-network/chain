package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"

	"github.com/lum-network/chain/x/millions/types"
)

// MarshalRedelegateCallbackArgs Marshal delegate RedelegateCallback arguments
func (k Keeper) MarshalRedelegateCallbackArgs(ctx sdk.Context, redelegateCallback types.RedelegateCallback) ([]byte, error) {
	out, err := k.cdc.Marshal(&redelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalRedelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// UnmarshalRedelegateCallbackArgs Marshal delegate callback arguments into a RedelegateCallback struct
func (k Keeper) UnmarshalRedelegateCallbackArgs(ctx sdk.Context, redelegateCallback []byte) (*types.RedelegateCallback, error) {
	unmarshalledRedelegateCallback := types.RedelegateCallback{}
	if err := k.cdc.Unmarshal(redelegateCallback, &unmarshalledRedelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalRedelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledRedelegateCallback, nil
}

// Get the latest completion time across each MsgBeginRedelegate in the ICA transaction
func (k Keeper) GetLatestRedelegationCompletionTime(ctx sdk.Context, msgResponses [][]byte) (*time.Time, error) {
	// Update the completion time using the latest completion time across each message within the transaction
	latestCompletionTime := time.Time{}

	for _, msgResponse := range msgResponses {
		// unmarshall the ack response into a MsgBeginRedelegateResponse and grab the completion time
		var beginRedelegateResponse stakingtypes.MsgBeginRedelegateResponse
		err := k.cdc.Unmarshal(msgResponse, &beginRedelegateResponse)
		if err != nil {
			return nil, errorsmod.Wrapf(types.ErrUnmarshalFailure, "Unable to unmarshal redelegation tx response: %s", err.Error())
		}
		if beginRedelegateResponse.CompletionTime.After(latestCompletionTime) {
			latestCompletionTime = beginRedelegateResponse.CompletionTime
		}
	}

	if latestCompletionTime.IsZero() {
		return nil, errorsmod.Wrapf(types.ErrInvalidPacketCompletionTime, "Invalid completion time (%s) from txMsg", latestCompletionTime.String())
	}
	return &latestCompletionTime, nil
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
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Debug("Received timeout for a redelegate packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Debug("Received failure for a redelegate packet")
		return k.OnRedelegateOnNativeChainCompleted(ctx, redelegateCallback.GetPoolId(), redelegateCallback.GetOperatorAddress(), redelegateCallback.GetSplitDelegations(), nil, true)
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_SUCCESS {
		k.Logger(ctx).Debug("Received success for a redelegate packet")
		redelegationEndsAt, err := k.GetLatestRedelegationCompletionTime(ctx, ackResponse.MsgResponses)
		if err != nil {
			return err
		}
		return k.OnRedelegateOnNativeChainCompleted(ctx, redelegateCallback.GetPoolId(), redelegateCallback.GetOperatorAddress(), redelegateCallback.GetSplitDelegations(), redelegationEndsAt, false)
	}
	return nil
}
