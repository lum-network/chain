package millions

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/millions/keeper"
	"github.com/lum-network/chain/x/millions/types"
)

func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetNextPoolID(ctx, genState.NextPoolId)
	k.SetNextDepositID(ctx, genState.NextDepositId)
	k.SetNextPrizeID(ctx, genState.NextPrizeId)
	k.SetNextWithdrawalID(ctx, genState.NextWithdrawalId)

	for _, pool := range genState.Pools {
		// Voluntary reset TvlAmount, DepositorsCount and SponsorshipAmount since they will be recomputed from deposits
		pool.TvlAmount = sdk.ZeroInt()
		pool.DepositorsCount = 0
		pool.SponsorshipAmount = sdk.ZeroInt()
		k.AddPool(ctx, &pool)
	}

	// Import deposits
	for _, deposit := range genState.Deposits {
		k.AddDeposit(ctx, &deposit)
	}

	// Import draws
	for _, draw := range genState.Draws {
		k.SetPoolDraw(ctx, draw)
	}

	// Import prizes
	for _, prize := range genState.Prizes {
		k.AddPrize(ctx, prize)
	}

	// Import withdrawals
	for _, withdrawal := range genState.Withdrawals {
		k.AddWithdrawal(ctx, withdrawal)
	}
}

func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)
	return &types.GenesisState{
		Params:           params,
		NextPoolId:       k.GetNextPoolID(ctx),
		NextDepositId:    k.GetNextDepositID(ctx),
		NextPrizeId:      k.GetNextPrizeID(ctx),
		NextWithdrawalId: k.GetNextWithdrawalID(ctx),
		Pools:            k.ListPools(ctx),
		Deposits:         k.ListDeposits(ctx),
		Draws:            k.ListDraws(ctx),
		Prizes:           k.ListPrizes(ctx),
		Withdrawals:      k.ListWithdrawals(ctx),
	}
}
