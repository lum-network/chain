package v162

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dfractkeeper "github.com/lum-network/chain/x/dfract/keeper"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
)

func Nuke(ctx sdk.Context, dk dfractkeeper.Keeper) error {
	// Destroy the queues
	dk.IterateDepositsPendingMint(ctx, func(deposit dfracttypes.Deposit) bool {
		dk.RemoveDepositPendingMint(ctx, sdk.MustAccAddressFromBech32(deposit.GetDepositorAddress()))
		return false
	})
	dk.IterateDepositsPendingWithdrawal(ctx, func(deposit dfracttypes.Deposit) bool {
		dk.RemoveDepositPendingWithdrawal(ctx, sdk.MustAccAddressFromBech32(deposit.GetDepositorAddress()))
		return false
	})

	// This one is a special case. To be able to burn coins, we need to move them to the module account first
	dk.IterateDepositsMinted(ctx, func(deposit dfracttypes.Deposit) bool {
		dk.RemoveDepositMinted(ctx, sdk.MustAccAddressFromBech32(deposit.GetDepositorAddress()))
		return false
	})

	// Burn the coins by looping over all the accounts balances
	if err := dk.UnsafeBurnAllBalance(ctx); err != nil {
		panic(err)
	}
	return nil
}

func Migrate(ctx sdk.Context, dk dfractkeeper.Keeper) error {
	if err := Nuke(ctx, dk); err != nil {
		return err
	}
	return nil
}
