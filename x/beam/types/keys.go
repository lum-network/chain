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
	PrefixBeamQueue       = []byte{0x01}
	PrefixOpenBeamQueue   = []byte{0x02}
	PrefixClosedBeamQueue = []byte{0x03}

	delimiter = []byte("/")
)

// BeamKey Get a beam ID as bytes array from string
func BeamKey(beamID string) []byte {
	key := append(PrefixBeamQueue, delimiter...)
	if len(beamID) > 0 {
		key = append(key, []byte(beamID)...)
		key = append(key, delimiter...)
	}
	return key
}

// OpenBeamQueueKey Return a key constructed for the active beams queue
func OpenBeamQueueKey(beamID string) []byte {
	key := append(PrefixOpenBeamQueue, delimiter...)
	if len(beamID) > 0 {
		key = append(key, []byte(beamID)...)
		key = append(key, delimiter...)
	}
	return key
}

// ClosedBeamQueueKey Return a key constructed for the closed beams queue
func ClosedBeamQueueKey(beamID string) []byte {
	key := append(PrefixClosedBeamQueue, delimiter...)
	if len(beamID) > 0 {
		key = append(key, []byte(beamID)...)
		key = append(key, delimiter...)
	}
	return key
}
