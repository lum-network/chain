package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgUnbond{}

// NewMsgUnbond creates a new MsgUnbond instance
func NewMsgUnbond(delegatorAddress string) *MsgUnbond {
	return &MsgUnbond{
		DelegatorAddress: delegatorAddress,
	}
}

// Route returns the message route for MsgUnbond
func (msg MsgUnbond) Route() string { return RouterKey }

// Type returns the message type for MsgUnbond
func (msg MsgUnbond) Type() string { return "Unbound" }

// Used to identify which accounts must sign a message in order for it to be considered valid
func (msg *MsgUnbond) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.GetDelegatorAddress())
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// Used to create a consistent representation of the message's data that can be signed
func (msg *MsgUnbond) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the MsgUnbond fields
func (msg *MsgUnbond) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.GetDelegatorAddress())

	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}
	return nil
}
