package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgDeposit{}

func NewMsgDeposit(id string, depositor string, amount sdk.Coin) *MsgDeposit {
	return &MsgDeposit{
		Id:               id,
		DepositorAddress: depositor,
		Amount:           amount,
	}
}

func (msg MsgDeposit) Route() string {
	return RouterKey
}

func (msg MsgDeposit) Type() string {
	return "Deposit"
}

func (msg *MsgDeposit) GetSigners() []sdk.AccAddress {
	depositor, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress())
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{depositor}
}

func (msg *MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeposit) ValidateBasic() error {
	if len(msg.Id) <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Invalid id supplied (%d)", len(msg.Id))
	}

	// Ensure the address is correct and that we are able to acquire it
	_, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress())
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid creator address (%s)", err)
	}

	// If we have an amount, make sure it is not negative nor zero
	if msg.GetAmount().IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "Invalid amount: must be greater than 0")
	}
	return nil
}
