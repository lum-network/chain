package types

const (
	// ModuleName defines the module name
	ModuleName = "interchainquery"

	// ModuleVersion defines the module version
	ModuleVersion = 1

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// prefix bytes for the icqueries persistent store
const (
	prefixData  = iota + 1
	prefixQuery = iota + 1
)

// keys for proof queries to various stores, note: there's an implicit assumption here that
// the stores on the counterparty chain are prefixed with the standard cosmos-sdk module names
// this might not be true for all IBC chains, and is something we should verify before onboarding a new chain

const (
	BANK_STORE_QUERY_WITH_PROOF = "store/bank/key"
)

var (
	KeyPrefixData  = []byte{prefixData}
	KeyPrefixQuery = []byte{prefixQuery}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
