package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/lum-network/chain/x/beam/types"
)

func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.BeamsPrefix):
			var beamA, beamB types.Beam
			cdc.MustUnmarshal(kvA.Value, &beamA)
			cdc.MustUnmarshal(kvB.Value, &beamB)
			return fmt.Sprintf("%v\n%v", beamA, beamB)

		default:
			panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
		}
	}
}
