package keeper_test

import (
	"encoding/hex"
	"github.com/lum-network/chain/utils"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/app"
	apptypes "github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/x/beam/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	queryClient types.QueryClient
	app         *app.App
	addrs       []sdk.AccAddress
}

// SetupTest Create our testing app, and make sure everything is correctly usable
func (suite *KeeperTestSuite) SetupTest() {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.BeamKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
	suite.addrs = apptesting.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))
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
	claimerFunds := app.BankKeeper.GetBalance(ctx, claimer, apptypes.CoinBondDenom)

	// Create a random token as claim secret
	claimSecret := utils.GenerateSecureToken(4)

	// Create a beam with 100 tokens
	msgVal := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
	msg := types.NewMsgOpenBeam(
		utils.GenerateSecureToken(12),
		creator.String(),
		"",
		&msgVal,
		hex.EncodeToString(utils.GenerateHashFromString(claimSecret)),
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
	require.Equal(suite.T(), claimerFunds, app.BankKeeper.GetBalance(ctx, claimer, apptypes.CoinBondDenom))

	// Acquire the beam and make sure props were updated
	beam, err := app.BeamKeeper.GetBeam(ctx, msg.GetId())
	require.NoError(suite.T(), err)
	require.True(suite.T(), beam.GetClaimed())
	require.Equal(suite.T(), beam.GetClaimAddress(), claimer.String())
	require.Equal(suite.T(), beam.GetStatus(), types.BeamState_StateOpen)
}

// TestClaimClosedBeam Test to claim a closed beam and make sure funds were transfered
func (suite *KeeperTestSuite) TestClaimClosedBeam() {
	app := suite.app
	ctx := suite.ctx

	// Create the required accounts
	creator := suite.addrs[0]
	claimer := suite.addrs[1]
	require.NotEqual(suite.T(), creator.String(), claimer.String())

	// We store the initial claimer funds
	claimerFunds := app.BankKeeper.GetBalance(ctx, claimer, apptypes.CoinBondDenom)

	// Create a random token as claim secret
	claimSecret := utils.GenerateSecureToken(4)

	// Create a beam with 100 tokens
	msgVal := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
	msg := types.NewMsgOpenBeam(
		utils.GenerateSecureToken(12),
		creator.String(),
		"",
		&msgVal,
		hex.EncodeToString(utils.GenerateHashFromString(claimSecret)),
		types.BEAM_SCHEMA_REVIEW,
		nil,
		0,
		0,
	)
	err := app.BeamKeeper.OpenBeam(ctx, *msg)
	require.NoError(suite.T(), err)
	require.True(suite.T(), app.BeamKeeper.HasBeam(ctx, msg.GetId()))

	// Close the beam
	msgClose := types.NewMsgUpdateBeam(
		creator.String(),
		msg.GetId(),
		nil,
		types.BeamState_StateClosed,
		nil,
		"",
		false,
		0,
		0,
	)
	err = app.BeamKeeper.UpdateBeam(ctx, *msgClose)
	require.NoError(suite.T(), err)

	// If we try to update again, should pass
	err = app.BeamKeeper.UpdateBeam(ctx, *msgClose)
	require.Error(suite.T(), err)

	// Get the beam and ensure properties
	beam, err := app.BeamKeeper.GetBeam(ctx, msg.GetId())
	require.NoError(suite.T(), err)
	require.False(suite.T(), beam.GetClaimed())
	require.False(suite.T(), beam.GetHideContent())
	require.Zero(suite.T(), beam.GetClosedAt())
	require.Zero(suite.T(), beam.GetClosesAtBlock())

	// Claim the beam
	err = app.BeamKeeper.ClaimBeam(ctx, *types.NewMsgClaimBeam(
		claimer.String(),
		msg.GetId(),
		claimSecret,
	))
	require.NoError(suite.T(), err)

	// Try to claim again
	err = app.BeamKeeper.ClaimBeam(ctx, *types.NewMsgClaimBeam(
		claimer.String(),
		msg.GetId(),
		claimSecret,
	))
	require.Error(suite.T(), err)

	// Now the funds should've been transfered
	require.Equal(suite.T(), claimerFunds.Add(beam.GetAmount()), app.BankKeeper.GetBalance(ctx, claimer, apptypes.CoinBondDenom))
}

// Test to cancel a beam and make sure funds were returned to the sender
func (suite *KeeperTestSuite) TestCancelBeam() {
	app := suite.app
	ctx := suite.ctx

	// Create the required accounts
	creator := suite.addrs[0]
	claimer := suite.addrs[1]
	require.NotEqual(suite.T(), creator.String(), claimer.String())

	// Create a random token as claim secret
	claimSecret := utils.GenerateSecureToken(4)

	// We store the initial claimer funds
	creatorFunds := app.BankKeeper.GetBalance(ctx, creator, apptypes.CoinBondDenom)

	// Create a beam with 100 tokens
	msgVal := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
	msg := types.NewMsgOpenBeam(
		utils.GenerateSecureToken(12),
		creator.String(),
		"",
		&msgVal,
		hex.EncodeToString(utils.GenerateHashFromString(claimSecret)),
		types.BEAM_SCHEMA_REVIEW,
		nil,
		0,
		0,
	)
	err := app.BeamKeeper.OpenBeam(ctx, *msg)
	require.NoError(suite.T(), err)
	require.True(suite.T(), app.BeamKeeper.HasBeam(ctx, msg.GetId()))

	// Make sure the creator was debited
	require.Equal(suite.T(), creatorFunds.SubAmount(sdk.NewInt(100)), app.BankKeeper.GetBalance(ctx, creator, apptypes.CoinBondDenom))

	// Cancel the beam
	msgCancel := types.NewMsgUpdateBeam(
		creator.String(),
		msg.GetId(),
		nil,
		types.BeamState_StateCanceled,
		nil,
		"Test Cancel",
		true,
		0,
		0,
	)
	err = app.BeamKeeper.UpdateBeam(ctx, *msgCancel)
	require.NoError(suite.T(), err)
	beam, err := app.BeamKeeper.GetBeam(ctx, msg.GetId())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), beam.GetStatus(), types.BeamState_StateCanceled)
	require.Equal(suite.T(), beam.GetCancelReason(), msgCancel.GetCancelReason())

	// Make sure the creator was credited back
	require.Equal(suite.T(), creatorFunds, app.BankKeeper.GetBalance(ctx, creator, apptypes.CoinBondDenom))

	// Try to cancel again and make sure it cannot happen
	msgCancel = types.NewMsgUpdateBeam(
		creator.String(),
		msg.GetId(),
		nil,
		types.BeamState_StateCanceled,
		nil,
		"Test Cancel",
		true,
		0,
		0,
	)
	err = app.BeamKeeper.UpdateBeam(ctx, *msgCancel)
	require.Error(suite.T(), err)

	// Make sure the beam is now present in the closed beams queue
	closedIterator := app.BeamKeeper.ClosedBeamsQueueIterator(ctx)
	require.NoError(suite.T(), closedIterator.Error())
	require.True(suite.T(), closedIterator.Valid())
	require.Equal(suite.T(), beam.GetId(), string(closedIterator.Value()))
	closedIterator.Close()
}

func (suite *KeeperTestSuite) TestOpenCloseIterators() {
	app := suite.app
	ctx := suite.ctx

	// Create the required accounts
	creator := suite.addrs[0]
	claimer := suite.addrs[1]
	require.NotEqual(suite.T(), creator.String(), claimer.String())

	// Create a random token as claim secret
	claimSecret := utils.GenerateSecureToken(4)

	// Create a beam with 100 tokens
	msgVal := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
	msg := types.NewMsgOpenBeam(
		utils.GenerateSecureToken(12),
		creator.String(),
		"",
		&msgVal,
		hex.EncodeToString(utils.GenerateHashFromString(claimSecret)),
		types.BEAM_SCHEMA_REVIEW,
		nil,
		0,
		0,
	)
	err := app.BeamKeeper.OpenBeam(ctx, *msg)
	require.NoError(suite.T(), err)
	require.True(suite.T(), app.BeamKeeper.HasBeam(ctx, msg.GetId()))

	// The beam should not be present since we disabled the auto close with 0 block height
	openQueue := app.BeamKeeper.OpenBeamsByBlockQueueIterator(ctx)
	require.False(suite.T(), openQueue.Valid())
	openQueue.Close()

	// But not on closed queue
	closedIterator := app.BeamKeeper.ClosedBeamsQueueIterator(ctx)
	require.Error(suite.T(), closedIterator.Error())
	require.False(suite.T(), closedIterator.Valid())
	closedIterator.Close()

	// Close the beam
	msgCancel := types.NewMsgUpdateBeam(
		creator.String(),
		msg.GetId(),
		nil,
		types.BeamState_StateCanceled,
		nil,
		"Test Cancel",
		true,
		0,
		0,
	)
	err = app.BeamKeeper.UpdateBeam(ctx, *msgCancel)
	require.NoError(suite.T(), err)

	// We should not have it inside open beams queue
	openQueue = app.BeamKeeper.OpenBeamsByBlockQueueIterator(ctx)
	require.False(suite.T(), openQueue.Valid())
	openQueue.Close()

	// But in the closed queue
	closedIterator = app.BeamKeeper.ClosedBeamsQueueIterator(ctx)
	require.NoError(suite.T(), closedIterator.Error())
	require.True(suite.T(), closedIterator.Valid())
	closedIterator.Close()

	// Create another beam
	msgVal = sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
	msg = types.NewMsgOpenBeam(
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
	err = app.BeamKeeper.OpenBeam(ctx, *msg)
	require.NoError(suite.T(), err)
	require.True(suite.T(), app.BeamKeeper.HasBeam(ctx, msg.GetId()))

	// Is the beam present in the open queue
	openQueue = app.BeamKeeper.OpenBeamsByBlockQueueIterator(ctx)
	require.True(suite.T(), openQueue.Valid())
	openQueue.Close()

	// But not on closed queue
	closedIterator = app.BeamKeeper.ClosedBeamsQueueIterator(ctx)
	require.NoError(suite.T(), closedIterator.Error())
	require.True(suite.T(), closedIterator.Valid())
	closedIterator.Close()
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
	msgVal := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
	msg := types.NewMsgOpenBeam(
		utils.GenerateSecureToken(12),
		owner.String(),
		owner.String(),
		&msgVal,
		utils.GenerateSecureToken(4),
		types.BEAM_SCHEMA_REVIEW,
		nil,
		120,
		0,
	)

	// Open the beam and make sure there was no error
	err := app.BeamKeeper.OpenBeam(ctx, *msg)
	require.NoError(suite.T(), err)
	require.True(suite.T(), app.BeamKeeper.HasBeam(ctx, msg.GetId()))

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

	// Make sure the beam is now present in the open beams queue
	openQueue := app.BeamKeeper.OpenBeamsByBlockQueueIterator(ctx)
	require.True(suite.T(), openQueue.Valid())
	openQueue.Close()
}

func (suite *KeeperTestSuite) TestOpenAutoCloseBeam() {

}

// TestFetchBeams Open a new beam and try to fetch it through the list
func (suite *KeeperTestSuite) TestFetchBeams() {
	app := suite.app
	ctx := suite.ctx

	// Create the original owner
	owner := suite.addrs[0]

	// Create value and the linked message
	msgVal := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
	msg := types.NewMsgOpenBeam(
		utils.GenerateSecureToken(12),
		owner.String(),
		owner.String(),
		&msgVal,
		utils.GenerateSecureToken(4),
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

	// Try to get the beam via the ID taken from list
	beam, err := app.BeamKeeper.GetBeam(ctx, beams[0].GetId())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), beam.GetId(), msg.GetId())
}

// TestIncorrectBeamId A beam id that contains a comma must be refused
func (suite *KeeperTestSuite) TestIncorrectBeamId() {
	app := suite.app
	ctx := suite.ctx

	// Create the original owner
	owner := suite.addrs[0]

	// Create value and the linked message
	msgVal := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
	msg := types.NewMsgOpenBeam(
		"i-am-a-beam-id-with-a,comma",
		owner.String(),
		owner.String(),
		&msgVal,
		utils.GenerateSecureToken(4),
		types.BEAM_SCHEMA_REVIEW,
		nil,
		120,
		0,
	)

	// Open the beam and make sure there was an error
	err := app.BeamKeeper.OpenBeam(ctx, *msg)
	require.Error(suite.T(), err)
	require.False(suite.T(), app.BeamKeeper.HasBeam(ctx, msg.GetId()))
}

// TestKeeperSuite Main entry point for the testing suite
func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
