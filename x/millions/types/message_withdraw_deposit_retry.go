package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgWithdrawDepositRetry{}

func NewMsgWithdrawDepositRetry(depositorAddr string, poolID, withdrawalID uint64) *MsgWithdrawDepositRetry {
	return &MsgWithdrawDepositRetry{
		PoolId:           poolID,
		WithdrawalId:     withdrawalID,
		DepositorAddress: depositorAddr,
	}
}

func (msg MsgWithdrawDepositRetry) Route() string {
	return RouterKey
}

func (msg MsgWithdrawDepositRetry) Type() string {
	return "WithdrawDepositRetry"
}

func (msg *MsgWithdrawDepositRetry) GetSigners() []sdk.AccAddress {
	depositor := sdk.MustAccAddressFromBech32(msg.GetDepositorAddress())
	return []sdk.AccAddress{depositor}
}

func (msg *MsgWithdrawDepositRetry) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgWithdrawDepositRetry) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	if _, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress()); err != nil {
		return ErrInvalidDepositorAddress
	}

	if msg.GetPoolId() == UnknownID || msg.GetWithdrawalId() == UnknownID {
		return ErrInvalidID
	}

	return nil
}
