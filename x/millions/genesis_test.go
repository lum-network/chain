package millions_test

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	apptypes "github.com/lum-network/chain/app"
	"github.com/lum-network/chain/x/millions"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

var (
	now      = time.Now().UTC()
	future   = now.Add(10 * time.Second)
	testAccs = []string{
		"lum19vp4wqsdyw8gftvkat0vrfqzvdrlp9wxp6fevr",
		"lum1s3zyjkd79tf83jew2f4cu4zhc9e5y53faqv3te",
		"lum1hag45rj53nfpj6357ql7nzq3p678s7lx3gpmyu",
		"lum1xutdwytajz6zn338wq7n4wx6m95076nepyu5zl",
		"lum10dtsxdl0uj0pw58yvlmsuy8p5mvzvewga5dgl8",
		"lum17p728nwzmkg5038pepfzeavxfu063aej0shwc8",
	}
	defaultValidators = []millionstypes.PoolValidator{{
		OperatorAddress: "lumvaloper1qx2dts3tglxcu0jh47k7ghstsn4nactufgmmlk",
		BondedAmount:    sdk.ZeroInt(),
	}}
	defaultSchedule   = millionstypes.DrawSchedule{DrawDelta: 404 * time.Hour, InitialDrawAt: now}
	defaultPrizeStrat = millionstypes.PrizeStrategy{PrizeBatches: []millionstypes.PrizeBatch{{PoolPercent: 100, Quantity: 1, DrawProbability: sdk.NewDec(1)}}}
)

var testGenesis = millionstypes.GenesisState{
	Params: millionstypes.Params{
		MinDepositAmount:        sdk.NewInt(1_000_000),
		MaxPrizeStrategyBatches: 10,
		MaxPrizeBatchQuantity:   1000,
		MinDrawScheduleDelta:    1 * time.Hour,
		MaxDrawScheduleDelta:    366 * 24 * time.Hour, // 366 days
		PrizeExpirationDelta:    30 * 24 * time.Hour,  // 30 days
		FeesStakers:             sdk.NewDec(0),
		MinDepositDrawDelta:     5 * time.Minute,
	},
	NextPoolId:       6,
	NextDepositId:    6,
	NextPrizeId:      6,
	NextWithdrawalId: 6,
	Pools: []millionstypes.Pool{
		{PoolId: 1, PoolType: millionstypes.PoolType_Staking, TvlAmount: sdk.NewInt(510), DepositorsCount: 3, SponsorshipAmount: sdk.ZeroInt(), Denom: "denom-1", NativeDenom: "denom-1", NextDrawId: 3,
			ChainId: "c1", Validators: defaultValidators, MinDepositAmount: sdk.NewInt(1_000_000), AvailablePrizePool: sdk.NewCoin("denom-1", sdk.ZeroInt()),
			DrawSchedule: defaultSchedule, PrizeStrategy: defaultPrizeStrat,
			State: millionstypes.PoolState_Created, Bech32PrefixAccAddr: "lum", Bech32PrefixValAddr: "lumvaloper",
		},
		{PoolId: 2, PoolType: millionstypes.PoolType_Staking, TvlAmount: sdk.NewInt(603), DepositorsCount: 2, SponsorshipAmount: sdk.NewInt(401), Denom: "denom-2", NativeDenom: "denom-2", NextDrawId: 2,
			ChainId: "c1", Validators: defaultValidators, MinDepositAmount: sdk.NewInt(1_000_000), AvailablePrizePool: sdk.NewCoin("denom-2", sdk.ZeroInt()),
			DrawSchedule: defaultSchedule, PrizeStrategy: defaultPrizeStrat,
			State: millionstypes.PoolState_Ready, Bech32PrefixAccAddr: "lum", Bech32PrefixValAddr: "lumvaloper",
		},
		{PoolId: 3, PoolType: millionstypes.PoolType_Staking, TvlAmount: sdk.NewInt(601), DepositorsCount: 1, SponsorshipAmount: sdk.ZeroInt(), Denom: "denom-3", NativeDenom: "denom-3", NextDrawId: 1,
			ChainId: "c1", Validators: defaultValidators, MinDepositAmount: sdk.NewInt(1_000_000), AvailablePrizePool: sdk.NewCoin("denom-3", sdk.ZeroInt()),
			DrawSchedule: defaultSchedule, PrizeStrategy: defaultPrizeStrat,
			State: millionstypes.PoolState_Killed, Bech32PrefixAccAddr: "lum", Bech32PrefixValAddr: "lumvaloper",
		},
		{PoolId: 4, PoolType: millionstypes.PoolType_Staking, TvlAmount: sdk.NewInt(400), DepositorsCount: 1, SponsorshipAmount: sdk.ZeroInt(), Denom: "denom-4", NativeDenom: "denom-4", NextDrawId: 1,
			ChainId: "c1", Validators: defaultValidators, MinDepositAmount: sdk.NewInt(1_000_000), AvailablePrizePool: sdk.NewCoin("denom-4", sdk.ZeroInt()),
			DrawSchedule: defaultSchedule, PrizeStrategy: defaultPrizeStrat,
			State: millionstypes.PoolState_Created, Bech32PrefixAccAddr: "lum", Bech32PrefixValAddr: "lumvaloper",
		},
		{PoolId: 5, PoolType: millionstypes.PoolType_Staking, TvlAmount: sdk.NewInt(0), DepositorsCount: 0, SponsorshipAmount: sdk.ZeroInt(), Denom: "denom-5", NativeDenom: "denom-5", NextDrawId: 1,
			ChainId: "c1", Validators: defaultValidators, MinDepositAmount: sdk.NewInt(1_000_000), AvailablePrizePool: sdk.NewCoin("denom-5", sdk.ZeroInt()),
			DrawSchedule: defaultSchedule, PrizeStrategy: defaultPrizeStrat,
			State: millionstypes.PoolState_Killed, Bech32PrefixAccAddr: "lum", Bech32PrefixValAddr: "lumvaloper",
		},
	},
	Deposits: []millionstypes.Deposit{
		{PoolId: 1, DepositId: 1, DepositorAddress: testAccs[0], WinnerAddress: testAccs[0], Amount: sdk.NewCoin("denom-1", sdk.NewInt(100)), State: millionstypes.DepositState_IbcTransfer},
		{PoolId: 1, DepositId: 2, DepositorAddress: testAccs[0], WinnerAddress: testAccs[0], Amount: sdk.NewCoin("denom-1", sdk.NewInt(101)), State: millionstypes.DepositState_IcaDelegate},
		{PoolId: 1, DepositId: 3, DepositorAddress: testAccs[1], WinnerAddress: testAccs[1], Amount: sdk.NewCoin("denom-1", sdk.NewInt(102)), State: millionstypes.DepositState_Success},
		{PoolId: 1, DepositId: 4, DepositorAddress: testAccs[1], WinnerAddress: testAccs[1], Amount: sdk.NewCoin("denom-1", sdk.NewInt(103)), State: millionstypes.DepositState_Failure},
		{PoolId: 1, DepositId: 5, DepositorAddress: testAccs[2], WinnerAddress: testAccs[2], Amount: sdk.NewCoin("denom-1", sdk.NewInt(104)), State: millionstypes.DepositState_IbcTransfer},
		{PoolId: 2, DepositId: 6, DepositorAddress: testAccs[2], WinnerAddress: testAccs[2], Amount: sdk.NewCoin("denom-2", sdk.NewInt(200)), IsSponsor: true, State: millionstypes.DepositState_IbcTransfer},
		{PoolId: 2, DepositId: 7, DepositorAddress: testAccs[3], WinnerAddress: testAccs[3], Amount: sdk.NewCoin("denom-2", sdk.NewInt(201)), IsSponsor: true, State: millionstypes.DepositState_IbcTransfer},
		{PoolId: 2, DepositId: 8, DepositorAddress: testAccs[3], WinnerAddress: testAccs[3], Amount: sdk.NewCoin("denom-2", sdk.NewInt(202)), State: millionstypes.DepositState_IbcTransfer},
		{PoolId: 3, DepositId: 9, DepositorAddress: testAccs[4], WinnerAddress: testAccs[4], Amount: sdk.NewCoin("denom-3", sdk.NewInt(300)), State: millionstypes.DepositState_Success},
		{PoolId: 3, DepositId: 10, DepositorAddress: testAccs[4], WinnerAddress: testAccs[4], Amount: sdk.NewCoin("denom-3", sdk.NewInt(301)), State: millionstypes.DepositState_Success},
		{PoolId: 4, DepositId: 11, DepositorAddress: testAccs[5], WinnerAddress: testAccs[5], Amount: sdk.NewCoin("denom-4", sdk.NewInt(400)), State: millionstypes.DepositState_Success},
	},
	Draws: []millionstypes.Draw{
		{PoolId: 1, DrawId: 1, RandSeed: 10, TotalWinCount: 100, TotalWinAmount: sdk.NewInt(1000)},
		{PoolId: 1, DrawId: 2, RandSeed: 20, TotalWinCount: 200, TotalWinAmount: sdk.NewInt(2000)},
		{PoolId: 2, DrawId: 1, RandSeed: 30, TotalWinCount: 300, TotalWinAmount: sdk.NewInt(3000)},
	},
	Prizes: []millionstypes.Prize{
		{PoolId: 1, DrawId: 1, PrizeId: 1, State: millionstypes.PrizeState_Pending, WinnerAddress: testAccs[0], Amount: sdk.NewCoin("denom-1", sdk.NewInt(1)), ExpiresAt: future},
		{PoolId: 1, DrawId: 2, PrizeId: 2, State: millionstypes.PrizeState_Pending, WinnerAddress: testAccs[1], Amount: sdk.NewCoin("denom-1", sdk.NewInt(1)), ExpiresAt: now},
		{PoolId: 2, DrawId: 1, PrizeId: 3, State: millionstypes.PrizeState_Pending, WinnerAddress: testAccs[2], Amount: sdk.NewCoin("denom-2", sdk.NewInt(1)), ExpiresAt: now},
	},
	Withdrawals: []millionstypes.Withdrawal{
		{PoolId: 1, DepositId: 3, WithdrawalId: 1, DepositorAddress: testAccs[1], ToAddress: testAccs[1], Amount: sdk.NewCoin("denom-1", sdk.NewInt(102)), State: millionstypes.WithdrawalState_IcaUnbonding, UnbondingEndsAt: &now},
		{PoolId: 3, DepositId: 9, WithdrawalId: 2, DepositorAddress: testAccs[4], ToAddress: testAccs[4], Amount: sdk.NewCoin("denom-3", sdk.NewInt(300)), State: millionstypes.WithdrawalState_IcaUnbonding, UnbondingEndsAt: &future},
		{PoolId: 3, DepositId: 10, WithdrawalId: 3, DepositorAddress: testAccs[4], ToAddress: testAccs[4], Amount: sdk.NewCoin("denom-3", sdk.NewInt(301)), State: millionstypes.WithdrawalState_IcaUnbonding, UnbondingEndsAt: &future},
		{PoolId: 4, DepositId: 11, WithdrawalId: 4, DepositorAddress: testAccs[5], ToAddress: testAccs[5], Amount: sdk.NewCoin("denom-4", sdk.NewInt(400)), State: millionstypes.WithdrawalState_IcaUnbonding, UnbondingEndsAt: &now},
	},
}

func TestInitGenesis(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	// Init the module genesis state
	millions.InitGenesis(ctx, *app.MillionsKeeper, testGenesis)

	// Make sure params match
	params := app.MillionsKeeper.GetParams(ctx)
	require.Equal(t, testGenesis.Params, params)
	require.Equal(t, testGenesis.Params.MinDepositAmount, params.MinDepositAmount)
	require.Equal(t, testGenesis.Params.MaxPrizeStrategyBatches, params.MaxPrizeStrategyBatches)
	require.Equal(t, testGenesis.Params.MaxPrizeBatchQuantity, params.MaxPrizeBatchQuantity)
	require.Equal(t, testGenesis.Params.MinDrawScheduleDelta, params.MinDrawScheduleDelta)
	require.Equal(t, testGenesis.Params.MaxDrawScheduleDelta, params.MaxDrawScheduleDelta)
	require.Equal(t, testGenesis.Params.PrizeExpirationDelta, params.PrizeExpirationDelta)
	require.Equal(t, testGenesis.Params.FeesStakers, params.FeesStakers)
	require.Equal(t, testGenesis.Params.MinDepositDrawDelta, params.MinDepositDrawDelta)

	// Make sure genesis next poolID matches
	nextPoolID := app.MillionsKeeper.GetNextPoolID(ctx)
	require.Equal(t, testGenesis.NextPoolId, nextPoolID)

	// Make sure genesis next depositID matches
	nextDepositID := app.MillionsKeeper.GetNextDepositID(ctx)
	require.Equal(t, testGenesis.NextDepositId, nextDepositID)

	// Make sure genesis next prizeID matches
	nextPrizeID := app.MillionsKeeper.GetNextPrizeID(ctx)
	require.Equal(t, testGenesis.NextPrizeId, nextPrizeID)

	// Make sure genesis next withdrawalID matches
	nextWithdrawalID := app.MillionsKeeper.GetNextWithdrawalID(ctx)
	require.Equal(t, testGenesis.NextWithdrawalId, nextWithdrawalID)

	// Pools should all have been imported
	pools := app.MillionsKeeper.ListPools(ctx)
	require.Len(t, pools, len(testGenesis.Pools))
	for i, p := range pools {
		require.Equal(t, testGenesis.Pools[i].PoolId, p.PoolId)
		require.Equal(t, testGenesis.Pools[i].PoolType, p.PoolType)
		require.Equal(t, testGenesis.Pools[i].TvlAmount.Int64(), p.TvlAmount.Int64())
		require.Equal(t, testGenesis.Pools[i].DepositorsCount, p.DepositorsCount)
		require.Equal(t, testGenesis.Pools[i].SponsorshipAmount.Int64(), p.SponsorshipAmount.Int64())
		require.Equal(t, testGenesis.Pools[i].Denom, p.Denom)
		require.Equal(t, testGenesis.Pools[i].NextDrawId, p.NextDrawId)
		// Pool deposits should all have been imported
		deposits := app.MillionsKeeper.ListPoolDeposits(ctx, p.PoolId)
		if p.PoolId == 1 {
			require.Len(t, deposits, 5)
		} else if p.PoolId == 2 {
			require.Len(t, deposits, 3)
		} else if p.PoolId == 3 {
			require.Len(t, deposits, 2)
		} else if p.PoolId == 4 {
			require.Len(t, deposits, 1)
		} else if p.PoolId == 5 {
			require.Len(t, deposits, 0)
		}
	}

	// Deposits should all have been imported
	deposits := app.MillionsKeeper.ListDeposits(ctx)
	require.Len(t, deposits, len(testGenesis.Deposits))
	sort.SliceStable(deposits, func(i, j int) bool {
		return deposits[i].DepositId < deposits[j].DepositId
	})
	for i, d := range deposits {
		require.Equal(t, testGenesis.Deposits[i].PoolId, d.PoolId)
		require.Equal(t, testGenesis.Deposits[i].DepositId, d.DepositId)
		require.Equal(t, testGenesis.Deposits[i].DepositorAddress, d.DepositorAddress)
		require.Equal(t, testGenesis.Deposits[i].WinnerAddress, d.WinnerAddress)
		require.Equal(t, testGenesis.Deposits[i].Amount, d.Amount)
		require.Equal(t, testGenesis.Deposits[i].IsSponsor, d.IsSponsor)
		require.Equal(t, testGenesis.Deposits[i].State, d.State)
	}

	// Account deposits should all have been imported
	for i := 0; i < len(testAccs); i++ {
		addr := sdk.MustAccAddressFromBech32(testAccs[i])
		deposits := app.MillionsKeeper.ListAccountDeposits(ctx, addr)
		if i == len(testAccs)-1 {
			require.Len(t, deposits, 1)
		} else {
			require.Len(t, deposits, 2)
		}
	}

	// Draws should have been imported
	draws := app.MillionsKeeper.ListDraws(ctx)
	require.Len(t, draws, len(testGenesis.Draws))
	for i, d := range draws {
		require.Equal(t, testGenesis.Draws[i].PoolId, d.PoolId)
		require.Equal(t, testGenesis.Draws[i].DrawId, d.DrawId)
		require.Equal(t, int64(10*(i+1)), d.RandSeed)
		require.Equal(t, uint64(100*(i+1)), d.TotalWinCount)
		require.Equal(t, sdk.NewInt(1000*int64(i+1)), d.TotalWinAmount)
	}

	// ListPoolDraws should have been imported
	for _, pd := range testGenesis.Draws {
		poolDraws := app.MillionsKeeper.ListPoolDraws(ctx, pd.PoolId)
		if pd.PoolId == 1 {
			require.Len(t, poolDraws, 2)
		} else {
			require.Len(t, poolDraws, 1)
		}
	}

	// Prizes should have been imported
	prizes := app.MillionsKeeper.ListPrizes(ctx)

	require.Len(t, prizes, len(testGenesis.Prizes))
	for i, pz := range prizes {
		require.Equal(t, testGenesis.Prizes[i].PoolId, pz.PoolId)
		require.Equal(t, testGenesis.Prizes[i].DrawId, pz.DrawId)
		require.Equal(t, testGenesis.Prizes[i].PrizeId, pz.PrizeId)
		require.Equal(t, testGenesis.Prizes[i].State, pz.State)
		require.Equal(t, testGenesis.Prizes[i].WinnerAddress, pz.WinnerAddress)
		require.Equal(t, testGenesis.Prizes[i].Amount, pz.Amount)
	}

	// ListAccountPrizes should have been imported
	for _, ap := range testGenesis.Prizes {
		addr := sdk.MustAccAddressFromBech32(ap.WinnerAddress)
		accPrizes := app.MillionsKeeper.ListAccountPrizes(ctx, addr)
		require.Len(t, accPrizes, 1)
	}

	// Test Expired Prizes Queue Import
	for i := 0; i < len(testGenesis.Withdrawals); i++ {
		col := app.MillionsKeeper.GetPrizeIDsEPCBQueue(ctx, now)
		require.Len(t, col.PrizesIds, 2)
		for i, pid := range col.PrizesIds {
			require.Equal(t, col.PrizesIds[i].PoolId, pid.PoolId)
			require.Equal(t, col.PrizesIds[i].PrizeId, pid.PrizeId)
		}
	}

	// Withdrawals
	withdrawals := app.MillionsKeeper.ListWithdrawals(ctx)
	require.Len(t, withdrawals, len(testGenesis.Withdrawals))
	for i, w := range withdrawals {
		require.Equal(t, testGenesis.Withdrawals[i].PoolId, w.PoolId)
		require.Equal(t, testGenesis.Withdrawals[i].DepositId, w.DepositId)
		require.Equal(t, testGenesis.Withdrawals[i].WithdrawalId, w.WithdrawalId)
		require.Equal(t, testGenesis.Withdrawals[i].WithdrawalId, w.WithdrawalId)
		require.Equal(t, testGenesis.Withdrawals[i].DepositorAddress, w.DepositorAddress)
		require.Equal(t, testGenesis.Withdrawals[i].ToAddress, w.ToAddress)
		require.Equal(t, testGenesis.Withdrawals[i].Amount, w.Amount)
		require.Equal(t, testGenesis.Withdrawals[i].State, w.State)
	}

	// Account withdrawals should have been imported
	for _, aw := range testGenesis.Withdrawals {
		addr := sdk.MustAccAddressFromBech32(aw.DepositorAddress)
		accWithdrawals := app.MillionsKeeper.ListAccountWithdrawals(ctx, addr)
		if aw.PoolId == 1 {
			require.Len(t, accWithdrawals, 1)
		} else if aw.PoolId == 3 {
			require.Len(t, accWithdrawals, 2)
		} else {
			require.Len(t, accWithdrawals, 1)
		}
	}

	// Test Matured Withdrawal Queue Import
	for i := 0; i < len(testGenesis.Withdrawals); i++ {
		col := app.MillionsKeeper.GetWithdrawalIDsMaturedQueue(ctx, now)
		require.Len(t, col.WithdrawalsIds, 2)
		for i, wid := range col.WithdrawalsIds {
			require.Equal(t, col.WithdrawalsIds[i].PoolId, wid.PoolId)
			require.Equal(t, col.WithdrawalsIds[i].WithdrawalId, wid.WithdrawalId)
		}
	}
}

func TestExportGenesis(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	millions.InitGenesis(ctx, *app.MillionsKeeper, testGenesis)
	exportGenesis := millions.ExportGenesis(ctx, *app.MillionsKeeper)

	// Test params export
	require.Equal(t, exportGenesis.Params, testGenesis.Params)
	require.Equal(t, exportGenesis.Params.MinDepositAmount, testGenesis.Params.MinDepositAmount)
	require.Equal(t, exportGenesis.Params.MaxPrizeStrategyBatches, testGenesis.Params.MaxPrizeStrategyBatches)
	require.Equal(t, exportGenesis.Params.MaxPrizeBatchQuantity, testGenesis.Params.MaxPrizeBatchQuantity)
	require.Equal(t, exportGenesis.Params.MinDrawScheduleDelta, testGenesis.Params.MinDrawScheduleDelta)
	require.Equal(t, exportGenesis.Params.MaxDrawScheduleDelta, testGenesis.Params.MaxDrawScheduleDelta)
	require.Equal(t, exportGenesis.Params.PrizeExpirationDelta, testGenesis.Params.PrizeExpirationDelta)
	require.Equal(t, exportGenesis.Params.FeesStakers, testGenesis.Params.FeesStakers)
	require.Equal(t, exportGenesis.Params.MinDepositDrawDelta, testGenesis.Params.MinDepositDrawDelta)

	// Test IDs export
	require.Equal(t, exportGenesis.NextPoolId, testGenesis.NextPoolId)
	require.Equal(t, exportGenesis.NextDepositId, testGenesis.NextDepositId)
	require.Equal(t, exportGenesis.NextPrizeId, testGenesis.NextPrizeId)
	require.Equal(t, exportGenesis.NextWithdrawalId, testGenesis.NextWithdrawalId)

	// Test Pools export
	require.Len(t, exportGenesis.Pools, len(testGenesis.Pools))
	for i, p := range testGenesis.Pools {
		require.Equal(t, exportGenesis.Pools[i].PoolId, p.PoolId)
		require.Equal(t, exportGenesis.Pools[i].PoolType, p.PoolType)
		require.Equal(t, exportGenesis.Pools[i].TvlAmount.Int64(), p.TvlAmount.Int64())
		require.Equal(t, exportGenesis.Pools[i].DepositorsCount, p.DepositorsCount)
		require.Equal(t, exportGenesis.Pools[i].SponsorshipAmount.Int64(), p.SponsorshipAmount.Int64())
		require.Equal(t, exportGenesis.Pools[i].Denom, p.Denom)
		require.Equal(t, exportGenesis.Pools[i].NextDrawId, p.NextDrawId)
		// Pool deposits should all have been exported
		deposits := app.MillionsKeeper.ListPoolDeposits(ctx, p.PoolId)
		if p.PoolId == 1 {
			require.Len(t, deposits, 5)
		} else if p.PoolId == 2 {
			require.Len(t, deposits, 3)
		} else if p.PoolId == 3 {
			require.Len(t, deposits, 2)
		} else if p.PoolId == 4 {
			require.Len(t, deposits, 1)
		} else if p.PoolId == 5 {
			require.Len(t, deposits, 0)
		}
	}

	// Test deposits export
	require.Len(t, exportGenesis.Deposits, len(testGenesis.Deposits))
	sort.SliceStable(exportGenesis.Deposits, func(i, j int) bool {
		return exportGenesis.Deposits[i].DepositId < exportGenesis.Deposits[j].DepositId
	})
	for i, d := range testGenesis.Deposits {
		require.Equal(t, exportGenesis.Deposits[i].PoolId, d.PoolId)
		require.Equal(t, exportGenesis.Deposits[i].DepositId, d.DepositId)
		require.Equal(t, exportGenesis.Deposits[i].DepositorAddress, d.DepositorAddress)
		require.Equal(t, exportGenesis.Deposits[i].WinnerAddress, d.WinnerAddress)
		require.Equal(t, exportGenesis.Deposits[i].Amount, d.Amount)
		require.Equal(t, exportGenesis.Deposits[i].IsSponsor, d.IsSponsor)
		require.Equal(t, exportGenesis.Deposits[i].State, d.State)
	}

	// Test draws export
	require.Len(t, exportGenesis.Draws, len(testGenesis.Draws))
	for i, dw := range testGenesis.Draws {
		require.Equal(t, exportGenesis.Draws[i].PoolId, dw.PoolId)
		require.Equal(t, exportGenesis.Draws[i].DrawId, dw.DrawId)
		require.Equal(t, int64(10*(i+1)), dw.RandSeed)
		require.Equal(t, uint64(100*(i+1)), dw.TotalWinCount)
		require.Equal(t, sdk.NewInt(1000*int64(i+1)), dw.TotalWinAmount)
	}

	// Test prizes export
	require.Len(t, exportGenesis.Prizes, len(testGenesis.Prizes))
	for i, pz := range testGenesis.Prizes {
		require.Equal(t, exportGenesis.Prizes[i].PoolId, pz.PoolId)
		require.Equal(t, exportGenesis.Prizes[i].DrawId, pz.DrawId)
		require.Equal(t, exportGenesis.Prizes[i].PrizeId, pz.PrizeId)
		require.Equal(t, exportGenesis.Prizes[i].State, pz.State)
		require.Equal(t, exportGenesis.Prizes[i].WinnerAddress, pz.WinnerAddress)
		require.Equal(t, exportGenesis.Prizes[i].Amount, pz.Amount)
	}

	// Test withdrawals export
	require.Len(t, exportGenesis.Withdrawals, len(testGenesis.Withdrawals))
	for i, w := range testGenesis.Withdrawals {
		require.Equal(t, exportGenesis.Withdrawals[i].PoolId, w.PoolId)
		require.Equal(t, exportGenesis.Withdrawals[i].DepositId, w.DepositId)
		require.Equal(t, exportGenesis.Withdrawals[i].WithdrawalId, w.WithdrawalId)
		require.Equal(t, exportGenesis.Withdrawals[i].WithdrawalId, w.WithdrawalId)
		require.Equal(t, exportGenesis.Withdrawals[i].DepositorAddress, w.DepositorAddress)
		require.Equal(t, exportGenesis.Withdrawals[i].ToAddress, w.ToAddress)
		require.Equal(t, exportGenesis.Withdrawals[i].Amount, w.Amount)
		require.Equal(t, exportGenesis.Withdrawals[i].State, w.State)
	}
}
