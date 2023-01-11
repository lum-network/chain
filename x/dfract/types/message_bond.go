package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgBond{}

// NewMsgBond creates a new MsgBond instance
func NewMsgBond(delegatorAddress string, bondedToken sdk.Coin) *MsgBond {
	return &MsgBond{
		DelegatorAddress: delegatorAddress,
		StakingToken:     bondedToken,
	}
}

// Route returns the message route for MsgBond
func (msg MsgBond) Route() string { return RouterKey }

// Type returns the message type for MsgBond
func (msg MsgBond) Type() string { return "Bond" }

// Used to identify which accounts must sign a message in order for it to be considered valid
func (msg *MsgBond) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.GetDelegatorAddress())
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// Used to create a consistent representation of the message's data that can be signed
func (msg *MsgBond) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the MsgUnbond fields
func (msg *MsgBond) ValidateBasic() error {
	// Ensure that the sender is in a correct format
	_, err := sdk.AccAddressFromBech32(msg.GetDelegatorAddress())

	if err != nil {
		return sdkerrors.ErrInvalidAddress
	}

	if msg.StakingToken.Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "Invalid amount: must be greater than 0")
	}

	return nil
}
