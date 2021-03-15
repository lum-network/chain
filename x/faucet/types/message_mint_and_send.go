package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"time"
)

var _ sdk.Msg = &MsgMintAndSend{}

func NewMsgMintAndSend(minter string, mintTime time.Time) *MsgMintAndSend {
	return &MsgMintAndSend{
		Minter:   minter,
		MintTime: mintTime,
	}
}

func (msg MsgMintAndSend) Route() string {
	return RouterKey
}

func (msg MsgMintAndSend) Type() string {
	return "MintAndSend"
}

func (msg *MsgMintAndSend) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Minter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgMintAndSend) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMintAndSend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Minter)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid minter address (%s)", err)
	}

	if msg.GetMintTime().After(time.Now()) {
		return ErrMintDateFuture
	}

	return nil
}
