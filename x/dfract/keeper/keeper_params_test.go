package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
			WithdrawalAddress: "invalidAddress",
		})
	}
	suite.Require().Panics(panicF)

	panicF = func() {
		app.DFractKeeper.SetParams(ctx, dfracttypes.Params{
			DepositDenoms:     []string{""},
			MinDepositAmount:  10000,
			WithdrawalAddress: "lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj",
		})
	}
	suite.Require().Panics(panicF)

	app.DFractKeeper.SetParams(ctx, dfracttypes.DefaultParams())
}

func (suite *KeeperTestSuite) TestParams_Update() {
	app := suite.app
	ctx := suite.ctx

	var emptyDepositDenoms []string
	validDepositDenoms := []string{dfracttypes.DefaultDenom, "udfr"}

	params := app.DFractKeeper.GetParams(ctx)
	suite.Require().Equal(dfracttypes.DefaultDenoms, params.DepositDenoms)
	suite.Require().Equal(uint32(dfracttypes.DefaultMinDepositAmount), params.MinDepositAmount)
	suite.Require().Equal("", params.WithdrawalAddress)
	suite.Require().Equal(true, params.IsDepositEnabled)

	err := app.DFractKeeper.UpdateParams(ctx, "lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", nil, emptyDepositDenoms, nil)
	suite.Require().NoError(err)
	// Update params with management address but no deposit enablement
	params = app.DFractKeeper.GetParams(ctx)
	suite.Require().Equal("lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", params.WithdrawalAddress)
	suite.Require().Equal(true, params.IsDepositEnabled)

	// Update with no management address but deposit enablement set to false
	err = app.DFractKeeper.UpdateParams(ctx, "", &gogotypes.BoolValue{Value: false}, emptyDepositDenoms, nil)
	suite.Require().NoError(err)
	params = app.DFractKeeper.GetParams(ctx)
	suite.Require().Equal("lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", params.WithdrawalAddress)
	suite.Require().Equal(false, params.IsDepositEnabled)

	newMinDepositAmount := sdk.NewInt(2000000)
	// Update deposit denoms and min deposit amount
	err = app.DFractKeeper.UpdateParams(ctx, "", nil, validDepositDenoms, &newMinDepositAmount)
	suite.Require().NoError(err)
	params = app.DFractKeeper.GetParams(ctx)
	suite.Require().Equal(validDepositDenoms, params.DepositDenoms)
	suite.Require().Equal(uint32(newMinDepositAmount.Uint64()), params.MinDepositAmount)

	newMinDepositAmount = sdk.NewInt(500000)
	// Should throw error as below min default amount
	err = app.DFractKeeper.UpdateParams(ctx, "", nil, validDepositDenoms, &newMinDepositAmount)
	suite.Require().Error(err)
}
