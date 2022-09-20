package types

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

	// ParamsKey defines the store key for claim module parameters
	ParamsKey = "dfract_params"
)

var (
	WaitingProposalDepositsPrefix = []byte{0x01}
	WaitingMintDepositsPrefix     = []byte{0x02}
	MintedDepositsPrefix          = []byte{0x03}
)

func GetWaitingProposalDepositsKey(depositorAddress string) []byte {
	return append(WaitingProposalDepositsPrefix, []byte(depositorAddress)...)
}

func GetWaitingMintDepositsKey(depositorAddress string) []byte {
	return append(WaitingMintDepositsPrefix, []byte(depositorAddress)...)
}

func GetMintedDepositsKey(depositorAddress string) []byte {
	return append(MintedDepositsPrefix, []byte(depositorAddress)...)
}
