package airdrop

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/airdrop/keeper"
	"github.com/lum-network/chain/x/airdrop/types"
)

// InitGenesis initializes the capability module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.CreateModuleAccount(ctx, genState.ModuleAccountBalance)
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
	if err := k.SetClaimRecords(ctx, genState.ClaimRecords); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	records := k.GetClaimRecords(ctx)

	return &types.GenesisState{
		ModuleAccountBalance: k.GetAirdropAccountBalance(ctx),
		Params:               params,
		ClaimRecords:         records,
	}
}
