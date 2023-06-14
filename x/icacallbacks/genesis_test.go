package icacallbacks_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	apptypes "github.com/lum-network/chain/app"

	"github.com/stretchr/testify/require"

	"github.com/lum-network/chain/x/icacallbacks"
	"github.com/lum-network/chain/x/icacallbacks/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		CallbackDataList: []types.CallbackData{
			{
				CallbackKey: "0",
			},
			{
				CallbackKey: "1",
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	icacallbacks.InitGenesis(ctx, *app.ICACallbacksKeeper, genesisState)
	got := icacallbacks.ExportGenesis(ctx, *app.ICACallbacksKeeper)
	require.NotNil(t, got)

	// TODO: upgrade fix
	// nullify.Fill(&genesisState)
	// nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	require.ElementsMatch(t, genesisState.CallbackDataList, got.CallbackDataList)
}
