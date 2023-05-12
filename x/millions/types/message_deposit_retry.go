package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgDepositRetry{}

func NewMsgDepositRetry(depositorAddr string, poolID uint64, depositID uint64) *MsgDepositRetry {
	return &MsgDepositRetry{
		PoolId:           poolID,
		DepositId:        depositID,
		DepositorAddress: depositorAddr,
	}
}

func (msg MsgDepositRetry) Route() string {
	return RouterKey
}

func (msg MsgDepositRetry) Type() string {
	return "DepositRetry"
}

func (msg *MsgDepositRetry) GetSigners() []sdk.AccAddress {
	depositor := sdk.MustAccAddressFromBech32(msg.GetDepositorAddress())
	return []sdk.AccAddress{depositor}
}

func (msg *MsgDepositRetry) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDepositRetry) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	if _, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress()); err != nil {
		return ErrInvalidDepositorAddress
	}

	if msg.GetPoolId() == UnknownID || msg.GetDepositId() == UnknownID {
		return ErrInvalidID
	}

	return nil
}
