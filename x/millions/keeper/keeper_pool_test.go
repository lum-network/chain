package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/lum-network/chain/x/epochs/types"
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

func (suite *KeeperTestSuite) TestPool_Helpers() {
	poolLocal := millionstypes.Pool{Bech32PrefixAccAddr: sdk.GetConfig().GetBech32AccountAddrPrefix()}
	cases := []struct {
		address        string
		isLocalAddress bool
		expectError    bool
	}{
		{address: "", isLocalAddress: false, expectError: true},
		{address: "not-an-address", isLocalAddress: false, expectError: true},
		{address: cosmosIcaDepositAddress, isLocalAddress: true, expectError: true},
		{address: suite.valAddrs[0].String(), isLocalAddress: false, expectError: true},
		{address: suite.addrs[0].String(), isLocalAddress: true, expectError: false},
		{address: suite.moduleAddrs[0].String(), isLocalAddress: true, expectError: false},
	}
	for i, c := range cases {
		isLocalAddress, addr, err := poolLocal.AccAddressFromBech32(c.address)
		if c.expectError {
			suite.Require().Error(err, "case %d", i)
			suite.Require().Nil(addr, "case %d", i)
		} else {
			suite.Require().Equal(c.isLocalAddress, isLocalAddress, "case %d", i)
			suite.Require().NotNil(addr, "case %d", i)
		}
	}

	poolRemote := millionstypes.Pool{Bech32PrefixAccAddr: "cosmos"}
	cases = []struct {
		address        string
		isLocalAddress bool
		expectError    bool
	}{
		{address: "", isLocalAddress: false, expectError: true},
		{address: "not-an-address", isLocalAddress: false, expectError: true},
		{address: suite.valAddrs[0].String(), isLocalAddress: false, expectError: true},
		{address: "osmo1clpqr4nrk4khgkxj78fcwwh6dl3uw4epasmvnj", isLocalAddress: false, expectError: true},
		{address: cosmosIcaDepositAddress, isLocalAddress: false, expectError: false},
		{address: suite.addrs[0].String(), isLocalAddress: true, expectError: false},
		{address: suite.moduleAddrs[0].String(), isLocalAddress: true, expectError: false},
	}
	for i, c := range cases {
		isLocalAddress, addr, err := poolRemote.AccAddressFromBech32(c.address)
		if c.expectError {
			suite.Require().Error(err, "case %d", i)
			suite.Require().Nil(addr, "case %d", i)
		} else {
			suite.Require().Equal(c.isLocalAddress, isLocalAddress, "case %d", i)
			suite.Require().NotNil(addr, "case %d", i)
		}
	}
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
		time.Duration(millionstypes.DefaultUnbondingDuration),
		sdk.NewInt(millionstypes.DefaultMaxUnbondingEntries),
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
	epochInfo, err := TriggerEpochUpdate(suite)
	suite.Require().NoError(err)
	_, err = TriggerEpochTrackerUpdate(suite, epochInfo)
	suite.Require().NoError(err)

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
		State:            millionstypes.WithdrawalState_Pending,
		DepositorAddress: d1.DepositorAddress,
		ToAddress:        d1.DepositorAddress,
		Amount:           d1.Amount,
	}
	app.MillionsKeeper.AddWithdrawal(ctx, w1)
	su1 := pool.ComputeSplitUndelegations(ctx, w1.Amount.Amount)
	suite.Require().Len(su1, 1)

	withdrawals := app.MillionsKeeper.ListWithdrawals(ctx)

	err = app.MillionsKeeper.AddEpochUnbonding(ctx, withdrawals[0], false)
	suite.Require().NoError(err)

	w2 := millionstypes.Withdrawal{
		PoolId:           poolID,
		DepositId:        d2.DepositId,
		WithdrawalId:     app.MillionsKeeper.GetNextWithdrawalIdAndIncrement(ctx),
		State:            millionstypes.WithdrawalState_Pending,
		DepositorAddress: d2.DepositorAddress,
		ToAddress:        d2.DepositorAddress,
		Amount:           d2.Amount,
	}
	app.MillionsKeeper.AddWithdrawal(ctx, w2)
	su2 := pool.ComputeSplitUndelegations(ctx, w2.Amount.Amount)
	suite.Require().Len(su2, 1)

	withdrawals = app.MillionsKeeper.ListWithdrawals(ctx)

	err = app.MillionsKeeper.AddEpochUnbonding(ctx, withdrawals[1], false)
	suite.Require().NoError(err)

	// Get the millions internal module tracker
	epochTracker, err := app.MillionsKeeper.GetEpochTracker(ctx, epochstypes.DAY_EPOCH, millionstypes.WithdrawalTrackerType)
	suite.Require().NoError(err)

	// Get epoch unbonding
	currentEpochUnbonding, err := app.MillionsKeeper.GetEpochPoolUnbonding(ctx, epochTracker.EpochNumber, poolID)
	suite.Require().NoError(err)

	// Simulate undelegate failure for d1 + d2
	t := time.Now()
	err = app.MillionsKeeper.RemoveEpochUnbonding(ctx, currentEpochUnbonding)
	suite.Require().NoError(err)

	err = app.MillionsKeeper.OnUndelegateWithdrawalsOnRemoteZoneCompleted(ctx, poolID, []uint64{w1.WithdrawalId, w2.WithdrawalId}, &t, true)
	suite.Require().Error(err)

	// Validators bounded amount should be d1 + d2
	pool, err = app.MillionsKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(d1.Amount.Amount.Add(d2.Amount.Amount), pool.Validators[0].BondedAmount)

	// Error is intended
	err = app.MillionsKeeper.UndelegateWithdrawalsOnRemoteZone(ctx, currentEpochUnbonding)
	suite.Require().Error(err)
	// Simulate undelegate success for d1 + d2
	err = app.MillionsKeeper.OnUndelegateWithdrawalsOnRemoteZoneCompleted(ctx, poolID, []uint64{w1.WithdrawalId, w2.WithdrawalId}, &t, false)
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
	goCtx := sdk.WrapSDKContext(ctx)
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)

	privKeys := make([]cryptotypes.PubKey, 5)
	addrs := make([]sdk.AccAddress, 5)
	validators := make([]stakingtypes.Validator, 5)

	drawDelta1 := 1 * time.Hour
	newDrawSchedule := millionstypes.DrawSchedule{
		InitialDrawAt: ctx.BlockTime().Add(2 * time.Hour),
		DrawDelta:     drawDelta1,
	}
	newDepositAmount := sdk.NewInt(2_000_000)
	newPrizeStrategy := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 90, DrawProbability: sdk.NewDec(1), Quantity: 10},
			{PoolPercent: 10, DrawProbability: sdk.NewDec(1), Quantity: 10},
		},
	}

	UnbondingDuration := time.Duration(millionstypes.DefaultUnbondingDuration)
	maxUnbondingEntries := sdk.NewInt(millionstypes.DefaultMaxUnbondingEntries)

	// create 5 validators
	for i := 0; i < 5; i++ {
		privKey := ed25519.GenPrivKey().PubKey()
		privKeys[i] = privKey
		addrs[i] = sdk.AccAddress(privKey.Address())

		validator, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[i]), privKeys[i], stakingtypes.Description{})
		suite.Require().NoError(err)
		validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, false)
		err = app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, validator.GetOperator())
		suite.Require().NoError(err)

		validators[i] = validator
		validators = append(validators, validator)
	}
	// Add pool with valSet
	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId: 1,
		Validators: []millionstypes.PoolValidator{
			{
				OperatorAddress: validators[0].OperatorAddress,
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: validators[1].OperatorAddress,
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: validators[2].OperatorAddress,
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: validators[3].OperatorAddress,
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: validators[4].OperatorAddress,
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
		},
	}))
	pool, err := app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)

	// Deposit to trigger delegation for local pool
	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		DepositorAddress: suite.addrs[0].String(),
		PoolId:           pool.PoolId,
		Amount:           sdk.NewCoin(localPoolDenom, sdk.NewInt(9_000_000)),
	})
	suite.Require().NoError(err)

	deposit, err := app.MillionsKeeper.GetPoolDeposit(ctx, uint64(1), uint64(1))
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.DepositState_Success, deposit.State)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)

	// Pool validators should have equal BondedAmount
	suite.Require().Equal(sdk.NewInt(1_800_000), pool.Validators[0].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_800_000), pool.Validators[1].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_800_000), pool.Validators[2].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_800_000), pool.Validators[3].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_800_000), pool.Validators[4].BondedAmount)

	// Simulate that 2 validators are bonded but inactive in the poolSet
	// - Validator 2 and 4 are being removed
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, []string{pool.Validators[0].GetOperatorAddress(), pool.Validators[2].GetOperatorAddress(), pool.Validators[4].GetOperatorAddress()}, nil, nil, nil, nil, nil, millionstypes.PoolState_Unspecified)
	suite.Require().NoError(err)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)

	// The remaining pool amount of 9_000_000 should be divided by 3 among the remaining active validators
	suite.Require().Equal(sdk.NewInt(3_000_000), pool.Validators[0].BondedAmount)
	suite.Require().Equal(sdk.NewInt(0), pool.Validators[1].BondedAmount)
	suite.Require().Equal(sdk.NewInt(3_000_000), pool.Validators[2].BondedAmount)
	suite.Require().Equal(sdk.NewInt(0), pool.Validators[3].BondedAmount)
	suite.Require().Equal(sdk.NewInt(3_000_000), pool.Validators[4].BondedAmount)
	suite.Require().Equal(true, pool.Validators[0].IsEnabled)
	suite.Require().Equal(false, pool.Validators[1].IsEnabled)
	suite.Require().Equal(true, pool.Validators[2].IsEnabled)
	suite.Require().Equal(false, pool.Validators[3].IsEnabled)
	suite.Require().Equal(true, pool.Validators[4].IsEnabled)

	// UpdatePool with new params
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, nil, &newDepositAmount, &UnbondingDuration, &maxUnbondingEntries, &newDrawSchedule, &newPrizeStrategy, millionstypes.PoolState_Paused)
	suite.Require().NoError(err)
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	// New Deposit amount should be 2_000_000
	suite.Require().Equal(newDepositAmount, pool.MinDepositAmount)
	// New UnbondingDuration
	suite.Require().Equal(UnbondingDuration, pool.UnbondingDuration)
	// New maxUnbondingEntries
	suite.Require().Equal(maxUnbondingEntries, pool.MaxUnbondingEntries)
	// PoolState should be paused
	suite.Require().Equal(millionstypes.PoolState_Paused, pool.State)
	// New prize strategy applied
	suite.Require().Equal(newPrizeStrategy, pool.PrizeStrategy)
	// New draw schedule applied
	suite.Require().Equal(newDrawSchedule, pool.DrawSchedule)

	// UpdatePool with invalid pool_state -> PoolState_Killed
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, nil, &newDepositAmount, &UnbondingDuration, &maxUnbondingEntries, &newDrawSchedule, &newPrizeStrategy, millionstypes.PoolState_Killed)
	suite.Require().ErrorIs(err, millionstypes.ErrPoolStateChangeNotAllowed)
	// Grab fresh pool instance
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.PoolState_Paused, pool.State)

	// UpdatePool with invalid pool_state -> PoolState_Unspecified
	// Should remain with the current poolState
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, nil, &newDepositAmount, &UnbondingDuration, &maxUnbondingEntries, &newDrawSchedule, &newPrizeStrategy, millionstypes.PoolState_Unspecified)
	suite.Require().NoError(err)
	// Grab fresh pool instance
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.PoolState_Paused, pool.State)

	// UpdatePool with valid pool_state -> PoolState_Ready
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, nil, &newDepositAmount, &UnbondingDuration, &maxUnbondingEntries, &newDrawSchedule, &newPrizeStrategy, millionstypes.PoolState_Ready)
	suite.Require().NoError(err)
	// Grab fresh pool instance
	pool, err = app.MillionsKeeper.GetPool(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(millionstypes.PoolState_Ready, pool.State)

	// Create remote pool with 5 remote valAddreses
	valAddrsRemote := []string{
		"cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c",
		"cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn",
		"cosmosvaloper1fsg635n5vgc7jazz9sx5725wnc3xqgr7awxaag",
		"cosmosvaloper1gpx52r9h3zeul45amvcy2pysgvcwddxrgx6cnv",
		"cosmosvaloper1vvwtk805lxehwle9l4yudmq6mn0g32px9xtkhc",
	}

	app.MillionsKeeper.AddPool(ctx, newValidPool(suite, millionstypes.Pool{
		PoolId:              2,
		Bech32PrefixValAddr: "cosmosvaloper",
		ChainId:             "cosmos",
		Denom:               "uatom",
		NativeDenom:         "uatom",
		ConnectionId:        "connection-id",
		TransferChannelId:   "transferChannel-id",
		Validators: []millionstypes.PoolValidator{
			{
				OperatorAddress: valAddrsRemote[0],
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: valAddrsRemote[1],
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: valAddrsRemote[2],
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: valAddrsRemote[3],
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: valAddrsRemote[4],
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
		},
		IcaDepositAddress:   cosmosIcaDepositAddress,
		IcaPrizepoolAddress: cosmosIcaPrizePoolAddress,
		State:               millionstypes.PoolState_Ready,
	}))
	// Add new deposit to pool
	app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
		DepositId:        1,
		PoolId:           2,
		State:            millionstypes.DepositState_IcaDelegate,
		Amount:           sdk.NewCoin(remotePoolDenom, sdk.NewInt(5_000_000)),
		DepositorAddress: suite.addrs[0].String(),
		WinnerAddress:    suite.addrs[0].String(),
	})
	pool, err = app.MillionsKeeper.GetPool(ctx, 2)
	suite.Require().NoError(err)
	deposit, err = app.MillionsKeeper.GetPoolDeposit(ctx, uint64(2), uint64(1))
	suite.Require().NoError(err)

	// ComputeSplitDelegations to ease the simulation of DelegateDeposit for remote pool
	splits := pool.ComputeSplitDelegations(ctx, deposit.Amount.Amount)
	suite.Require().Len(splits, 5)
	// Simulate successful delegation
	err = app.MillionsKeeper.OnDelegateDepositOnNativeChainCompleted(ctx, pool.PoolId, deposit.DepositId, splits, false)
	suite.Require().NoError(err)
	pool, err = app.MillionsKeeper.GetPool(ctx, 2)
	suite.Require().NoError(err)

	// Pool validators should have equal BondedAmount
	suite.Require().Equal(sdk.NewInt(1_000_000), pool.Validators[0].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_000_000), pool.Validators[1].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_000_000), pool.Validators[2].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_000_000), pool.Validators[3].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_000_000), pool.Validators[4].BondedAmount)

	// Simulate validator 1 that gets removed from the valSet
	err = app.MillionsKeeper.UpdatePool(ctx, pool.PoolId, []string{pool.Validators[1].GetOperatorAddress(), pool.Validators[2].GetOperatorAddress(), pool.Validators[3].GetOperatorAddress(), pool.Validators[4].GetOperatorAddress()}, nil, nil, nil, nil, nil, millionstypes.PoolState_Unspecified)
	suite.Require().Error(err)
	// No active channel for this owner
	pool, err = app.MillionsKeeper.GetPool(ctx, 2)
	suite.Require().NoError(err)

	// The remaining pool amount of 5_000_000 should be divided by 4 among the remaining active validators
	suite.Require().Equal(sdk.NewInt(0), pool.Validators[0].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_250_000), pool.Validators[1].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_250_000), pool.Validators[2].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_250_000), pool.Validators[3].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_250_000), pool.Validators[4].BondedAmount)
	suite.Require().Equal(false, pool.Validators[0].IsEnabled)
	suite.Require().Equal(true, pool.Validators[1].IsEnabled)
	suite.Require().Equal(true, pool.Validators[2].IsEnabled)
	suite.Require().Equal(true, pool.Validators[3].IsEnabled)
	suite.Require().Equal(true, pool.Validators[4].IsEnabled)

	// Simulate failed ICA callback
	// The initial splits amount for the inactive validator 1 was 1_000_000
	splits = pool.ComputeSplitDelegations(ctx, sdk.NewInt(1_000_000))
	err = app.MillionsKeeper.OnRedelegateToRemoteZoneCompleted(ctx, pool.PoolId, pool.Validators[0].GetOperatorAddress(), splits, true)
	suite.Require().NoError(err)
	pool, err = app.MillionsKeeper.GetPool(ctx, 2)
	suite.Require().NoError(err)

	// The BondedAmount should be back to normal as for pre-op Redelegate
	suite.Require().Equal(sdk.NewInt(1_000_000), pool.Validators[0].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_000_000), pool.Validators[1].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_000_000), pool.Validators[2].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_000_000), pool.Validators[3].BondedAmount)
	suite.Require().Equal(sdk.NewInt(1_000_000), pool.Validators[4].BondedAmount)
	suite.Require().Equal(false, pool.Validators[0].IsEnabled)
	suite.Require().Equal(true, pool.Validators[1].IsEnabled)
	suite.Require().Equal(true, pool.Validators[2].IsEnabled)
	suite.Require().Equal(true, pool.Validators[3].IsEnabled)
	suite.Require().Equal(true, pool.Validators[4].IsEnabled)
}
