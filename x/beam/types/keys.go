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
	BeamsPrefix            = []byte{0x01}
	OpenBeamsQueuePrefix   = []byte{0x02}
	ClosedBeamsQueuePrefix = []byte{0x03}
)

func GetBeamIDBytes(beamID string) []byte {
	return []byte(beamID)
}

func GetBeamIDFromBytes(beamID []byte) string {
	return string(beamID)
}

func GetBeamKey(beamID string) []byte {
	return append(BeamsPrefix, GetBeamIDBytes(beamID)...)
}

func GetOpenBeamQueueKey(beamID string) []byte {
	return append(OpenBeamsQueuePrefix, GetBeamIDBytes(beamID)...)
}

func GetClosedBeamQueueKey(beamID string) []byte {
	return append(ClosedBeamsQueuePrefix, GetBeamIDBytes(beamID)...)
}

func SplitBeamKey(key []byte) []byte {
	return key[1:]
}

func SplitOpenBeamQueueKey(key []byte) (beamID []byte) {
	return key[1:]
}

func SplitClosedBeamQueueKey(key []byte) (beamID []byte) {
	return key[1:]
}
