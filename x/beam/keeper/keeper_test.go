package keeper_test

import (
	"encoding/hex"
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

// SetupTest Create our testing app, and make sure everything is correctly usable
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

// TestClaimNewBeam Try to create a beam and claim it using another account
func (suite *KeeperTestSuite) TestClaimOpenBeam() {
	app := suite.app
	ctx := suite.ctx

	// Create the required accounts
	creator := suite.addrs[0]
	claimer := suite.addrs[1]
	require.NotEqual(suite.T(), creator.String(), claimer.String())

	// We store the initial claimer funds
	claimerFunds := app.BankKeeper.GetBalance(ctx, claimer, "stake")

	// Create a random token as claim secret
	claimSecret := types.GenerateSecureToken(4)

	// Create a beam with 100 tokens
	msgVal := sdk.NewCoin("stake", sdk.NewInt(100))
	msg := types.NewMsgOpenBeam(
		types.GenerateSecureToken(12),
		creator.String(),
		"",
		&msgVal,
		hex.EncodeToString(types.GenerateHashFromString(claimSecret)),
		types.BEAM_SCHEMA_REVIEW,
		nil,
		0,
		0,
	)
	err := app.BeamKeeper.OpenBeam(ctx, *msg)
	require.NoError(suite.T(), err)
	require.True(suite.T(), app.BeamKeeper.HasBeam(ctx, msg.GetId()))

	// Try to claim unknown beam
	err = app.BeamKeeper.ClaimBeam(ctx, *types.NewMsgClaimBeam(
		claimer.String(),
		"qksjbdnqsjhbdjsq122112",
		"qskjbdq",
	))
	require.Error(suite.T(), err)

	// Try to claim using bad secret
	err = app.BeamKeeper.ClaimBeam(ctx, *types.NewMsgClaimBeam(
		claimer.String(),
		msg.GetId(),
		"test_1234",
	))
	require.Error(suite.T(), err)

	// Claim and make sure funds weren't released
	err = app.BeamKeeper.ClaimBeam(ctx, *types.NewMsgClaimBeam(
		claimer.String(),
		msg.GetId(),
		claimSecret,
	))
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), claimerFunds, app.BankKeeper.GetBalance(ctx, claimer, "stake"))

	// Acquire the beam and make sure props were updated
	beam, err := app.BeamKeeper.GetBeam(ctx, msg.GetId())
	require.NoError(suite.T(), err)
	require.True(suite.T(), beam.GetClaimed())
	require.Equal(suite.T(), beam.GetClaimAddress(), claimer.String())
}

// TestUnknownBeam Make sure we cannot get an unknown beam
func (suite *KeeperTestSuite) TestUnknownBeam() {
	app := suite.app
	ctx := suite.ctx

	_, err := app.BeamKeeper.GetBeam(ctx, "kjqsdjkqsd")
	require.Error(suite.T(), err)
}

// TestOpenNewBeam Try to create a new beam and make sure the stored entity matches the original one
func (suite *KeeperTestSuite) TestOpenNewBeam() {
	app := suite.app
	ctx := suite.ctx

	// Create the original owner
	owner := suite.addrs[0]

	// Create value and the linked message
	msgVal := sdk.NewCoin("stake", sdk.NewInt(100))
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

	// Open the beam and make sure there was no error
	err := app.BeamKeeper.OpenBeam(ctx, *msg)
	require.NoError(suite.T(), err)
	require.True(suite.T(), app.BeamKeeper.HasBeam(ctx, msg.GetId()))

	// Ask for the list of beams and make sure we have it
	beams := app.BeamKeeper.ListBeams(ctx)
	require.GreaterOrEqual(suite.T(), len(beams), 1)

	// Make sure we can get it
	beam, err := app.BeamKeeper.GetBeam(ctx, msg.GetId())
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), beam)

	require.Equal(suite.T(), msg.GetId(), beam.GetId())
	require.Equal(suite.T(), msg.GetCreatorAddress(), beam.GetCreatorAddress())
	require.Equal(suite.T(), msg.GetClaimAddress(), beam.GetClaimAddress())
	require.Equal(suite.T(), msg.GetSchema(), beam.GetSchema())
	require.Equal(suite.T(), msg.GetData(), beam.GetData())
	require.Equal(suite.T(), msg.GetClosesAtBlock(), beam.GetClosesAtBlock())
	require.Equal(suite.T(), msg.GetClaimExpiresAtBlock(), beam.GetClaimExpiresAtBlock())
	require.Equal(suite.T(), beam.GetStatus(), types.BeamState_StateOpen)
}

// TestKeeperSuite Main entry point for the testing suite
func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
