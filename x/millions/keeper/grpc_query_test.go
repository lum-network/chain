package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

// TestGRPC_Params runs simple Params GRPCs integration tests.
func (suite *KeeperTestSuite) TestGRPC_Query_Params() {
	app := suite.app
	ctx := suite.ctx
	queryServer := millionskeeper.NewQueryServerImpl(*app.MillionsKeeper)

	// Get ctx params
	initialParams := app.MillionsKeeper.GetParams(ctx)

	// Get query params
	paramsRes, err := queryServer.Params(ctx, &millionstypes.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(initialParams.FeesStakers, paramsRes.Params.FeesStakers)
	suite.Require().Equal(initialParams.MaxDrawScheduleDelta, paramsRes.Params.MaxDrawScheduleDelta)
	suite.Require().Equal(initialParams.MaxPrizeBatchQuantity, paramsRes.Params.MaxPrizeBatchQuantity)
	suite.Require().Equal(initialParams.MinDepositAmount, paramsRes.Params.MinDepositAmount)
	suite.Require().Equal(initialParams.MinDepositDrawDelta, paramsRes.Params.MinDepositDrawDelta)
	suite.Require().Equal(initialParams.MinDrawScheduleDelta, paramsRes.Params.MinDrawScheduleDelta)
	suite.Require().Equal(initialParams.PrizeExpirationDelta, paramsRes.Params.PrizeExpirationDelta)
}

// TestGRPC_Query_Pool runs simple integration tests for pools and pool.
func (suite *KeeperTestSuite) TestGRPC_Query_Pool() {
	app := suite.app
	ctx := suite.ctx
	queryServer := millionskeeper.NewQueryServerImpl(*app.MillionsKeeper)

	nbrItems := 4
	// Test run generates
	// - 5 pools
	for i := 0; i < nbrItems; i++ {
		pool := newValidPool(suite, millionstypes.Pool{
			PrizeStrategy: millionstypes.PrizeStrategy{
				PrizeBatches: []millionstypes.PrizeBatch{
					{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
					{PoolPercent: 50, Quantity: 4, DrawProbability: floatToDec(1.00)},
				},
			},
		})
		// Force the available pool prize
		pool.AvailablePrizePool = sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000))
		err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pool.IcaPrizepoolAddress), sdk.NewCoins(pool.AvailablePrizePool))
		suite.Require().NoError(err)

		app.MillionsKeeper.AddPool(ctx, pool)
	}

	// Test Pools
	poolsRes, err := queryServer.Pools(ctx, &millionstypes.QueryPoolsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(poolsRes.GetPools(), nbrItems)
	suite.Require().Equal(uint64(1), poolsRes.GetPools()[0].PoolId)
	suite.Require().Equal(uint64(nbrItems), poolsRes.GetPools()[nbrItems-1].PoolId)
	// Test Pools with pagination
	poolsResPaginated, err := queryServer.Pools(ctx, &millionstypes.QueryPoolsRequest{Pagination: &query.PageRequest{Offset: 1, Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().Len(poolsResPaginated.GetPools(), 2)
	suite.Require().Equal(poolsRes.GetPools()[1].PoolId, poolsResPaginated.GetPools()[0].PoolId)
	suite.Require().Equal(poolsRes.GetPools()[2].PoolId, poolsResPaginated.GetPools()[1].PoolId)

	// Test Pool with wrong poolID
	_, err = queryServer.Pool(ctx, &millionstypes.QueryPoolRequest{PoolId: 0})
	suite.Require().Error(err)
	// Test Pool
	poolRes, err := queryServer.Pool(ctx, &millionstypes.QueryPoolRequest{PoolId: 1})
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), poolRes.GetPool().PoolId)
	pools := app.MillionsKeeper.ListPools(ctx)
	suite.Require().Equal(sdk.NewCoin(pools[0].Denom, sdk.NewInt(1_000_000)), pools[0].AvailablePrizePool)
}

// TestGRPC_Query_Deposit runs simple integration tests for deposits and pools.
func (suite *KeeperTestSuite) TestGRPC_Query_Deposit() {
	app := suite.app
	ctx := suite.ctx
	queryServer := millionskeeper.NewQueryServerImpl(*app.MillionsKeeper)

	nbrItems := 4
	// Test run generates
	// - 5 pools
	// - 5 accounts
	// - 32 deposits
	for i := 0; i < nbrItems; i++ {
		pool := newValidPool(suite, millionstypes.Pool{
			PrizeStrategy: millionstypes.PrizeStrategy{
				PrizeBatches: []millionstypes.PrizeBatch{
					{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
					{PoolPercent: 50, Quantity: 4, DrawProbability: floatToDec(1.00)},
				},
			},
		})

		// Force the available pool prize
		pool.AvailablePrizePool = sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000))
		err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pool.IcaPrizepoolAddress), sdk.NewCoins(pool.AvailablePrizePool))
		suite.Require().NoError(err)

		app.MillionsKeeper.AddPool(ctx, pool)

		// Create deposits
		for i := 0; i < nbrItems; i++ {
			app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
				DepositorAddress: suite.addrs[i].String(),
				WinnerAddress:    suite.addrs[i].String(),
				PoolId:           pool.PoolId,
				Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
				State:            millionstypes.DepositState_Success,
			})
			app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
				DepositorAddress: suite.addrs[i].String(),
				WinnerAddress:    suite.addrs[i].String(),
				PoolId:           pool.PoolId,
				Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(500_000)),
				State:            millionstypes.DepositState_Success,
			})
		}
	}

	// Test Deposits
	depositorsRes, err := queryServer.Deposits(ctx, &millionstypes.QueryDepositsRequest{})
	suite.Require().NoError(err)
	poolsFound := map[uint64]interface{}{}
	depositsFound := map[string]interface{}{}
	for _, depositor := range depositorsRes.GetDeposits() {
		poolsFound[depositor.PoolId] = true
		depositsFound[fmt.Sprintf("%s-%d-%d", depositor.DepositorAddress, depositor.GetPoolId(), depositor.GetAmount().Amount.Int64())] = true
	}
	suite.Require().Len(poolsFound, nbrItems)               // 5 pools
	suite.Require().Len(depositsFound, nbrItems*nbrItems*2) // 5 pools * 5 accounts * 2 deposits

	depositorsRes, err = queryServer.Deposits(ctx, &millionstypes.QueryDepositsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(depositorsRes.GetDeposits(), nbrItems*nbrItems*2)
	// Test Deposits with pagination
	depositorsResPaginated, err := queryServer.Deposits(ctx, &millionstypes.QueryDepositsRequest{Pagination: &query.PageRequest{Offset: 1, Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().Len(depositorsResPaginated.GetDeposits(), 2)
	suite.Require().Equal(depositorsRes.GetDeposits()[1].DepositId, depositorsResPaginated.GetDeposits()[0].DepositId)
	suite.Require().Equal(depositorsRes.GetDeposits()[2].DepositId, depositorsResPaginated.GetDeposits()[1].DepositId)

	// Test AccountDeposits with wrong depositorAddr
	_, err = queryServer.AccountDeposits(ctx, &millionstypes.QueryAccountDepositsRequest{DepositorAddress: ""})
	suite.Require().Error(err)
	// Test AccountDeposits
	accountDeposits, err := queryServer.AccountDeposits(ctx, &millionstypes.QueryAccountDepositsRequest{DepositorAddress: suite.addrs[0].String()})
	suite.Require().NoError(err)
	suite.Require().Len(accountDeposits.GetDeposits(), nbrItems*2) // 1 account 8 deposits
	// Test AccountDeposits with pagination
	accountDepositsPaginated, err := queryServer.AccountDeposits(ctx, &millionstypes.QueryAccountDepositsRequest{DepositorAddress: suite.addrs[0].String(), Pagination: &query.PageRequest{Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().Len(accountDepositsPaginated.GetDeposits(), 2)
	suite.Require().Equal(accountDeposits.GetDeposits()[0].DepositId, accountDepositsPaginated.GetDeposits()[0].DepositId)
	suite.Require().Equal(accountDeposits.GetDeposits()[1].DepositId, accountDepositsPaginated.GetDeposits()[1].DepositId)

	// Test PoolDeposits with wrong poolID
	_, err = queryServer.PoolDeposits(ctx, &millionstypes.QueryPoolDepositsRequest{PoolId: 0})
	suite.Require().Error(err)
	// Test PoolDeposits
	poolDepositsRes, err := queryServer.PoolDeposits(ctx, &millionstypes.QueryPoolDepositsRequest{PoolId: 1})
	suite.Require().NoError(err)
	suite.Require().Len(poolDepositsRes.GetDeposits(), nbrItems*2) // 1 pool * 8 deposits
	// Test PoolDeposits with pagination
	poolDepositsResPaginated, err := queryServer.PoolDeposits(ctx, &millionstypes.QueryPoolDepositsRequest{PoolId: 1, Pagination: &query.PageRequest{Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().Len(poolDepositsResPaginated.GetDeposits(), 2)
	suite.Require().Equal(poolDepositsRes.GetDeposits()[0].DepositId, poolDepositsResPaginated.GetDeposits()[0].DepositId)
	suite.Require().Equal(poolDepositsRes.GetDeposits()[1].DepositId, poolDepositsResPaginated.GetDeposits()[1].DepositId)

	// Test PoolDeposit with wrong poolID
	_, err = queryServer.PoolDeposit(ctx, &millionstypes.QueryPoolDepositRequest{PoolId: 0, DepositId: 1})
	suite.Require().Error(err)
	// Test PoolDeposit with wrong depositID
	_, err = queryServer.PoolDeposit(ctx, &millionstypes.QueryPoolDepositRequest{PoolId: 1, DepositId: 0})
	suite.Require().Error(err)
	// Test PoolDeposit
	poolDeposit, err := queryServer.PoolDeposit(ctx, &millionstypes.QueryPoolDepositRequest{PoolId: 1, DepositId: 1})
	suite.Require().NoError(err)
	suite.Require().Equal(poolDeposit.Deposit.DepositId, uint64(1))

	// Test AccountPoolDeposits with wrong depositorAddr
	_, err = queryServer.AccountPoolDeposits(ctx, &millionstypes.QueryAccountPoolDepositsRequest{DepositorAddress: "", PoolId: 1})
	suite.Require().Error(err)
	// Test AccountPoolDeposits with wrong poolID
	_, err = queryServer.AccountPoolDeposits(ctx, &millionstypes.QueryAccountPoolDepositsRequest{DepositorAddress: suite.addrs[0].String(), PoolId: 0})
	suite.Require().Error(err)
	// Test AccountPoolDeposits
	accountPoolDeposits, err := queryServer.AccountPoolDeposits(ctx, &millionstypes.QueryAccountPoolDepositsRequest{DepositorAddress: suite.addrs[0].String(), PoolId: 1})
	suite.Require().NoError(err)
	suite.Require().Len(accountPoolDeposits.GetDeposits(), 2) // 1 account * 2 deposits
	// Test AccountPoolDeposits with pagination
	accountPoolDepositsPaginated, err := queryServer.AccountPoolDeposits(ctx, &millionstypes.QueryAccountPoolDepositsRequest{DepositorAddress: suite.addrs[0].String(), PoolId: 1, Pagination: &query.PageRequest{Offset: 0, Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().Len(accountPoolDepositsPaginated.GetDeposits(), 2)
	suite.Require().Equal(accountPoolDeposits.GetDeposits()[0].DepositId, accountPoolDepositsPaginated.GetDeposits()[0].DepositId)
	suite.Require().Equal(accountPoolDeposits.GetDeposits()[1].DepositId, accountPoolDepositsPaginated.GetDeposits()[1].DepositId)
}

// TestGRPC_Query_Draw runs simple integration tests on draws and pools.
func (suite *KeeperTestSuite) TestGRPC_Query_Draw() {
	app := suite.app
	ctx := suite.ctx
	queryServer := millionskeeper.NewQueryServerImpl(*app.MillionsKeeper)

	nbrItems := 4
	// Test run generates
	// - 5 pools
	// - 5 accounts
	// - 32 deposits
	// - 16 draws
	for i := 0; i < nbrItems; i++ {
		pool := newValidPool(suite, millionstypes.Pool{
			PrizeStrategy: millionstypes.PrizeStrategy{
				PrizeBatches: []millionstypes.PrizeBatch{
					{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
					{PoolPercent: 50, Quantity: 4, DrawProbability: floatToDec(1.00)},
				},
			},
		})

		// Force the available pool prize
		pool.AvailablePrizePool = sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000))
		err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pool.IcaPrizepoolAddress), sdk.NewCoins(pool.AvailablePrizePool))
		suite.Require().NoError(err)

		app.MillionsKeeper.AddPool(ctx, pool)

		// Create deposits
		for i := 0; i < nbrItems; i++ {
			app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
				DepositorAddress: suite.addrs[i].String(),
				WinnerAddress:    suite.addrs[i].String(),
				PoolId:           pool.PoolId,
				Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
				State:            millionstypes.DepositState_Success,
			})
			app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
				DepositorAddress: suite.addrs[i].String(),
				WinnerAddress:    suite.addrs[i].String(),
				PoolId:           pool.PoolId,
				Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(500_000)),
				State:            millionstypes.DepositState_Success,
			})
		}

		// Launch new draw
		for i := 0; i < nbrItems; i++ {
			draw, err := app.MillionsKeeper.LaunchNewDraw(ctx, pool.PoolId)
			suite.Require().NoError(err)
			suite.Require().Equal(draw.PrizePool.Amount.Int64(), draw.TotalWinAmount.Int64())
		}
	}

	// Test Draws
	drawsRes, err := queryServer.Draws(ctx, &millionstypes.QueryDrawsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(drawsRes.GetDraws(), nbrItems*nbrItems)
	suite.Require().Equal(uint64(1), drawsRes.GetDraws()[0].PoolId)
	suite.Require().Equal(uint64(nbrItems), drawsRes.GetDraws()[nbrItems*nbrItems-1].PoolId)
	// Test Draws with pagination
	drawsResPaginated, err := queryServer.Draws(ctx, &millionstypes.QueryDrawsRequest{Pagination: &query.PageRequest{Offset: 1, Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().Len(drawsResPaginated.GetDraws(), 2)
	suite.Require().Equal(drawsRes.GetDraws()[1].DrawId, drawsResPaginated.GetDraws()[0].DrawId)
	suite.Require().Equal(drawsRes.GetDraws()[2].DrawId, drawsResPaginated.GetDraws()[1].DrawId)

	// Test PoolDraws with wrong poolID
	_, err = queryServer.PoolDraws(ctx, &millionstypes.QueryPoolDrawsRequest{PoolId: 0})
	suite.Require().Error(err)
	// Test PoolDraws
	poolDrawsRes, err := queryServer.PoolDraws(ctx, &millionstypes.QueryPoolDrawsRequest{PoolId: 1})
	suite.Require().NoError(err)
	suite.Require().Len(poolDrawsRes.GetDraws(), nbrItems)
	suite.Require().Equal(uint64(1), poolDrawsRes.GetDraws()[0].PoolId)
	suite.Require().Equal(uint64(1), poolDrawsRes.GetDraws()[nbrItems-1].PoolId)
	// Test PoolDraws with pagination
	poolDrawsResPaginated, err := queryServer.PoolDraws(ctx, &millionstypes.QueryPoolDrawsRequest{PoolId: 1, Pagination: &query.PageRequest{Offset: 2, Limit: 3}})
	suite.Require().NoError(err)
	suite.Require().Len(poolDrawsResPaginated.GetDraws(), 2)
	suite.Require().Equal(poolDrawsRes.GetDraws()[2].DrawId, poolDrawsResPaginated.GetDraws()[0].DrawId)
	suite.Require().Equal(poolDrawsRes.GetDraws()[3].DrawId, poolDrawsResPaginated.GetDraws()[1].DrawId)

	// Test PoolDraw with wrong poolID
	_, err = queryServer.PoolDraw(ctx, &millionstypes.QueryPoolDrawRequest{PoolId: 0, DrawId: 1})
	suite.Require().Error(err)
	// Test PoolDraw with wrong drawID
	_, err = queryServer.PoolDraw(ctx, &millionstypes.QueryPoolDrawRequest{PoolId: 1, DrawId: 0})
	suite.Require().Error(err)
	// Test PoolDraw
	poolDrawRes, err := queryServer.PoolDraw(ctx, &millionstypes.QueryPoolDrawRequest{PoolId: 1, DrawId: 1})
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), poolDrawRes.GetDraw().PoolId)
	suite.Require().Equal(uint64(1), poolDrawRes.GetDraw().DrawId)
	// Test PoolDraw with different id
	poolDrawRes, err = queryServer.PoolDraw(ctx, &millionstypes.QueryPoolDrawRequest{PoolId: uint64(nbrItems), DrawId: uint64(nbrItems)})
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(nbrItems), poolDrawRes.GetDraw().PoolId)
	suite.Require().Equal(uint64(nbrItems), poolDrawRes.GetDraw().DrawId)
}

// TestGRPC_Query_Prize runs simple integration tests on prizes, pools and draws.
func (suite *KeeperTestSuite) TestGRPC_Query_Prize() {
	app := suite.app
	ctx := suite.ctx
	queryServer := millionskeeper.NewQueryServerImpl(*app.MillionsKeeper)

	nbrItems := 4
	// Test run generates
	// - 5 pools
	// - 5 accounts
	// - 32 deposits
	// - 16 draws
	for i := 0; i < nbrItems; i++ {
		pool := newValidPool(suite, millionstypes.Pool{
			PrizeStrategy: millionstypes.PrizeStrategy{
				PrizeBatches: []millionstypes.PrizeBatch{
					{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
					{PoolPercent: 50, Quantity: 4, DrawProbability: floatToDec(1.00)},
				},
			},
		})

		// Force the available pool prize
		pool.AvailablePrizePool = sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000))
		err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pool.IcaPrizepoolAddress), sdk.NewCoins(pool.AvailablePrizePool))
		suite.Require().NoError(err)

		app.MillionsKeeper.AddPool(ctx, pool)

		// Create deposits
		for i := 0; i < nbrItems; i++ {
			app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
				DepositorAddress: suite.addrs[i].String(),
				WinnerAddress:    suite.addrs[i].String(),
				PoolId:           pool.PoolId,
				Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
				State:            millionstypes.DepositState_Success,
			})
			app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
				DepositorAddress: suite.addrs[i].String(),
				WinnerAddress:    suite.addrs[i].String(),
				PoolId:           pool.PoolId,
				Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(500_000)),
				State:            millionstypes.DepositState_Success,
			})
		}

		// Launch new draw
		for i := 0; i < nbrItems; i++ {
			draw, err := app.MillionsKeeper.LaunchNewDraw(ctx, pool.PoolId)
			suite.Require().NoError(err)
			suite.Require().Equal(draw.PrizePool.Amount.Int64(), draw.TotalWinAmount.Int64())
		}
	}

	// Test Prizes
	prizeRes, err := queryServer.Prizes(ctx, &millionstypes.QueryPrizesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(prizeRes.GetPrizes(), nbrItems*5)
	// Test Prizes with pagination
	prizeResPaginated, err := queryServer.Prizes(ctx, &millionstypes.QueryPrizesRequest{Pagination: &query.PageRequest{Offset: 1, Limit: 3}})
	suite.Require().NoError(err)
	suite.Require().Len(prizeResPaginated.GetPrizes(), 3)
	suite.Require().Equal(prizeRes.GetPrizes()[1].PrizeId, prizeResPaginated.GetPrizes()[0].PrizeId)
	suite.Require().Equal(prizeRes.GetPrizes()[2].PrizeId, prizeResPaginated.GetPrizes()[1].PrizeId)
	suite.Require().Equal(prizeRes.GetPrizes()[3].PrizeId, prizeResPaginated.GetPrizes()[2].PrizeId)

	// Test AccountPrizes with wrong winnerAddr
	_, err = queryServer.AccountPrizes(ctx, &millionstypes.QueryAccountPrizesRequest{WinnerAddress: ""})
	suite.Require().Error(err)
	// Test AccountPrizes
	accountPrizes, err := queryServer.AccountPrizes(ctx, &millionstypes.QueryAccountPrizesRequest{WinnerAddress: suite.addrs[2].String()})
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(accountPrizes.GetPrizes()), 0)
	// Test AccountPrizes with pagination
	accountPrizesPaginated, err := queryServer.AccountPrizes(ctx, &millionstypes.QueryAccountPrizesRequest{WinnerAddress: suite.addrs[2].String(), Pagination: &query.PageRequest{Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(accountPrizesPaginated.GetPrizes()), 0)

	// Test PoolPrizes with wrong poolID
	_, err = queryServer.PoolPrizes(ctx, &millionstypes.QueryPoolPrizesRequest{PoolId: 0})
	suite.Require().Error(err)
	// Test PoolPrizes
	poolPrizes, err := queryServer.PoolPrizes(ctx, &millionstypes.QueryPoolPrizesRequest{PoolId: 3})
	suite.Require().NoError(err)
	suite.Require().Len(poolPrizes.GetPrizes(), 5)
	// Test PoolPrizes with pagination
	poolPrizesPaginated, err := queryServer.PoolPrizes(ctx, &millionstypes.QueryPoolPrizesRequest{PoolId: 3, Pagination: &query.PageRequest{Offset: 1, Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().Len(poolPrizesPaginated.GetPrizes(), 2)
	suite.Require().Equal(poolPrizes.GetPrizes()[1].PrizeId, poolPrizesPaginated.GetPrizes()[0].PrizeId)
	suite.Require().Equal(poolPrizes.GetPrizes()[2].PrizeId, poolPrizesPaginated.GetPrizes()[1].PrizeId)

	// Test AccountPoolPrizes with wrong winnerAddr
	_, err = queryServer.AccountPoolPrizes(ctx, &millionstypes.QueryAccountPoolPrizesRequest{WinnerAddress: "", PoolId: 1})
	suite.Require().Error(err)
	// Test AccountPoolPrizes with wrong poolID
	_, err = queryServer.AccountPoolPrizes(ctx, &millionstypes.QueryAccountPoolPrizesRequest{WinnerAddress: suite.addrs[0].String(), PoolId: 0})
	suite.Require().Error(err)
	// Test AccountPoolPrizes
	accountPoolPrizes, err := queryServer.AccountPoolPrizes(ctx, &millionstypes.QueryAccountPoolPrizesRequest{WinnerAddress: suite.addrs[0].String(), PoolId: 1})
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(accountPoolPrizes.GetPrizes()), 0)
	// Test AccountPoolPrizes with pagination
	accountPoolPrizesPaginated, err := queryServer.AccountPoolPrizes(ctx, &millionstypes.QueryAccountPoolPrizesRequest{WinnerAddress: suite.addrs[0].String(), PoolId: 1, Pagination: &query.PageRequest{Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(accountPoolPrizesPaginated.GetPrizes()), 0)

	// Test PoolDrawPrizes with wrong poolID
	_, err = queryServer.PoolDrawPrizes(ctx, &millionstypes.QueryPoolDrawPrizesRequest{PoolId: 0, DrawId: 1})
	suite.Require().Error(err)
	// Test PoolDrawPrizes with wrong drawID
	_, err = queryServer.PoolDrawPrizes(ctx, &millionstypes.QueryPoolDrawPrizesRequest{PoolId: 1, DrawId: 0})
	suite.Require().Error(err)
	// Test PoolDrawPrizes
	poolDrawPrizes, err := queryServer.PoolDrawPrizes(ctx, &millionstypes.QueryPoolDrawPrizesRequest{PoolId: 1, DrawId: 1})
	suite.Require().NoError(err)
	suite.Require().Len(poolDrawPrizes.GetPrizes(), 5)
	// Test PoolDrawPrizes with pagination
	poolDrawPrizesPaginated, err := queryServer.PoolDrawPrizes(ctx, &millionstypes.QueryPoolDrawPrizesRequest{PoolId: 1, DrawId: 1, Pagination: &query.PageRequest{Offset: 1, Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().Len(poolDrawPrizesPaginated.GetPrizes(), 2)
	suite.Require().Equal(poolDrawPrizes.GetPrizes()[1].PrizeId, poolDrawPrizesPaginated.GetPrizes()[0].PrizeId)
	suite.Require().Equal(poolDrawPrizes.GetPrizes()[2].PrizeId, poolDrawPrizesPaginated.GetPrizes()[1].PrizeId)

	// Test PoolDrawPrize with wrong poolID
	_, err = queryServer.PoolDrawPrize(ctx, &millionstypes.QueryPoolDrawPrizeRequest{PoolId: 0, DrawId: 1, PrizeId: 1})
	suite.Require().Error(err)
	// Test PoolDrawPrize with wrong drawID
	_, err = queryServer.PoolDrawPrize(ctx, &millionstypes.QueryPoolDrawPrizeRequest{PoolId: 1, DrawId: 0, PrizeId: 1})
	suite.Require().Error(err)
	// Test PoolDrawPrize with wrong prizeID
	_, err = queryServer.PoolDrawPrize(ctx, &millionstypes.QueryPoolDrawPrizeRequest{PoolId: 1, DrawId: 1, PrizeId: 0})
	suite.Require().Error(err)
	// Test PoolDrawPrize
	poolDrawPrize, err := queryServer.PoolDrawPrize(ctx, &millionstypes.QueryPoolDrawPrizeRequest{PoolId: 1, DrawId: 1, PrizeId: 1})
	suite.Require().NoError(err)
	suite.Require().Equal(poolDrawPrize.Prize.PrizeId, uint64(1))

	// Test AccountPoolDrawPrizes with wrong winnerAddr
	_, err = queryServer.AccountPoolDrawPrizes(ctx, &millionstypes.QueryAccountPoolDrawPrizesRequest{WinnerAddress: "", PoolId: 1, DrawId: 1})
	suite.Require().Error(err)
	// Test AccountPoolDrawPrizes with wrong poolID
	_, err = queryServer.AccountPoolDrawPrizes(ctx, &millionstypes.QueryAccountPoolDrawPrizesRequest{WinnerAddress: suite.addrs[1].String(), PoolId: 0, DrawId: 1})
	suite.Require().Error(err)
	// Test AccountPoolDrawPrizes with wrong drawID
	_, err = queryServer.AccountPoolDrawPrizes(ctx, &millionstypes.QueryAccountPoolDrawPrizesRequest{WinnerAddress: suite.addrs[1].String(), PoolId: 1, DrawId: 0})
	suite.Require().Error(err)
	// Test AccountPoolDrawPrizes
	accountPoolDrawPrizes, err := queryServer.AccountPoolDrawPrizes(ctx, &millionstypes.QueryAccountPoolDrawPrizesRequest{WinnerAddress: suite.addrs[1].String(), PoolId: 1, DrawId: 1})
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(accountPoolDrawPrizes.GetPrizes()), 0)
	// Test AccountPoolDrawPrizes with pagination
	accountPoolDrawPrizesPaginated, err := queryServer.AccountPoolDrawPrizes(ctx, &millionstypes.QueryAccountPoolDrawPrizesRequest{WinnerAddress: suite.addrs[1].String(), PoolId: 1, DrawId: 1, Pagination: &query.PageRequest{Limit: 2}})
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(accountPoolDrawPrizesPaginated.GetPrizes()), 0)
}

// TestGRPC_Query_Withdrawal runs simple integration tests for withdrawals and pools.
func (suite *KeeperTestSuite) TestGRPC_Query_Withdrawal() {
	app := suite.app
	ctx := suite.ctx
	queryServer := millionskeeper.NewQueryServerImpl(*app.MillionsKeeper)

	nbrItems := 4
	// Test run generates
	// - 5 pools
	// - 5 accounts
	// - 32 deposits
	// - 6 withdrawals
	for i := 0; i < nbrItems; i++ {
		pool := newValidPool(suite, millionstypes.Pool{
			PrizeStrategy: millionstypes.PrizeStrategy{
				PrizeBatches: []millionstypes.PrizeBatch{
					{PoolPercent: 50, Quantity: 1, DrawProbability: floatToDec(1.00)},
					{PoolPercent: 50, Quantity: 4, DrawProbability: floatToDec(1.00)},
				},
			},
		})

		// Force the available pool prize
		pool.AvailablePrizePool = sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000))
		err := app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(pool.IcaPrizepoolAddress), sdk.NewCoins(pool.AvailablePrizePool))
		suite.Require().NoError(err)

		app.MillionsKeeper.AddPool(ctx, pool)

		// Create deposits
		for i := 0; i < nbrItems; i++ {
			app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
				DepositorAddress: suite.addrs[i].String(),
				WinnerAddress:    suite.addrs[i].String(),
				PoolId:           pool.PoolId,
				Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(1_000_000)),
				State:            millionstypes.DepositState_Success,
			})
			app.MillionsKeeper.AddDeposit(ctx, &millionstypes.Deposit{
				DepositorAddress: suite.addrs[i].String(),
				WinnerAddress:    suite.addrs[i].String(),
				PoolId:           pool.PoolId,
				Amount:           sdk.NewCoin(pool.Denom, sdk.NewInt(500_000)),
				State:            millionstypes.DepositState_Success,
			})
		}
	}

	// For this test case, we get the deposits and take the first 6 to generate withdrawals
	deposits := app.MillionsKeeper.ListDeposits(ctx)
	deposits = deposits[:6]
	for i, deposit := range deposits {
		app.MillionsKeeper.AddWithdrawal(ctx, millionstypes.Withdrawal{
			PoolId:           deposit.PoolId,
			DepositId:        deposit.DepositId,
			DepositorAddress: suite.addrs[i].String(),
			Amount:           sdk.NewCoin(deposit.Amount.Denom, deposit.Amount.Amount),
			State:            millionstypes.WithdrawalState_IbcTransfer,
		})
	}

	// Test Withdrawals
	withdrawalRes, err := queryServer.Withdrawals(ctx, &millionstypes.QueryWithdrawalsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(withdrawalRes.GetWithdrawals(), 6)
	// Test Withdrawals with pagination
	withdrawalResPaginated, err := queryServer.Withdrawals(ctx, &millionstypes.QueryWithdrawalsRequest{Pagination: &query.PageRequest{Offset: 1, Limit: 3}})
	suite.Require().NoError(err)
	suite.Require().Len(withdrawalResPaginated.GetWithdrawals(), 3)
	suite.Require().Equal(withdrawalRes.GetWithdrawals()[1].WithdrawalId, withdrawalResPaginated.GetWithdrawals()[0].WithdrawalId)
	suite.Require().Equal(withdrawalRes.GetWithdrawals()[2].WithdrawalId, withdrawalResPaginated.GetWithdrawals()[1].WithdrawalId)
	suite.Require().Equal(withdrawalRes.GetWithdrawals()[3].WithdrawalId, withdrawalResPaginated.GetWithdrawals()[2].WithdrawalId)

	// Test PoolWithdrawals with wrong poolID
	_, err = queryServer.PoolWithdrawals(ctx, &millionstypes.QueryPoolWithdrawalsRequest{PoolId: 0})
	suite.Require().Error(err)
	// Test PoolWithdrawals
	poolWithdrawals, err := queryServer.PoolWithdrawals(ctx, &millionstypes.QueryPoolWithdrawalsRequest{PoolId: 1})
	suite.Require().NoError(err)
	suite.Require().Len(poolWithdrawals.GetWithdrawals(), 6)
	// Test PoolWithdrawals with pagination
	poolWithdrawalsPaginated, err := queryServer.PoolWithdrawals(ctx, &millionstypes.QueryPoolWithdrawalsRequest{PoolId: 1, Pagination: &query.PageRequest{Offset: 1, Limit: 3}})
	suite.Require().NoError(err)
	suite.Require().Len(poolWithdrawalsPaginated.GetWithdrawals(), 3)
	suite.Require().Equal(poolWithdrawals.GetWithdrawals()[1].WithdrawalId, poolWithdrawalsPaginated.GetWithdrawals()[0].WithdrawalId)
	suite.Require().Equal(poolWithdrawals.GetWithdrawals()[2].WithdrawalId, poolWithdrawalsPaginated.GetWithdrawals()[1].WithdrawalId)
	suite.Require().Equal(poolWithdrawals.GetWithdrawals()[3].WithdrawalId, poolWithdrawalsPaginated.GetWithdrawals()[2].WithdrawalId)

	// Test PoolWithdrawal with wrong poolID
	_, err = queryServer.PoolWithdrawal(ctx, &millionstypes.QueryPoolWithdrawalRequest{PoolId: 0})
	suite.Require().Error(err)
	// Test PoolWithdrawal
	poolWithdrawal, err := queryServer.PoolWithdrawal(ctx, &millionstypes.QueryPoolWithdrawalRequest{PoolId: 1, WithdrawalId: 1})
	suite.Require().NoError(err)
	suite.Require().Equal(poolWithdrawal.Withdrawal.State, millionstypes.WithdrawalState_IbcTransfer)

	// Test AccountWithdrawals with wrong depositorAddr
	_, err = queryServer.AccountWithdrawals(ctx, &millionstypes.QueryAccountWithdrawalsRequest{DepositorAddress: ""})
	suite.Require().Error(err)
	// Test AccountWithdrawals
	accountWithdrawals, err := queryServer.AccountWithdrawals(ctx, &millionstypes.QueryAccountWithdrawalsRequest{DepositorAddress: suite.addrs[0].String()})
	suite.Require().NoError(err)
	suite.Require().Len(accountWithdrawals.GetWithdrawals(), 1)
	// Test AccountWithdrawals with pagination
	accountWithdrawalsPaginated, err := queryServer.AccountWithdrawals(ctx, &millionstypes.QueryAccountWithdrawalsRequest{DepositorAddress: suite.addrs[0].String(), Pagination: &query.PageRequest{Limit: 1}})
	suite.Require().NoError(err)
	suite.Require().Len(accountWithdrawalsPaginated.GetWithdrawals(), 1)
	suite.Require().Equal(accountWithdrawals.GetWithdrawals()[0].WithdrawalId, accountWithdrawalsPaginated.GetWithdrawals()[0].WithdrawalId)

	// Test AccountPoolWithdrawals with wrong depositorAddr
	_, err = queryServer.AccountPoolWithdrawals(ctx, &millionstypes.QueryAccountPoolWithdrawalsRequest{DepositorAddress: "", PoolId: 1})
	suite.Require().Error(err)
	// Test AccountPoolWithdrawals with wrong poolID
	_, err = queryServer.AccountPoolWithdrawals(ctx, &millionstypes.QueryAccountPoolWithdrawalsRequest{DepositorAddress: suite.addrs[0].String(), PoolId: 0})
	suite.Require().Error(err)
	// Test AccountPoolWithdrawals
	accountPoolWithdrawals, err := queryServer.AccountPoolWithdrawals(ctx, &millionstypes.QueryAccountPoolWithdrawalsRequest{DepositorAddress: suite.addrs[0].String(), PoolId: 1})
	suite.Require().NoError(err)
	suite.Require().Len(accountPoolWithdrawals.GetWithdrawals(), 1)
	// Test AccountPoolWithdrawals with pagination
	accountPoolWithdrawalsPaginated, err := queryServer.AccountPoolWithdrawals(ctx, &millionstypes.QueryAccountPoolWithdrawalsRequest{DepositorAddress: suite.addrs[0].String(), PoolId: 1, Pagination: &query.PageRequest{Limit: 1}})
	suite.Require().NoError(err)
	suite.Require().Len(accountPoolWithdrawalsPaginated.GetWithdrawals(), 1)
	suite.Require().Equal(accountPoolWithdrawals.GetWithdrawals()[0].WithdrawalId, accountPoolWithdrawalsPaginated.GetWithdrawals()[0].WithdrawalId)
}
