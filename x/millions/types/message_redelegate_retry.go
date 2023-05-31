package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRedelegateRetry{}

func NewMsgRedelegateRetry(RedelegateRetryAddr string, poolID uint64, operatorAddr string) *MsgRedelegateRetry {
	return &MsgRedelegateRetry{
		PoolId:                 poolID,
		OperatetorAddress:      operatorAddr,
		RedelegateRetryAddress: RedelegateRetryAddr,
	}
}

func (msg MsgRedelegateRetry) Route() string {
	return RouterKey
}

func (msg MsgRedelegateRetry) Type() string {
	return "RedelegateRetry"
}

func (msg *MsgRedelegateRetry) GetSigners() []sdk.AccAddress {
	depositor := sdk.MustAccAddressFromBech32(msg.GetRedelegateRetryAddress())
	return []sdk.AccAddress{depositor}
}

func (msg *MsgRedelegateRetry) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRedelegateRetry) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	if _, err := sdk.AccAddressFromBech32(msg.GetRedelegateRetryAddress()); err != nil {
		return ErrInvalidRedelegateRetryAddress
	}

	if msg.GetPoolId() == UnknownID {
		return ErrInvalidID
	}

	return nil
}
