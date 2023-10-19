package v162_test

import (
	"encoding/hex"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	apptypes "github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/utils"
	v162 "github.com/lum-network/chain/x/beam/migrations/v162"
	"github.com/lum-network/chain/x/beam/types"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type StoreMigrationTestSuite struct {
	suite.Suite

	ctx   sdk.Context
	app   *apptypes.App
	addrs []sdk.AccAddress
}

func (suite *StoreMigrationTestSuite) SetupTest() {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Set up the default application
	suite.app = app
	suite.ctx = ctx.WithChainID("lum-network-devnet").WithBlockTime(time.Now().UTC())

	// Set up the default addresses
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 6, sdk.NewInt(10_000_000_000), app.StakingKeeper.BondDenom(ctx))
}

// TestDeleteBeamData tests the migration that deletes all beam data
func (suite *StoreMigrationTestSuite) TestDeleteBeamOpenByBlockQueue() {
	// Generate 100 fake beams
	for i := 0; i < 100; i++ {
		// Create our beam ID
		coin := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
		id := utils.GenerateSecureToken(12)
		msg := types.NewMsgOpenBeam(
			id,
			suite.addrs[0].String(),
			suite.addrs[1].String(),
			&coin,
			hex.EncodeToString(utils.GenerateHashFromString("test_1234")),
			types.BEAM_SCHEMA_REVIEW,
			nil,
			int32(suite.ctx.BlockHeight()+int64(10)),
			int32(suite.ctx.BlockHeight()+int64(20)),
		)

		// Create the beam
		_, err := suite.app.BeamKeeper.OpenBeam(suite.ctx, msg)
		suite.Require().NoError(err)
	}

	// Our beam queues should not be empty (valid)
	openByBlockQueue := suite.app.BeamKeeper.OpenBeamsByBlockQueueIterator(suite.ctx)
	suite.Require().True(openByBlockQueue.Valid())
	openByBlockQueue.Close()

	// Run the migration
	err := v162.DeleteBeamsData(suite.ctx, *suite.app.BeamKeeper)
	suite.Require().NoError(err)

	// Now it must be empty
	openByBlockQueue = suite.app.BeamKeeper.OpenBeamsByBlockQueueIterator(suite.ctx)
	suite.Require().False(openByBlockQueue.Valid())
	openByBlockQueue.Close()
}

// TestDeleteBeamClosedQueue tests the migration that deletes all beam data
func (suite *StoreMigrationTestSuite) TestDeleteBeamClosedQueue() {
	// Generate 100 fake beams
	for i := 0; i < 100; i++ {
		// Create our beam ID
		coin := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
		id := utils.GenerateSecureToken(12)
		msg := types.NewMsgOpenBeam(
			id,
			suite.addrs[0].String(),
			suite.addrs[1].String(),
			&coin,
			hex.EncodeToString(utils.GenerateHashFromString("test_1234")),
			types.BEAM_SCHEMA_REVIEW,
			nil,
			0,
			0,
		)

		// Create the beam
		_, err := suite.app.BeamKeeper.OpenBeam(suite.ctx, msg)
		suite.Require().NoError(err)

		// Automatically close the beam
		err = suite.app.BeamKeeper.UpdateBeamStatus(suite.ctx, id, types.BeamState_StateCanceled)
		suite.Require().NoError(err)
	}

	// Our beam queues should be empty (invalid)
	openQueue := suite.app.BeamKeeper.OpenBeamsByBlockQueueIterator(suite.ctx)
	suite.Require().False(openQueue.Valid())
	openQueue.Close()

	// Our closed beam queues should be empty (invalid)
	closedQueue := suite.app.BeamKeeper.ClosedBeamsQueueIterator(suite.ctx)
	suite.Require().True(closedQueue.Valid())
	closedQueue.Close()

	// Run the migration
	err := v162.DeleteBeamsData(suite.ctx, *suite.app.BeamKeeper)
	suite.Require().NoError(err)

	// Now it must be empty
	closedQueue = suite.app.BeamKeeper.ClosedBeamsQueueIterator(suite.ctx)
	suite.Require().False(closedQueue.Valid())
	closedQueue.Close()
}

// TestDeleteBeamData tests the migration that deletes all beam data
func (suite *StoreMigrationTestSuite) TestDeleteBeamData() {
	var beamIDs []string

	// Generate 100 fake beams
	for i := 0; i < 100; i++ {
		// Create our beam ID
		coin := sdk.NewCoin(apptypes.CoinBondDenom, sdk.NewInt(100))
		id := utils.GenerateSecureToken(12)
		msg := types.NewMsgOpenBeam(
			id,
			suite.addrs[0].String(),
			suite.addrs[1].String(),
			&coin,
			hex.EncodeToString(utils.GenerateHashFromString("test_1234")),
			types.BEAM_SCHEMA_REVIEW,
			nil,
			int32(suite.ctx.BlockHeight()+int64(10)),
			int32(suite.ctx.BlockHeight()+int64(20)),
		)

		// Push the beam id inside the array
		beamIDs = append(beamIDs, id)

		// Create the beam
		_, err := suite.app.BeamKeeper.OpenBeam(suite.ctx, msg)
		suite.Require().NoError(err)
	}

	// Run the migration
	err := v162.DeleteBeamsData(suite.ctx, *suite.app.BeamKeeper)
	suite.Require().NoError(err)

	// Loop over the beam id array and make sure they do not exist anymore
	for _, id := range beamIDs {
		_, err := suite.app.BeamKeeper.GetBeam(suite.ctx, id)
		suite.Require().Error(err)
	}
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(StoreMigrationTestSuite))
}
