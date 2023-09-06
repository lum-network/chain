package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var _ sdk.Msg = &MsgGenerateSeed{}

func NewMsgGenerateSeed(requesterAddress string) *MsgGenerateSeed {
	return &MsgGenerateSeed{
		RequesterAddress: requesterAddress,
	}
}

func (msg MsgGenerateSeed) Route() string {
	return RouterKey
}

func (msg MsgGenerateSeed) Type() string {
	return "GenerateSeed"
}

func (msg *MsgGenerateSeed) GetSigners() []sdk.AccAddress {
	depositor := sdk.MustAccAddressFromBech32(msg.GetRequesterAddress())
	return []sdk.AccAddress{depositor}
}

func (msg *MsgGenerateSeed) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgGenerateSeed) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	if _, err := sdk.AccAddressFromBech32(msg.GetRequesterAddress()); err != nil {
		return ErrInvalidRequesterAddress
	}

	return nil
}
