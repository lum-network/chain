package beam_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	apptypes "github.com/lum-network/chain/app"

	"github.com/lum-network/chain/x/beam/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	acc := app.AccountKeeper.GetModuleAccount(ctx, types.ModuleName)
	require.NotNil(t, acc)
}
