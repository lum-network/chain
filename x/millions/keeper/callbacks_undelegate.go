package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	"github.com/lum-network/chain/x/millions/types"
)

// MarshalUndelegateCallbackArgs Marshal delegate UndelegateCallback arguments
func (k Keeper) MarshalUndelegateCallbackArgs(ctx sdk.Context, undelegateCallback types.UndelegateCallback) ([]byte, error) {
	out, err := k.cdc.Marshal(&undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalUndelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// UnmarshalUndelegateCallbackArgs Marshal delegate callback arguments into a UndelegateCallback struct
func (k Keeper) UnmarshalUndelegateCallbackArgs(ctx sdk.Context, undelegateCallback []byte) (*types.UndelegateCallback, error) {
	unmarshalledUndelegateCallback := types.UndelegateCallback{}
	if err := k.cdc.Unmarshal(undelegateCallback, &unmarshalledUndelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalUndelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledUndelegateCallback, nil
}

// Get the latest completion time across each MsgUndelegate in the ICA transaction
func (k Keeper) GetUnbondingCompletionTime(ctx sdk.Context, msgResponses [][]byte) (*time.Time, error) {
	// Update the completion time using the latest completion time across each message within the transaction
	latestCompletionTime := time.Time{}

	for _, msgResponse := range msgResponses {
		// unmarshall the ack response into a MsgUndelegateResponse and grab the completion time
		var undelegateResponse stakingtypes.MsgUndelegateResponse
		err := k.cdc.Unmarshal(msgResponse, &undelegateResponse)
		if err != nil {
			return nil, errorsmod.Wrapf(types.ErrUnmarshalFailure, "Unable to unmarshal undelegation tx response: %s", err.Error())
		}
		if undelegateResponse.CompletionTime.After(latestCompletionTime) {
			latestCompletionTime = undelegateResponse.CompletionTime
		}
	}

	if latestCompletionTime.IsZero() {
		return nil, errorsmod.Wrapf(types.ErrInvalidPacketCompletionTime, "Invalid completion time (%s) from txMsg", latestCompletionTime.String())
	}
	return &latestCompletionTime, nil
}

func UndelegateCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Create a custom temporary cache
	cacheCtx, writeCache := ctx.CacheContext()

	// Deserialize the callback args
	undelegateCallback, err := k.UnmarshalUndelegateCallbackArgs(cacheCtx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal undelegate callback args: %s", err.Error()))
	}

	// Acquire the pool instance from the callback
	_, err = k.GetPool(cacheCtx, undelegateCallback.GetPoolId())
	if err != nil {
		return err
	}

	// If the response status is a timeout, that's not an "error" since the relayer will retry then fail or succeed.
	// We just log it out and return no error
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(cacheCtx).Debug("Received timeout for an undelegate packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(cacheCtx).Debug("Received failure for an undelegate packet")
		// Failed OnUndelegateEpochUnbondingOnRemoteZoneCompleted
		err := k.OnUndelegateWithdrawalsOnRemoteZoneCompleted(
			cacheCtx,
			undelegateCallback.PoolId,
			undelegateCallback.WithdrawalIds,
			nil,
			true,
		)
		if err != nil {
			return err
		}

		// Commit the cache
		writeCache()
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_SUCCESS {
		k.Logger(cacheCtx).Debug("Received success for an undelegate packet")

		// Update the completion time using the latest completion time across each message within the transaction
		unbondingEndsAt, err := k.GetUnbondingCompletionTime(cacheCtx, ackResponse.MsgResponses)
		if err != nil {
			return err
		}

		// Call the callback handler
		err = k.OnUndelegateWithdrawalsOnRemoteZoneCompleted(
			cacheCtx,
			undelegateCallback.PoolId,
			undelegateCallback.WithdrawalIds,
			unbondingEndsAt,
			false,
		)

		if err != nil {
			return err
		}

		// Commit the cache
		writeCache()
	}
	return nil
}
