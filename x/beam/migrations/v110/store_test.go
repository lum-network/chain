package v110_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	apptypes "github.com/lum-network/chain/app"
	"github.com/lum-network/chain/utils"
	v110 "github.com/lum-network/chain/x/beam/migrations/v110"
	"github.com/lum-network/chain/x/beam/types"
)

type StoreMigrationTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *apptypes.App
}

func (suite *StoreMigrationTestSuite) SetupTest() {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	suite.app = app
	suite.ctx = ctx
}

func (suite *StoreMigrationTestSuite) TestDisabledAutoCloseMigration() {
	// Simulate legacy address
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// Simulate legacy beam entries with auto close disabled (otherwise they would've been added automatically to the new queue)
	beam1id := utils.GenerateSecureToken(8)
	_, err := suite.app.BeamKeeper.OpenBeam(suite.ctx, &types.MsgOpenBeam{
		Id:                  beam1id,
		CreatorAddress:      addr1.String(),
		Secret:              utils.GenerateSecureToken(10),
		Schema:              "",
		Data:                nil,
		ClosesAtBlock:       0,
		ClaimExpiresAtBlock: 0,
	})
	suite.Require().NoError(err)
	beam2id := utils.GenerateSecureToken(8)
	_, err = suite.app.BeamKeeper.OpenBeam(suite.ctx, &types.MsgOpenBeam{
		Id:                  beam2id,
		CreatorAddress:      addr2.String(),
		Secret:              utils.GenerateSecureToken(10),
		Schema:              "",
		Data:                nil,
		ClosesAtBlock:       0,
		ClaimExpiresAtBlock: 0,
	})
	suite.Require().NoError(err)
	beam3id := utils.GenerateSecureToken(8)
	_, err = suite.app.BeamKeeper.OpenBeam(suite.ctx, &types.MsgOpenBeam{
		Id:                  beam3id,
		CreatorAddress:      addr3.String(),
		Secret:              utils.GenerateSecureToken(10),
		Schema:              "",
		Data:                nil,
		ClosesAtBlock:       0,
		ClaimExpiresAtBlock: 0,
	})
	suite.Require().NoError(err)

	// Make sure the open beam iterator is invalid
	openedIterator := suite.app.BeamKeeper.OpenBeamsQueueIterator(suite.ctx)
	suite.Require().Error(openedIterator.Error())
	suite.Require().False(openedIterator.Valid())
	openedIterator.Close()

	// Manually append the entries into the old queue system
	suite.app.BeamKeeper.InsertOpenBeamQueue(suite.ctx, beam1id)
	suite.app.BeamKeeper.InsertOpenBeamQueue(suite.ctx, beam2id)
	suite.app.BeamKeeper.InsertOpenBeamQueue(suite.ctx, beam3id)

	// Make sure the open beam queue is valid
	openedIterator = suite.app.BeamKeeper.OpenBeamsQueueIterator(suite.ctx)
	suite.Require().NoError(openedIterator.Error())
	suite.Require().True(openedIterator.Valid())
	openedIterator.Close()

	// Run the migration
	err = v110.MigrateBeamQueues(suite.ctx, *suite.app.BeamKeeper)
	suite.Require().NoError(err)

	// Make sure the open beam queue is empty and now invalid
	openedIterator = suite.app.BeamKeeper.OpenBeamsQueueIterator(suite.ctx)
	suite.Require().Error(openedIterator.Error())
	suite.Require().False(openedIterator.Valid())
	openedIterator.Close()

	// Make sure the open by height queue is invalid
	openedIterator = suite.app.BeamKeeper.OpenBeamsByBlockQueueIterator(suite.ctx)
	suite.Require().Error(openedIterator.Error())
	suite.Require().False(openedIterator.Valid())
	openedIterator.Close()
}

func (suite *StoreMigrationTestSuite) TestEnabledAutoCloseMigration() {
	// Simulate legacy address
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// Simulate legacy beam entries with auto close disabled (otherwise they would've been added automatically to the new queue)
	beam1id := utils.GenerateSecureToken(8)
	_, err := suite.app.BeamKeeper.OpenBeam(suite.ctx, &types.MsgOpenBeam{
		Id:                  beam1id,
		CreatorAddress:      addr1.String(),
		Secret:              utils.GenerateSecureToken(10),
		Schema:              "",
		Data:                nil,
		ClosesAtBlock:       10,
		ClaimExpiresAtBlock: 10,
	})
	suite.Require().NoError(err)
	beam2id := utils.GenerateSecureToken(8)
	_, err = suite.app.BeamKeeper.OpenBeam(suite.ctx, &types.MsgOpenBeam{
		Id:                  beam2id,
		CreatorAddress:      addr2.String(),
		Secret:              utils.GenerateSecureToken(10),
		Schema:              "",
		Data:                nil,
		ClosesAtBlock:       10,
		ClaimExpiresAtBlock: 10,
	})
	suite.Require().NoError(err)
	beam3id := utils.GenerateSecureToken(8)
	_, err = suite.app.BeamKeeper.OpenBeam(suite.ctx, &types.MsgOpenBeam{
		Id:                  beam3id,
		CreatorAddress:      addr3.String(),
		Secret:              utils.GenerateSecureToken(10),
		Schema:              "",
		Data:                nil,
		ClosesAtBlock:       10,
		ClaimExpiresAtBlock: 10,
	})
	suite.Require().NoError(err)

	// Manually append the entries into the old queue system
	suite.app.BeamKeeper.InsertOpenBeamQueue(suite.ctx, beam1id)
	suite.app.BeamKeeper.InsertOpenBeamQueue(suite.ctx, beam2id)
	suite.app.BeamKeeper.InsertOpenBeamQueue(suite.ctx, beam3id)

	// Make sure we have three IDs at the height in the new system
	entries := suite.app.BeamKeeper.GetBeamIDsFromBlockQueue(suite.ctx, 10)
	suite.Require().Equal(3, len(entries))

	// Manually remove the entries from the new system
	suite.app.BeamKeeper.RemoveFromOpenBeamByBlockQueue(suite.ctx, 10, beam1id)
	suite.app.BeamKeeper.RemoveFromOpenBeamByBlockQueue(suite.ctx, 10, beam2id)
	suite.app.BeamKeeper.RemoveFromOpenBeamByBlockQueue(suite.ctx, 10, beam3id)

	entries = suite.app.BeamKeeper.GetBeamIDsFromBlockQueue(suite.ctx, 10)
	suite.Require().Equal(0, len(entries))

	// Make sure the open beam iterator is valid
	openedIterator := suite.app.BeamKeeper.OpenBeamsQueueIterator(suite.ctx)
	suite.Require().NoError(openedIterator.Error())
	suite.Require().True(openedIterator.Valid())
	openedIterator.Close()

	// Run the migration
	err = v110.MigrateBeamQueues(suite.ctx, *suite.app.BeamKeeper)
	suite.Require().NoError(err)

	// Make sure the legacy open beam iterator is invalid
	openedIterator = suite.app.BeamKeeper.OpenBeamsQueueIterator(suite.ctx)
	suite.Require().Error(openedIterator.Error())
	suite.Require().False(openedIterator.Valid())
	openedIterator.Close()

	// Make sure the new open beam by height queue is valid
	openedIterator = suite.app.BeamKeeper.OpenBeamsByBlockQueueIterator(suite.ctx)
	suite.Require().NoError(openedIterator.Error())
	suite.Require().True(openedIterator.Valid())
	openedIterator.Close()

	// Make sure we have three IDs at the height in the new queue
	entries = suite.app.BeamKeeper.GetBeamIDsFromBlockQueue(suite.ctx, 10)
	suite.Require().Equal(3, len(entries))
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(StoreMigrationTestSuite))
}
