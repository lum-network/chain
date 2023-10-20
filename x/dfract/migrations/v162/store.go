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
	accounts := dk.AuthKeeper.GetAllAccounts(ctx)
	for _, account := range accounts {
		supply := dk.BankKeeper.GetBalance(ctx, account.GetAddress(), dfracttypes.MintDenom)
		if supply.IsPositive() {
			if err := dk.BankKeeper.SendCoinsFromAccountToModule(ctx, account.GetAddress(), dfracttypes.ModuleName, sdk.NewCoins(supply)); err != nil {
				panic(err)
			}
		}
	}

	// Get the exact balance of the module account for the mint denom
	moduleAccount := dk.AuthKeeper.GetModuleAccount(ctx, dfracttypes.ModuleName)
	supply := dk.BankKeeper.GetBalance(ctx, moduleAccount.GetAddress(), dfracttypes.MintDenom)

	// Burn the coins
	if err := dk.BankKeeper.BurnCoins(ctx, dfracttypes.ModuleName, sdk.NewCoins(supply)); err != nil {
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
