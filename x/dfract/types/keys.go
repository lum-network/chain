package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "dfract"

	// ModuleVersion defines the current module version
	ModuleVersion = 1

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_capability"

	// MintDenom defines the denom of the minted token
	MintDenom = "udfr"
)

var (
	DepositsPendingWithdrawalPrefix = []byte{0x01}
	DepositsPendingMintPrefix       = []byte{0x02}
	DepositsMintedPrefix            = []byte{0x03}
	StakedTokenPrefix               = []byte{0x04}
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

func GetStakedTokenKey(senderAddress sdk.AccAddress) []byte {
	return append(StakedTokenPrefix, senderAddress.Bytes()...)
}
