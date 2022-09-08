package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/dfract/types"
)

func (k Keeper) InsertIntoWaitingProposalQueue(ctx sdk.Context, depositID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetWaitingProposalDepositsQueueKey(depositID), []byte(depositID))
}

func (k Keeper) RemoveFromWaitingProposalQueue(ctx sdk.Context, depositID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetWaitingProposalDepositsQueueKey(depositID))
}

func (k Keeper) InsertIntoWaitingMintQueue(ctx sdk.Context, depositID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetWaitingMintDepositsQueueKey(depositID), []byte(depositID))
}

func (k Keeper) RemoveFromWaitingMintQueue(ctx sdk.Context, depositID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetWaitingMintDepositsQueueKey(depositID))
}

func (k Keeper) InsertIntoMintedQueue(ctx sdk.Context, depositID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetMintedDepositsQueueKey(depositID), []byte(depositID))
}

func (k Keeper) RemoveFromMintedQueue(ctx sdk.Context, depositID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetMintedDepositsQueueKey(depositID))
}

func (k Keeper) IterateWaitingProposalDepositsQueue(ctx sdk.Context, cb func(depositID string) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.WaitingProposalDepositsQueuePrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		depositID := string(iterator.Value())
		if cb(depositID) {
			break
		}
	}
}

func (k Keeper) ListWaitingProposalDeposits(ctx sdk.Context) (deposits []*types.Deposit) {
	k.IterateWaitingProposalDepositsQueue(ctx, func(depositID string) bool {
		deposit, _ := k.GetDeposit(ctx, depositID)
		deposits = append(deposits, &deposit)
		return false
	})
	return deposits
}

func (k Keeper) IterateWaitingMintDepositsQueue(ctx sdk.Context, cb func(depositID string) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.WaitingMintDepositsQueuePrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		depositID := string(iterator.Value())
		if cb(depositID) {
			break
		}
	}
}

func (k Keeper) ListWaitingMintDeposits(ctx sdk.Context) (deposits []*types.Deposit) {
	k.IterateWaitingMintDepositsQueue(ctx, func(depositID string) bool {
		deposit, _ := k.GetDeposit(ctx, depositID)
		deposits = append(deposits, &deposit)
		return false
	})
	return deposits
}

func (k Keeper) IterateMintedDepositsQueue(ctx sdk.Context, cb func(depositID string) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.MintedDepositsQueuePrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		depositID := string(iterator.Value())
		if cb(depositID) {
			break
		}
	}
}

func (k Keeper) ListMintedDeposits(ctx sdk.Context) (deposits []*types.Deposit) {
	k.IterateMintedDepositsQueue(ctx, func(depositID string) bool {
		deposit, _ := k.GetDeposit(ctx, depositID)
		deposits = append(deposits, &deposit)
		return false
	})
	return deposits
}
