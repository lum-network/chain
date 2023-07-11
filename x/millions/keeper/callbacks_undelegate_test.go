package keeper_test

import (
	"cosmossdk.io/math"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func (suite *KeeperTestSuite) TestCallbacks_Undelegates() {
	pool := newValidPool(suite, millionstypes.Pool{PoolId: 1})

	// Prepare delegations split
	splits := pool.ComputeSplitDelegations(suite.ctx, math.NewInt(100))
	suite.Require().Greater(len(splits), 0)
	suite.Require().Equal(uint64(100), splits[0].Amount.Uint64())
	suite.Require().Greater(len(splits[0].ValidatorAddress), 0)

	portID := "icacontroller-pool1"
	channelID := "channel-0"
	sequence := uint64(5)

	// Construct our callback data
	callbackData := millionstypes.UndelegateCallback{
		PoolId:        pool.GetPoolId(),
		WithdrawalIds: []uint64{1, 2, 3},
	}

	// Serialize our callback
	marshalledCallbackArgs, err := suite.app.MillionsKeeper.MarshalUndelegateCallbackArgs(suite.ctx, callbackData)
	suite.Require().NoError(err)

	// Store inside the local datastore
	callback := icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(portID, channelID, sequence),
		PortId:       portID,
		ChannelId:    channelID,
		Sequence:     sequence,
		CallbackId:   millionskeeper.ICACallbackID_Undelegate,
		CallbackArgs: marshalledCallbackArgs,
	}
	suite.app.ICACallbacksKeeper.SetCallbackData(suite.ctx, callback)

	// Grab from the local datastore
	data, found := suite.app.ICACallbacksKeeper.GetCallbackData(suite.ctx, icacallbackstypes.PacketID(portID, channelID, sequence))
	suite.Require().True(found)

	// Deserialize our callback data
	unmarshalledCallbackData, err := suite.app.MillionsKeeper.UnmarshalUndelegateCallbackArgs(suite.ctx, data.CallbackArgs)
	suite.Require().NoError(err)

	// Make sure it matches
	suite.Require().Equal(callbackData.PoolId, unmarshalledCallbackData.PoolId)
	suite.Require().Equal(callbackData.WithdrawalIds[0], unmarshalledCallbackData.WithdrawalIds[0])
	suite.Require().Equal(callbackData.WithdrawalIds[1], unmarshalledCallbackData.WithdrawalIds[1])
	suite.Require().Equal(callbackData.WithdrawalIds[2], unmarshalledCallbackData.WithdrawalIds[2])
}
