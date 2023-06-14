package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/dfract/types"
)

func (k Keeper) SetDepositPendingWithdrawal(ctx sdk.Context, depositorAddress sdk.AccAddress, deposit types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	encodedDeposit := k.cdc.MustMarshal(&deposit)
	store.Set(types.GetDepositsPendingWithdrawalKey(depositorAddress), encodedDeposit)
}

func (k Keeper) GetDepositPendingWithdrawal(ctx sdk.Context, depositorAddress sdk.AccAddress) (deposit types.Deposit, found bool) {
	store := ctx.KVStore(k.storeKey)
	if store.Has(types.GetDepositsPendingWithdrawalKey(depositorAddress)) {
		err := k.cdc.Unmarshal(store.Get(types.GetDepositsPendingWithdrawalKey(depositorAddress)), &deposit)
		if err != nil {
			return deposit, false
		}
		return deposit, true
	}
	return deposit, false
}

func (k Keeper) RemoveDepositPendingWithdrawal(ctx sdk.Context, depositorAddress sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDepositsPendingWithdrawalKey(depositorAddress))
}

func (k Keeper) SetDepositPendingMint(ctx sdk.Context, depositorAddress sdk.AccAddress, deposit types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	encodedDeposit := k.cdc.MustMarshal(&deposit)
	store.Set(types.GetDepositsPendingMintKey(depositorAddress), encodedDeposit)
}

func (k Keeper) GetDepositPendingMint(ctx sdk.Context, depositorAddress sdk.AccAddress) (deposit types.Deposit, found bool) {
	store := ctx.KVStore(k.storeKey)
	if store.Has(types.GetDepositsPendingMintKey(depositorAddress)) {
		err := k.cdc.Unmarshal(store.Get(types.GetDepositsPendingMintKey(depositorAddress)), &deposit)
		if err != nil {
			return deposit, false
		}
		return deposit, true
	}
	return deposit, false
}

func (k Keeper) RemoveDepositPendingMint(ctx sdk.Context, depositorAddress sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDepositsPendingMintKey(depositorAddress))
}

func (k Keeper) SetDepositMinted(ctx sdk.Context, depositorAddress sdk.AccAddress, deposit types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	encodedDeposit := k.cdc.MustMarshal(&deposit)
	store.Set(types.GetDepositsMintedKey(depositorAddress), encodedDeposit)
}

// AddDepositMinted same as SetDepositMinted but takes into account the existing value.
func (k Keeper) AddDepositMinted(ctx sdk.Context, depositorAddress sdk.AccAddress, deposit types.Deposit) {
	previousDeposit, found := k.GetDepositMinted(ctx, depositorAddress)
	if !found {
		k.SetDepositMinted(ctx, depositorAddress, deposit)
	} else {
		deposit.Amount = deposit.Amount.Add(previousDeposit.Amount)
		k.SetDepositMinted(ctx, depositorAddress, deposit)
	}
}

func (k Keeper) GetDepositMinted(ctx sdk.Context, depositorAddress sdk.AccAddress) (deposit types.Deposit, found bool) {
	store := ctx.KVStore(k.storeKey)
	if store.Has(types.GetDepositsMintedKey(depositorAddress)) {
		err := k.cdc.Unmarshal(store.Get(types.GetDepositsMintedKey(depositorAddress)), &deposit)
		if err != nil {
			return deposit, false
		}
		return deposit, true
	}
	return deposit, false
}

func (k Keeper) RemoveDepositMinted(ctx sdk.Context, depositorAddress sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDepositsMintedKey(depositorAddress))
}

func (k Keeper) IterateDepositsPendingWithdrawal(ctx sdk.Context, cb func(deposit types.Deposit) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.DepositsPendingWithdrawalPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

func (k Keeper) ListDepositsPendingWithdrawal(ctx sdk.Context) (deposits []*types.Deposit) {
	k.IterateDepositsPendingWithdrawal(ctx, func(deposit types.Deposit) bool {
		deposits = append(deposits, &deposit)
		return false
	})
	return deposits
}

func (k Keeper) IterateDepositsPendingMint(ctx sdk.Context, cb func(deposit types.Deposit) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.DepositsPendingMintPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

func (k Keeper) ListDepositsPendingMint(ctx sdk.Context) (deposits []*types.Deposit) {
	k.IterateDepositsPendingMint(ctx, func(deposit types.Deposit) bool {
		deposits = append(deposits, &deposit)
		return false
	})
	return deposits
}

func (k Keeper) IterateDepositsMinted(ctx sdk.Context, cb func(deposit types.Deposit) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.DepositsMintedPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

func (k Keeper) ListDepositsMinted(ctx sdk.Context) (deposits []*types.Deposit) {
	k.IterateDepositsMinted(ctx, func(deposit types.Deposit) bool {
		deposits = append(deposits, &deposit)
		return false
	})
	return deposits
}
