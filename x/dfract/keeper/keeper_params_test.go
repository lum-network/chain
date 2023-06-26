package keeper_test

import (
	gogotypes "github.com/cosmos/gogoproto/types"

	dfracttypes "github.com/lum-network/chain/x/dfract/types"
)

func (suite *KeeperTestSuite) TestInvalidParams() {
	app := suite.app
	ctx := suite.ctx

	panicF := func() {
		app.DFractKeeper.SetParams(ctx, dfracttypes.Params{
			DepositDenoms:    []string{""},
			MinDepositAmount: 0,
		})
	}
	suite.Require().Panics(panicF)

	panicF = func() {
		app.DFractKeeper.SetParams(ctx, dfracttypes.Params{
			DepositDenoms:    []string{""},
			MinDepositAmount: 1,
		})
	}
	suite.Require().Panics(panicF)

	panicF = func() {
		app.DFractKeeper.SetParams(ctx, dfracttypes.Params{
			DepositDenoms:     []string{""},
			MinDepositAmount:  10000,
			ManagementAddress: "invalidAddress",
		})
	}
	suite.Require().Panics(panicF)

	panicF = func() {
		app.DFractKeeper.SetParams(ctx, dfracttypes.Params{
			DepositDenoms:     []string{""},
			MinDepositAmount:  10000,
			ManagementAddress: "lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj",
		})
	}
	suite.Require().Panics(panicF)

	app.DFractKeeper.SetParams(ctx, dfracttypes.DefaultParams())
}

func (suite *KeeperTestSuite) TestParams_Update() {
	app := suite.app
	ctx := suite.ctx

	params := app.DFractKeeper.GetParams(ctx)
	suite.Require().Equal(dfracttypes.DefaultDenoms, params.DepositDenoms)
	suite.Require().Equal(uint32(dfracttypes.DefaultMinDepositAmount), params.MinDepositAmount)
	suite.Require().Equal("", params.ManagementAddress)
	suite.Require().Equal(true, params.IsDepositEnabled)

	err := app.DFractKeeper.UpdateParams(ctx, "lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", nil)
	suite.Require().NoError(err)
	// Update params with management address but no deposit enablement
	params = app.DFractKeeper.GetParams(ctx)
	suite.Require().Equal("lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", params.ManagementAddress)
	suite.Require().Equal(true, params.IsDepositEnabled)

	// Update with no management address but deposit enablement set to false
	err = app.DFractKeeper.UpdateParams(ctx, "", &gogotypes.BoolValue{Value: false})
	suite.Require().NoError(err)
	params = app.DFractKeeper.GetParams(ctx)
	suite.Require().Equal("lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", params.ManagementAddress)
	suite.Require().Equal(false, params.IsDepositEnabled)
}
