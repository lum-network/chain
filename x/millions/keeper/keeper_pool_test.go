package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

func splitDelegationSliceToMap(vals []*millionstypes.SplitDelegation) map[string]*millionstypes.SplitDelegation {
	m := map[string]*millionstypes.SplitDelegation{}
	for _, v := range vals {
		m[v.ValidatorAddress] = v
	}
	return m
}

func (suite *KeeperTestSuite) TestPool_IDsGeneration() {
	app := suite.app
	ctx := suite.ctx

	for i := 0; i < 10; i++ {
		nextPoolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
		app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{PoolId: nextPoolID}))
		suite.Require().Equal(uint64(i+1), nextPoolID)
	}
}

func (suite *KeeperTestSuite) TestPool_ValidatorsBasics() {
	params := suite.app.MillionsKeeper.GetParams(suite.ctx)

	pool := newValidPool(suite, millionstypes.Pool{PoolId: 1})

	// Both prefixes are required
	pool.Bech32PrefixAccAddr = ""
	pool.Bech32PrefixValAddr = ""
	suite.Require().Error(pool.ValidateBasic(params))
	suite.Require().ErrorIs(pool.ValidateBasic(params), millionstypes.ErrInvalidPoolParams)
	pool.Bech32PrefixAccAddr = "lum"
	suite.Require().Error(pool.ValidateBasic(params))
	suite.Require().ErrorIs(pool.ValidateBasic(params), millionstypes.ErrInvalidPoolParams)
	pool.Bech32PrefixValAddr = "lumvaloper"
	suite.Require().NoError(pool.ValidateBasic(params))

	// Prefix should match validators config
	pool.Bech32PrefixValAddr = "cosmosvaloper"
	suite.Require().Error(pool.ValidateBasic(params))
	suite.Require().ErrorIs(pool.ValidateBasic(params), millionstypes.ErrInvalidPoolParams)
	pool.Validators = []millionstypes.PoolValidator{{
		OperatorAddress: "cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn",
		BondedAmount:    sdk.ZeroInt(),
	}}
	suite.Require().NoError(pool.ValidateBasic(params))
}

// TestPool_DrawSchedule validates draw schedule configuration and implementation
func (suite *KeeperTestSuite) TestPool_DrawScheduleBasics() {
	ctx := suite.ctx

	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)
	params := suite.app.MillionsKeeper.GetParams(ctx)
	params.MinDrawScheduleDelta = 24 * time.Hour
	suite.app.MillionsKeeper.SetParams(ctx, params)

	// Draw delta must be >= 24 hours
	ds := millionstypes.DrawSchedule{InitialDrawAt: ctx.BlockTime().Add(24 * time.Hour), DrawDelta: 0}
	suite.Require().Error(ds.ValidateNew(ctx, params))
	ds = millionstypes.DrawSchedule{InitialDrawAt: ctx.BlockTime().Add(24 * time.Hour), DrawDelta: 1 * time.Second}
	suite.Require().Error(ds.ValidateNew(ctx, params))

	// Initial Draw must be in the future + delta hours
	ds = millionstypes.DrawSchedule{InitialDrawAt: ctx.BlockTime().Add(-1 * time.Second), DrawDelta: 24 * time.Hour}
	suite.Require().Error(ds.ValidateNew(ctx, params))
	ds = millionstypes.DrawSchedule{InitialDrawAt: ctx.BlockTime().Add(1 * time.Second), DrawDelta: 24 * time.Hour}
	suite.Require().Error(ds.ValidateNew(ctx, params))
	ds = millionstypes.DrawSchedule{InitialDrawAt: ctx.BlockTime().Add(23 * time.Hour), DrawDelta: 24 * time.Hour}
	suite.Require().Error(ds.ValidateNew(ctx, params))

	// Initial Draw gets rounded to the minute upon sanitization
	ds = millionstypes.DrawSchedule{InitialDrawAt: ctx.BlockTime().Add(23 * time.Hour), DrawDelta: 24 * time.Hour}
	dsSan := ds.Sanitized()
	suite.Require().Equal(ds.GetInitialDrawAt().Year(), dsSan.GetInitialDrawAt().Year())
	suite.Require().Equal(ds.GetInitialDrawAt().Month(), dsSan.GetInitialDrawAt().Month())
	suite.Require().Equal(ds.GetInitialDrawAt().Day(), dsSan.GetInitialDrawAt().Day())
	suite.Require().Equal(ds.GetInitialDrawAt().Hour(), dsSan.GetInitialDrawAt().Hour())
	suite.Require().Equal(ds.GetInitialDrawAt().Minute(), dsSan.GetInitialDrawAt().Minute())
	suite.Require().Equal(0, dsSan.GetInitialDrawAt().Second())
	suite.Require().Equal(0, dsSan.GetInitialDrawAt().Nanosecond())

	// Subsequent draw should trigger as soon as the draw delta is elapsed
	// In other words, we should be able to launch a draw even if the elasped time is not passed draw delta depending
	// on the initial launch time to prevent pool draw time drift (ex: if chain stalling for a while)
	dsNow := time.Date(2000, 10, 10, 10, 10, 0, 0, time.UTC)
	dsCtx := ctx.WithBlockTime(dsNow)
	dsSan = millionstypes.DrawSchedule{InitialDrawAt: dsNow.Add(1 * time.Hour), DrawDelta: 1 * time.Hour}.Sanitized()
	suite.Require().False(dsSan.ShouldDraw(dsCtx, nil))
	suite.Require().False(dsSan.ShouldDraw(dsCtx.WithBlockTime(dsNow.Add(45*time.Minute)), nil))
	suite.Require().False(dsSan.ShouldDraw(dsCtx.WithBlockTime(dsNow.Add(50*time.Minute)), nil))
	// First draw schedule should be exactly the stanitized InitialDrawAt
	suite.Require().True(dsSan.ShouldDraw(dsCtx.WithBlockTime(dsNow.Add(60*time.Minute)), nil))
	// Subsequent draws should be able to be scheduled earlier than the expected draw delta to re-align timing with initial draw
	// simulate last draw drifted by 5 min from expected time - should allow next draw 5min earlier
	dsNow = dsNow.Add(65 * time.Minute)
	suite.Require().False(dsSan.ShouldDraw(dsCtx.WithBlockTime(dsNow.Add(45*time.Minute)), &dsNow))
	suite.Require().False(dsSan.ShouldDraw(dsCtx.WithBlockTime(dsNow.Add(50*time.Minute)), &dsNow))
	suite.Require().True(dsSan.ShouldDraw(dsCtx.WithBlockTime(dsNow.Add(55*time.Minute)), &dsNow))
	suite.Require().True(dsSan.ShouldDraw(dsCtx.WithBlockTime(dsNow.Add(155*time.Minute)), &dsNow))

	// Pools draw should be defined by their state and the draw schedule
	pool := newValidPool(suite, millionstypes.Pool{
		DrawSchedule: millionstypes.DrawSchedule{InitialDrawAt: ctx.BlockTime().Add(24 * time.Hour), DrawDelta: 24 * time.Hour},
		CreatedAt:    ctx.BlockTime(),
	})

	// Fresh pool should have 0 elapsed time and should not draw
	suite.Require().NoError(pool.DrawSchedule.ValidateNew(ctx, params))
	suite.Require().False(pool.ShouldDraw(ctx))

	// Moving forward by 12 hours - same situation with a duration of 12h
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().False(pool.ShouldDraw(ctx))

	// Moving forward by 12 hours - first draw should be able to occur
	ctx = ctx.WithBlockTime(now.Add(24 * time.Hour))
	suite.Require().True(pool.ShouldDraw(ctx))

	// Simulate draw status which should never allow another draw
	ld := ctx.BlockTime()
	pool.LastDrawCreatedAt = &ld
	for _, s := range []millionstypes.DrawState{
		millionstypes.DrawState_Failure,
		millionstypes.DrawState_IcaWithdrawRewards,
		millionstypes.DrawState_IbcTransfer,
	} {
		pool.LastDrawState = s
		suite.Require().False(pool.ShouldDraw(ctx.WithBlockTime(now.Add(48 * time.Hour))))
	}

	// Simulate first draw success which should make the pool accept a new draw depending on the previous one creation time
	pool.LastDrawState = millionstypes.DrawState_Success
	suite.Require().False(pool.ShouldDraw(ctx.WithBlockTime((*pool.LastDrawCreatedAt).Add(-5 * time.Hour))))
	suite.Require().False(pool.ShouldDraw(ctx.WithBlockTime((*pool.LastDrawCreatedAt).Add(0 * time.Hour))))
	suite.Require().False(pool.ShouldDraw(ctx.WithBlockTime((*pool.LastDrawCreatedAt).Add(12 * time.Hour))))
	suite.Require().True(pool.ShouldDraw(ctx.WithBlockTime((*pool.LastDrawCreatedAt).Add(24 * time.Hour))))
	suite.Require().True(pool.ShouldDraw(ctx.WithBlockTime((*pool.LastDrawCreatedAt).Add(48 * time.Hour))))

	// Test pool not ready states which should not allow any draw
	pool.State = millionstypes.PoolState_Killed
	suite.Require().False(pool.ShouldDraw(ctx.WithBlockTime(now.Add(48 * time.Hour))))

	pool.State = millionstypes.PoolState_Created
	pool.LastDrawCreatedAt = nil
	pool.LastDrawState = millionstypes.DrawState_Unspecified
	suite.Require().False(pool.ShouldDraw(ctx.WithBlockTime(now.Add(48 * time.Hour))))
}

func (suite *KeeperTestSuite) TestPool_PrizeStrategiesBasics() {
	params := suite.app.MillionsKeeper.GetParams(suite.ctx)

	// PoolPercent should sum up to 100
	err := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{},
	}.Validate(params)
	suite.Require().Error(err)

	err = millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 25, Quantity: 1, DrawProbability: floatToDec(0.1)},
		},
	}.Validate(params)
	suite.Require().Error(err)

	err = millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 101, Quantity: 1, DrawProbability: floatToDec(0.1)},
		},
	}.Validate(params)
	suite.Require().Error(err)

	err = millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 25, Quantity: 1, DrawProbability: floatToDec(0.1)},
			{PoolPercent: 80, Quantity: 1, DrawProbability: floatToDec(0.1)},
		},
	}.Validate(params)
	suite.Require().Error(err)

	// DrawProbability should be [0, 1]
	err = millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 20, Quantity: 1, DrawProbability: floatToDec(0.01)},
			{PoolPercent: 30, Quantity: 1, DrawProbability: floatToDec(0.11)},
			{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.01)},
		},
	}.Validate(params)
	suite.Require().Error(err)

	err = millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 20, Quantity: 1, DrawProbability: floatToDec(0.01)},
			{PoolPercent: 30, Quantity: 1, DrawProbability: floatToDec(0.11)},
			{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(-1.01)},
		},
	}.Validate(params)
	suite.Require().Error(err)

	// Quantity should be > 0
	err = millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 20, Quantity: 0, DrawProbability: floatToDec(0.01)},
			{PoolPercent: 30, Quantity: 1, DrawProbability: floatToDec(0.11)},
			{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
		},
	}.Validate(params)
	suite.Require().Error(err)

	err = millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 20, Quantity: 0, DrawProbability: floatToDec(0.01)},
			{PoolPercent: 30, Quantity: 1, DrawProbability: floatToDec(0.11)},
			{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
		},
	}.Validate(params)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestPool_DepositorsCountAndTVL() {
	app := suite.app
	ctx := suite.ctx
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)

	denom := app.StakingKeeper.BondDenom(ctx)
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
	pool, err := app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)

	// Deposit with 3 accounts
	for i := 0; i < 5; i++ {
		_, err := msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
			DepositorAddress: suite.addrs[0].String(),
			PoolId:           pool.PoolId,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(1_000_000)),
		})
		suite.Require().NoError(err)
	}
	pool, err = app.MillionsKeeper.GetPool(ctx, pool.PoolId)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), pool.DepositorsCount)
	suite.Require().Equal(sdk.NewInt(5_000_000), pool.TvlAmount)

	for i := 0; i < 10; i++ {
		_, err := msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
			DepositorAddress: suite.addrs[1].String(),
			PoolId:           pool.PoolId,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(2_000_000)),
		})
		suite.Require().NoError(err)
	}
	pool, err = app.MillionsKeeper.GetPool(ctx, pool.PoolId)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(2), pool.DepositorsCount)
	suite.Require().Equal(sdk.NewInt(5_000_000+20_000_000), pool.TvlAmount)

	for i := 0; i < 20; i++ {
		_, err := msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
			DepositorAddress: suite.addrs[2].String(),
			PoolId:           pool.PoolId,
			Amount:           sdk.NewCoin(denom, sdk.NewInt(4_000_000)),
		})
		suite.Require().NoError(err)
	}
	pool, err = app.MillionsKeeper.GetPool(ctx, pool.PoolId)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(3), pool.DepositorsCount)
	suite.Require().Equal(sdk.NewInt(5_000_000+20_000_000+80_000_000), pool.TvlAmount)

	// Remove all deposits from all accounts
	deposits := app.MillionsKeeper.ListAccountPoolDeposits(ctx, suite.addrs[0], pool.PoolId)
	for i, d := range deposits {
		app.MillionsKeeper.RemoveDeposit(ctx, &d)
		pool, err = app.MillionsKeeper.GetPool(ctx, pool.PoolId)
		suite.Require().NoError(err)
		if i == len(deposits)-1 {
			suite.Require().Equal(uint64(2), pool.DepositorsCount)
		} else {
			suite.Require().Equal(uint64(3), pool.DepositorsCount)
		}
		suite.Require().Equal(sdk.NewInt(1_000_000*int64(5-i-1)+20_000_000+80_000_000), pool.TvlAmount)
	}

	deposits = app.MillionsKeeper.ListAccountPoolDeposits(ctx, suite.addrs[1], pool.PoolId)
	for i, d := range deposits {
		app.MillionsKeeper.RemoveDeposit(ctx, &d)
		pool, err = app.MillionsKeeper.GetPool(ctx, pool.PoolId)
		suite.Require().NoError(err)
		if i == len(deposits)-1 {
			suite.Require().Equal(uint64(1), pool.DepositorsCount)
		} else {
			suite.Require().Equal(uint64(2), pool.DepositorsCount)
		}
		suite.Require().Equal(sdk.NewInt(2_000_000*int64(10-i-1)+80_000_000), pool.TvlAmount)
	}

	deposits = app.MillionsKeeper.ListAccountPoolDeposits(ctx, suite.addrs[1], pool.PoolId)
	for i, d := range deposits {
		app.MillionsKeeper.RemoveDeposit(ctx, &d)
		pool, err = app.MillionsKeeper.GetPool(ctx, pool.PoolId)
		suite.Require().NoError(err)
		if i == len(deposits)-1 {
			suite.Require().Equal(uint64(0), pool.DepositorsCount)
		} else {
			suite.Require().Equal(uint64(1), pool.DepositorsCount)
		}
		suite.Require().Equal(sdk.NewInt(4_000_000*int64(20-i-1)), pool.TvlAmount)
	}
}

// TestPool_ValidatorsSplitDelegate test pool validator set split delegations
func (suite *KeeperTestSuite) TestPool_ValidatorsSplitDelegate() {
	app := suite.app
	ctx := suite.ctx

	valAddrs := []string{
		"cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c",
		"cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn",
		"cosmosvaloper1fsg635n5vgc7jazz9sx5725wnc3xqgr7awxaag",
		"cosmosvaloper1gpx52r9h3zeul45amvcy2pysgvcwddxrgx6cnv",
		"cosmosvaloper1vvwtk805lxehwle9l4yudmq6mn0g32px9xtkhc",
	}

	// valSet1 make validator 0 inactive
	valSet1 := []millionstypes.PoolValidator{
		{
			OperatorAddress: valAddrs[0],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       false,
		},
	}
	// valSet2 make validator 0 active
	valSet2 := []millionstypes.PoolValidator{
		{
			OperatorAddress: valAddrs[0],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
	}
	// valSet3 adds 3 more validators active and one inactive
	valSet3 := []millionstypes.PoolValidator{
		{
			OperatorAddress: valAddrs[0],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[1],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[2],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[3],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[4],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       false,
		},
	}
	// valSet4 disables validators 0 and 1
	valSet4 := []millionstypes.PoolValidator{
		{
			OperatorAddress: valAddrs[0],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       false,
		},
		{
			OperatorAddress: valAddrs[1],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       false,
		},
		{
			OperatorAddress: valAddrs[2],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[3],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[4],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       false,
		},
	}

	pool := newValidPool(suite, millionstypes.Pool{
		PoolId:              app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx),
		Bech32PrefixValAddr: "cosmosvaloper",
		ChainId:             "cosmos",
		Denom:               "uatom",
		NativeDenom:         "uatom",
		ConnectionId:        "connection-id",
		TransferChannelId:   "transferChannel-id",
		Validators:          valSet1,
		IcaDepositAddress:   cosmosIcaDepositAddress,
		IcaPrizepoolAddress: cosmosIcaPrizePoolAddress,
		State:               millionstypes.PoolState_Ready,
	})

	amount := sdk.NewInt(1_000_000)

	// Splits should return nil (error case) values if no validator set available
	sDel := pool.ComputeSplitDelegations(ctx, amount)
	suite.Require().Nil(sDel)

	// Update pool to make val[0] active
	pool.Validators = valSet2

	// Split delegate should evenly distribute amount
	sDel = pool.ComputeSplitDelegations(ctx, amount)
	suite.Require().Len(sDel, 1)
	suite.Require().Equal(valAddrs[0], sDel[0].ValidatorAddress)
	suite.Require().Equal(amount, sDel[0].Amount)

	// Update pool to add 3 more validators active and one inactive
	pool.Validators = valSet3

	// Split delegate should evenly distribute amount to active validators
	sDel = pool.ComputeSplitDelegations(ctx, amount)
	suite.Require().Len(sDel, 4)
	sdMap := splitDelegationSliceToMap(sDel)
	suite.Require().Equal(amount.Quo(sdk.NewInt(4)), sdMap[valAddrs[0]].Amount)
	suite.Require().Equal(amount.Quo(sdk.NewInt(4)), sdMap[valAddrs[1]].Amount)
	suite.Require().Equal(amount.Quo(sdk.NewInt(4)), sdMap[valAddrs[2]].Amount)
	suite.Require().Equal(amount.Quo(sdk.NewInt(4)), sdMap[valAddrs[3]].Amount)

	// Update pool to add make val[2] and val[3] the only active ones
	pool.Validators = valSet4

	// Split delegate should evenly distribute amount to active validators
	sDel = pool.ComputeSplitDelegations(ctx, amount)
	suite.Require().Len(sDel, 2)
	sdMap = splitDelegationSliceToMap(sDel)
	suite.Require().Equal(amount.Quo(sdk.NewInt(2)), sdMap[valAddrs[2]].Amount)
	suite.Require().Equal(amount.Quo(sdk.NewInt(2)), sdMap[valAddrs[3]].Amount)

	// Split delegate should distribute rests to the last validator in the set in case of rounding approximation
	amountNotMultipleOf2 := sdk.NewInt(1_000_001)
	sDel = pool.ComputeSplitDelegations(ctx, amountNotMultipleOf2)
	suite.Require().Len(sDel, 2)
	sdMap = splitDelegationSliceToMap(sDel)
	suite.Require().GreaterOrEqual(sdMap[valAddrs[2]].Amount.Int64(), amountNotMultipleOf2.Quo(sdk.NewInt(2)).Int64())
	suite.Require().GreaterOrEqual(sdMap[valAddrs[3]].Amount.Int64(), amountNotMultipleOf2.Quo(sdk.NewInt(2)).Int64())
	suite.Require().Equal(amountNotMultipleOf2, sdMap[valAddrs[2]].Amount.Add(sdMap[valAddrs[3]].Amount))
}

// TestPool_ValidatorsSplitUndelegate test pool validator set split undelegations
func (suite *KeeperTestSuite) TestPool_ValidatorsSplitUndelegate() {
	app := suite.app
	ctx := suite.ctx

	valAddrs := []string{
		"cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c",
		"cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn",
		"cosmosvaloper1fsg635n5vgc7jazz9sx5725wnc3xqgr7awxaag",
		"cosmosvaloper1gpx52r9h3zeul45amvcy2pysgvcwddxrgx6cnv",
		"cosmosvaloper1vvwtk805lxehwle9l4yudmq6mn0g32px9xtkhc",
	}

	// valSet1 make validator 0 inactive
	valSet1 := []millionstypes.PoolValidator{
		{
			OperatorAddress: valAddrs[0],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       false,
		},
	}
	// valSet2 make validator 0 active
	valSet2 := []millionstypes.PoolValidator{
		{
			OperatorAddress: valAddrs[0],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
	}
	// valSet3 adds 3 more validators active and one inactive
	valSet3 := []millionstypes.PoolValidator{
		{
			OperatorAddress: valAddrs[0],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[1],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[2],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[3],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[4],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       false,
		},
	}
	// valSet4 disables validators 0 and 1
	valSet4 := []millionstypes.PoolValidator{
		{
			OperatorAddress: valAddrs[0],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       false,
		},
		{
			OperatorAddress: valAddrs[1],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       false,
		},
		{
			OperatorAddress: valAddrs[2],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[3],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		},
		{
			OperatorAddress: valAddrs[4],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       false,
		},
	}

	pool := newValidPool(suite, millionstypes.Pool{
		PoolId:              app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx),
		Bech32PrefixValAddr: "cosmosvaloper",
		ChainId:             "cosmos",
		Denom:               "uatom",
		NativeDenom:         "uatom",
		ConnectionId:        "connection-id",
		TransferChannelId:   "transferChannel-id",
		Validators:          valSet1,
		IcaDepositAddress:   cosmosIcaDepositAddress,
		IcaPrizepoolAddress: cosmosIcaPrizePoolAddress,
		State:               millionstypes.PoolState_Ready,
	})

	amount := sdk.NewInt(1_000_000)

	// Splits should return nil (error case) values if no validator bonded
	sUndel := pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Nil(sUndel)

	// Splits should return nil (error case) values if bonded validators don't have enough to undelegate
	pool.Validators[0].BondedAmount = amount.Sub(sdk.NewInt(1))
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Nil(sUndel)

	// Splits should return ok values if bonded amount is enough even if validator set is fully disabled
	pool.Validators[0].BondedAmount = amount
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Len(sUndel, 1)
	suite.Require().Equal(valAddrs[0], sUndel[0].ValidatorAddress)
	suite.Require().Equal(amount, sUndel[0].Amount)

	// Update pool to make val[0] active
	pool.Validators = valSet2

	// Splits should have the same behaviour regardless of the activation status of the validator
	pool.Validators[0].BondedAmount = amount.Sub(sdk.NewInt(1))
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Nil(sUndel)
	pool.Validators[0].BondedAmount = amount
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Len(sUndel, 1)
	suite.Require().Equal(valAddrs[0], sUndel[0].ValidatorAddress)
	suite.Require().Equal(amount, sUndel[0].Amount)

	// Update pool to add 3 more validators active and one inactive
	pool.Validators = valSet3

	// Splits should return nil (error case) values if no bonded valitors or not enough bonded amount
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Nil(sUndel)

	pool.Validators[0].BondedAmount = sdk.NewInt(1)
	pool.Validators[1].BondedAmount = sdk.NewInt(1)
	pool.Validators[2].BondedAmount = sdk.NewInt(1)
	pool.Validators[3].BondedAmount = sdk.NewInt(1)
	pool.Validators[4].BondedAmount = sdk.NewInt(1)

	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Nil(sUndel)

	// Update pool to add make val[2] and val[3] the only active ones
	pool.Validators = valSet4

	// Splits should prioritize inactive validators
	// Disabled set
	pool.Validators[0].BondedAmount = sdk.NewInt(1_000)
	pool.Validators[1].BondedAmount = sdk.NewInt(100)
	pool.Validators[4].BondedAmount = sdk.NewInt(10)
	// Enabled set
	pool.Validators[2].BondedAmount = sdk.NewInt(100)
	pool.Validators[3].BondedAmount = sdk.NewInt(10)

	// Will pick up all val[0]
	amount = sdk.NewInt(1_000)
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Len(sUndel, 1)
	suMap := splitDelegationSliceToMap(sUndel)
	suite.Require().Equal(amount, suMap[valAddrs[0]].Amount)

	// Will pick up all val[0] then val[1]
	amount = sdk.NewInt(1_105)
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Len(sUndel, 3)
	suMap = splitDelegationSliceToMap(sUndel)
	suite.Require().Equal(pool.Validators[0].BondedAmount, suMap[valAddrs[0]].Amount)
	suite.Require().Equal(pool.Validators[1].BondedAmount, suMap[valAddrs[1]].Amount)
	suite.Require().Equal(sdk.NewInt(5), suMap[valAddrs[4]].Amount)

	// Will pick up inactive vals then split the remaining on all active validators
	amount = sdk.NewInt(1_111)
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Len(sUndel, 4)
	suMap = splitDelegationSliceToMap(sUndel)
	suite.Require().Equal(pool.Validators[0].BondedAmount, suMap[valAddrs[0]].Amount)
	suite.Require().Equal(pool.Validators[1].BondedAmount, suMap[valAddrs[1]].Amount)
	suite.Require().Equal(pool.Validators[4].BondedAmount, suMap[valAddrs[4]].Amount)
	suite.Require().Equal(sdk.NewInt(1), suMap[valAddrs[2]].Amount)

	amount = sdk.NewInt(1_112)
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Len(sUndel, 5)
	suMap = splitDelegationSliceToMap(sUndel)
	suite.Require().Equal(pool.Validators[0].BondedAmount, suMap[valAddrs[0]].Amount)
	suite.Require().Equal(pool.Validators[1].BondedAmount, suMap[valAddrs[1]].Amount)
	suite.Require().Equal(pool.Validators[4].BondedAmount, suMap[valAddrs[4]].Amount)
	suite.Require().Equal(sdk.NewInt(1), suMap[valAddrs[2]].Amount)
	suite.Require().Equal(sdk.NewInt(1), suMap[valAddrs[3]].Amount)

	amount = sdk.NewInt(1_122)
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Len(sUndel, 5)
	suMap = splitDelegationSliceToMap(sUndel)
	suite.Require().Equal(pool.Validators[0].BondedAmount, suMap[valAddrs[0]].Amount)
	suite.Require().Equal(pool.Validators[1].BondedAmount, suMap[valAddrs[1]].Amount)
	suite.Require().Equal(pool.Validators[4].BondedAmount, suMap[valAddrs[4]].Amount)
	suite.Require().Equal(sdk.NewInt(6), suMap[valAddrs[2]].Amount)
	suite.Require().Equal(sdk.NewInt(6), suMap[valAddrs[3]].Amount)

	amount = sdk.NewInt(1_150)
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Len(sUndel, 5)
	suMap = splitDelegationSliceToMap(sUndel)
	suite.Require().Equal(pool.Validators[0].BondedAmount, suMap[valAddrs[0]].Amount)
	suite.Require().Equal(pool.Validators[1].BondedAmount, suMap[valAddrs[1]].Amount)
	suite.Require().Equal(pool.Validators[4].BondedAmount, suMap[valAddrs[4]].Amount)
	suite.Require().Equal(sdk.NewInt(30), suMap[valAddrs[2]].Amount)
	suite.Require().Equal(sdk.NewInt(10), suMap[valAddrs[3]].Amount)

	amount = sdk.NewInt(1_220)
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Len(sUndel, 5)
	suMap = splitDelegationSliceToMap(sUndel)
	suite.Require().Equal(pool.Validators[0].BondedAmount, suMap[valAddrs[0]].Amount)
	suite.Require().Equal(pool.Validators[1].BondedAmount, suMap[valAddrs[1]].Amount)
	suite.Require().Equal(pool.Validators[4].BondedAmount, suMap[valAddrs[4]].Amount)
	suite.Require().Equal(pool.Validators[2].BondedAmount, suMap[valAddrs[2]].Amount)
	suite.Require().Equal(pool.Validators[3].BondedAmount, suMap[valAddrs[3]].Amount)

	// Will fail due to missing amount
	amount = sdk.NewInt(1_221)
	sUndel = pool.ComputeSplitUndelegations(ctx, amount)
	suite.Require().Nil(sUndel)
}

// TestPool_ValidatorsSplitConsistency test pool validator set bonded amount consistency and associated race conditions
func (suite *KeeperTestSuite) TestPool_ValidatorsSplitConsistency() {
	app := suite.app
	ctx := suite.ctx

	valAddr := "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c"
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:              poolID,
		Bech32PrefixValAddr: "cosmosvaloper",
		ChainId:             "cosmos",
		Denom:               "uatom",
		NativeDenom:         "uatom",
		ConnectionId:        "connection-id",
		TransferChannelId:   "transferChannel-id",
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: valAddr,
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
		}},
		IcaDepositAddress:   cosmosIcaDepositAddress,
		IcaPrizepoolAddress: cosmosIcaPrizePoolAddress,
		State:               millionstypes.PoolState_Ready,
	}))

	// Start deposit d1, d2 and d3
	d1 := millionstypes.Deposit{
		PoolId:           poolID,
		DepositId:        app.MillionsKeeper.GetNextDepositIdAndIncrement(ctx),
		State:            millionstypes.DepositState_IbcTransfer,
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
		Amount:           sdk.NewCoin("uatom", sdk.NewInt(111)),
	}
	app.MillionsKeeper.AddDeposit(ctx, &d1)
	d2 := millionstypes.Deposit{
		PoolId:           poolID,
		DepositId:        app.MillionsKeeper.GetNextDepositIdAndIncrement(ctx),
		State:            millionstypes.DepositState_IbcTransfer,
		DepositorAddress: suite.addrs[1].String(),
		WinnerAddress:    suite.addrs[1].String(),
		Amount:           sdk.NewCoin("uatom", sdk.NewInt(222)),
	}
	app.MillionsKeeper.AddDeposit(ctx, &d2)
	d3 := millionstypes.Deposit{
		PoolId:           poolID,
		DepositId:        app.MillionsKeeper.GetNextDepositIdAndIncrement(ctx),
		State:            millionstypes.DepositState_IbcTransfer,
		DepositorAddress: suite.addrs[2].String(),
		WinnerAddress:    suite.addrs[2].String(),
		Amount:           sdk.NewCoin("uatom", sdk.NewInt(333)),
	}
	app.MillionsKeeper.AddDeposit(ctx, &d3)

	// Validators bounded amount should still be 0 (not delegated yet)
	pool, err := app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroInt(), pool.Validators[0].BondedAmount)

	// Simulate coin transfer for d1 and d2 + failure on d3
	err = app.MillionsKeeper.OnTransferDepositToNativeChainCompleted(ctx, poolID, d1.DepositId, false)
	suite.Require().NoError(err)
	err = app.MillionsKeeper.OnTransferDepositToNativeChainCompleted(ctx, poolID, d2.DepositId, false)
	suite.Require().NoError(err)
	err = app.MillionsKeeper.OnTransferDepositToNativeChainCompleted(ctx, poolID, d3.DepositId, true)
	suite.Require().NoError(err)

	// Force fix deposits statuses since they will fail due to missing remote chain
	app.MillionsKeeper.UpdateDepositStatus(ctx, poolID, d1.DepositId, millionstypes.DepositState_IcaDelegate, false)
	app.MillionsKeeper.UpdateDepositStatus(ctx, poolID, d2.DepositId, millionstypes.DepositState_IcaDelegate, false)
	app.MillionsKeeper.UpdateDepositStatus(ctx, poolID, d3.DepositId, millionstypes.DepositState_IcaDelegate, false)

	// Validators bounded amount should still be 0 (not delegated yet)
	pool, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroInt(), pool.Validators[0].BondedAmount)

	// Simulate delegate success for d1 + failure on d2
	sd := pool.ComputeSplitDelegations(ctx, d1.Amount.Amount)
	suite.Require().Len(sd, 1)
	err = app.MillionsKeeper.OnDelegateDepositOnNativeChainCompleted(ctx, poolID, d1.DepositId, sd, false)
	suite.Require().NoError(err)
	sd = pool.ComputeSplitDelegations(ctx, d2.Amount.Amount)
	suite.Require().Len(sd, 1)
	err = app.MillionsKeeper.OnDelegateDepositOnNativeChainCompleted(ctx, poolID, d2.DepositId, sd, true)
	suite.Require().NoError(err)

	// Validators bounded amount should be d1 amount now
	pool, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(d1.Amount.Amount, pool.Validators[0].BondedAmount)

	// Simulate delegate success for d2
	sd = pool.ComputeSplitDelegations(ctx, d2.Amount.Amount)
	suite.Require().Len(sd, 1)
	app.MillionsKeeper.UpdateDepositStatus(ctx, poolID, d2.DepositId, millionstypes.DepositState_IcaDelegate, false)
	err = app.MillionsKeeper.OnDelegateDepositOnNativeChainCompleted(ctx, poolID, d2.DepositId, sd, false)
	suite.Require().NoError(err)

	// Validators bounded amount should be d1+d2 amount now
	pool, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(d1.Amount.Amount.Add(d2.Amount.Amount), pool.Validators[0].BondedAmount)

	// Start withdrawal d1 and d2
	w1 := millionstypes.Withdrawal{
		PoolId:           poolID,
		DepositId:        d1.DepositId,
		WithdrawalId:     app.MillionsKeeper.GetNextWithdrawalIdAndIncrement(ctx),
		State:            millionstypes.WithdrawalState_IcaUndelegate,
		DepositorAddress: d1.DepositorAddress,
		ToAddress:        d1.DepositorAddress,
		Amount:           d1.Amount,
	}
	app.MillionsKeeper.AddWithdrawal(ctx, w1)
	su1 := pool.ComputeSplitUndelegations(ctx, w1.Amount.Amount)
	suite.Require().Len(su1, 1)

	w2 := millionstypes.Withdrawal{
		PoolId:           poolID,
		DepositId:        d2.DepositId,
		WithdrawalId:     app.MillionsKeeper.GetNextWithdrawalIdAndIncrement(ctx),
		State:            millionstypes.WithdrawalState_IcaUndelegate,
		DepositorAddress: d2.DepositorAddress,
		ToAddress:        d2.DepositorAddress,
		Amount:           d2.Amount,
	}
	app.MillionsKeeper.AddWithdrawal(ctx, w2)
	su2 := pool.ComputeSplitUndelegations(ctx, w2.Amount.Amount)
	suite.Require().Len(su2, 1)

	// Simulate undelegate launched with success for d1 + d2 (ignore error voluntarely here)
	err = app.MillionsKeeper.UndelegateWithdrawalOnNativeChain(ctx, poolID, w1.WithdrawalId)
	suite.Require().Error(err)
	err = app.MillionsKeeper.UndelegateWithdrawalOnNativeChain(ctx, poolID, w2.WithdrawalId)
	suite.Require().Error(err)

	// Force fix withdrawals statuses since they will fail due to missing remote chain
	app.MillionsKeeper.UpdateWithdrawalStatus(ctx, poolID, w1.WithdrawalId, millionstypes.WithdrawalState_IcaUndelegate, nil, false)
	app.MillionsKeeper.UpdateWithdrawalStatus(ctx, poolID, w2.WithdrawalId, millionstypes.WithdrawalState_IcaUndelegate, nil, false)

	// Validators bounded amount should be 0 now
	pool, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroInt(), pool.Validators[0].BondedAmount)

	// Simulate undelegate success for d1 + failure on d2
	t := time.Now()
	err = app.MillionsKeeper.OnUndelegateWithdrawalOnNativeChainCompleted(ctx, poolID, w1.WithdrawalId, su1, &t, false)
	suite.Require().NoError(err)
	err = app.MillionsKeeper.OnUndelegateWithdrawalOnNativeChainCompleted(ctx, poolID, w2.WithdrawalId, su2, &t, true)
	suite.Require().NoError(err)

	// Validators bounded amount should be d2 now since d2 undelegate request failed
	pool, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(d2.Amount.Amount, pool.Validators[0].BondedAmount)

	// Simulate undelegate success for d2
	app.MillionsKeeper.UpdateWithdrawalStatus(ctx, poolID, w2.WithdrawalId, millionstypes.WithdrawalState_IcaUndelegate, nil, false)
	err = app.MillionsKeeper.UndelegateWithdrawalOnNativeChain(ctx, poolID, w2.WithdrawalId)
	suite.Require().Error(err)
	err = app.MillionsKeeper.OnUndelegateWithdrawalOnNativeChainCompleted(ctx, poolID, w2.WithdrawalId, su2, &t, false)
	suite.Require().NoError(err)

	// Validators bounded amount should be 0 now
	pool, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroInt(), pool.Validators[0].BondedAmount)
}

// TestPool_UpdatePool tests the different conditions for a proposal pool update
func (suite *KeeperTestSuite) TestPool_UpdatePool() {
	app := suite.app
	ctx := suite.ctx

	valAddrs := []string{
		"lumvaloper16rlynj5wvzwts5lqep0je5q4m3eaepn5cqj38s",
	}

	drawDelta1 := 1 * time.Hour
	newDrawSchedule := millionstypes.DrawSchedule{
		InitialDrawAt: ctx.BlockTime().Add(2 * time.Hour),
		DrawDelta:     drawDelta1,
	}
	newPrizeStrategy := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 90, DrawProbability: sdk.NewDec(1), Quantity: 10},
			{PoolPercent: 10, DrawProbability: sdk.NewDec(1), Quantity: 10},
		},
	}

	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: 1,
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
		State:              millionstypes.PoolState_Ready,
		MinDepositAmount:   sdk.NewInt(1_000_000),
	}))

	pool, err := app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)

	// State should be unchanged as status ready and incoming state is killed
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, valAddrs, &pool.MinDepositAmount, &pool.DrawSchedule, &pool.PrizeStrategy, millionstypes.PoolState_Killed)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolStateChangeNotAllowed)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.PoolState_Ready, pool.State)

	// State should be updated to status to PoolState_Paused
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, valAddrs, &pool.MinDepositAmount, &pool.DrawSchedule, &pool.PrizeStrategy, millionstypes.PoolState_Paused)
	suite.Require().NoError(err)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.PoolState_Paused, pool.State)

	// State should remain to PoolState_Paused
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, valAddrs, nil, nil, nil, millionstypes.PoolState_Unspecified)
	suite.Require().NoError(err)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.PoolState_Paused, pool.State)

	// State should remain to PoolState_Paused
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, valAddrs, nil, nil, nil, millionstypes.PoolState_Created)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolStateChangeNotAllowed)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.PoolState_Paused, pool.State)

	// Test that the same pool can be ready again
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, valAddrs, &pool.MinDepositAmount, &pool.DrawSchedule, &pool.PrizeStrategy, millionstypes.PoolState_Ready)
	suite.Require().NoError(err)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.PoolState_Ready, pool.State)

	// Test new drawSchedule
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, valAddrs, &pool.MinDepositAmount, &newDrawSchedule, &pool.PrizeStrategy, millionstypes.PoolState_Paused)
	suite.Require().NoError(err)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(newDrawSchedule, pool.DrawSchedule)

	// Test new prizeStrategy
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, valAddrs, &pool.MinDepositAmount, &newDrawSchedule, &newPrizeStrategy, millionstypes.PoolState_Ready)
	suite.Require().NoError(err)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(newPrizeStrategy, pool.PrizeStrategy)
}
