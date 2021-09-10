package keeper_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/simapp"
	"github.com/lum-network/chain/x/beam/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	queryClient types.QueryClient
	app         *simapp.SimApp
	addrs       []sdk.AccAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.BeamKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
	suite.addrs = simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))
}

func (suite *KeeperTestSuite) TestOpenNewBeam() {
	app := suite.app
	ctx := suite.ctx

	// Create the original owner
	owner := suite.addrs[0]

	// Fund him


	// Create value and the linked message
	msgVal := sdk.NewCoin("lum", sdk.NewInt(100))
	msg := types.NewMsgOpenBeam(
		types.GenerateSecureToken(12),
		owner.String(),
		owner.String(),
		&msgVal,
		types.GenerateSecureToken(4),
		types.BEAM_SCHEMA_REVIEW,
		nil,
		0,
		0,
	)

	// Open the beam
	err := app.BeamKeeper.OpenBeam(ctx, *msg)
	require.NoError(suite.T(), err)
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
