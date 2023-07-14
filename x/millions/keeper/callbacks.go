package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
)

const (
	ICACallbackID_Delegate           = "delegate"
	ICACallbackID_Claim              = "claim"
	ICACallbackID_Undelegate         = "undelegate"
	ICACallbackID_Redelegate         = "redelegate"
	ICACallbackID_TransferToNative   = "transfer_to_native"
	ICACallbackID_TransferFromNative = "transfer_from_native"
	ICACallbackID_SetWithdrawAddress = "set_withdraw_address"
	ICACallbackID_BankSend           = "bank_send"
)

// ICACallback wrapper struct for millions keeper
type ICACallback func(Keeper, sdk.Context, channeltypes.Packet, *icacallbackstypes.AcknowledgementResponse, []byte) error

type ICACallbacks struct {
	k         Keeper
	callbacks map[string]ICACallback
}

var _ icacallbackstypes.ICACallbackHandler = ICACallbacks{}

func (k Keeper) ICACallbackHandler() ICACallbacks {
	return ICACallbacks{k, make(map[string]ICACallback)}
}

func (c ICACallbacks) CallICACallback(ctx sdk.Context, id string, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	return c.callbacks[id](c.k, ctx, packet, ackResponse, args)
}

func (c ICACallbacks) HasICACallback(id string) bool {
	_, found := c.callbacks[id]
	return found
}

func (c ICACallbacks) AddICACallback(id string, fn interface{}) icacallbackstypes.ICACallbackHandler {
	c.callbacks[id] = fn.(ICACallback)
	return c
}

func (c ICACallbacks) RegisterICACallbacks() icacallbackstypes.ICACallbackHandler {
	a := c.AddICACallback(ICACallbackID_Delegate, ICACallback(DelegateCallback)).
		AddICACallback(ICACallbackID_Undelegate, ICACallback(UndelegateCallback)).
		AddICACallback(ICACallbackID_Redelegate, ICACallback(RedelegateCallback)).
		AddICACallback(ICACallbackID_Claim, ICACallback(ClaimCallback)).
		AddICACallback(ICACallbackID_TransferFromNative, ICACallback(TransferFromNativeCallback)).
		AddICACallback(ICACallbackID_TransferToNative, ICACallback(TransferToNativeCallback)).
		AddICACallback(ICACallbackID_SetWithdrawAddress, ICACallback(SetWithdrawAddressCallback)).
		AddICACallback(ICACallbackID_BankSend, ICACallback(BankSendCallback))
	return a.(ICACallbacks)
}
