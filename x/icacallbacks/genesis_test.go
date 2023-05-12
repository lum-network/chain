package icacallbacks_test

import (
	apptypes "github.com/lum-network/chain/app"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v6/testutil/nullify"
	"github.com/lum-network/chain/x/icacallbacks"
	"github.com/lum-network/chain/x/icacallbacks/types"
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

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	require.ElementsMatch(t, genesisState.CallbackDataList, got.CallbackDataList)
}
