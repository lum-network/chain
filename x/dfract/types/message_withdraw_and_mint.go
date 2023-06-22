package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgWithdrawAndMint{}

func NewMsgWithdrawAndMint(address string, microMintRate int64) *MsgWithdrawAndMint {
	return &MsgWithdrawAndMint{
		Address:       address,
		MicroMintRate: microMintRate,
	}
}

func (msg MsgWithdrawAndMint) Route() string {
	return RouterKey
}

func (msg MsgWithdrawAndMint) Type() string {
	return "WithdrawAndMint"
}

func (msg *MsgWithdrawAndMint) GetSigners() []sdk.AccAddress {
	depositor := sdk.MustAccAddressFromBech32(msg.GetAddress())
	return []sdk.AccAddress{depositor}
}

func (msg *MsgWithdrawAndMint) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgWithdrawAndMint) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	_, err := sdk.AccAddressFromBech32(msg.GetAddress())
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidSignerAddress, err.Error())
	}

	// Check the micro min rate
	if msg.MicroMintRate < 0 {
		return ErrEmptyMicroMintRate
	}

	return nil
}
