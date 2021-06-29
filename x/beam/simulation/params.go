package simulation

import (
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"math/rand"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{}
}
