package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgDepositEdit{}

func NewMsgDepositEdit(depositorAddr string, poolID uint64, depositID uint64) *MsgDepositEdit {
	return &MsgDepositEdit{
		PoolId:           poolID,
		DepositId:        depositID,
		DepositorAddress: depositorAddr,
	}
}

func (msg MsgDepositEdit) Route() string {
	return RouterKey
}

func (msg MsgDepositEdit) Type() string {
	return "DepositEdit"
}

func (msg *MsgDepositEdit) GetSigners() []sdk.AccAddress {
	depositor := sdk.MustAccAddressFromBech32(msg.GetDepositorAddress())
	return []sdk.AccAddress{depositor}
}

func (msg *MsgDepositEdit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDepositEdit) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	if _, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress()); err != nil {
		return ErrInvalidDepositorAddress
	}

	if msg.GetPoolId() == UnknownID || msg.GetDepositId() == UnknownID {
		return ErrInvalidID
	}
	return nil
}
