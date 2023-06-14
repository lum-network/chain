package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	"github.com/lum-network/chain/x/millions/types"
)

// MarshalSetWithdrawAddressCallbackArgs Marshal delegate RedelegateCallback arguments.
func (k Keeper) MarshalSetWithdrawAddressCallbackArgs(ctx sdk.Context, setWithdrawAddrCallback types.SetWithdrawAddressCallback) ([]byte, error) {
	out, err := k.cdc.Marshal(&setWithdrawAddrCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalSetWithdrawAddressCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// UnmarshalSetWithdrawAddressCallbackArgs Marshal delegate callback arguments into a RedelegateCallback struct.
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

	// If the response status is a timeout, that's not an "error" since the relayer will retry then fail or succeed.
	// We just log it out and return no error
	if ackResponse.Status == icacallbackstypes.AckResponseStatusTimeout {
		k.Logger(ctx).Debug("Received timeout for a set withdraw address packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatusFailure {
		k.Logger(ctx).Debug("Received failure for a set withdraw address packet")
	} else if ackResponse.Status == icacallbackstypes.AckResponseStatusSuccess {
		k.Logger(ctx).Debug("Received success for a set withdraw address packet")
		_, err := k.OnSetupPoolWithdrawalAddressCompleted(ctx, pool.PoolId)
		return err
	}
	return nil
}
