package beam_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	apptypes "github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"

	"github.com/lum-network/chain/utils"
	"github.com/lum-network/chain/x/beam"
	"github.com/lum-network/chain/x/beam/keeper"
	"github.com/lum-network/chain/x/beam/types"
)

type ABCITestSuite struct {
	suite.Suite

	ctx         sdk.Context
	queryClient types.QueryClient
	app         *apptypes.App
	addrs       []sdk.AccAddress
}

// SetupTest Create our testing app, and make sure everything is correctly usable
func (suite *ABCITestSuite) SetupTest() {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(*app.BeamKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
	suite.addrs = apptesting.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))
}

func (suite *ABCITestSuite) TestTickBeamAutoClose() {
	app := suite.app
	ctx := suite.ctx

	// Simulate the begin block call
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Make sure the open beam queue is empty
	openQueue := app.BeamKeeper.OpenBeamsQueueIterator(ctx)
	suite.Require().False(openQueue.Valid())
	openQueue.Close()

	// Create the required accounts
	creator := suite.addrs[0]
	claimer := suite.addrs[1]
	suite.Require().NotEqual(creator.String(), claimer.String())

	// Create a random token as claim secret
	claimSecret := utils.GenerateSecureToken(4)

	// Create a beam
	msgVal := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
	msg := types.NewMsgOpenBeam(
		utils.GenerateSecureToken(12),
		creator.String(),
		"",
		&msgVal,
		hex.EncodeToString(utils.GenerateHashFromString(claimSecret)),
		types.BEAM_SCHEMA_REVIEW,
		nil,
		10,
		0,
	)
	_, err := app.BeamKeeper.OpenBeam(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().True(app.BeamKeeper.HasBeam(ctx, msg.GetId()))

	// Make sure the open beam queue is now valid
	openQueue = app.BeamKeeper.OpenBeamsByBlockQueueIterator(ctx)
	suite.Require().True(openQueue.Valid())
	openQueue.Close()

	// Simulate a call with a faked block header at a height of 10
	newHeader := ctx.BlockHeader()
	newHeader.Height = 10
	ctx = ctx.WithBlockHeader(newHeader)

	// Make sure the closed beams queue is invalid
	closedQueue := app.BeamKeeper.ClosedBeamsQueueIterator(ctx)
	suite.Require().False(closedQueue.Valid())
	closedQueue.Close()

	// Call the end blocker function to trigger beam expiration
	beam.EndBlocker(ctx, *app.BeamKeeper)

	// Make sure the open beam queue is now invalid
	openQueue = app.BeamKeeper.OpenBeamsQueueIterator(ctx)
	suite.Require().False(openQueue.Valid())
	openQueue.Close()

	closedQueue = app.BeamKeeper.ClosedBeamsQueueIterator(ctx)
	suite.Require().True(closedQueue.Valid())
	closedQueue.Close()
}

func TestABCISuite(t *testing.T) {
	suite.Run(t, new(ABCITestSuite))
}
