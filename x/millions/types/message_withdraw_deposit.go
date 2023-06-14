package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

var _ sdk.Msg = &MsgWithdrawDeposit{}

func NewMsgWithdrawDeposit(depositor string, toAddress string, poolID uint64, depositID uint64) *MsgWithdrawDeposit {
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

	// Verify that a valid bech32 address is passed regardless of its prefix
	// because the toAddress can be targeting the local zone or the pool remote zone
	if _, _, err := bech32.DecodeAndConvert(msg.GetToAddress()); err != nil {
		return ErrInvalidDestinationAddress
	}

	if msg.GetPoolId() == UnknownID || msg.GetDepositId() == UnknownID {
		return ErrInvalidID
	}

	return nil
}
