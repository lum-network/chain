package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgClaimPrize{}

func NewMsgMsgClaimPrize(winnerAddress string, poolID, drawID, prizeID uint64) *MsgClaimPrize {
	return &MsgClaimPrize{
		PoolId:        poolID,
		DrawId:        drawID,
		PrizeId:       prizeID,
		WinnerAddress: winnerAddress,
	}
}

func (msg MsgClaimPrize) Route() string {
	return RouterKey
}

func (msg MsgClaimPrize) Type() string {
	return "ClaimPrize"
}

func (msg *MsgClaimPrize) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.GetWinnerAddress())
	return []sdk.AccAddress{addr}
}

func (msg *MsgClaimPrize) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClaimPrize) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	_, err := sdk.AccAddressFromBech32(msg.GetWinnerAddress())
	if err != nil {
		return ErrInvalidWinnerAddress
	}

	if msg.GetPoolId() == UnknownID || msg.GetDrawId() == UnknownID || msg.GetPrizeId() == UnknownID {
		return ErrInvalidID
	}

	return nil
}
