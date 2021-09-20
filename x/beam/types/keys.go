package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
	BeamsIndex            = []byte{0x01}
	OpenBeamsQueueIndex   = []byte{0x02}
	ClosedBeamsQueueIndex = []byte{0x03}
)

func GetBeamIDBytes(beamID string) []byte {
	return []byte(beamID)
}

func GetBeamIDFromBytes(beamID []byte) string {
	return string(beamID)
}

func GetBeamKey(beamID string) []byte {
	return append(BeamsIndex, GetBeamIDBytes(beamID)...)
}

func GetOpenBeamQueueKey(beamID string, height int64) []byte {
	return append(append(OpenBeamsQueueIndex, sdk.Uint64ToBigEndian(uint64(height))...), GetBeamIDBytes(beamID)...)
}

func GetClosedBeamQueueKey(beamID string, height int64) []byte {
	return append(append(ClosedBeamsQueueIndex, sdk.Uint64ToBigEndian(uint64(height))...), GetBeamIDBytes(beamID)...)
}
