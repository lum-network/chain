package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var _ sdk.Msg = &MsgRestoreInterchainAccounts{}

func NewMsgRestoreInterchainAccounts(restorerAddress string, poolID uint64) *MsgRestoreInterchainAccounts {
	return &MsgRestoreInterchainAccounts{
		RestorerAddress: restorerAddress,
		PoolId:          poolID,
	}
}

func (msg MsgRestoreInterchainAccounts) Route() string {
	return RouterKey
}

func (msg MsgRestoreInterchainAccounts) Type() string {
	return "RestoreInterchainAccounts"
}

func (msg *MsgRestoreInterchainAccounts) GetSigners() []sdk.AccAddress {
	depositor := sdk.MustAccAddressFromBech32(msg.GetRestorerAddress())
	return []sdk.AccAddress{depositor}
}

func (msg *MsgRestoreInterchainAccounts) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRestoreInterchainAccounts) ValidateBasic() error {
	// Ensure the address is correct and that we are able to acquire it
	if _, err := sdk.AccAddressFromBech32(msg.GetRestorerAddress()); err != nil {
		return ErrInvalidRestorerAddress
	}

	if msg.GetPoolId() == UnknownID {
		return ErrInvalidID
	}

	return nil
}
