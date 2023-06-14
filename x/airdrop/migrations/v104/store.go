package v104

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	airdroptypes "github.com/lum-network/chain/x/airdrop/types"
)

// MigrateModuleBalance performs in-place store migrations from v1.0.3 to v1.0.4.
// The migration includes:
// - Fix airdrop module account balance.
func MigrateModuleBalance(ctx sdk.Context, ak accountkeeper.AccountKeeper, bk bankkeeper.Keeper, claimRecords []airdroptypes.ClaimRecord) error {
	// Compute how much of each coins the airdrop module account should have
	moduleCoins := sdk.Coins{}
	for _, rec := range claimRecords {
		for _, c := range rec.InitialClaimableAmount {
			moduleCoins = moduleCoins.Add(c)
		}
	}

	// Create the module account if needed
	moduleAcc := ak.GetModuleAccount(ctx, airdroptypes.ModuleName)
	if moduleAcc == nil {
		moduleAcc = authtypes.NewEmptyModuleAccount(airdroptypes.ModuleName, authtypes.Minter)
		ak.SetModuleAccount(ctx, moduleAcc)
	}

	// Get current module account balance
	moduleAccAddr := ak.GetModuleAddress(airdroptypes.ModuleName)
	currentBalances := bk.GetAllBalances(ctx, moduleAccAddr)

	// For each coin compute the missing amount and mint it
	for _, c := range moduleCoins {
		currentBalance := sdk.NewCoin(c.Denom, sdk.NewInt(0))
		for _, b := range currentBalances {
			if b.Denom == c.Denom {
				currentBalance = currentBalance.Add(b)
			}
		}
		if currentBalance.Amount.LT(c.Amount) {
			if err := bk.MintCoins(ctx, airdroptypes.ModuleName, sdk.Coins{c.Sub(currentBalance)}); err != nil {
				return err
			}
		}
	}

	return nil
}
