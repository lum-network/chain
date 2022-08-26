package beam_test

import (
	apptypes "github.com/lum-network/chain/app"
	"github.com/lum-network/chain/x/beam/types"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.InitChain(
		abcitypes.RequestInitChain{
			AppStateBytes: []byte("{}"),
			ChainId:       "lumnetwork-testnet",
		},
	)

	acc := app.AccountKeeper.GetModuleAccount(ctx, types.ModuleName)
	require.NotNil(t, acc)
}
