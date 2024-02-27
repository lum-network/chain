package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"

	"github.com/lum-network/chain/x/millions/types"
)

// MarshalClaimCallbackArgs Marshal claim ClaimCallback arguments
func (k Keeper) MarshalClaimCallbackArgs(ctx sdk.Context, claimCallback types.ClaimRewardsCallback) ([]byte, error) {
	out, err := k.cdc.Marshal(&claimCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalClaimCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// UnmarshalClaimCallbackArgs Marshal claim callback arguments into a ClaimCallback struct
func (k Keeper) UnmarshalClaimCallbackArgs(ctx sdk.Context, claimCallback []byte) (*types.ClaimRewardsCallback, error) {
	unmarshalledClaimCallback := types.ClaimRewardsCallback{}
	if err := k.cdc.Unmarshal(claimCallback, &unmarshalledClaimCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalClaimCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledClaimCallback, nil
}

func ClaimCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Deserialize the callback args
	claimCallback, err := k.UnmarshalClaimCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal claim callback args: %s", err.Error()))
	}

	// Scenarios:
	// - Timeout: Does nothing, handled when restoring the ICA channel
	// - Error: Put entity in error state to allow users to retry
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Debug("Received timeout for a claim packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Debug("Received failure for a claim packet")
		_, err = k.OnClaimYieldOnRemoteZoneCompleted(ctx, claimCallback.GetPoolId(), claimCallback.GetDrawId(), true)
		return err

	} else if ackResponse.Status == icacallbackstypes.AckResponseStatus_SUCCESS {
		k.Logger(ctx).Debug("Received success for a claim packet")
		_, err = k.OnClaimYieldOnRemoteZoneCompleted(ctx, claimCallback.GetPoolId(), claimCallback.GetDrawId(), false)
		return err
	}
	return nil
}
