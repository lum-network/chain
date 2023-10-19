package v162

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dfractkeeper "github.com/lum-network/chain/x/dfract/keeper"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
)

func Nuke(ctx sdk.Context, dk dfractkeeper.Keeper) error {
	dk.IterateDepositsPendingMint(ctx, func(deposit dfracttypes.Deposit) bool {
		dk.RemoveDepositPendingMint(ctx, sdk.MustAccAddressFromBech32(deposit.GetDepositorAddress()))
		return false
	})
	dk.IterateDepositsPendingWithdrawal(ctx, func(deposit dfracttypes.Deposit) bool {
		dk.RemoveDepositPendingWithdrawal(ctx, sdk.MustAccAddressFromBech32(deposit.GetDepositorAddress()))
		return false
	})
	dk.IterateDepositsMinted(ctx, func(deposit dfracttypes.Deposit) bool {
		// Get the supply for the given depositor
		supply := dk.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(deposit.GetDepositorAddress()), dfracttypes.MintDenom)

		// Move the supply to the module account
		if err := dk.BankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(deposit.GetDepositorAddress()), dfracttypes.ModuleName, sdk.NewCoins(supply)); err != nil {
			panic(err)
		}

		// Burn the coins
		if err := dk.BankKeeper.BurnCoins(ctx, dfracttypes.ModuleName, sdk.NewCoins(supply)); err != nil {
			panic(err)
		}

		dk.RemoveDepositMinted(ctx, sdk.MustAccAddressFromBech32(deposit.GetDepositorAddress()))
		return false
	})
	return nil
}

func Migrate(ctx sdk.Context, dk dfractkeeper.Keeper) error {
	if err := Nuke(ctx, dk); err != nil {
		return err
	}
	return nil
}
