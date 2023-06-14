package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgWithdrawDeposit{}

func NewMsgWithdrawDeposit(depositor, toAddress string, poolID, depositID uint64) *MsgWithdrawDeposit {
	return &MsgWithdrawDeposit{
		PoolId:           poolID,
		DepositId:        depositID,
		DepositorAddress: depositor,
		ToAddress:        toAddress,
	}
}

func (msg MsgWithdrawDeposit) Route() string {
	return RouterKey
}

func (msg MsgWithdrawDeposit) Type() string {
	return "WithdrawDeposit"
}

func (msg *MsgWithdrawDeposit) GetSigners() []sdk.AccAddress {
	depositor := sdk.MustAccAddressFromBech32(msg.GetDepositorAddress())
	return []sdk.AccAddress{depositor}
}

func (msg *MsgWithdrawDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgWithdrawDeposit) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	if _, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress()); err != nil {
		return ErrInvalidDepositorAddress
	}

	if _, err := sdk.AccAddressFromBech32(msg.GetToAddress()); err != nil {
		return ErrInvalidDestinationAddress
	}

	if msg.GetPoolId() == UnknownID || msg.GetDepositId() == UnknownID {
		return ErrInvalidID
	}

	return nil
}
