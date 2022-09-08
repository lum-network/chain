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
	DepositsPrefix                     = []byte{0x01}
	WaitingProposalDepositsQueuePrefix = []byte{0x02}
	WaitingMintDepositsQueuePrefix     = []byte{0x03}
	MintedDepositsQueuePrefix          = []byte{0x04}
)

func GetDepositKey(depositId string) []byte {
	return append(DepositsPrefix, []byte(depositId)...)
}

func GetWaitingProposalDepositsQueueKey(depositId string) []byte {
	return append(WaitingProposalDepositsQueuePrefix, []byte(depositId)...)
}

func GetWaitingMintDepositsQueueKey(depositId string) []byte {
	return append(WaitingMintDepositsQueuePrefix, []byte(depositId)...)
}

func GetMintedDepositsQueueKey(depositId string) []byte {
	return append(MintedDepositsQueuePrefix, []byte(depositId)...)
}
