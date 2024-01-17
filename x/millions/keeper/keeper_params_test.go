package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func (suite *KeeperTestSuite) TestParams_Validation() {
	app := suite.app
	ctx := suite.ctx

	params := app.MillionsKeeper.GetParams(ctx)
	suite.Require().NoError(params.ValidateBasics())

	// MinDepositAmount should always be gte than MinAcceptableDepositAmount
	params.MinDepositAmount = sdk.NewInt(millionstypes.MinAcceptableDepositAmount - 1)
	suite.Require().Error(params.ValidateBasics())
	params.MinDepositAmount = sdk.NewInt(millionstypes.MinAcceptableDepositAmount)
	suite.Require().NoError(params.ValidateBasics())

	// MaxPrizeStrategyBatches should always be gt 0
	params.MaxPrizeStrategyBatches = 0
	suite.Require().Error(params.ValidateBasics())
	params.MaxPrizeStrategyBatches = 1
	suite.Require().NoError(params.ValidateBasics())

	// MaxPrizeBatchQuantity should always be gt 0
	params.MaxPrizeBatchQuantity = 0
	suite.Require().Error(params.ValidateBasics())
	params.MaxPrizeBatchQuantity = 1
	suite.Require().NoError(params.ValidateBasics())

	// MinDrawScheduleDelta should always be gte MinAcceptableDrawDelta
	params.MinDrawScheduleDelta = millionstypes.MinAcceptableDrawDelta - 1
	suite.Require().Error(params.ValidateBasics())
	params.MinDrawScheduleDelta = millionstypes.MinAcceptableDrawDelta
	suite.Require().NoError(params.ValidateBasics())

	// MaxDrawScheduleDelta should always be gte MinDrawScheduleDelta
	params.MaxDrawScheduleDelta = params.MinDrawScheduleDelta - 1
	suite.Require().Error(params.ValidateBasics())
	params.MaxDrawScheduleDelta = params.MinDrawScheduleDelta
	suite.Require().NoError(params.ValidateBasics())

	// PrizeExpirationDelta should always be gte MinAcceptablePrizeExpirationDelta
	params.PrizeExpirationDelta = millionstypes.MinAcceptablePrizeExpirationDelta - 1
	suite.Require().Error(params.ValidateBasics())
	params.PrizeExpirationDelta = millionstypes.MinAcceptablePrizeExpirationDelta
	suite.Require().NoError(params.ValidateBasics())

	// MinDepositDrawDelta should always be gte MinAcceptableDepositDrawDelta
	params.MinDepositDrawDelta = millionstypes.MinAcceptableDepositDrawDelta - 1
	suite.Require().Error(params.ValidateBasics())
	params.MinDepositDrawDelta = millionstypes.MinAcceptableDepositDrawDelta
	suite.Require().NoError(params.ValidateBasics())
}
