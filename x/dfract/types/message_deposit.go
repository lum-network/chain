package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgDeposit{}

func NewMsgDeposit(depositor string, amount sdk.Coin) *MsgDeposit {
	return &MsgDeposit{
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
	depositor := sdk.MustAccAddressFromBech32(msg.GetDepositorAddress())
	return []sdk.AccAddress{depositor}
}

func (msg *MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeposit) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	_, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress())
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, err.Error())
	}

	// If we have an amount, make sure it is not negative nor zero
	if msg.GetAmount().IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "Invalid amount: must be greater than 0")
	}
	return nil
}
