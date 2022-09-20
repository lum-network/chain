package dfract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/dfract/keeper"
	"github.com/lum-network/chain/x/dfract/types"
)

// InitGenesis initializes the dfract module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.CreateModuleAccount(ctx, genState.ModuleAccountBalance)
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// Process the minted deposits
	for _, deposit := range genState.MintedDeposits {
		k.InsertIntoMintedDeposits(ctx, deposit.GetDepositorAddress(), *deposit)
	}

	// Process the waiting mint deposits
	for _, deposit := range genState.WaitingMintDeposits {
		k.InsertIntoWaitingMintDeposits(ctx, deposit.GetDepositorAddress(), *deposit)
	}

	// Process the waiting proposal deposits
	for _, deposit := range genState.WaitingProposalDeposits {
		k.InsertIntoWaitingProposalDeposits(ctx, deposit.GetDepositorAddress(), *deposit)
	}
}

// ExportGenesis returns the dfract module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		ModuleAccountBalance:    k.GetModuleAccountBalance(ctx),
		Params:                  params,
		MintedDeposits:          k.ListMintedDeposits(ctx),
		WaitingMintDeposits:     k.ListWaitingMintDeposits(ctx),
		WaitingProposalDeposits: k.ListWaitingProposalDeposits(ctx),
	}
}
