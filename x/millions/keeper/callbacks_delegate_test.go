package keeper_test

import (
	"cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func (suite *KeeperTestSuite) TestCallbacks_Delegates() {
	pool := newValidPool(suite, millionstypes.Pool{PoolId: 1})

	// Prepare delegations split
	splits := pool.ComputeSplitDelegations(suite.Ctx, math.NewInt(100))
	suite.Require().Greater(len(splits), 0)
	suite.Require().Equal(uint64(100), splits[0].Amount.Uint64())
	suite.Require().Greater(len(splits[0].ValidatorAddress), 0)

	portID := "icacontroller-pool1"
	sequence := uint64(5)

	// Construct our callback data
	callbackData := millionstypes.DelegateCallback{
		PoolId:           pool.GetPoolId(),
		DepositId:        1,
		SplitDelegations: splits,
	}

	// Serialize our callback
	marshalledCallbackArgs, err := suite.App.MillionsKeeper.MarshalDelegateCallbackArgs(suite.Ctx, callbackData)
	suite.Require().NoError(err)

	// Store inside the local datastore
	callback := icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(portID, ibctesting.FirstChannelID, sequence),
		PortId:       portID,
		ChannelId:    ibctesting.FirstChannelID,
		Sequence:     sequence,
		CallbackId:   millionskeeper.ICACallbackID_Delegate,
		CallbackArgs: marshalledCallbackArgs,
	}
	suite.App.ICACallbacksKeeper.SetCallbackData(suite.Ctx, callback)

	// Grab from the local datastore
	data, found := suite.App.ICACallbacksKeeper.GetCallbackData(suite.Ctx, icacallbackstypes.PacketID(portID, ibctesting.FirstChannelID, sequence))
	suite.Require().True(found)

	// Deserialize our callback data
	unmarshalledCallbackData, err := suite.App.MillionsKeeper.UnmarshalDelegateCallbackArgs(suite.Ctx, data.CallbackArgs)
	suite.Require().NoError(err)

	// Make sure it matches
	suite.Require().Equal(callbackData.PoolId, unmarshalledCallbackData.PoolId)
	suite.Require().Equal(callbackData.DepositId, unmarshalledCallbackData.DepositId)
	suite.Require().Equal(len(callbackData.SplitDelegations), len(unmarshalledCallbackData.SplitDelegations))
	suite.Require().Equal(callbackData.SplitDelegations[0].ValidatorAddress, unmarshalledCallbackData.SplitDelegations[0].ValidatorAddress)
	suite.Require().Equal(callbackData.SplitDelegations[0].Amount.Int64(), unmarshalledCallbackData.SplitDelegations[0].Amount.Int64())
}
