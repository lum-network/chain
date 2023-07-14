package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx   sdk.Context
	app   *app.App
	addrs []sdk.AccAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	app := app.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Setup the default application
	suite.app = app
	suite.ctx = ctx

	// Setup the module params
	app.DFractKeeper.SetParams(ctx, dfracttypes.DefaultParams())

	params := app.DFractKeeper.GetParams(ctx)

	// Iterate over the array of deposit denoms
	for _, denom := range params.DepositDenoms {
		suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 6, sdk.NewInt(300000000), denom)
	}

}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
