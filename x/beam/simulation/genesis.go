package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/lum-network/chain/x/beam/types"
)

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	mintGenesis := types.DefaultGenesis()

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}
