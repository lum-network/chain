package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/dfract/types"
)

func (k Keeper) InsertIntoWaitingProposalDeposits(ctx sdk.Context, depositorAddress string, deposit types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	encodedDeposit := k.cdc.MustMarshal(&deposit)
	store.Set(types.GetWaitingProposalDepositsKey(depositorAddress), encodedDeposit)
}

func (k Keeper) GetFromWaitingProposalDeposits(ctx sdk.Context, depositorAddress string) (deposit types.Deposit, found bool) {
	store := ctx.KVStore(k.storeKey)
	if store.Has(types.GetWaitingProposalDepositsKey(depositorAddress)) {
		err := k.cdc.Unmarshal(store.Get(types.GetWaitingProposalDepositsKey(depositorAddress)), &deposit)
		if err != nil {
			return deposit, false
		}
		return deposit, true
	}
	return deposit, false
}

func (k Keeper) RemoveFromWaitingProposalDeposits(ctx sdk.Context, depositorAddress string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetWaitingProposalDepositsKey(depositorAddress))
}

func (k Keeper) InsertIntoWaitingMintDeposits(ctx sdk.Context, depositorAddress string, deposit types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	encodedDeposit := k.cdc.MustMarshal(&deposit)
	store.Set(types.GetWaitingMintDepositsKey(depositorAddress), encodedDeposit)
}

func (k Keeper) GetFromWaitingMintDeposits(ctx sdk.Context, depositorAddress string) (deposit types.Deposit, found bool) {
	store := ctx.KVStore(k.storeKey)
	if store.Has(types.GetWaitingMintDepositsKey(depositorAddress)) {
		err := k.cdc.Unmarshal(store.Get(types.GetWaitingMintDepositsKey(depositorAddress)), &deposit)
		if err != nil {
			return deposit, false
		}
		return deposit, true
	}
	return deposit, false
}

func (k Keeper) RemoveFromWaitingMintDeposits(ctx sdk.Context, depositorAddress string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetWaitingMintDepositsKey(depositorAddress))
}

func (k Keeper) InsertIntoMintedDeposits(ctx sdk.Context, depositorAddress string, deposit types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	encodedDeposit := k.cdc.MustMarshal(&deposit)
	store.Set(types.GetMintedDepositsKey(depositorAddress), encodedDeposit)
}

func (k Keeper) GetFromMintedDeposits(ctx sdk.Context, depositorAddress string) (deposit types.Deposit, found bool) {
	store := ctx.KVStore(k.storeKey)
	if store.Has(types.GetMintedDepositsKey(depositorAddress)) {
		err := k.cdc.Unmarshal(store.Get(types.GetMintedDepositsKey(depositorAddress)), &deposit)
		if err != nil {
			return deposit, false
		}
		return deposit, true
	}
	return deposit, false
}

func (k Keeper) RemoveFromMintedDeposits(ctx sdk.Context, depositorAddress string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetMintedDepositsKey(depositorAddress))
}

func (k Keeper) IterateWaitingProposalDeposits(ctx sdk.Context, cb func(deposit types.Deposit) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.WaitingProposalDepositsPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

func (k Keeper) ListWaitingProposalDeposits(ctx sdk.Context) (deposits []*types.Deposit) {
	k.IterateWaitingProposalDeposits(ctx, func(deposit types.Deposit) bool {
		deposits = append(deposits, &deposit)
		return false
	})
	return deposits
}

func (k Keeper) IterateWaitingMintDeposits(ctx sdk.Context, cb func(deposit types.Deposit) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.WaitingMintDepositsPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

func (k Keeper) ListWaitingMintDeposits(ctx sdk.Context) (deposits []*types.Deposit) {
	k.IterateWaitingMintDeposits(ctx, func(deposit types.Deposit) bool {
		deposits = append(deposits, &deposit)
		return false
	})
	return deposits
}

func (k Keeper) IterateMintedDeposits(ctx sdk.Context, cb func(deposit types.Deposit) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.MintedDepositsPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

func (k Keeper) ListMintedDeposits(ctx sdk.Context) (deposits []*types.Deposit) {
	k.IterateMintedDeposits(ctx, func(deposit types.Deposit) bool {
		deposits = append(deposits, &deposit)
		return false
	})
	return deposits
}
