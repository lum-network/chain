package beam_test

import (
	apptypes "github.com/lum-network/chain/app"
	"github.com/lum-network/chain/x/beam/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	acc := app.AccountKeeper.GetModuleAccount(ctx, types.ModuleName)
	require.NotNil(t, acc)
}
