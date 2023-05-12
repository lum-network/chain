package keeper_test

import (
	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func (suite *KeeperTestSuite) TestCallbacks_TransfersToNative() {
	pool := newValidPool(suite, millionstypes.Pool{PoolId: 1})

	portID := "icacontroller-pool1"
	channelID := "channel-0"
	sequence := uint64(5)

	// Construct our callback data
	callbackData := millionstypes.TransferToNativeCallback{
		PoolId:    pool.GetPoolId(),
		DepositId: 1,
	}

	// Serialize our callback
	marshalledCallbackArgs, err := suite.app.MillionsKeeper.MarshalTransferToNativeCallbackArgs(suite.ctx, callbackData)
	suite.Require().NoError(err)

	// Store inside the local datastore
	callback := icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(portID, channelID, sequence),
		PortId:       portID,
		ChannelId:    channelID,
		Sequence:     sequence,
		CallbackId:   millionskeeper.ICACallbackID_TransferToNative,
		CallbackArgs: marshalledCallbackArgs,
	}
	suite.app.ICACallbacksKeeper.SetCallbackData(suite.ctx, callback)

	// Grab from the local datastore
	data, found := suite.app.ICACallbacksKeeper.GetCallbackData(suite.ctx, icacallbackstypes.PacketID(portID, channelID, sequence))
	suite.Require().True(found)

	// Deserialize our callback data
	unmarshalledCallbackData, err := suite.app.MillionsKeeper.UnmarshalTransferToNativeCallbackArgs(suite.ctx, data.CallbackArgs)
	suite.Require().NoError(err)

	// Make sure it matches
	suite.Require().Equal(callbackData.PoolId, unmarshalledCallbackData.PoolId)
	suite.Require().Equal(callbackData.DepositId, unmarshalledCallbackData.DepositId)
}
