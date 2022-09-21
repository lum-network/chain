package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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

	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 6, sdk.NewInt(300000000), "ulum")
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

	// We try to deposit 100000000 of another denom != from the mintDenom
	err = app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin("uatom", sdk.NewInt(100000000)),
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

	// The total final deposit should reflect the two deposits done
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

	// We store the initial depositor balances
	depositorAvailableBalance := app.BankKeeper.GetBalance(ctx, depositor, params.DepositDenom)
	require.GreaterOrEqual(suite.T(), depositorAvailableBalance.Amount.Int64(), int64(300000000))
	depositorMintedBalance := app.BankKeeper.GetBalance(ctx, depositor, params.MintDenom)
	require.Equal(suite.T(), depositorMintedBalance.Amount.Int64(), int64(0))

	// We try to deposit 100000000 of the deposit denom
	err := app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
	})
	require.NoError(suite.T(), err)

	// Depositor balance should reflect this transfer
	depositorAvailableBalance = app.BankKeeper.GetBalance(ctx, depositor, params.DepositDenom)
	require.Equal(suite.T(), depositorAvailableBalance.Amount.Int64(), int64(200000000))

	// Depositor should have its deposit available in the waiting proposal queue
	deposit, found := app.DFractKeeper.GetFromWaitingProposalDeposits(ctx, depositor.String())
	require.True(suite.T(), found)
	require.Equal(suite.T(), depositor.String(), deposit.DepositorAddress)
	require.Equal(suite.T(), deposit.Amount, sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)))

	// Making another deposit should have the same effect
	err = app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(100000000)),
	})
	require.NoError(suite.T(), err)
	depositorAvailableBalance = app.BankKeeper.GetBalance(ctx, depositor, params.DepositDenom)
	require.Equal(suite.T(), depositorAvailableBalance.Amount.Int64(), int64(100000000))
	deposit, found = app.DFractKeeper.GetFromWaitingProposalDeposits(ctx, depositor.String())
	require.True(suite.T(), found)
	require.Equal(suite.T(), depositor.String(), deposit.DepositorAddress)
	require.Equal(suite.T(), deposit.Amount, sdk.NewCoin(params.DepositDenom, sdk.NewInt(200000000)))
}

func (suite *KeeperTestSuite) TestChainedFullProcess() {
	app := suite.app
	ctx := suite.ctx
	params, _ := app.DFractKeeper.GetParams(ctx)
	destAddr := suite.addrs[0].String()
	moduleAddr := app.DFractKeeper.GetModuleAccount(ctx).String()
	depositorsAddrs := []string{suite.addrs[1].String(), suite.addrs[2].String(), suite.addrs[3].String(), suite.addrs[4].String(), suite.addrs[5].String()}

	require.Equal(suite.T(), 5, len(depositorsAddrs))

	type stageDef struct {
		depositorsAddrs   []string
		depositorsAmounts []int64
		mintRate          int64
	}

	// Build stages to run with voluntary empty stages
	stages := []stageDef{{
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		mintRate:          2,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		mintRate:          3,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		mintRate:          4,
	}, {
		depositorsAddrs:   []string{depositorsAddrs[0]},
		depositorsAmounts: []int64{100000},
		mintRate:          5,
	}, {
		depositorsAddrs:   []string{depositorsAddrs[0], depositorsAddrs[1]},
		depositorsAmounts: []int64{2000000, 3000000},
		mintRate:          6,
	}, {
		depositorsAddrs:   []string{depositorsAddrs[0], depositorsAddrs[1], depositorsAddrs[2], depositorsAddrs[3]},
		depositorsAmounts: []int64{4000000, 5000000, 6000000, 7000000},
		mintRate:          7,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		mintRate:          8,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		mintRate:          9,
	}, {
		depositorsAddrs:   []string{depositorsAddrs[1], depositorsAddrs[2], depositorsAddrs[3], depositorsAddrs[4]},
		depositorsAmounts: []int64{8000000, 9000000, 10000000, 11000000},
		mintRate:          10,
	}, {
		depositorsAddrs:   []string{depositorsAddrs[3], depositorsAddrs[3], depositorsAddrs[4], depositorsAddrs[4], depositorsAddrs[4]},
		depositorsAmounts: []int64{12000000, 13000000, 14000000, 15000000, 16000000},
		mintRate:          11,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		mintRate:          12,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		mintRate:          13,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		mintRate:          14,
	}}

	// getAllStates acquires all relevant balances
	// use map["module"] for module balances
	// coin[0]: deposit balance
	// coin[1]: mint balance
	// coin[2]: waiting proposal balance
	// coin[3]: waiting mint balance
	getAllStates := func() map[string][]sdk.Coin {
		res := map[string][]sdk.Coin{}
		addrs := append(depositorsAddrs, destAddr, app.DFractKeeper.GetModuleAccount(ctx).String())

		for _, addr := range addrs {
			accAddr, err := sdk.AccAddressFromBech32(addr)
			require.NoError(suite.T(), err)
			res[addr] = []sdk.Coin{
				app.BankKeeper.GetBalance(ctx, accAddr, params.DepositDenom),
				app.BankKeeper.GetBalance(ctx, accAddr, params.MintDenom),
			}
			deposit, found := app.DFractKeeper.GetFromWaitingProposalDeposits(ctx, addr)
			if found {
				res[addr] = append(res[addr], deposit.Amount)
			} else {
				res[addr] = append(res[addr], sdk.NewCoin(params.DepositDenom, sdk.NewInt(0)))
			}
			deposit, found = app.DFractKeeper.GetFromWaitingMintDeposits(ctx, addr)
			if found {
				res[addr] = append(res[addr], deposit.Amount)
			} else {
				res[addr] = append(res[addr], sdk.NewCoin(params.DepositDenom, sdk.NewInt(0)))
			}
		}

		return res
	}

	preRunStates := getAllStates()
	preRunSupplies := []sdk.Coin{
		app.BankKeeper.GetSupply(ctx, params.DepositDenom),
		app.BankKeeper.GetSupply(ctx, params.MintDenom),
	}

	for i, stage := range stages {
		// Acquire current accounts balances states
		initialStates := getAllStates()
		initialSupplies := []sdk.Coin{
			app.BankKeeper.GetSupply(ctx, params.DepositDenom),
			app.BankKeeper.GetSupply(ctx, params.MintDenom),
		}
		require.Len(suite.T(), initialStates, len(depositorsAddrs)+2)

		// Make deposits
		var totalDepositedAmount int64
		depositedAmounts := map[string]int64{}
		for d, depositor := range stage.depositorsAddrs {
			err := app.DFractKeeper.CreateDeposit(ctx, types.MsgDeposit{
				DepositorAddress: depositor,
				Amount:           sdk.NewCoin(params.DepositDenom, sdk.NewInt(stage.depositorsAmounts[d])),
			})
			require.NoError(suite.T(), err, fmt.Sprintf("stage %d", i))
			depositedAmounts[depositor] += stage.depositorsAmounts[d]
			totalDepositedAmount += stage.depositorsAmounts[d]
		}

		// Control new account states
		postDepositStates := getAllStates()
		require.Len(suite.T(), postDepositStates, len(depositorsAddrs)+2)
		// Check depositors and non depositors balances changes
		for _, depositor := range depositorsAddrs {
			initialState := initialStates[depositor]
			postDepositState := postDepositStates[depositor]
			if depositedAmounts[depositor] > 0 {
				// Deposited amount(s) should have been removed from deposit balance
				require.True(suite.T(), postDepositState[0].Amount.Equal(initialState[0].Amount.Sub(sdk.NewInt(depositedAmounts[depositor]))), fmt.Sprintf("stage %d", i))
				// Mint balance should not have changed
				require.True(suite.T(), postDepositState[1].Amount.Equal(initialState[1].Amount), fmt.Sprintf("stage %d", i))
				// Waiting proposal balance should have inrease by the deposited amount
				require.True(suite.T(), postDepositState[2].Amount.Equal(initialState[2].Amount.Add(sdk.NewInt(depositedAmounts[depositor]))), fmt.Sprintf("stage %d", i))
				// Waiting mint balance should not have changed
				require.True(suite.T(), postDepositState[3].Amount.Equal(initialState[3].Amount), fmt.Sprintf("stage %d", i))
			} else {
				// Non depositor should have no change
				require.True(suite.T(), postDepositState[0].Amount.Equal(initialState[0].Amount), fmt.Sprintf("stage %d", i))
				require.True(suite.T(), postDepositState[1].Amount.Equal(initialState[1].Amount), fmt.Sprintf("stage %d", i))
				require.True(suite.T(), postDepositState[2].Amount.Equal(initialState[2].Amount), fmt.Sprintf("stage %d", i))
				require.True(suite.T(), postDepositState[3].Amount.Equal(initialState[3].Amount), fmt.Sprintf("stage %d", i))
			}
		}
		// Module account should have received the new deposits
		initialState := initialStates[moduleAddr]
		postDepositState := postDepositStates[moduleAddr]
		require.True(suite.T(), postDepositState[0].Amount.Equal(initialState[0].Amount.Add(sdk.NewInt(totalDepositedAmount))), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postDepositState[1].Amount.Equal(initialState[1].Amount), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postDepositState[2].Amount.Equal(initialState[2].Amount), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postDepositState[3].Amount.Equal(initialState[3].Amount), fmt.Sprintf("stage %d", i))
		// Destination address should not have changed
		initialState = initialStates[destAddr]
		postDepositState = postDepositStates[destAddr]
		require.True(suite.T(), postDepositState[0].Amount.Equal(initialState[0].Amount), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postDepositState[1].Amount.Equal(initialState[1].Amount), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postDepositState[2].Amount.Equal(initialState[2].Amount), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postDepositState[3].Amount.Equal(initialState[3].Amount), fmt.Sprintf("stage %d", i))
		// Control supply (no change expected)
		postDepositsSupplies := []sdk.Coin{
			app.BankKeeper.GetSupply(ctx, params.DepositDenom),
			app.BankKeeper.GetSupply(ctx, params.MintDenom),
		}
		require.True(suite.T(), postDepositsSupplies[0].Equal(initialSupplies[0]))
		require.True(suite.T(), postDepositsSupplies[1].Equal(initialSupplies[1]))

		// Run proposal
		err := app.DFractKeeper.ProcessSpendAndAdjustProposal(ctx, &types.SpendAndAdjustProposal{
			MintRate:         stage.mintRate,
			Title:            "test",
			Description:      "test",
			SpendDestination: destAddr,
		})
		require.NoError(suite.T(), err)

		// Control new account states
		postProposalStates := getAllStates()
		require.Len(suite.T(), postProposalStates, len(depositorsAddrs)+2)
		totalMintAmount := sdk.NewInt(0)
		// Check depositors and non depositors balances changes
		for _, depositor := range depositorsAddrs {
			postDepositState := postDepositStates[depositor]
			postProposalState := postProposalStates[depositor]
			// Deposit balance should not have change since the deposit
			require.True(suite.T(), postProposalState[0].Amount.Equal(postDepositState[0].Amount), fmt.Sprintf("stage %d", i))
			// Mint balance should have received the amount available in the waiting mint * mintRate
			require.True(suite.T(), postProposalState[1].Amount.Equal(postDepositState[1].Amount.Add(postDepositState[3].Amount.MulRaw(stage.mintRate))), fmt.Sprintf("stage %d", i))
			// Waiting proposal balance should be empty
			require.True(suite.T(), postProposalState[2].Amount.Equal(sdk.NewInt(0)), fmt.Sprintf("stage %d", i))
			// Waiting mint balance should have received the amount in the waiting proposal balance
			require.True(suite.T(), postProposalState[3].Amount.Equal(postDepositState[2].Amount), fmt.Sprintf("stage %d", i))
			totalMintAmount = totalMintAmount.Add(postDepositState[3].Amount)
		}
		// Module account should have sent the new deposits to the destAddr (empty balance)
		postDepositState = postDepositStates[moduleAddr]
		postProposalState := postProposalStates[moduleAddr]
		require.True(suite.T(), postProposalState[0].Amount.Equal(sdk.NewInt(0)), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postProposalState[1].Amount.Equal(postDepositState[1].Amount), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postProposalState[2].Amount.Equal(postDepositState[2].Amount), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postProposalState[3].Amount.Equal(postDepositState[3].Amount), fmt.Sprintf("stage %d", i))
		// Destination address should have received the deposits
		postDepositState = postDepositStates[destAddr]
		postProposalState = postProposalStates[destAddr]
		require.True(suite.T(), postProposalState[0].Amount.Equal(postDepositState[0].Amount.Add(sdk.NewInt(totalDepositedAmount))), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postProposalState[1].Amount.Equal(postDepositState[1].Amount), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postProposalState[2].Amount.Equal(postDepositState[2].Amount), fmt.Sprintf("stage %d", i))
		require.True(suite.T(), postProposalState[3].Amount.Equal(postDepositState[3].Amount), fmt.Sprintf("stage %d", i))
		// Control supply (mint should have increased by the minted amount)
		postProposalSupplies := []sdk.Coin{
			app.BankKeeper.GetSupply(ctx, params.DepositDenom),
			app.BankKeeper.GetSupply(ctx, params.MintDenom),
		}
		require.True(suite.T(), postProposalSupplies[0].Equal(postDepositsSupplies[0]))
		require.True(suite.T(), postProposalSupplies[1].Amount.Equal(postDepositsSupplies[1].Amount.Add(totalMintAmount.MulRaw(stage.mintRate))))

	}

	postRunStates := getAllStates()
	postRunSupplies := []sdk.Coin{
		app.BankKeeper.GetSupply(ctx, params.DepositDenom),
		app.BankKeeper.GetSupply(ctx, params.MintDenom),
	}

	// Check final state compared to our stage set
	// Simulate deposits and proposals to the preRun states
	for _, stage := range stages {
		for d, depositor := range stage.depositorsAddrs {
			// Simulate deposit
			preRunStates[depositor][0].Amount = preRunStates[depositor][0].Amount.Sub(sdk.NewInt(stage.depositorsAmounts[d]))
			preRunStates[depositor][2].Amount = preRunStates[depositor][2].Amount.Add(sdk.NewInt(stage.depositorsAmounts[d]))
		}
		for _, depositor := range depositorsAddrs {
			// Simulate proposal mint
			preRunStates[depositor][1].Amount = preRunStates[depositor][1].Amount.Add(preRunStates[depositor][3].Amount.MulRaw(stage.mintRate))
			preRunSupplies[1].Amount = preRunSupplies[1].Amount.Add(preRunStates[depositor][3].Amount.MulRaw(stage.mintRate))
			// Simulate proposal send
			preRunStates[destAddr][0].Amount = preRunStates[destAddr][0].Amount.Add(preRunStates[depositor][2].Amount)
			preRunStates[depositor][3].Amount = preRunStates[depositor][2].Amount
			preRunStates[depositor][2].Amount = sdk.NewInt(0)
		}
	}
	// Compare simulation to actual results
	for addr := range postRunStates {
		for i := range postRunStates[addr] {
			require.True(suite.T(), postRunStates[addr][i].Equal(preRunStates[addr][i]), fmt.Sprintf("Compare %s balance for state coin[%d]", addr, i))
		}
	}
	require.True(suite.T(), postRunSupplies[0].Equal(preRunSupplies[0]))
	require.True(suite.T(), postRunSupplies[0].Equal(preRunSupplies[0]))
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
