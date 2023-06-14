package types

import (
	"testing"

	"github.com/lum-network/chain/utils"
	"github.com/stretchr/testify/require"
)

func TestBeamKey(t *testing.T) {
	beamID := utils.GenerateSecureToken(8)
	key := GetBeamKey(beamID)
	require.NotEqual(t, beamID, string(key))
	parsedBeamID := BytesKeyToString(SplitBeamKey(key))
	require.Equal(t, beamID, parsedBeamID)
}

func TestOpenBeamsQueueKey(t *testing.T) {
	beamID := utils.GenerateSecureToken(8)
	key := GetOpenBeamQueueKey(beamID)
	require.NotEqual(t, beamID, string(key))
	parsedBeamID := BytesKeyToString(SplitBeamKey(key))
	require.Equal(t, beamID, parsedBeamID)
}

func TestClosedBeamsQueueKey(t *testing.T) {
	beamID := utils.GenerateSecureToken(8)
	key := GetClosedBeamQueueKey(beamID)
	require.NotEqual(t, beamID, string(key))
	parsedBeamID := BytesKeyToString(SplitBeamKey(key))
	require.Equal(t, beamID, parsedBeamID)
}
