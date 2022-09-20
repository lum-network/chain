package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
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
	err := app.DFractKeeper.SetParams(ctx, types.DefaultParams())
	suite.Require().NoError(err)

	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 3, sdk.NewInt(300000000), "ulum")
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
	})
	require.Error(suite.T(), err)
	require.Equal(suite.T(), err, types.ErrEmptyDepositAmount)
}

func (suite *KeeperTestSuite) TestDoubleDeposit() {
	app := suite.app
	ctx := suite.ctx
	params, _ := app.DFractKeeper.GetParams(ctx)

	// Obtain the required accounts
	depositor := suite.addrs[0]

	// We try to deposit 100000000 of the deposit denom
	err := app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
	})
	require.NoError(suite.T(), err)

	// We try to deposit 100000000 of the deposit denom
	err = app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
	})
	require.NoError(suite.T(), err)

	// Acquire the deposit
	deposit, found := app.DFractKeeper.GetFromWaitingProposalDeposits(ctx, depositor.String())
	require.True(suite.T(), found)
	require.Equal(suite.T(), deposit.Amount, sdk.NewCoin(params.DepositDenom, sdk.NewInt(200000000)))
}

func (suite *KeeperTestSuite) TestValidDeposit() {
	app := suite.app
	ctx := suite.ctx
	params, _ := app.DFractKeeper.GetParams(ctx)

	// Obtain the required accounts
	depositor := suite.addrs[0]

	// We store the initial claimer funds
	depositorDepositFund := app.BankKeeper.GetBalance(ctx, depositor, params.DepositDenom)
	require.GreaterOrEqual(suite.T(), depositorDepositFund.Amount.Int64(), int64(300000000))
	depositorMintFund := app.BankKeeper.GetBalance(ctx, depositor, params.MintDenom)
	require.Equal(suite.T(), depositorMintFund.Amount.Int64(), int64(0))

	// We try to deposit 100000000 of the deposit denom
	err := app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
	})
	require.NoError(suite.T(), err)

	// We check the final claimer funds
	depositorDepositFund = app.BankKeeper.GetBalance(ctx, depositor, params.DepositDenom)
	require.Equal(suite.T(), depositorDepositFund.Amount.Int64(), int64(200000000))

	// Make sure the deposit actually exists
	deposit, found := app.DFractKeeper.GetFromWaitingProposalDeposits(ctx, depositor.String())
	require.True(suite.T(), found)
	require.Equal(suite.T(), depositor.String(), deposit.DepositorAddress)
	require.Equal(suite.T(), deposit.Amount, sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)))
}

func (suite *KeeperTestSuite) TestFullProcess() {
	app := suite.app
	ctx := suite.ctx
	params, _ := app.DFractKeeper.GetParams(ctx)

	// Define our mint rate to use
	var mintRate int64 = 2

	// Obtain the required accounts
	depositor1 := suite.addrs[0]
	depositor2 := suite.addrs[1]
	destinationAccount := suite.addrs[2]

	// We store the initial claimer funds
	require.GreaterOrEqual(suite.T(), app.BankKeeper.GetBalance(ctx, depositor1, params.DepositDenom).Amount.Int64(), int64(300000000))
	require.GreaterOrEqual(suite.T(), app.BankKeeper.GetBalance(ctx, depositor2, params.DepositDenom).Amount.Int64(), int64(300000000))
	require.GreaterOrEqual(suite.T(), app.BankKeeper.GetBalance(ctx, destinationAccount, params.MintDenom).Amount.Int64(), int64(0))
	require.GreaterOrEqual(suite.T(), app.BankKeeper.GetBalance(ctx, destinationAccount, params.DepositDenom).Amount.Int64(), int64(300000000))

	// We store the initial module account balance
	moduleAccountDepositBalance := app.BankKeeper.GetBalance(ctx, app.DFractKeeper.GetModuleAccount(ctx), params.DepositDenom)
	moduleAccountMintBalance := app.BankKeeper.GetBalance(ctx, app.DFractKeeper.GetModuleAccount(ctx), params.MintDenom)
	require.Equal(suite.T(), moduleAccountDepositBalance.Amount.Int64(), int64(0))
	require.Equal(suite.T(), moduleAccountMintBalance.Amount.Int64(), int64(0))

	// We try to deposit 100000000 of the deposit denom, and make sure our user is actually debited while the other one remain intact
	err := app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor1.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
	})
	afterFirstDepositBalance1 := app.BankKeeper.GetBalance(ctx, depositor1, params.DepositDenom)
	afterFirstDepositBalance2 := app.BankKeeper.GetBalance(ctx, depositor2, params.DepositDenom)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), afterFirstDepositBalance1.Amount.Int64(), int64(200000000))
	require.Equal(suite.T(), afterFirstDepositBalance2.Amount.Int64(), int64(300000000))

	// We make sure the module account balance received the amounts
	afterDepositModuleAccount := app.BankKeeper.GetBalance(ctx, app.DFractKeeper.GetModuleAccount(ctx), params.DepositDenom)
	require.Equal(suite.T(), afterDepositModuleAccount.Amount.Int64(), int64(100000000))

	// Make the second deposit
	err = app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor2.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
	})
	require.NoError(suite.T(), err)

	// We make sure the module account balance received the amounts
	afterDepositModuleAccount = app.BankKeeper.GetBalance(ctx, app.DFractKeeper.GetModuleAccount(ctx), params.DepositDenom)
	require.Equal(suite.T(), afterDepositModuleAccount.Amount.Int64(), int64(200000000))

	// We make sure our users have no mint balance
	depositor1MintBalance := app.BankKeeper.GetBalance(ctx, depositor1, params.MintDenom)
	depositor2MintBalance := app.BankKeeper.GetBalance(ctx, depositor2, params.MintDenom)
	require.Equal(suite.T(), depositor1MintBalance.Amount.Int64(), int64(0))
	require.Equal(suite.T(), depositor2MintBalance.Amount.Int64(), int64(0))

	// Acquire the list of waiting proposal & waiting mint deposits
	waitingProposalDeposits := app.DFractKeeper.ListWaitingProposalDeposits(ctx)
	waitingMintDeposits := app.DFractKeeper.ListWaitingMintDeposits(ctx)
	require.Equal(suite.T(), 2, len(waitingProposalDeposits))
	require.Equal(suite.T(), 0, len(waitingMintDeposits))

	// Process the first proposal iteration
	err = app.DFractKeeper.ProcessSpendAndAdjustProposal(ctx, &types.SpendAndAdjustProposal{
		MintRate:         mintRate,
		Title:            "test",
		Description:      "test",
		SpendDestination: destinationAccount.String(),
	})
	require.NoError(suite.T(), err)

	// Now the module account should be empty and the destination account must contain initial allocation plus deposited tokens
	afterFirstProposalModuleAccount := app.BankKeeper.GetBalance(ctx, app.DFractKeeper.GetModuleAccount(ctx), params.DepositDenom)
	afterFirstProposalDestinationAccount := app.BankKeeper.GetBalance(ctx, destinationAccount, params.DepositDenom)
	require.Equal(suite.T(), afterFirstProposalModuleAccount.Amount.Int64(), int64(0))
	require.Equal(suite.T(), afterFirstProposalDestinationAccount.Amount.Int64(), int64(300000000+200000000))

	// We make sure our users have no mint balance
	depositor1MintBalance = app.BankKeeper.GetBalance(ctx, depositor1, params.MintDenom)
	depositor2MintBalance = app.BankKeeper.GetBalance(ctx, depositor2, params.MintDenom)
	require.Equal(suite.T(), int64(0), depositor1MintBalance.Amount.Int64())
	require.Equal(suite.T(), int64(0), depositor2MintBalance.Amount.Int64())

	// Ensure the queues state
	waitingProposalDeposits = app.DFractKeeper.ListWaitingProposalDeposits(ctx)
	waitingMintDeposits = app.DFractKeeper.ListWaitingMintDeposits(ctx)
	mintedDepositsBefore := app.DFractKeeper.ListMintedDeposits(ctx)
	require.Equal(suite.T(), 0, len(waitingProposalDeposits))
	require.Equal(suite.T(), 2, len(waitingMintDeposits))
	require.Equal(suite.T(), 0, len(mintedDepositsBefore))

	// Process our second proposal
	err = app.DFractKeeper.ProcessSpendAndAdjustProposal(ctx, &types.SpendAndAdjustProposal{
		MintRate:         mintRate,
		Title:            "test 2",
		Description:      "test",
		SpendDestination: destinationAccount.String(),
	})
	require.NoError(suite.T(), err)

	// Our module account balance must not have changed
	afterSecondProposalModuleAccount := app.BankKeeper.GetBalance(ctx, app.DFractKeeper.GetModuleAccount(ctx), params.DepositDenom)
	afterSecondProposalDestinationAccount := app.BankKeeper.GetBalance(ctx, destinationAccount, params.DepositDenom)
	require.Equal(suite.T(), afterFirstProposalModuleAccount.Amount.Int64(), afterSecondProposalModuleAccount.Amount.Int64())
	require.Equal(suite.T(), afterFirstProposalDestinationAccount.Amount.Int64(), afterSecondProposalDestinationAccount.Amount.Int64())

	// Make sure the queues swapped
	waitingProposalDeposits = app.DFractKeeper.ListWaitingProposalDeposits(ctx)
	waitingMintDeposits = app.DFractKeeper.ListWaitingMintDeposits(ctx)
	mintedDepositsAfter := app.DFractKeeper.ListMintedDeposits(ctx)
	require.Equal(suite.T(), 0, len(waitingProposalDeposits))
	require.Equal(suite.T(), 0, len(waitingMintDeposits))
	require.Equal(suite.T(), 2, len(mintedDepositsAfter))

	// Make sure our users have their tokens
	depositor1MintBalance = app.BankKeeper.GetBalance(ctx, depositor1, params.MintDenom)
	depositor2MintBalance = app.BankKeeper.GetBalance(ctx, depositor2, params.MintDenom)
	require.Equal(suite.T(), int64(100000000*mintRate), depositor1MintBalance.Amount.Int64())
	require.Equal(suite.T(), int64(100000000*mintRate), depositor2MintBalance.Amount.Int64())
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
