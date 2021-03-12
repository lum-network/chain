package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sandblockio/chain/x/faucet/types"
	"time"
)

func (k Keeper) HasMining(ctx sdk.Context, minter sdk.AccAddress) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FaucetKey))
	return store.Has(types.KeyPrefix(types.FaucetKey + minter.String()))
}

func (k Keeper) GetMining(ctx sdk.Context, minter sdk.AccAddress) types.Mining {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FaucetKey))

	// If this user never requested allocation, construct from the staking module parameters
	if !k.HasMining(ctx, minter) {
		denom := k.StakingKeeper.BondDenom(ctx)
		return types.Mining{
			Minter:   minter.String(),
			Total:    sdk.NewCoin(denom, sdk.NewInt(0)),
			LastTime: time.Time{},
		}
	}

	//
	var mining types.Mining
	k.cdc.MustUnmarshalBinaryBare(store.Get(types.KeyPrefix(types.FaucetKey+minter.String())), &mining)
	return mining
}

func (k Keeper) SetMining(ctx sdk.Context, miner sdk.AccAddress, mining types.Mining) {
	if len(mining.Minter) <= 0 {
		return
	}

	if !mining.Total.IsPositive() {
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FaucetKey))
	store.Set(types.KeyPrefix(types.FaucetKey+miner.String()), k.cdc.MustMarshalBinaryBare(&mining))
}

func (k Keeper) MintAndSend(ctx sdk.Context, msg types.MsgMintAndSend) error {
	accAddress := sdk.AccAddress(msg.Minter)
	// Acquire mining instance
	mining := k.GetMining(ctx, accAddress)

	// Refuse mint if it's too early
	if k.HasMining(ctx, accAddress) && mining.LastTime.Add(k.defaultLimit).UTC().After(msg.MintTime) {
		return types.ErrMintTooOften
	}

	denom := k.StakingKeeper.BondDenom(ctx)
	newCoin := sdk.NewCoin(denom, sdk.NewInt(k.defaultAmount))

	// Mint the coins from the supply
	err := k.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(newCoin))
	if err != nil {
		return err
	}

	// Send the minted coins
	err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accAddress, sdk.NewCoins(newCoin))
	if err != nil {
		return err
	}

	// Update the local kv store mining once everything was minted
	mining.Total = mining.Total.Add(newCoin)
	mining.LastTime = msg.MintTime
	k.SetMining(ctx, accAddress, mining)

	return nil
}
