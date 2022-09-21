package dfract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/lum-network/chain/x/dfract/keeper"
	"github.com/lum-network/chain/x/dfract/types"
)

// InitGenesis initializes the dfract module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) error {
	k.CreateModuleAccount(ctx, genState.ModuleAccountBalance)
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// Process the minted deposits
	for _, deposit := range genState.DepositsMinted {
		depositorAddress, err := sdk.AccAddressFromBech32(deposit.GetDepositorAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}
		k.SetDepositMinted(ctx, depositorAddress, *deposit)
	}

	// Process the waiting mint deposits
	for _, deposit := range genState.DepositsPendingMint {
		depositorAddress, err := sdk.AccAddressFromBech32(deposit.GetDepositorAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}
		k.SetDepositPendingMint(ctx, depositorAddress, *deposit)
	}

	// Process the waiting proposal deposits
	for _, deposit := range genState.DepositsPendingWithdrawal {
		depositorAddress, err := sdk.AccAddressFromBech32(deposit.GetDepositorAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}
		k.SetDepositPendingWithdrawal(ctx, depositorAddress, *deposit)
	}

	return nil
}

// ExportGenesis returns the dfract module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		ModuleAccountBalance:      k.GetModuleAccountBalance(ctx),
		Params:                    params,
		DepositsMinted:            k.ListDepositsMinted(ctx),
		DepositsPendingMint:       k.ListDepositsPendingMint(ctx),
		DepositsPendingWithdrawal: k.ListDepositsPendingWithdrawal(ctx),
	}
}
