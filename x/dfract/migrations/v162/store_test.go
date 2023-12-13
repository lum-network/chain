package v162_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	apptypes "github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/x/dfract/keeper"
	v162 "github.com/lum-network/chain/x/dfract/migrations/v162"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
)

type StoreMigrationTestSuite struct {
	suite.Suite

	ctx   sdk.Context
	app   *apptypes.App
	addrs []sdk.AccAddress
}

func (suite *StoreMigrationTestSuite) SetupTest() {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Set up the default application
	suite.app = app
	suite.ctx = ctx.WithChainID("lum-network-devnet").WithBlockTime(time.Now().UTC())

	// Iterate over the array of deposit denoms
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 6, sdk.NewInt(300000000), dfracttypes.DefaultDenom)

	// Set up the default params
	params := dfracttypes.DefaultParams()
	params.WithdrawalAddress = suite.addrs[1].String()
	app.DFractKeeper.SetParams(ctx, params)
}

func (suite *StoreMigrationTestSuite) TestMigration() {
	msgServer := keeper.NewMsgServerImpl(*suite.app.DFractKeeper)

	// Simulate 10 deposits
	for i := 0; i < 10; i++ {
		_, err := msgServer.Deposit(sdk.WrapSDKContext(suite.ctx), &dfracttypes.MsgDeposit{
			DepositorAddress: suite.addrs[0].String(),
			Amount:           sdk.NewCoin(dfracttypes.DefaultDenom, sdk.NewInt(1000000)),
		})
		suite.Require().NoError(err)
	}

	// Check the states
	_, found := suite.app.DFractKeeper.GetDepositPendingWithdrawal(suite.ctx, suite.addrs[0])
	suite.Require().True(found)
	_, found = suite.app.DFractKeeper.GetDepositMinted(suite.ctx, suite.addrs[0])
	suite.Require().False(found)
	_, found = suite.app.DFractKeeper.GetDepositPendingMint(suite.ctx, suite.addrs[0])
	suite.Require().False(found)

	// Run the process of withdraw & mint
	_, err := msgServer.WithdrawAndMint(sdk.WrapSDKContext(suite.ctx), &dfracttypes.MsgWithdrawAndMint{
		Address:       suite.addrs[1].String(),
		MicroMintRate: 2,
	})
	suite.Require().NoError(err)

	// Check the states
	_, found = suite.app.DFractKeeper.GetDepositPendingWithdrawal(suite.ctx, suite.addrs[0])
	suite.Require().False(found)
	_, found = suite.app.DFractKeeper.GetDepositMinted(suite.ctx, suite.addrs[0])
	suite.Require().False(found)
	_, found = suite.app.DFractKeeper.GetDepositPendingMint(suite.ctx, suite.addrs[0])
	suite.Require().True(found)

	// Run the process of withdraw & mint
	_, err = msgServer.WithdrawAndMint(sdk.WrapSDKContext(suite.ctx), &dfracttypes.MsgWithdrawAndMint{
		Address:       suite.addrs[1].String(),
		MicroMintRate: 2,
	})
	suite.Require().NoError(err)

	// Check the states
	_, found = suite.app.DFractKeeper.GetDepositPendingWithdrawal(suite.ctx, suite.addrs[0])
	suite.Require().False(found)
	_, found = suite.app.DFractKeeper.GetDepositMinted(suite.ctx, suite.addrs[0])
	suite.Require().True(found)
	_, found = suite.app.DFractKeeper.GetDepositPendingMint(suite.ctx, suite.addrs[0])
	suite.Require().False(found)

	// Initialize a transfer out of the scope of the module
	err = suite.app.BankKeeper.SendCoins(suite.ctx, suite.addrs[0], suite.addrs[1], sdk.NewCoins(sdk.NewCoin(dfracttypes.MintDenom, sdk.NewInt(5))))
	suite.Require().NoError(err)

	// Get the destination balance
	cheaterBalance := suite.app.BankKeeper.GetBalance(suite.ctx, suite.addrs[1], dfracttypes.MintDenom)
	suite.Require().Equal(cheaterBalance.Amount.Int64(), int64(5))

	// Get the total supply
	beforeSupply := suite.app.BankKeeper.GetSupply(suite.ctx, dfracttypes.MintDenom)
	suite.Require().NoError(err)
	suite.Require().Equal(beforeSupply.Amount.Int64(), int64(20))

	// Run the migration
	err = v162.Migrate(suite.ctx, *suite.app.DFractKeeper)
	suite.Require().NoError(err)

	// Cheater balance must be empty as well
	cheaterBalance = suite.app.BankKeeper.GetBalance(suite.ctx, suite.addrs[1], dfracttypes.MintDenom)
	suite.Require().Equal(cheaterBalance.Amount.Int64(), int64(0))

	// Queues must be empty
	_, found = suite.app.DFractKeeper.GetDepositPendingWithdrawal(suite.ctx, suite.addrs[0])
	suite.Require().False(found)
	_, found = suite.app.DFractKeeper.GetDepositMinted(suite.ctx, suite.addrs[0])
	suite.Require().False(found)
	_, found = suite.app.DFractKeeper.GetDepositPendingMint(suite.ctx, suite.addrs[0])
	suite.Require().False(found)

	// Get the total supply
	afterSupply := suite.app.BankKeeper.GetSupply(suite.ctx, dfracttypes.MintDenom)
	suite.Require().NoError(err)

	// Get the module account balance, and ensure it contains the total supply
	moduleAccount := suite.app.DFractKeeper.GetModuleAccount(suite.ctx)
	moduleAccountBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, moduleAccount)
	suite.Require().Equal(moduleAccountBalance.AmountOf(dfracttypes.MintDenom).Int64(), afterSupply.Amount.Int64())
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(StoreMigrationTestSuite))
}
