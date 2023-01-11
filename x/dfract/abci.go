package dfract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/dfract/keeper"
	"github.com/lum-network/chain/x/dfract/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Get all staked tokens
	stakedTokens := k.ListStakedTokens(ctx)

	// Iterate through the staked tokens
	for _, stakedToken := range stakedTokens {
		// Check if the token has reached the unbonding time
		if stakedToken.Status == types.StakingState_StateUnbonding && ctx.BlockTime().After(stakedToken.UnbondedAt) {

			delegatorAddr, _ := sdk.AccAddressFromBech32(stakedToken.DelegatorAddress)

			// Will later be updated with the adjusted inflation rewards

			err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegatorAddr, sdk.NewCoins(stakedToken.StakingToken))
			if err != nil {
				panic(err)
			}

			// Update the token status to Unbonded
			k.SetStakedToken(ctx, sdk.AccAddress(stakedToken.DelegatorAddress), *stakedToken)
		}
	}
}
