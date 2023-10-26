package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "dfract"

	// ModuleVersion defines the current module version
	ModuleVersion = 2

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MintDenom defines the denom of the minted token
	MintDenom = "udfr"
)

var (
	DepositsPendingWithdrawalPrefix = []byte{0x01}
	DepositsPendingMintPrefix       = []byte{0x02}
	DepositsMintedPrefix            = []byte{0x03}
	ParamsPrefix                    = []byte{0x10}
)

func GetDepositsPendingWithdrawalKey(depositorAddress sdk.AccAddress) []byte {
	return append(DepositsPendingWithdrawalPrefix, depositorAddress.Bytes()...)
}

func GetDepositsPendingMintKey(depositorAddress sdk.AccAddress) []byte {
	return append(DepositsPendingMintPrefix, depositorAddress.Bytes()...)
}

func GetDepositsMintedKey(depositorAddress sdk.AccAddress) []byte {
	return append(DepositsMintedPrefix, depositorAddress.Bytes()...)
}
