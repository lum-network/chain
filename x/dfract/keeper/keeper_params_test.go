package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
)

func (suite *KeeperTestSuite) TestInvalidParams() {
	app := suite.app
	ctx := suite.ctx

	panicF := func() {
		app.DFractKeeper.SetParams(ctx, dfracttypes.Params{
			DepositDenoms:    []string{""},
			MinDepositAmount: math.NewInt(0),
		})
	}
	suite.Require().Panics(panicF)

	panicF = func() {
		app.DFractKeeper.SetParams(ctx, dfracttypes.Params{
			DepositDenoms:    []string{""},
			MinDepositAmount: math.NewInt(1),
		})
	}
	suite.Require().Panics(panicF)

	panicF = func() {
		app.DFractKeeper.SetParams(ctx, dfracttypes.Params{
			DepositDenoms:     []string{""},
			MinDepositAmount:  math.NewInt(10000),
			WithdrawalAddress: "invalidAddress",
		})
	}
	suite.Require().Panics(panicF)

	panicF = func() {
		app.DFractKeeper.SetParams(ctx, dfracttypes.Params{
			DepositDenoms:     []string{""},
			MinDepositAmount:  math.NewInt(10000),
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
	suite.Require().Equal([]string{dfracttypes.DefaultDenom}, params.DepositDenoms)
	suite.Require().Equal(math.NewInt(dfracttypes.DefaultMinDepositAmount), params.MinDepositAmount)
	suite.Require().Equal("", params.WithdrawalAddress)
	suite.Require().Equal(true, params.IsDepositEnabled)

	// Update params with management address but no deposit enablement
	params.WithdrawalAddress = "lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj"
	params.DepositDenoms = emptyDepositDenoms
	app.DFractKeeper.SetParams(ctx, params)
	params = app.DFractKeeper.GetParams(ctx)
	suite.Require().Equal("lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", params.WithdrawalAddress)
	suite.Require().Equal(true, params.IsDepositEnabled)

	// Update with no management address but deposit enablement set to false
	params.IsDepositEnabled = false
	app.DFractKeeper.SetParams(ctx, params)
	params = app.DFractKeeper.GetParams(ctx)
	suite.Require().Equal("lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", params.WithdrawalAddress)
	suite.Require().Equal(false, params.IsDepositEnabled)

	// Update deposit denoms and min deposit amount
	newMinDepositAmount := sdk.NewInt(2000000)
	params.MinDepositAmount = newMinDepositAmount
	params.DepositDenoms = validDepositDenoms
	app.DFractKeeper.SetParams(ctx, params)
	params = app.DFractKeeper.GetParams(ctx)
	suite.Require().Equal(validDepositDenoms, params.DepositDenoms)
	suite.Require().Equal(newMinDepositAmount, params.MinDepositAmount)

	// Should throw error as below min default amount
	newMinDepositAmount = sdk.NewInt(500000)
	params.MinDepositAmount = newMinDepositAmount
	suite.Require().Panics(func() {
		app.DFractKeeper.SetParams(ctx, params)
	})
}
