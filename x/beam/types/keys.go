package types

const (
	// ModuleName defines the module name
	ModuleName = "beam"

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
	)

var (
	PrefixBeam = []byte{0x01}

	delimiter = []byte("/")
)

func KeyBeam(beamID string) []byte {
	key := append(PrefixBeam, delimiter...)
	if len(beamID) > 0 {
		key = append(key, []byte(beamID)...)
		key = append(key, delimiter...)
	}
	return key
}
