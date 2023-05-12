package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgDeposit{}

func NewMsgDeposit(depositor string, amount sdk.Coin, poolID uint64) *MsgDeposit {
	return &MsgDeposit{
		PoolId:           poolID,
		Amount:           amount,
		DepositorAddress: depositor,
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
	if _, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress()); err != nil {
		return ErrInvalidDepositorAddress
	}

	// ensure denom is valid
	if err := sdk.ValidateDenom(msg.Amount.Denom); err != nil {
		return ErrInvalidDepositDenom
	}

	// If we have an amount, make sure it is not negative nor zero
	if msg.Amount.IsNil() || msg.Amount.Amount.LT(sdk.NewInt(MinAcceptableDepositAmount)) {
		return ErrInvalidDepositAmount
	}

	if msg.GetWinnerAddress() != "" {
		if _, err := sdk.AccAddressFromBech32(msg.GetWinnerAddress()); err != nil {
			return ErrInvalidWinnerAddress
		}
	}

	if msg.WinnerAddress != "" && msg.WinnerAddress != msg.DepositorAddress && msg.IsSponsor {
		return ErrInvalidSponsorWinnerCombo
	}

	if msg.GetPoolId() == UnknownID {
		return ErrInvalidID
	}

	return nil
}
