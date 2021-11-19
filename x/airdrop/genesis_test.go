package airdrop_test

import (
	"testing"

	keepertest "github.com/lum-network/chain/testutil/keeper"
	"github.com/lum-network/chain/x/airdrop"
	"github.com/lum-network/chain/x/airdrop/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.AirdropKeeper(t)
	airdrop.InitGenesis(ctx, *k, genesisState)
	got := airdrop.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	// this line is used by starport scaffolding # genesis/test/assert
}
