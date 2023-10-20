package types

import "fmt"

const (
	// ModuleName defines the module name
	ModuleName = "beam"

	// ModuleVersion defines the current module version
	ModuleVersion = 2

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_capability"

	// MemStoreQueueSeparator Used to separate content in the by-height-open-beams queue
	MemStoreQueueSeparator = ","
)

var (
	BeamsPrefix            = []byte{0x01}
	OpenBeamsQueuePrefix   = []byte{0x02}
	ClosedBeamsQueuePrefix = []byte{0x03}

	OpenBeamsByBlockQueuePrefix = []byte{0x04}
)

func StringKeyToBytes(key string) []byte {
	return []byte(key)
}

func BytesKeyToString(key []byte) string {
	return string(key)
}

func IntKeyToString(key int) string {
	return fmt.Sprintf("%d", key)
}

func GetBeamKey(beamID string) []byte {
	return append(BeamsPrefix, StringKeyToBytes(beamID)...)
}

func GetOpenBeamQueueKey(beamID string) []byte {
	return append(OpenBeamsQueuePrefix, StringKeyToBytes(beamID)...)
}

func GetClosedBeamQueueKey(beamID string) []byte {
	return append(ClosedBeamsQueuePrefix, StringKeyToBytes(beamID)...)
}

func GetOpenBeamsByBlockQueueKey(height int) []byte {
	return append(OpenBeamsByBlockQueuePrefix, IntKeyToString(height)...)
}

func SplitBeamKey(key []byte) []byte {
	return key[1:]
}
