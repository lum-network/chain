package keeper_test

import (
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func (suite *KeeperTestSuite) TestCallbacks_WithdrawAddress() {
	pool := newValidPool(suite, millionstypes.Pool{PoolId: 1})

	portID := "icacontroller-pool1"
	sequence := uint64(5)

	// Construct our callback data
	callbackData := millionstypes.SetWithdrawAddressCallback{
		PoolId: pool.GetPoolId(),
	}

	// Serialize our callback
	marshalledCallbackArgs, err := suite.App.MillionsKeeper.MarshalSetWithdrawAddressCallbackArgs(suite.Ctx, callbackData)
	suite.Require().NoError(err)

	// Store inside the local datastore
	callback := icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(portID, ibctesting.FirstChannelID, sequence),
		PortId:       portID,
		ChannelId:    ibctesting.FirstChannelID,
		Sequence:     sequence,
		CallbackId:   millionskeeper.ICACallbackID_SetWithdrawAddress,
		CallbackArgs: marshalledCallbackArgs,
	}
	suite.App.ICACallbacksKeeper.SetCallbackData(suite.Ctx, callback)

	// Grab from the local datastore
	data, found := suite.App.ICACallbacksKeeper.GetCallbackData(suite.Ctx, icacallbackstypes.PacketID(portID, ibctesting.FirstChannelID, sequence))
	suite.Require().True(found)

	// Deserialize our callback data
	unmarshalledCallbackData, err := suite.App.MillionsKeeper.UnmarshalSetWithdrawAddressCallbackArgs(suite.Ctx, data.CallbackArgs)
	suite.Require().NoError(err)

	// Make sure it matches
	suite.Require().Equal(callbackData.PoolId, unmarshalledCallbackData.PoolId)
}
