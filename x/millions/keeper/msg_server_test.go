package keeper_test

import (
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	apptesting "github.com/lum-network/chain/app/testing"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

// TestMsgServer_DrawRetry runs draw retry related tests
func (suite *KeeperTestSuite) TestMsgServer_DrawRetry() {
	// Set the app context
	app := suite.App
	ctx := suite.Ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)

	// Add 5 pools
	for i := 0; i < 5; i++ {
		poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
		drawDelta1 := 1 * time.Hour
		app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
			PoolId: poolID,
			PrizeStrategy: millionstypes.PrizeStrategy{
				PrizeBatches: []millionstypes.PrizeBatch{
					{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
				},
			},
			DrawSchedule: millionstypes.DrawSchedule{
				InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
				DrawDelta:     drawDelta1,
			},
			AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
		}))
	}

	// Retrieve the pool from the state
	pools := app.MillionsKeeper.ListPools(ctx)

	// Create a new draw
	draw1 := millionstypes.Draw{
		PoolId:                 pools[0].PoolId,
		DrawId:                 uint64(10),
		State:                  millionstypes.DrawState_Failure,
		ErrorState:             millionstypes.DrawState_IcaWithdrawRewards,
		PrizePoolRemainsAmount: sdk.NewInt(1_000_000),
		PrizePool:              sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	}
	app.MillionsKeeper.SetPoolDraw(ctx, draw1)

	// Test GetPoolDraw
	draw, err := app.MillionsKeeper.GetPoolDraw(ctx, draw1.PoolId, draw1.DrawId)
	suite.Require().NoError(err)
	// - Test GetPoolDraw with wrong poolID
	_, err = app.MillionsKeeper.GetPoolDraw(ctx, uint64(0), draw.DrawId)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolDrawNotFound)
	// - Test GetPoolDraw with wrong drawID
	_, err = app.MillionsKeeper.GetPoolDraw(ctx, draw.PoolId, uint64(0))
	suite.Require().ErrorIs(err, millionstypes.ErrPoolDrawNotFound)

	// Test ValidateDrawRetryBasic
	// - Test unknown poolID
	_, err = msgServer.DrawRetry(goCtx, &millionstypes.MsgDrawRetry{
		PoolId: uint64(0),
		DrawId: draw.DrawId,
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)
	// - Test unknown drawID
	_, err = msgServer.DrawRetry(goCtx, &millionstypes.MsgDrawRetry{
		PoolId: draw.PoolId,
		DrawId: uint64(0),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)

	// - Test unknown drawID
	_, err = msgServer.DrawRetry(goCtx, &millionstypes.MsgDrawRetry{
		PoolId: uint64(10),
		DrawId: draw.DrawId,
	})
	suite.Require().ErrorIs(err, millionstypes.ErrPoolNotFound)

	draw, err = app.MillionsKeeper.GetPoolDraw(ctx, draw.PoolId, draw.DrawId)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DrawState_Failure, draw.State)
	suite.Require().Equal(millionstypes.DrawState_IcaWithdrawRewards, draw.ErrorState)

	pools = app.MillionsKeeper.ListPools(ctx)

	// Create a new draw to test DrawState_IcaWithdrawRewards ErrorState
	draw2 := millionstypes.Draw{
		PoolId:                 pools[1].PoolId,
		DrawId:                 uint64(20),
		State:                  millionstypes.DrawState_Failure,
		ErrorState:             millionstypes.DrawState_IcaWithdrawRewards,
		PrizePoolRemainsAmount: sdk.NewInt(1_000_000),
		PrizePool:              sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	}
	app.MillionsKeeper.SetPoolDraw(ctx, draw2)

	draw2, err = app.MillionsKeeper.GetPoolDraw(ctx, draw2.PoolId, draw2.DrawId)
	suite.Require().NoError(err)

	_, err = msgServer.DrawRetry(goCtx, &millionstypes.MsgDrawRetry{
		PoolId: draw2.PoolId,
		DrawId: draw2.DrawId,
	})
	suite.Require().NoError(err)

	draw2, err = app.MillionsKeeper.GetPoolDraw(ctx, draw2.PoolId, draw2.DrawId)
	suite.Require().NoError(err)

	// Test that the msg retry is success
	suite.Require().Equal(millionstypes.DrawState_Success, draw2.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw2.ErrorState)
	suite.Require().Equal(ctx.BlockHeight(), draw2.UpdatedAtHeight)
	suite.Require().Equal(ctx.BlockTime(), draw2.UpdatedAt)

	pools = app.MillionsKeeper.ListPools(ctx)

	// Create a new draw to test DrawState_IbcTransfer ErrorState
	draw3 := millionstypes.Draw{
		PoolId:                 pools[2].PoolId,
		DrawId:                 uint64(30),
		State:                  millionstypes.DrawState_Failure,
		ErrorState:             millionstypes.DrawState_IbcTransfer,
		PrizePoolRemainsAmount: sdk.NewInt(1_000_000),
		PrizePool:              sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	}
	app.MillionsKeeper.SetPoolDraw(ctx, draw3)

	draw3, err = app.MillionsKeeper.GetPoolDraw(ctx, draw3.PoolId, draw3.DrawId)
	suite.Require().NoError(err)

	_, err = msgServer.DrawRetry(goCtx, &millionstypes.MsgDrawRetry{
		PoolId: draw3.PoolId,
		DrawId: draw3.DrawId,
	})
	suite.Require().NoError(err)

	draw3, err = app.MillionsKeeper.GetPoolDraw(ctx, draw3.PoolId, draw3.DrawId)
	suite.Require().NoError(err)

	// Test that the msg retry is success
	suite.Require().Equal(millionstypes.DrawState_Success, draw3.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw3.ErrorState)
	suite.Require().Equal(ctx.BlockHeight(), draw3.UpdatedAtHeight)
	suite.Require().Equal(ctx.BlockTime(), draw3.UpdatedAt)

	// Create a new draw to test DrawState_Drawing ErrorState
	draw4 := millionstypes.Draw{
		PoolId:                 pools[3].PoolId,
		DrawId:                 uint64(40),
		State:                  millionstypes.DrawState_Failure,
		ErrorState:             millionstypes.DrawState_Drawing,
		PrizePoolRemainsAmount: sdk.NewInt(1_000_000),
		PrizePool:              sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	}
	app.MillionsKeeper.SetPoolDraw(ctx, draw4)

	draw4, err = app.MillionsKeeper.GetPoolDraw(ctx, draw4.PoolId, draw4.DrawId)
	suite.Require().NoError(err)

	_, err = msgServer.DrawRetry(goCtx, &millionstypes.MsgDrawRetry{
		PoolId: draw4.PoolId,
		DrawId: draw4.DrawId,
	})
	suite.Require().NoError(err)

	draw4, err = app.MillionsKeeper.GetPoolDraw(ctx, draw4.PoolId, draw4.DrawId)
	suite.Require().NoError(err)

	suite.Require().Equal(millionstypes.DrawState_Success, draw4.State)
	suite.Require().Equal(millionstypes.DrawState_Unspecified, draw4.ErrorState)
	suite.Require().Equal(ctx.BlockHeight(), draw4.UpdatedAtHeight)
	suite.Require().Equal(ctx.BlockTime(), draw4.UpdatedAt)

	// Create a new draw to test DrawState_Unspecified ErrorState
	draw5 := millionstypes.Draw{
		PoolId:                 pools[4].PoolId,
		DrawId:                 uint64(50),
		State:                  millionstypes.DrawState_Failure,
		ErrorState:             millionstypes.DrawState_Unspecified,
		PrizePoolRemainsAmount: sdk.NewInt(1_000_000),
		PrizePool:              sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	}
	app.MillionsKeeper.SetPoolDraw(ctx, draw5)

	draw5, err = app.MillionsKeeper.GetPoolDraw(ctx, draw5.PoolId, draw5.DrawId)
	suite.Require().NoError(err)

	_, err = msgServer.DrawRetry(goCtx, &millionstypes.MsgDrawRetry{
		PoolId: draw5.PoolId,
		DrawId: draw5.DrawId,
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidDrawState)
}

// TestMsgServer_Deposit runs deposit related tests through ICA
// This test case is not intended to test deposit process (which is done on another case) but only the fact that it also works through ICA / IBC conditions
func (suite *KeeperTestSuite) TestMsgServer_Deposit_Remote() {
	app := suite.App
	ctx := suite.Ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)

	// Create pool ID and ICA channels
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	icaDepositPortName := string(millionstypes.NewPoolName(poolID, millionstypes.ICATypeDeposit))
	icaPrizepoolPortName := string(millionstypes.NewPoolName(poolID, millionstypes.ICATypePrizePool))
	suite.CreateICAChannel(icaDepositPortName)
	suite.CreateICAChannel(icaPrizepoolPortName)
	hostChainID := suite.HostChain.ChainID

	// Create a remote pool entity
	drawDelta1 := 1 * time.Hour
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:              poolID,
		ChainId:             hostChainID,
		IcaDepositPortId:    icaDepositPortName,
		IcaDepositAddress:   suite.ICAAddresses[icaDepositPortName],
		IcaPrizepoolPortId:  icaPrizepoolPortName,
		NativeDenom:         remotePoolDenom,
		Denom:               remotePoolDenomIBC,
		IcaPrizepoolAddress: suite.ICAAddresses[icaPrizepoolPortName],
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(remotePoolDenomIBC, math.NewInt(1000)),
	}))

	// Grab our pool entity
	pool, err := app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)

	// Make a working deposit and ensure no error
	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           pool.GetPoolId(),
		Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(int64(1_000_000))),
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().NoError(err)
}

// TestMsgServer_Deposit runs deposit related tests
func (suite *KeeperTestSuite) TestMsgServer_Deposit() {
	app := suite.App
	ctx := suite.Ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)

	// Create pool
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta1 := 1 * time.Hour
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: poolID,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	}))

	pools := app.MillionsKeeper.ListPools(ctx)

	// Invalid pool ID
	_, err := msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           uint64(0),
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolNotFound)

	// Invalid depositor address
	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           poolID,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(int64(1_000_000))),
		DepositorAddress: "",
	})
	suite.Require().Error(err)
	suite.Require().Equal(err, millionstypes.ErrInvalidDepositorAddress)

	// Invalid deposit denom
	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           poolID,
		Amount:           sdk.NewCoin("ufaketoken", sdk.NewInt(1_000_000)),
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidDepositDenom)

	// Invalid amount
	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           poolID,
		Amount:           sdk.NewCoin(localPoolDenom, pools[0].MinDepositAmount.Sub(sdk.NewInt(1))),
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, millionstypes.ErrInsufficientDepositAmount)

	// Invalid winner address
	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           poolID,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(int64(1_000_000))),
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    "no-address",
	})
	suite.Require().Error(err)
	suite.Require().Equal(err, millionstypes.ErrInvalidWinnerAddress)

	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           poolID,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(int64(123456))),
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[1].String(),
		IsSponsor:        true,
	})
	suite.Require().Error(err)
	suite.Require().Equal(err, millionstypes.ErrInvalidSponsorWinnerCombo)

	// Register real pool and apply real deposits
	poolID, err = app.MillionsKeeper.RegisterPool(ctx,
		"ulum",
		"ulum",
		testChainID,
		"",
		"",
		[]string{suite.valAddrs[0].String()},
		"lum",
		"lumvaloper",
		app.MillionsKeeper.GetParams(ctx).MinDepositAmount,
		millionstypes.DrawSchedule{DrawDelta: 24 * time.Hour, InitialDrawAt: ctx.BlockTime().Add(24 * time.Hour)},
		millionstypes.PrizeStrategy{PrizeBatches: []millionstypes.PrizeBatch{{PoolPercent: 100, Quantity: 100, DrawProbability: sdk.NewDec(1)}}},
	)
	suite.Require().NoError(err)
	_, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	// Create pool and deposits
	// Operate basic checks to verify if deposit is created
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: app.MillionsKeeper.GetNextPoolID(ctx),
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	}))
	pools = app.MillionsKeeper.ListPools(ctx)

	pool, err := app.MillionsKeeper.GetPool(ctx, pools[1].PoolId)
	suite.Require().NoError(err)

	balanceBefore := app.BankKeeper.GetBalance(ctx, suite.addrs[0], localPoolDenom)
	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           pool.GetPoolId(),
		Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(int64(1_000_000))),
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().NoError(err)
	balanceNow := app.BankKeeper.GetBalance(ctx, suite.addrs[0], localPoolDenom)
	suite.Require().Equal(balanceBefore.Amount.Int64()-1_000_000, balanceNow.Amount.Int64())

	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           pool.GetPoolId(),
		Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(int64(2_000_000))),
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[1].String(),
	})
	suite.Require().NoError(err)
	balanceNow = app.BankKeeper.GetBalance(ctx, suite.addrs[0], localPoolDenom)
	suite.Require().Equal(balanceBefore.Amount.Int64()-3_000_000, balanceNow.Amount.Int64())

	balanceBefore = app.BankKeeper.GetBalance(ctx, suite.addrs[2], localPoolDenom)
	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           pool.GetPoolId(),
		Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(int64(5_000_000))),
		DepositorAddress: suite.addrs[2].String(),
		IsSponsor:        true,
	})
	suite.Require().NoError(err)
	balanceNow = app.BankKeeper.GetBalance(ctx, suite.addrs[2], localPoolDenom)
	suite.Require().Equal(balanceBefore.Amount.Int64()-5_000_000, balanceNow.Amount.Int64())

	// Compare pool tvl and depositors count state
	suite.Require().Equal(sdk.NewInt(0), pool.TvlAmount)
	suite.Require().Equal(uint64(0), pool.DepositorsCount)
	pool, err = app.MillionsKeeper.GetPool(ctx, pool.GetPoolId())
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(8_000_000), pool.TvlAmount)
	suite.Require().Equal(uint64(2), pool.DepositorsCount)
	suite.Require().Equal(sdk.NewInt(int64(5_000_000)), pool.SponsorshipAmount)

	// Test deposit for a pool that is not in ready state
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: 4,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
		State:              millionstypes.PoolState_Paused,
	}))

	pool, err = app.MillionsKeeper.GetPool(ctx, 4)
	suite.Require().NoError(err)

	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		PoolId:           pool.PoolId,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(int64(1_000_000))),
		DepositorAddress: suite.addrs[0].String(),
	})
	// Deposit should be allowed for paused pools
	suite.Require().NoError(err)
	pool, err = app.MillionsKeeper.GetPool(ctx, 4)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(1_000_000), pool.TvlAmount.Int64())
}

// TestMsgServer_DepositRetry tests the retry of a failed deposit
func (suite *KeeperTestSuite) TestMsgServer_DepositRetry() {
	// Set the app context
	app := suite.App
	ctx := suite.Ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)

	// Set the denom
	denom := app.StakingKeeper.BondDenom(ctx)
	// Initialize the pool
	poolID, err := app.MillionsKeeper.RegisterPool(ctx,
		denom,
		denom,
		testChainID,
		"",
		"",
		[]string{suite.valAddrs[0].String()},
		"lum",
		"lumvaloper",
		app.MillionsKeeper.GetParams(ctx).MinDepositAmount,
		millionstypes.DrawSchedule{DrawDelta: 24 * time.Hour, InitialDrawAt: ctx.BlockTime().Add(24 * time.Hour)},
		millionstypes.PrizeStrategy{PrizeBatches: []millionstypes.PrizeBatch{{PoolPercent: 100, Quantity: 100, DrawProbability: sdk.NewDec(1)}}},
	)
	suite.Require().NoError(err)
	_, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	// Add 3 pools
	for i := 0; i < 3; i++ {
		poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
		drawDelta1 := 1 * time.Hour
		app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
			PoolId: poolID,
			PrizeStrategy: millionstypes.PrizeStrategy{
				PrizeBatches: []millionstypes.PrizeBatch{
					{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
				},
			},
			DrawSchedule: millionstypes.DrawSchedule{
				InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
				DrawDelta:     drawDelta1,
			},
			AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
		}))
	}

	// Retrieve the pool from the state
	pools := app.MillionsKeeper.ListPools(ctx)

	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[0].PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_Failure,
		ErrorState:       millionstypes.DepositState_IbcTransfer,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)

	deposits := app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	// - Test GetPoolDraw with wrong poolID
	_, err = app.MillionsKeeper.GetPoolDeposit(ctx, uint64(0), deposits[0].DepositId)
	suite.Require().ErrorIs(err, millionstypes.ErrDepositNotFound)
	// - Test GetPoolDraw with wrong drawID
	_, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, uint64(0))
	suite.Require().ErrorIs(err, millionstypes.ErrDepositNotFound)

	// Test validate Basics
	// - Test unknown poolID
	_, err = msgServer.DepositRetry(goCtx, &millionstypes.MsgDepositRetry{
		PoolId:           uint64(0),
		DepositId:        deposits[0].DepositId,
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)
	// - Test unknown depositID
	_, err = msgServer.DepositRetry(goCtx, &millionstypes.MsgDepositRetry{
		PoolId:           deposits[0].PoolId,
		DepositId:        uint64(0),
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)

	// Should throw error if depositor Addr msg is different than entity msg
	_, err = msgServer.DepositRetry(goCtx, &millionstypes.MsgDepositRetry{
		PoolId:           deposits[0].PoolId,
		DepositId:        deposits[0].DepositId,
		DepositorAddress: "different-address",
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidDepositorAddress)

	// Test DepositState_IbcTransfer ErrorState
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	_, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)

	_, err = msgServer.DepositRetry(goCtx, &millionstypes.MsgDepositRetry{
		PoolId:           deposits[0].PoolId,
		DepositId:        deposits[0].DepositId,
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().NoError(err)
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Equal(millionstypes.DepositState_Success, deposits[0].State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposits[0].ErrorState)

	// Test DepositState_IcaDelegate ErrorState
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[1].PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_Failure,
		ErrorState:       millionstypes.DepositState_IcaDelegate,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[1].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)

	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	_, err = msgServer.DepositRetry(goCtx, &millionstypes.MsgDepositRetry{
		PoolId:           deposits[1].PoolId,
		DepositId:        deposits[1].DepositId,
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().NoError(err)
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Equal(millionstypes.DepositState_Success, deposits[1].State)
	suite.Require().Equal(millionstypes.DepositState_Unspecified, deposits[1].ErrorState)

	// Test DepositState_Unspecified ErrorState
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[2].PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_Failure,
		ErrorState:       millionstypes.DepositState_Unspecified,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], suite.moduleAddrs[0], sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)

	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])

	_, err = msgServer.DepositRetry(goCtx, &millionstypes.MsgDepositRetry{
		PoolId:           deposits[2].PoolId,
		DepositId:        deposits[2].DepositId,
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidDepositState)
}

// TestMsgServer_DepositEdit tests the edition of a winnerAddr and sponsor
func (suite *KeeperTestSuite) TestMsgServer_DepositEdit() {
	// Set the app context
	app := suite.App
	ctx := suite.Ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)

	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta1 := 1 * time.Hour
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: poolID,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	}))

	// Retrieve the pool from the state
	pools := app.MillionsKeeper.ListPools(ctx)

	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pools[0].PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_Success,
		ErrorState:       millionstypes.DepositState_Unspecified,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)

	deposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, 1, 1)
	suite.Require().NoError(err)
	// - Test GetPoolDraw with wrong poolID
	_, err = app.MillionsKeeper.GetPoolDeposit(ctx, uint64(0), deposit.DepositId)
	suite.Require().ErrorIs(err, millionstypes.ErrDepositNotFound)
	// - Test GetPoolDraw with wrong drawID
	_, err = app.MillionsKeeper.GetPoolDeposit(ctx, deposit.PoolId, uint64(0))
	suite.Require().ErrorIs(err, millionstypes.ErrDepositNotFound)

	// Test validate Basics
	// - Test unknown poolID
	_, err = msgServer.DepositEdit(goCtx, &millionstypes.MsgDepositEdit{
		PoolId:           uint64(0),
		DepositId:        deposit.DepositId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		IsSponsor: &gogotypes.BoolValue{
			Value: false,
		},
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)
	// - Test unknown depositID
	_, err = msgServer.DepositEdit(goCtx, &millionstypes.MsgDepositEdit{
		PoolId:           deposit.PoolId,
		DepositId:        uint64(0),
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		IsSponsor: &gogotypes.BoolValue{
			Value: false,
		},
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)

	// Should throw error if depositor Addr msg is different than entity msg
	_, err = msgServer.DepositEdit(goCtx, &millionstypes.MsgDepositEdit{
		PoolId:           deposit.PoolId,
		DepositId:        deposit.DepositId,
		DepositorAddress: "different-address",
		WinnerAddress:    suite.addrs[0].String(),
		IsSponsor: &gogotypes.BoolValue{
			Value: false,
		},
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidDepositorAddress)

	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, 1, 1)
	suite.Require().NoError(err)

	// Should throw error for an invalid winner address
	_, err = msgServer.DepositEdit(goCtx, &millionstypes.MsgDepositEdit{
		PoolId:           deposit.PoolId,
		DepositId:        deposit.DepositId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    "fake-address",
		IsSponsor: &gogotypes.BoolValue{
			Value: false,
		},
	})
	suite.Require().Error(err)

	// Should update winnerAddress valid param
	_, err = msgServer.DepositEdit(goCtx, &millionstypes.MsgDepositEdit{
		PoolId:           deposit.PoolId,
		DepositId:        deposit.DepositId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[1].String(),
		IsSponsor: &gogotypes.BoolValue{
			Value: false,
		},
	})
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, 1, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.addrs[1].String(), deposit.WinnerAddress)

	// Sponsor can be nil and default to initial deposit isSponsor
	_, err = msgServer.DepositEdit(goCtx, &millionstypes.MsgDepositEdit{
		PoolId:           deposit.PoolId,
		DepositId:        deposit.DepositId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[1].String(),
		IsSponsor:        nil,
	})
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, 1, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(false, deposit.IsSponsor)
	pool, err := app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(0), pool.SponsorshipAmount.Int64())

	// Should update sponsor for deposit and pool SponsorshipAmount for valid sponsor param
	_, err = msgServer.DepositEdit(goCtx, &millionstypes.MsgDepositEdit{
		PoolId:           deposit.PoolId,
		DepositId:        deposit.DepositId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		IsSponsor: &gogotypes.BoolValue{
			Value: true,
		},
	})
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, 1, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(true, deposit.IsSponsor)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(1_000_000), pool.SponsorshipAmount.Int64())

	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)

	// A sponsor cannot designate another winner address
	_, err = msgServer.DepositEdit(goCtx, &millionstypes.MsgDepositEdit{
		PoolId:           deposit.PoolId,
		DepositId:        deposit.DepositId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[1].String(),
		IsSponsor:        nil,
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidSponsorWinnerCombo)

	// Sponsor can be nil and default to initial deposit isSponsor -> In this case true
	_, err = msgServer.DepositEdit(goCtx, &millionstypes.MsgDepositEdit{
		PoolId:           deposit.PoolId,
		DepositId:        deposit.DepositId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		IsSponsor:        nil,
	})
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, 1, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(true, deposit.IsSponsor)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(1_000_000), pool.SponsorshipAmount.Int64())

	// Should sub sponsor amount if user is not willing to sponsor anymore
	_, err = msgServer.DepositEdit(goCtx, &millionstypes.MsgDepositEdit{
		PoolId:           deposit.PoolId,
		DepositId:        deposit.DepositId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[1].String(),
		IsSponsor: &gogotypes.BoolValue{
			Value: false,
		},
	})
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, 1, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(false, deposit.IsSponsor)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(0), pool.SponsorshipAmount.Int64())
}

// TestMsgServer_WithdrawDeposit runs withdrawal deposit related tests
func (suite *KeeperTestSuite) TestMsgServer_WithdrawDeposit() {
	// Set the app context
	app := suite.App
	ctx := suite.Ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)
	uatomAddresses := apptesting.AddTestAddrsWithDenom(app, ctx, 7, sdk.NewCoins(sdk.NewCoin(remotePoolDenom, sdk.NewInt(1_000_0000_000))))

	// Initialize the pool
	poolID, err := app.MillionsKeeper.RegisterPool(ctx,
		"ulum",
		"ulum",
		testChainID,
		"",
		"",
		[]string{suite.valAddrs[0].String()},
		"lum",
		"lumvaloper",
		app.MillionsKeeper.GetParams(ctx).MinDepositAmount,
		millionstypes.DrawSchedule{DrawDelta: 24 * time.Hour, InitialDrawAt: ctx.BlockTime().Add(24 * time.Hour)},
		millionstypes.PrizeStrategy{PrizeBatches: []millionstypes.PrizeBatch{{PoolPercent: 100, Quantity: 100, DrawProbability: sdk.NewDec(1)}}},
	)
	suite.Require().NoError(err)
	_, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	poolID = app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta1 := 1 * time.Hour
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: poolID,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	}))
	pools := app.MillionsKeeper.ListPools(ctx)
	pool, err := app.MillionsKeeper.GetPool(ctx, pools[0].PoolId)
	suite.Require().NoError(err)

	// Create 2 successful deposits
	for i := 0; i < 2; i++ {
		// Create second deposit
		app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
			PoolId:           pool.PoolId,
			DepositorAddress: suite.addrs[0].String(),
			WinnerAddress:    suite.addrs[0].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
		})
		deposits := app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
		err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pools[0].IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000))})
		suite.Require().NoError(err)
		// Trigger the Transfer deposit to native chain
		err = app.MillionsKeeper.TransferDepositToNativeChain(ctx, deposits[i].PoolId, deposits[i].DepositId)
		suite.Require().NoError(err)
	}
	// Create deposit with failure state
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pool.PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_Failure,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	// List deposits
	deposits := app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	suite.Require().Len(deposits, 3)

	// Create invalid Withdraw request
	// - Wrong depositorAddress
	_, err = msgServer.WithdrawDeposit(goCtx, &millionstypes.MsgWithdrawDeposit{
		PoolId:           deposits[0].PoolId,
		DepositId:        deposits[0].DepositId,
		DepositorAddress: "",
		ToAddress:        suite.addrs[0].String(),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidDepositorAddress)

	// - Wrong to_address
	_, err = msgServer.WithdrawDeposit(goCtx, &millionstypes.MsgWithdrawDeposit{
		PoolId:           deposits[0].PoolId,
		DepositId:        deposits[0].DepositId,
		DepositorAddress: suite.addrs[0].String(),
		ToAddress:        "",
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidDestinationAddress)

	// - Wrong deposit_id
	_, err = msgServer.WithdrawDeposit(goCtx, &millionstypes.MsgWithdrawDeposit{
		PoolId:           pool.PoolId,
		DepositId:        uint64(0),
		DepositorAddress: suite.addrs[0].String(),
		ToAddress:        suite.addrs[0].String(),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrDepositNotFound)

	// - Wrong state
	_, err = msgServer.WithdrawDeposit(goCtx, &millionstypes.MsgWithdrawDeposit{
		PoolId:           deposits[2].PoolId,
		DepositId:        deposits[2].DepositId,
		DepositorAddress: deposits[2].DepositorAddress,
		ToAddress:        suite.addrs[0].String(),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidDepositState)

	// - Wront depositor and withdraw address
	_, err = msgServer.WithdrawDeposit(goCtx, &millionstypes.MsgWithdrawDeposit{
		PoolId:           pool.PoolId,
		DepositId:        deposits[1].DepositId,
		DepositorAddress: suite.addrs[3].String(),
		ToAddress:        suite.addrs[3].String(),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidWithdrawalDepositorAddress)

	// Add a successful withdrawal should remove a deposit
	_, err = msgServer.WithdrawDeposit(goCtx, &millionstypes.MsgWithdrawDeposit{
		PoolId:           deposits[0].PoolId,
		DepositId:        deposits[0].DepositId,
		DepositorAddress: suite.addrs[0].String(),
		ToAddress:        suite.addrs[0].String(),
	})
	suite.Require().NoError(err)
	// Make sure we have the correct number of withdrawals
	withdrawals := app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 1)
	// State should be unbonding
	suite.Require().Equal(millionstypes.WithdrawalState_IcaUnbonding, withdrawals[0].State)
	suite.Require().Equal(millionstypes.WithdrawalState_Unspecified, withdrawals[0].ErrorState)

	// Two deposits only should remain
	deposits = app.MillionsKeeper.ListDeposits(ctx)
	suite.Require().Len(deposits, 2)

	// Add a successful withdrawal should remove a deposit
	_, err = msgServer.WithdrawDeposit(goCtx, &millionstypes.MsgWithdrawDeposit{
		PoolId:           deposits[0].PoolId,
		DepositId:        deposits[0].DepositId,
		DepositorAddress: suite.addrs[0].String(),
		ToAddress:        suite.addrs[0].String(),
	})
	suite.Require().NoError(err)

	// Make sure we have the correct number of withdrawals
	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 2)
	// State should be unbonding
	suite.Require().Equal(millionstypes.WithdrawalState_IcaUnbonding, withdrawals[1].State)
	suite.Require().Equal(millionstypes.WithdrawalState_Unspecified, withdrawals[1].ErrorState)

	// One deposit only should remain
	deposits = app.MillionsKeeper.ListDeposits(ctx)
	suite.Require().Len(deposits, 1)

	// Initialize a remote pool
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:              3,
		Bech32PrefixValAddr: remoteBech32PrefixValAddr,
		Bech32PrefixAccAddr: remoteBech32PrefixAccAddr,
		ChainId:             remoteChainId,
		Denom:               remotePoolDenom,
		NativeDenom:         remotePoolDenom,
		ConnectionId:        ibctesting.FirstConnectionID,
		TransferChannelId:   ibctesting.FirstChannelID,
		Validators: []millionstypes.PoolValidator{
			{
				OperatorAddress: cosmosPoolValidator,
				BondedAmount:    sdk.NewInt(1_000_000),
				IsEnabled:       true,
			},
		},
		IcaDepositAddress:   cosmosIcaDepositAddress,
		IcaPrizepoolAddress: cosmosIcaPrizePoolAddress,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(remotePoolDenom, math.NewInt(1000)),
	}))
	pools = app.MillionsKeeper.ListPools(ctx)
	pool, err = app.MillionsKeeper.GetPool(ctx, pools[2].PoolId)
	suite.Require().NoError(err)

	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pool.PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_Success,
		Amount:           sdk.NewCoin(remotePoolDenom, sdk.NewInt(1_000_000)),
	})
	deposits = app.MillionsKeeper.ListAccountDeposits(ctx, suite.addrs[0])
	err = app.BankKeeper.SendCoins(ctx, uatomAddresses[0], sdk.MustAccAddressFromBech32(pools[2].LocalAddress), sdk.Coins{sdk.NewCoin(remotePoolDenom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)

	pools = app.MillionsKeeper.ListPools(ctx)
	pool, err = app.MillionsKeeper.GetPool(ctx, pools[2].PoolId)
	suite.Require().NoError(err)

	_, err = msgServer.WithdrawDeposit(goCtx, &millionstypes.MsgWithdrawDeposit{
		PoolId:           deposits[1].PoolId,
		DepositId:        deposits[1].DepositId,
		DepositorAddress: suite.addrs[0].String(),
		ToAddress:        "osmo11qggwr7vze9x7ustru9dctf6krhk2285lkf89g7",
	})
	// Error is triggered as toAddress does not match pool bech prefix and neither lum AccAddressFromBech32
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidDestinationAddress)

	_, err = msgServer.WithdrawDeposit(goCtx, &millionstypes.MsgWithdrawDeposit{
		PoolId:           deposits[1].PoolId,
		DepositId:        deposits[1].DepositId,
		DepositorAddress: suite.addrs[0].String(),
		ToAddress:        cosmosAccAddress,
	})
	// Error is triggered as failed to retrieve open active channel port (intended error)
	// but we passed all the toAddr validations which is the intention of this test
	suite.Require().ErrorIs(err, icatypes.ErrActiveChannelNotFound)
}

// TestMsgServer_WithdrawDepositRetry runs withdrawal retry related tests
func (suite *KeeperTestSuite) TestMsgServer_WithdrawDepositRetry() {
	// Set the app context
	app := suite.App
	ctx := suite.Ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)
	drawDelta1 := 1 * time.Hour
	var now = time.Now().UTC()

	pool := newValidPool(suite, millionstypes.Pool{
		PoolId:      uint64(1),
		Denom:       localPoolDenom,
		NativeDenom: localPoolDenom,
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: suite.valAddrs[0].String(),
			BondedAmount:    sdk.NewInt(1_000_000),
			IsEnabled:       true,
		}},
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: floatToDec(0.00)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(localPoolDenom, math.NewInt(1000)),
	})
	app.MillionsKeeper.AddPool(ctx, pool)

	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		PoolId:           pool.PoolId,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		State:            millionstypes.DepositState_IbcTransfer,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000)),
	})
	err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pool.IcaDepositAddress), sdk.Coins{sdk.NewCoin(localPoolDenom, sdk.NewInt(1_000_000))})
	suite.Require().NoError(err)
	deposits := app.MillionsKeeper.ListDeposits(ctx)
	// Simulate transfer deposit and delegate to native chain
	err = app.MillionsKeeper.TransferDepositToNativeChain(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	deposits = app.MillionsKeeper.ListDeposits(ctx)
	err = app.MillionsKeeper.DelegateDepositOnNativeChain(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().Error(err)

	deposits = app.MillionsKeeper.ListDeposits(ctx)

	app.MillionsKeeper.AddWithdrawal(ctx, millionstypes.Withdrawal{
		PoolId:           deposits[0].PoolId,
		DepositId:        deposits[0].DepositId,
		State:            millionstypes.WithdrawalState_Failure,
		ErrorState:       millionstypes.WithdrawalState_IcaUndelegate,
		DepositorAddress: suite.addrs[0].String(),
		ToAddress:        suite.addrs[0].String(),
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(int64(1_000_000))),
		UnbondingEndsAt:  &time.Time{},
	})

	deposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, deposits[0].PoolId, deposits[0].DepositId)
	suite.Require().NoError(err)
	// Remove deposit to respect the flow
	app.MillionsKeeper.RemoveDeposit(ctx, &deposit)

	withdrawals := app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 1)

	// Test ValidateWithdrawDepositRetryBasic
	// Test withdrawal retry with invalid poolID
	// - Test unknown poolID
	_, err = msgServer.WithdrawDepositRetry(goCtx, &millionstypes.MsgWithdrawDepositRetry{
		PoolId:           uint64(0),
		WithdrawalId:     withdrawals[0].WithdrawalId,
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)

	// - Test unknown withdrawalID
	_, err = msgServer.WithdrawDepositRetry(goCtx, &millionstypes.MsgWithdrawDepositRetry{
		PoolId:           withdrawals[0].PoolId,
		WithdrawalId:     uint64(0),
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidID)

	// Test GetPoolWithdrawal
	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 1)

	// Test GetPoolWithdrawal with wrong poolID
	_, err = app.MillionsKeeper.GetPoolWithdrawal(ctx, uint64(0), withdrawals[0].WithdrawalId)
	suite.Require().ErrorIs(err, millionstypes.ErrWithdrawalNotFound)

	// Test GetPoolWithdrawal with wrong withdrawalID
	_, err = app.MillionsKeeper.GetPoolWithdrawal(ctx, withdrawals[0].PoolId, uint64(0))
	suite.Require().ErrorIs(err, millionstypes.ErrWithdrawalNotFound)

	// Test if depositorAddr msg is the same as pool withdrawal depositor entity
	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)
	_, err = msgServer.WithdrawDepositRetry(goCtx, &millionstypes.MsgWithdrawDepositRetry{
		PoolId:           withdrawals[0].PoolId,
		WithdrawalId:     withdrawals[0].PoolId,
		DepositorAddress: "different-address",
	})
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidDepositorAddress)

	// Test WithdrawalState_IcaUndelegate ErrorState
	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 1)

	_, err = msgServer.WithdrawDepositRetry(goCtx, &millionstypes.MsgWithdrawDepositRetry{
		PoolId:           withdrawals[0].PoolId,
		WithdrawalId:     withdrawals[0].WithdrawalId,
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().NoError(err)

	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)
	suite.Require().Len(withdrawals, 1)

	// Should have the correct state
	suite.Require().Equal(millionstypes.WithdrawalState_IcaUnbonding, withdrawals[0].State)
	suite.Require().Equal(millionstypes.WithdrawalState_Unspecified, withdrawals[0].ErrorState)

	app.MillionsKeeper.UpdateWithdrawalStatus(ctx, withdrawals[0].PoolId, withdrawals[0].WithdrawalId, millionstypes.WithdrawalState_IbcTransfer, withdrawals[0].UnbondingEndsAt, true)

	// move 21 days unbonding onwards and force unbonding
	ctx = ctx.WithBlockTime(now.Add(21 * 24 * time.Hour))
	goCtx = ctx
	_, err = app.StakingKeeper.CompleteUnbonding(ctx, sdk.MustAccAddressFromBech32(pool.IcaDepositAddress), suite.valAddrs[0])
	suite.Require().NoError(err)

	// Test WithdrawalState_IbcTransfer ErrorState
	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)
	_, err = msgServer.WithdrawDepositRetry(goCtx, &millionstypes.MsgWithdrawDepositRetry{
		PoolId:           withdrawals[0].PoolId,
		WithdrawalId:     withdrawals[0].WithdrawalId,
		DepositorAddress: suite.addrs[0].String(),
	})
	suite.Require().NoError(err)
	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)
	// Withdrawals should be removed upon completion
	suite.Require().Len(withdrawals, 0)
}

// TestMsgServer_ClaimPrize runs prize claim related tests
func (suite *KeeperTestSuite) TestMsgServer_ClaimPrize() {
	app := suite.App
	ctx := suite.Ctx
	params := app.MillionsKeeper.GetParams(ctx)
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)

	// Create pool and simulate module account address
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	pool := newValidPool(suite, millionstypes.Pool{
		PoolId: poolID,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 50, Quantity: 100, DrawProbability: floatToDec(0.5)},
				{PoolPercent: 50, Quantity: 200, DrawProbability: floatToDec(0.7)},
			},
		},
	})
	app.MillionsKeeper.AddPool(ctx, pool)

	// Create prizes at various stages
	app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
		PoolId:        poolID,
		DrawId:        pool.NextDrawId,
		PrizeId:       1,
		Amount:        sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
		WinnerAddress: suite.addrs[1].String(),
		CreatedAt:     ctx.BlockTime(),
		ExpiresAt:     ctx.BlockTime().Add(params.PrizeExpirationDelta),
		State:         millionstypes.PrizeState_Pending,
	})
	app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
		PoolId:        poolID,
		DrawId:        pool.NextDrawId,
		PrizeId:       2,
		Amount:        sdk.NewCoin(pool.Denom, sdk.NewInt(2_000_000)),
		WinnerAddress: suite.addrs[1].String(),
		CreatedAt:     ctx.BlockTime(),
		ExpiresAt:     ctx.BlockTime().Add(params.PrizeExpirationDelta),
		State:         millionstypes.PrizeState_Pending,
	})
	app.MillionsKeeper.AddPrize(ctx, millionstypes.Prize{
		PoolId:        poolID,
		DrawId:        pool.NextDrawId,
		PrizeId:       3,
		Amount:        sdk.NewCoin(pool.Denom, sdk.NewInt(3_000_000)),
		WinnerAddress: suite.addrs[1].String(),
		CreatedAt:     ctx.BlockTime(),
		ExpiresAt:     ctx.BlockTime().Add(params.PrizeExpirationDelta),
		State:         millionstypes.PrizeState_Pending,
	})

	// List Prizes
	prizes := app.MillionsKeeper.ListPrizes(ctx)
	suite.Require().Len(prizes, 3)
	suite.Require().Equal(millionstypes.PrizeState_Pending, prizes[0].State)
	suite.Require().Equal(millionstypes.PrizeState_Pending, prizes[1].State)
	suite.Require().Equal(millionstypes.PrizeState_Pending, prizes[2].State)

	// Send required amounts to pool account
	modAddr, err := sdk.AccAddressFromBech32(pool.GetLocalAddress())
	suite.Require().NoError(err)
	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], modAddr, sdk.NewCoins(prizes[0].Amount.Add(prizes[1].Amount).Add(prizes[2].Amount)))
	suite.Require().NoError(err)

	// Test invalid claim requests
	_, err = msgServer.ClaimPrize(goCtx, millionstypes.NewMsgMsgClaimPrize("", prizes[0].PoolId, prizes[0].DrawId, prizes[0].PrizeId))
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidWinnerAddress)
	_, err = msgServer.ClaimPrize(goCtx, millionstypes.NewMsgMsgClaimPrize(suite.addrs[0].String(), prizes[0].PoolId, prizes[0].DrawId, prizes[0].PrizeId))
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, millionstypes.ErrInvalidWinnerAddress)

	_, err = msgServer.ClaimPrize(goCtx, millionstypes.NewMsgMsgClaimPrize(suite.addrs[1].String(), prizes[0].PoolId+1, prizes[0].DrawId, prizes[0].PrizeId))
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, millionstypes.ErrPrizeNotFound)
	_, err = msgServer.ClaimPrize(goCtx, millionstypes.NewMsgMsgClaimPrize(suite.addrs[1].String(), prizes[0].PoolId, prizes[0].DrawId+1, prizes[0].PrizeId))
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, millionstypes.ErrPrizeNotFound)
	_, err = msgServer.ClaimPrize(goCtx, millionstypes.NewMsgMsgClaimPrize(suite.addrs[1].String(), prizes[0].PoolId, prizes[0].DrawId, prizes[2].PrizeId+1))
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, millionstypes.ErrPrizeNotFound)

	prizes = app.MillionsKeeper.ListPrizes(ctx)

	for _, prize := range prizes {
		addr := sdk.MustAccAddressFromBech32(prize.WinnerAddress)
		// Get the module balance before claimPrize
		moduleBalanceBefore := app.BankKeeper.GetBalance(ctx, modAddr, pool.Denom)
		// Get the balance before claimPrize
		balanceBefore := app.BankKeeper.GetBalance(ctx, addr, pool.Denom)
		// claimPrize
		_, err = msgServer.ClaimPrize(goCtx, millionstypes.NewMsgMsgClaimPrize(addr.String(), prize.PoolId, prize.DrawId, prize.PrizeId))
		suite.Require().NoError(err)
		// Compare the new balance with the expected balance
		balance := app.BankKeeper.GetBalance(ctx, addr, pool.Denom)
		expectedBalance := balanceBefore.Add(prize.Amount)
		suite.Require().True(expectedBalance.Equal(balance))
		// Compare the new module balance with the expected module balance
		expectedModuleBalance := moduleBalanceBefore.Sub(prize.Amount)
		moduleBalance := app.BankKeeper.GetBalance(ctx, modAddr, pool.Denom)
		suite.Require().True(expectedModuleBalance.Equal(moduleBalance))
	}
	// There should be no more prizes
	prizes = app.MillionsKeeper.ListPrizes(ctx)
	suite.Require().Len(prizes, 0)
}
