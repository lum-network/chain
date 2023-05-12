package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgDrawRetry{}

func NewMsgDrawRetry(drawRetryAddr string, poolID uint64, drawID uint64) *MsgDrawRetry {
	return &MsgDrawRetry{
		DrawRetryAddress: drawRetryAddr,
		PoolId:           poolID,
		DrawId:           drawID,
	}
}

func (msg MsgDrawRetry) Route() string {
	return RouterKey
}

func (msg MsgDrawRetry) Type() string {
	return "DrawRetry"
}

func (msg *MsgDrawRetry) GetSigners() []sdk.AccAddress {
	depositor := sdk.MustAccAddressFromBech32(msg.GetDrawRetryAddress())
	return []sdk.AccAddress{depositor}
}

func (msg *MsgDrawRetry) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDrawRetry) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	if _, err := sdk.AccAddressFromBech32(msg.GetDrawRetryAddress()); err != nil {
		return ErrInvalidDrawRetryAddress
	}

	if msg.GetPoolId() == UnknownID || msg.GetDrawId() == UnknownID {
		return ErrInvalidID
	}

	return nil
}
