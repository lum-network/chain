package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gogotypes "github.com/cosmos/gogoproto/types"

	dFractKeeper "github.com/lum-network/chain/x/dfract/keeper"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
)

// TestMsgServer_DepositEnablement tests if a user can deposit based on its enablement
func (suite *KeeperTestSuite) TestMsgServer_DepositEnablement() {
	app := suite.app
	ctx := suite.ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := dFractKeeper.NewMsgServerImpl(*app.DFractKeeper)
	var emptyDepositDenoms []string
	// Obtain the required accounts
	depositor := suite.addrs[0]
	// Simulate that the deposit mode is not enabled
	err := app.DFractKeeper.UpdateParams(ctx, depositor.String(), &gogotypes.BoolValue{Value: false}, emptyDepositDenoms, nil)
	suite.Require().NoError(err)
	params := app.DFractKeeper.GetParams(ctx)

	_, err = msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenoms[0], sdk.NewInt(100000000)),
	})
	suite.Require().ErrorIs(err, dfracttypes.ErrDepositNotEnabled)

	err = app.DFractKeeper.UpdateParams(ctx, "", &gogotypes.BoolValue{Value: true}, emptyDepositDenoms, nil)
	suite.Require().NoError(err)

	_, err = msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(params.DepositDenoms[0], sdk.NewInt(100000000)),
	})
	suite.Require().NoError(err)
}

// TestMsgServer_InvalidDenomDeposit rejects deposit with invalid denoms
func (suite *KeeperTestSuite) TestMsgServer_InvalidDenomDeposit() {
	app := suite.app
	ctx := suite.ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := dFractKeeper.NewMsgServerImpl(*app.DFractKeeper)

	// Obtain the required accounts
	depositor := suite.addrs[0]

	// TestInvalidDenomDeposit
	// Try to deposit 100000000 of the mint denom
	_, err := msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin(dfracttypes.MintDenom, sdk.NewInt(100000000)),
	})
	suite.Require().Error(err)
	suite.Require().Equal(err, dfracttypes.ErrUnauthorizedDepositDenom)

	// Try to deposit 100000000 of another denom != from the mintDenom
	_, err = msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
		DepositorAddress: depositor.String(),
		Amount:           sdk.NewCoin("uatom", sdk.NewInt(100000000)),
	})
	suite.Require().Error(err)
	suite.Require().Equal(err, dfracttypes.ErrUnauthorizedDepositDenom)
}

// TestMsgServer_InvalidAmountDeposit rejects invalid deposit amount
func (suite *KeeperTestSuite) TestMsgServer_InvalidAmountDeposit() {
	app := suite.app
	ctx := suite.ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := dFractKeeper.NewMsgServerImpl(*app.DFractKeeper)
	params := app.DFractKeeper.GetParams(ctx)

	// Obtain the required accounts
	depositor := suite.addrs[0]

	// Iterate over the array of deposit denoms
	for _, denom := range params.DepositDenoms {
		// Try to deposit 0 of the deposit denom
		_, err := msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
			DepositorAddress: depositor.String(),
			Amount:           sdk.NewCoin(denom, sdk.NewInt(0)),
		})
		suite.Require().Error(err)
		suite.Require().Equal(err, dfracttypes.ErrEmptyDepositAmount)

		// Try to deposit below the min deposit amount
		_, err = msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
			DepositorAddress: depositor.String(),
			Amount:           sdk.NewCoin(denom, sdk.NewInt(int64(params.MinDepositAmount)-1)),
		})
		suite.Require().Error(err)
		suite.Require().Equal(err, dfracttypes.ErrInsufficientDepositAmount)
	}
}

// TestMsgServer_DoubleDeposit validates that a depositor can multiple times
func (suite *KeeperTestSuite) TestMsgServer_DoubleDeposit() {
	app := suite.app
	ctx := suite.ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := dFractKeeper.NewMsgServerImpl(*app.DFractKeeper)
	params := app.DFractKeeper.GetParams(ctx)

	// Obtain the required accounts
	depositor := suite.addrs[0]

	// Iterate over the array of deposit denoms
	for _, denom := range params.DepositDenoms {
		// Try to deposit 100000000 of the deposit denom
		_, err := msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
			DepositorAddress: depositor.String(),
			Amount:           sdk.NewCoin(denom, sdk.NewInt(100000000)),
		})
		suite.Require().NoError(err)

		// Try to deposit 100000000 of the deposit denom
		_, err = msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
			DepositorAddress: depositor.String(),
			Amount:           sdk.NewCoin(denom, sdk.NewInt(100000000)),
		})
		suite.Require().NoError(err)

		// The total final deposit should reflect the two deposits done
		deposit, found := app.DFractKeeper.GetDepositPendingWithdrawal(ctx, depositor)
		suite.Require().True(found)
		suite.Require().Equal(deposit.Amount, sdk.NewCoin(denom, sdk.NewInt(200000000)))
	}
}

// TestMsgServer_ValidDeposit validates deposit
func (suite *KeeperTestSuite) TestMsgServer_ValidDeposit() {
	app := suite.app
	ctx := suite.ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := dFractKeeper.NewMsgServerImpl(*app.DFractKeeper)
	params := app.DFractKeeper.GetParams(ctx)

	// Obtain the required accounts
	depositor := suite.addrs[0]

	// Iterate over the array of deposit denoms
	for _, denom := range params.DepositDenoms {
		// We store the initial depositor balances
		depositorAvailableBalance := app.BankKeeper.GetBalance(ctx, depositor, denom)
		suite.Require().GreaterOrEqual(depositorAvailableBalance.Amount.Int64(), int64(300000000))
		depositorMintedBalance := app.BankKeeper.GetBalance(ctx, depositor, dfracttypes.MintDenom)
		suite.Require().Equal(depositorMintedBalance.Amount.Int64(), int64(0))

		// We try to deposit 100000000 of the deposit denom
		_, err := msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
			DepositorAddress: depositor.String(),
			Amount:           sdk.NewCoin(denom, sdk.NewInt(100000000)),
		})
		suite.Require().NoError(err)

		// Depositor balance should reflect this transfer
		depositorAvailableBalance = app.BankKeeper.GetBalance(ctx, depositor, denom)
		suite.Require().Equal(depositorAvailableBalance.Amount.Int64(), int64(200000000))

		// Depositor should have its deposit available in the pending withdrawal queue
		deposit, found := app.DFractKeeper.GetDepositPendingWithdrawal(ctx, depositor)
		suite.Require().True(found)
		suite.Require().Equal(depositor.String(), deposit.DepositorAddress)
		suite.Require().Equal(deposit.Amount, sdk.NewCoin(denom, sdk.NewInt(100000000)))

		// Making another deposit should have the same effect
		_, err = msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
			DepositorAddress: depositor.String(),
			Amount:           sdk.NewCoin(denom, sdk.NewInt(100000000)),
		})
		suite.Require().NoError(err)
		depositorAvailableBalance = app.BankKeeper.GetBalance(ctx, depositor, denom)
		suite.Require().Equal(depositorAvailableBalance.Amount.Int64(), int64(100000000))
		deposit, found = app.DFractKeeper.GetDepositPendingWithdrawal(ctx, depositor)
		suite.Require().True(found)
		suite.Require().Equal(depositor.String(), deposit.DepositorAddress)
		suite.Require().Equal(deposit.Amount, sdk.NewCoin(denom, sdk.NewInt(200000000)))
	}
}

// TestMsgServer_MintAccuracy tests if the msg server WithdrawAndMint is accurately triggering a mint process
func (suite *KeeperTestSuite) TestMsgServer_MintAccuracy() {
	app := suite.app
	ctx := suite.ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := dFractKeeper.NewMsgServerImpl(*app.DFractKeeper)
	var emptyDepositDenoms []string
	err := app.DFractKeeper.UpdateParams(ctx, suite.addrs[0].String(), &gogotypes.BoolValue{Value: true}, emptyDepositDenoms, nil)
	suite.Require().NoError(err)
	params := app.DFractKeeper.GetParams(ctx)

	// Iterate over the array of deposit denoms
	for _, denom := range params.DepositDenoms {
		testAccuracy := func(depositor sdk.AccAddress, depositAmount int64, microMintRate int64, expectedMintedAmount int64) {
			balanceBeforeMint := app.BankKeeper.GetBalance(ctx, depositor, dfracttypes.MintDenom)
			app.DFractKeeper.SetDepositPendingMint(ctx, depositor, dfracttypes.Deposit{
				DepositorAddress: depositor.String(),
				Amount:           sdk.NewCoin(denom, sdk.NewInt(depositAmount)),
			})
			// make sure the signer is the management address
			_, err := msgServer.WithdrawAndMint(goCtx, &dfracttypes.MsgWithdrawAndMint{
				Address:       suite.addrs[1].String(),
				MicroMintRate: microMintRate,
			})
			suite.Require().ErrorIs(err, dfracttypes.ErrInvalidSignerAddress)

			_, err = msgServer.WithdrawAndMint(goCtx, &dfracttypes.MsgWithdrawAndMint{
				Address:       params.GetWithdrawalAddress(),
				MicroMintRate: microMintRate,
			})
			suite.Require().NoError(err)
			balance := app.BankKeeper.GetBalance(ctx, depositor, dfracttypes.MintDenom)
			suite.Require().Equal(expectedMintedAmount, balance.Amount.Int64()-balanceBeforeMint.Amount.Int64())
		}

		// Test micro minting
		testAccuracy(suite.addrs[1], 1_000_000, 2, 2)
		testAccuracy(suite.addrs[1], 1_000_001, 2, 2)
		testAccuracy(suite.addrs[1], 1_000_010, 2, 2)
		testAccuracy(suite.addrs[1], 1_000_100, 2, 2)
		testAccuracy(suite.addrs[1], 1_000_100, 2000, 2000)
		testAccuracy(suite.addrs[1], 1_001_000, 2000, 2002)

		// Test mint rate <= 1.0
		testAccuracy(suite.addrs[1], 1_000_000, 1_000_000, 1_000_000)
		testAccuracy(suite.addrs[1], 1_000_000, 500_000, 500_000)
		testAccuracy(suite.addrs[1], 1_000_000, 250, 250)
		// Test mint rate > 1.0
		testAccuracy(suite.addrs[1], 1_000_000, 2_000_000, 2_000_000)
		testAccuracy(suite.addrs[1], 1_234_567, 2_000_000, 2_469_134)
		testAccuracy(suite.addrs[1], 5_000_000, 200_000_000, 10_00_000_000)
		// Test large amount mint
		testAccuracy(suite.addrs[1], 100_000_000, 10_000_000_000_000_000, 1_000_000_000_000_000_000)
	}
}

// TestMsgServer_ChainedFullProcess tests the full dfract module chain process
func (suite *KeeperTestSuite) TestMsgServer_ChainedFullProcess() {
	app := suite.app
	ctx := suite.ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := dFractKeeper.NewMsgServerImpl(*app.DFractKeeper)
	// Simulate that the address management is set
	var emptyDepositDenoms []string
	err := app.DFractKeeper.UpdateParams(ctx, suite.addrs[0].String(), &gogotypes.BoolValue{Value: true}, emptyDepositDenoms, nil)
	suite.Require().NoError(err)
	params := app.DFractKeeper.GetParams(ctx)
	withdrawAddr := params.WithdrawalAddress
	moduleAddr := app.DFractKeeper.GetModuleAccount(ctx).String()
	depositorsAddrs := []string{suite.addrs[1].String(), suite.addrs[2].String(), suite.addrs[3].String(), suite.addrs[4].String(), suite.addrs[5].String()}

	suite.Require().Equal(5, len(depositorsAddrs))

	type stageDef struct {
		depositorsAddrs   []string
		depositorsAmounts []int64
		microMintRate     int64
	}

	// Build stages to run with voluntary empty stages
	stages := []stageDef{{
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		microMintRate:     2,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		microMintRate:     3,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		microMintRate:     4,
	}, {
		depositorsAddrs:   []string{depositorsAddrs[0]},
		depositorsAmounts: []int64{1000000},
		microMintRate:     5,
	}, {
		depositorsAddrs:   []string{depositorsAddrs[0], depositorsAddrs[1]},
		depositorsAmounts: []int64{2000000, 3000000},
		microMintRate:     6,
	}, {
		depositorsAddrs:   []string{depositorsAddrs[0], depositorsAddrs[1], depositorsAddrs[2], depositorsAddrs[3]},
		depositorsAmounts: []int64{4000000, 5000000, 6000000, 7000000},
		microMintRate:     7,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		microMintRate:     8,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		microMintRate:     9,
	}, {
		depositorsAddrs:   []string{depositorsAddrs[1], depositorsAddrs[2], depositorsAddrs[3], depositorsAddrs[4]},
		depositorsAmounts: []int64{8000000, 9000000, 10000000, 11000000},
		microMintRate:     10,
	}, {
		depositorsAddrs:   []string{depositorsAddrs[3], depositorsAddrs[3], depositorsAddrs[4], depositorsAddrs[4], depositorsAddrs[4]},
		depositorsAmounts: []int64{12000000, 13000000, 14000000, 15000000, 16000000},
		microMintRate:     11,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		microMintRate:     12,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		microMintRate:     13,
	}, {
		depositorsAddrs:   []string{},
		depositorsAmounts: []int64{},
		microMintRate:     14,
	}}

	// getAllStates acquires all relevant balances
	// use map["module"] for module balances
	// coin[0]: deposit balance
	// coin[1]: minted balance
	// coin[2]: pending withdrawal balance
	// coin[3]: pending mint balance
	// coin[4]: deposit minted

	// Iterate over the array of deposit denoms
	for _, denom := range params.DepositDenoms {
		getAllStates := func() map[string][]sdk.Coin {
			res := map[string][]sdk.Coin{}
			addrs := append(depositorsAddrs, withdrawAddr, app.DFractKeeper.GetModuleAccount(ctx).String())

			for _, addr := range addrs {
				accAddr, err := sdk.AccAddressFromBech32(addr)
				suite.Require().NoError(err)
				res[addr] = []sdk.Coin{
					app.BankKeeper.GetBalance(ctx, accAddr, denom),
					app.BankKeeper.GetBalance(ctx, accAddr, dfracttypes.MintDenom),
				}
				deposit, found := app.DFractKeeper.GetDepositPendingWithdrawal(ctx, accAddr)
				if found {
					res[addr] = append(res[addr], deposit.Amount)
				} else {
					res[addr] = append(res[addr], sdk.NewCoin(denom, sdk.NewInt(0)))
				}
				deposit, found = app.DFractKeeper.GetDepositPendingMint(ctx, accAddr)
				if found {
					res[addr] = append(res[addr], deposit.Amount)
				} else {
					res[addr] = append(res[addr], sdk.NewCoin(denom, sdk.NewInt(0)))
				}
				deposit, found = app.DFractKeeper.GetDepositMinted(ctx, accAddr)
				if found {
					res[addr] = append(res[addr], deposit.Amount)
				} else {
					res[addr] = append(res[addr], sdk.NewCoin(denom, sdk.NewInt(0)))
				}
			}

			return res
		}

		preRunStates := getAllStates()
		preRunSupplies := []sdk.Coin{
			app.BankKeeper.GetSupply(ctx, denom),
			app.BankKeeper.GetSupply(ctx, dfracttypes.MintDenom),
		}

		for i, stage := range stages {
			// Acquire current accounts balances states
			initialStates := getAllStates()
			initialSupplies := []sdk.Coin{
				app.BankKeeper.GetSupply(ctx, denom),
				app.BankKeeper.GetSupply(ctx, dfracttypes.MintDenom),
			}
			suite.Require().Len(initialStates, len(depositorsAddrs)+2)

			// Make deposits
			var totalDepositedAmount int64
			depositedAmounts := map[string]int64{}
			for d, depositor := range stage.depositorsAddrs {
				_, err := msgServer.Deposit(goCtx, &dfracttypes.MsgDeposit{
					DepositorAddress: depositor,
					Amount:           sdk.NewCoin(denom, sdk.NewInt(stage.depositorsAmounts[d])),
				})
				suite.Require().NoError(err, fmt.Sprintf("stage %d", i))
				depositedAmounts[depositor] += stage.depositorsAmounts[d]
				totalDepositedAmount += stage.depositorsAmounts[d]
			}

			// Control new account states
			postDepositStates := getAllStates()
			suite.Require().Len(postDepositStates, len(depositorsAddrs)+2)
			// Check depositors and non depositors balances changes
			for _, depositor := range depositorsAddrs {
				initialState := initialStates[depositor]
				postDepositState := postDepositStates[depositor]
				if depositedAmounts[depositor] > 0 {
					// Deposited amount(s) should have been removed from deposit balance
					suite.Require().True(postDepositState[0].Amount.Equal(initialState[0].Amount.Sub(sdk.NewInt(depositedAmounts[depositor]))), fmt.Sprintf("stage %d", i))
					// Minted balance should not have changed
					suite.Require().True(postDepositState[1].Amount.Equal(initialState[1].Amount), fmt.Sprintf("stage %d", i))
					// Pending withdrawal balance should have inrease by the deposited amount
					suite.Require().True(postDepositState[2].Amount.Equal(initialState[2].Amount.Add(sdk.NewInt(depositedAmounts[depositor]))), fmt.Sprintf("stage %d", i))
					// Pending mint balance should not have changed
					suite.Require().True(postDepositState[3].Amount.Equal(initialState[3].Amount), fmt.Sprintf("stage %d", i))
					// Minted balance should not have changed
					suite.Require().True(postDepositState[4].Amount.Equal(initialState[4].Amount), fmt.Sprintf("stage %d", i))
				} else {
					// Non depositor should have no change
					suite.Require().True(postDepositState[0].Amount.Equal(initialState[0].Amount), fmt.Sprintf("stage %d", i))
					suite.Require().True(postDepositState[1].Amount.Equal(initialState[1].Amount), fmt.Sprintf("stage %d", i))
					suite.Require().True(postDepositState[2].Amount.Equal(initialState[2].Amount), fmt.Sprintf("stage %d", i))
					suite.Require().True(postDepositState[3].Amount.Equal(initialState[3].Amount), fmt.Sprintf("stage %d", i))
					suite.Require().True(postDepositState[4].Amount.Equal(initialState[4].Amount), fmt.Sprintf("stage %d", i))
				}
			}
			// Module account should have received the new deposits
			initialState := initialStates[moduleAddr]
			postDepositState := postDepositStates[moduleAddr]
			suite.Require().True(postDepositState[0].Amount.Equal(initialState[0].Amount.Add(sdk.NewInt(totalDepositedAmount))), fmt.Sprintf("stage %d", i))
			suite.Require().True(postDepositState[1].Amount.Equal(initialState[1].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postDepositState[2].Amount.Equal(initialState[2].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postDepositState[3].Amount.Equal(initialState[3].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postDepositState[4].Amount.Equal(initialState[4].Amount), fmt.Sprintf("stage %d", i))
			// Destination address should not have changed
			initialState = initialStates[withdrawAddr]
			postDepositState = postDepositStates[withdrawAddr]
			suite.Require().True(postDepositState[0].Amount.Equal(initialState[0].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postDepositState[1].Amount.Equal(initialState[1].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postDepositState[2].Amount.Equal(initialState[2].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postDepositState[3].Amount.Equal(initialState[3].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postDepositState[4].Amount.Equal(initialState[4].Amount), fmt.Sprintf("stage %d", i))
			// Control supply (no change expected)
			postDepositsSupplies := []sdk.Coin{
				app.BankKeeper.GetSupply(ctx, denom),
				app.BankKeeper.GetSupply(ctx, dfracttypes.MintDenom),
			}
			suite.Require().True(postDepositsSupplies[0].Equal(initialSupplies[0]))
			suite.Require().True(postDepositsSupplies[1].Equal(initialSupplies[1]))

			// Run the tx
			_, err := msgServer.WithdrawAndMint(goCtx, &dfracttypes.MsgWithdrawAndMint{
				Address:       params.GetWithdrawalAddress(),
				MicroMintRate: stage.microMintRate,
			})
			suite.Require().NoError(err)

			// Control new account states
			postProposalStates := getAllStates()
			suite.Require().Len(postProposalStates, len(depositorsAddrs)+2)
			totalMintAmount := sdk.NewInt(0)
			// Check depositors and non depositors balances changes
			for _, depositor := range depositorsAddrs {
				postDepositState := postDepositStates[depositor]
				postProposalState := postProposalStates[depositor]
				// Deposit balance should not have change since the deposit
				suite.Require().True(postProposalState[0].Amount.Equal(postDepositState[0].Amount), fmt.Sprintf("stage %d", i))
				// Minted balance should have received the amount available in the waiting mint * microMintRate
				suite.Require().True(postProposalState[1].Amount.Equal(postDepositState[1].Amount.Add(postDepositState[3].Amount.MulRaw(stage.microMintRate).QuoRaw(dFractKeeper.MicroPrecision))), fmt.Sprintf("stage %d", i))
				// Pending withdrawal balance should be empty
				suite.Require().True(postProposalState[2].Amount.Equal(sdk.NewInt(0)), fmt.Sprintf("stage %d", i))
				// Pending mint balance should have received the amount in the pending withdrawal balance
				suite.Require().True(postProposalState[3].Amount.Equal(postDepositState[2].Amount), fmt.Sprintf("stage %d", i))
				// Minted balance should have increased by the amount in the pending mint balance
				suite.Require().True(postProposalState[4].Amount.Equal(postDepositState[4].Amount.Add(postDepositState[3].Amount)), fmt.Sprintf("stage %d", i))
				totalMintAmount = totalMintAmount.Add(postDepositState[3].Amount)
			}
			// Module account should have sent the new deposits to the withdrawAddr (empty balance)
			postDepositState = postDepositStates[moduleAddr]
			postProposalState := postProposalStates[moduleAddr]
			suite.Require().True(postProposalState[0].Amount.Equal(sdk.NewInt(0)), fmt.Sprintf("stage %d", i))
			suite.Require().True(postProposalState[1].Amount.Equal(postDepositState[1].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postProposalState[2].Amount.Equal(postDepositState[2].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postProposalState[3].Amount.Equal(postDepositState[3].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postProposalState[4].Amount.Equal(postDepositState[4].Amount), fmt.Sprintf("stage %d", i))
			// Withdrawal address should have received the deposits
			postDepositState = postDepositStates[withdrawAddr]
			postProposalState = postProposalStates[withdrawAddr]
			suite.Require().True(postProposalState[0].Amount.Equal(postDepositState[0].Amount.Add(sdk.NewInt(totalDepositedAmount))), fmt.Sprintf("stage %d", i))
			suite.Require().True(postProposalState[1].Amount.Equal(postDepositState[1].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postProposalState[2].Amount.Equal(postDepositState[2].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postProposalState[3].Amount.Equal(postDepositState[3].Amount), fmt.Sprintf("stage %d", i))
			suite.Require().True(postProposalState[4].Amount.Equal(postDepositState[4].Amount), fmt.Sprintf("stage %d", i))
			// Control supply (mint should have increased by the minted amount)
			postProposalSupplies := []sdk.Coin{
				app.BankKeeper.GetSupply(ctx, denom),
				app.BankKeeper.GetSupply(ctx, dfracttypes.MintDenom),
			}
			suite.Require().True(postProposalSupplies[0].Equal(postDepositsSupplies[0]))
			suite.Require().True(postProposalSupplies[1].Amount.Equal(postDepositsSupplies[1].Amount.Add(totalMintAmount.MulRaw(stage.microMintRate).QuoRaw(dFractKeeper.MicroPrecision))))
		}

		postRunStates := getAllStates()
		postRunSupplies := []sdk.Coin{
			app.BankKeeper.GetSupply(ctx, denom),
			app.BankKeeper.GetSupply(ctx, dfracttypes.MintDenom),
		}
		// We expect the module account balances to be completely empty
		suite.Require().Equal(int64(0), postRunStates[moduleAddr][0].Amount.Int64())
		suite.Require().Equal(int64(0), postRunStates[moduleAddr][0].Amount.Int64())
		suite.Require().Equal(int64(0), postRunStates[moduleAddr][0].Amount.Int64())
		suite.Require().Equal(int64(0), postRunStates[moduleAddr][0].Amount.Int64())

		// Check final state compared to our stage set
		// Simulate deposits and proposals to the preRun states
		for _, stage := range stages {
			for d, depositor := range stage.depositorsAddrs {
				// Simulate deposit
				preRunStates[depositor][0].Amount = preRunStates[depositor][0].Amount.Sub(sdk.NewInt(stage.depositorsAmounts[d]))
				preRunStates[depositor][2].Amount = preRunStates[depositor][2].Amount.Add(sdk.NewInt(stage.depositorsAmounts[d]))
			}
			for _, depositor := range depositorsAddrs {
				// Simulate proposal mint stage
				preRunStates[depositor][1].Amount = preRunStates[depositor][1].Amount.Add(preRunStates[depositor][3].Amount.MulRaw(stage.microMintRate).QuoRaw(dFractKeeper.MicroPrecision))
				preRunSupplies[1].Amount = preRunSupplies[1].Amount.Add(preRunStates[depositor][3].Amount.MulRaw(stage.microMintRate).QuoRaw(dFractKeeper.MicroPrecision))
				// Simulate proposal withdraw stage
				preRunStates[withdrawAddr][0].Amount = preRunStates[withdrawAddr][0].Amount.Add(preRunStates[depositor][2].Amount)
				preRunStates[depositor][3].Amount = preRunStates[depositor][2].Amount
				preRunStates[depositor][2].Amount = sdk.NewInt(0)
				preRunStates[depositor][4].Amount = preRunStates[depositor][4].Amount.Add(preRunStates[depositor][3].Amount)
			}
		}
		// Compare simulation to actual results
		for addr := range postRunStates {
			for i := range postRunStates[addr] {
				suite.Require().True(postRunStates[addr][i].Equal(preRunStates[addr][i]), fmt.Sprintf("Compare %s balance for state coin[%d]", addr, i))
			}
		}
		suite.Require().True(postRunSupplies[0].Equal(preRunSupplies[0]))
		suite.Require().True(postRunSupplies[0].Equal(preRunSupplies[0]))
	}
}
