package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/utils"
	"github.com/lum-network/chain/x/dfract/keeper"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx   sdk.Context
	app   *app.App
	addrs []sdk.AccAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	app := app.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Setup the default application

	suite.app = app
	suite.ctx = ctx

	// Setup the module params
	err := app.DFractKeeper.SetParams(ctx, keeper.DefaultParams())
	suite.Require().NoError(err)

	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 2, sdk.NewInt(300000000), "ulum")
}

func (suite *KeeperTestSuite) TestInvalidDenomDeposit() {
	app := suite.app
	ctx := suite.ctx
	params, _ := app.DFractKeeper.GetParams(ctx)

	// Obtain the required accounts
	depositor := suite.addrs[0]

	// We try to deposit 100000000 of the mint denom
	err := app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.MintDenom, sdk.NewInt(100000000)),
		Id:               utils.GenerateSecureToken(4),
	})
	require.Error(suite.T(), err)
	require.Equal(suite.T(), err, types.ErrUnauthorizedDenom)
}

func (suite *KeeperTestSuite) TestInvalidAmountDeposit() {
	app := suite.app
	ctx := suite.ctx
	params, _ := app.DFractKeeper.GetParams(ctx)

	// Obtain the required accounts
	depositor := suite.addrs[0]

	// We try to deposit 0 of the deposit denom
	err := app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(0)),
		Id:               utils.GenerateSecureToken(4),
	})
	require.Error(suite.T(), err)
	require.Equal(suite.T(), err, types.ErrEmptyDepositAmount)
}

func (suite *KeeperTestSuite) TestInvalidIdDeposit() {
	app := suite.app
	ctx := suite.ctx
	params, _ := app.DFractKeeper.GetParams(ctx)

	// Obtain the required accounts
	depositor := suite.addrs[0]

	// We try to deposit 100000000 of the deposit denom
	err := app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
		Id:               "",
	})
	require.Error(suite.T(), err)
	require.Equal(suite.T(), err, types.ErrDepositAlreadyExists)
}

func (suite *KeeperTestSuite) TestDoubleDepositId() {
	app := suite.app
	ctx := suite.ctx
	params, _ := app.DFractKeeper.GetParams(ctx)

	// Obtain the required accounts
	depositor := suite.addrs[0]
	depositId := utils.GenerateSecureToken(4)

	// We try to deposit 100000000 of the deposit denom
	err := app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
		Id:               depositId,
	})
	require.NoError(suite.T(), err)

	// We try to deposit 100000000 of the deposit denom
	err = app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
		Id:               depositId,
	})
	require.Error(suite.T(), err)
	require.Equal(suite.T(), err, types.ErrDepositAlreadyExists)
}

func (suite *KeeperTestSuite) TestValidDeposit() {
	app := suite.app
	ctx := suite.ctx
	params, _ := app.DFractKeeper.GetParams(ctx)

	// Obtain the required accounts
	depositor := suite.addrs[0]
	depositId := utils.GenerateSecureToken(4)

	// We store the initial claimer funds
	depositorDepositFund := app.BankKeeper.GetBalance(ctx, depositor, params.DepositDenom)
	require.GreaterOrEqual(suite.T(), depositorDepositFund.Amount.Int64(), int64(300000000))
	depositorMintFund := app.BankKeeper.GetBalance(ctx, depositor, params.MintDenom)
	require.Equal(suite.T(), depositorMintFund.Amount.Int64(), int64(0))

	// We try to deposit 100000000 of the deposit denom
	err := app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
		Id:               depositId,
	})
	require.NoError(suite.T(), err)

	// We check the final claimer funds
	depositorDepositFund = app.BankKeeper.GetBalance(ctx, depositor, params.DepositDenom)
	require.Equal(suite.T(), depositorDepositFund.Amount.Int64(), int64(200000000))

	// Make sure the deposit actually exists
	require.True(suite.T(), app.DFractKeeper.HasDeposit(ctx, depositId))
	deposit, err := app.DFractKeeper.GetDeposit(ctx, depositId)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), deposit.DepositorAddress, depositor.String())
	require.Equal(suite.T(), deposit.Amount, sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)))
	require.Equal(suite.T(), deposit.Id, depositId)
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
