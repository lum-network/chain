package keeper

import (
	"fmt"
	"time"

	gogotypes "github.com/gogo/protobuf/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

// ClawBackPrize claw backs a prize by adding its amount to the clawback prize pool
func (k Keeper) ClawBackPrize(ctx sdk.Context, poolID uint64, drawID uint64, prizeID uint64) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}
	prize, err := k.GetPoolDrawPrize(ctx, poolID, drawID, prizeID)
	if err != nil {
		return err
	}
	if prize.State != types.PrizeState_Pending {
		return errorsmod.Wrapf(types.ErrIllegalStateOperation, "expecting %s but state is %s", types.PrizeState_Pending, prize.State)
	}

	pool.AvailablePrizePool = pool.AvailablePrizePool.Add(prize.Amount)
	k.updatePool(ctx, &pool)

	if err := k.RemovePrize(ctx, prize); err != nil {
		return err
	}

	return nil
}

// AddPrizes adds a prize to a pool and an account
// A new prizeID is generated if not provided
// - adds it to the pool {pool_id, draw_id, prize_id}
// - adds it to the account {winner_address, pool_id, draw_id} prizes
func (k Keeper) AddPrize(ctx sdk.Context, prize types.Prize) {
	// Automatically affect ID if missing
	if prize.GetPrizeId() == types.UnknownID {
		prize.PrizeId = k.GetNextPrizeIdAndIncrement(ctx)
	}

	// Ensure payload is valid
	if err := prize.ValidateBasic(); err != nil {
		panic(err)
	}
	// Ensure we never override an existing entity
	if _, err := k.GetPoolDrawPrize(ctx, prize.GetPoolId(), prize.GetDrawId(), prize.GetPrizeId()); err == nil {
		panic(errorsmod.Wrapf(types.ErrEntityOverride, "ID %d", prize.GetPrizeId()))
	}

	// Update pool prize
	k.setPoolPrize(ctx, prize)
	// Update account prize
	k.setAccountPrize(ctx, prize)
	// Add prize to EPCB queue if needed
	if prize.State == types.PrizeState_Pending {
		k.addPrizeToEPCBQueue(ctx, prize)
	}
}

// GetNextPrizeID gets the next prize ID
func (k Keeper) GetNextPrizeID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextPrizeId := gogotypes.UInt64Value{}

	b := store.Get(types.NextPrizePrefix)
	if b == nil {
		panic(fmt.Errorf("getting at key (%v) should not have been nil", types.NextPrizePrefix))
	}
	k.cdc.MustUnmarshal(b, &nextPrizeId)
	return nextPrizeId.GetValue()
}

// GetNextPrizeIdAndIncrement gets the next prize ID and store the incremented ID
func (k Keeper) GetNextPrizeIdAndIncrement(ctx sdk.Context) uint64 {
	nextPrizeId := k.GetNextPrizeID(ctx)
	k.SetNextPrizeID(ctx, nextPrizeId+1)
	return nextPrizeId
}

// SetNextPrizeID sets next prize ID
func (k Keeper) SetNextPrizeID(ctx sdk.Context, prizeID uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: prizeID})
	store.Set(types.NextPrizePrefix, bz)
}

// GetPoolDrawPrize returns a prize by poolID, drawID, prizeID
func (k Keeper) GetPoolDrawPrize(ctx sdk.Context, poolID uint64, drawID uint64, prizeID uint64) (types.Prize, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolDrawPrizeKey(poolID, drawID, prizeID))
	if bz == nil {
		return types.Prize{}, types.ErrPrizeNotFound
	}

	var prize types.Prize
	if err := k.cdc.Unmarshal(bz, &prize); err != nil {
		return types.Prize{}, err
	}

	return prize, nil
}

// RemovePrize removes a prize from the store
func (k Keeper) RemovePrize(ctx sdk.Context, prize types.Prize) error {
	// Ensure payload is valid
	if err := prize.ValidateBasic(); err != nil {
		return err
	}

	// Ensure prize entity exists
	prize, err := k.GetPoolDrawPrize(ctx, prize.PoolId, prize.DrawId, prize.PrizeId)
	if err != nil {
		return err
	}

	k.removePrizeFromEPCBQueue(ctx, prize)

	store := ctx.KVStore(k.storeKey)

	addr := sdk.MustAccAddressFromBech32(prize.WinnerAddress)
	store.Delete(types.GetPoolDrawPrizeKey(prize.PoolId, prize.DrawId, prize.PrizeId))
	store.Delete(types.GetAccountPoolDrawPrizeKey(addr, prize.PoolId, prize.DrawId, prize.PrizeId))

	return nil
}

// setPoolPrize sets a prize to the pool prize key
func (k Keeper) setPoolPrize(ctx sdk.Context, prize types.Prize) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPoolDrawPrizeKey(prize.PoolId, prize.DrawId, prize.PrizeId)
	encodedPrize := k.cdc.MustMarshal(&prize)
	store.Set(key, encodedPrize)
}

// setAccountPrize set the prize for the Account prize key
func (k Keeper) setAccountPrize(ctx sdk.Context, prize types.Prize) {
	store := ctx.KVStore(k.storeKey)
	winnerAddress := sdk.MustAccAddressFromBech32(prize.WinnerAddress)
	key := types.GetAccountPoolDrawPrizeKey(winnerAddress, prize.PoolId, prize.DrawId, prize.PrizeId)
	encodedPrize := k.cdc.MustMarshal(&prize)
	store.Set(key, encodedPrize)
}

// addPrizeToEPCBQueue adds a prize to the EPCB queue
func (k Keeper) addPrizeToEPCBQueue(ctx sdk.Context, prize types.Prize) {
	prizeStore := ctx.KVStore(k.storeKey)
	col := k.GetPrizeIDsEPCBQueue(ctx, prize.ExpiresAt)
	col.PrizesIds = append(col.PrizesIds, types.PrizeIDs{
		PoolId:  prize.PoolId,
		DrawId:  prize.DrawId,
		PrizeId: prize.PrizeId,
	})
	prizeStore.Set(types.GetExpiringPrizeTimeKey(prize.ExpiresAt), k.cdc.MustMarshal(&col))
}

// removePrizeFromEPCBQueue removes a prize from the EPCB queue
func (k Keeper) removePrizeFromEPCBQueue(ctx sdk.Context, prize types.Prize) {
	prizeStore := ctx.KVStore(k.storeKey)
	col := k.GetPrizeIDsEPCBQueue(ctx, prize.ExpiresAt)
	for i, pid := range col.PrizesIds {
		if pid.PoolId == prize.PoolId && pid.DrawId == prize.DrawId && pid.PrizeId == prize.PrizeId {
			col.PrizesIds = append(col.PrizesIds[:i], col.PrizesIds[i+1:]...)
			prizeStore.Set(types.GetExpiringPrizeTimeKey(prize.ExpiresAt), k.cdc.MustMarshal(&col))
			break
		}
	}
	prizeStore.Set(types.GetExpiringPrizeTimeKey(prize.ExpiresAt), k.cdc.MustMarshal(&col))
}

// GetPrizeIDsEPCBQueue gets a prize IDs collection for the expiring timestamp
func (k Keeper) GetPrizeIDsEPCBQueue(ctx sdk.Context, timestamp time.Time) (col types.PrizeIDsCollection) {
	prizeStore := ctx.KVStore(k.storeKey)
	bz := prizeStore.Get(types.GetExpiringPrizeTimeKey(timestamp))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &col)
	return
}

// EPCBQueueIterator returns an iterator for the EPCB queue up to the specified endTime
func (k Keeper) EPCBQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	prizeStore := ctx.KVStore(k.storeKey)
	// Add end bytes to ensure the last item gets included in the iterator
	return prizeStore.Iterator(types.PrizeExpirationTimePrefix, append(types.GetExpiringPrizeTimeKey(endTime), byte(0x00)))
}

// DequeueEPCBQueue return all the Expired Prizes to Claw Back and remove them from the queue
func (k Keeper) DequeueEPCBQueue(ctx sdk.Context, endTime time.Time) (prizes []types.PrizeIDs) {
	prizeStore := ctx.KVStore(k.storeKey)

	iterator := k.EPCBQueueIterator(ctx, endTime)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var col types.PrizeIDsCollection
		k.cdc.MustUnmarshal(iterator.Value(), &col)
		prizes = append(prizes, col.PrizesIds...)
		prizeStore.Delete(iterator.Key())
	}

	return
}

// ListAccountPoolPrizes return all the prizes for an account and a pool
// Warning: expensive operation
func (k Keeper) ListAccountPoolPrizes(ctx sdk.Context, addr sdk.Address, poolID uint64) (prizes []types.Prize) {
	prizeStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(prizeStore, types.GetAccountPoolPrizesKey(addr, poolID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var prize types.Prize
		k.cdc.MustUnmarshal(iterator.Value(), &prize)
		prizes = append(prizes, prize)
	}
	return
}

// ListAccountPrizes return all the prizes for an account
// Warning: expensive operation
func (k Keeper) ListAccountPrizes(ctx sdk.Context, addr sdk.Address) (prizes []types.Prize) {
	prizeStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(prizeStore, types.GetAccountPrizesKey(addr))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var prize types.Prize
		k.cdc.MustUnmarshal(iterator.Value(), &prize)
		prizes = append(prizes, prize)
	}
	return
}

// ListPoolDrawPrizes return all the prizes for a pool draw
// Warning: expensive operation
func (k Keeper) ListPoolDrawPrizes(ctx sdk.Context, poolID uint64, drawID uint64) (prizes []types.Prize) {
	prizeStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(prizeStore, types.GetPoolDrawPrizesKey(poolID, drawID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var prize types.Prize
		k.cdc.MustUnmarshal(iterator.Value(), &prize)
		prizes = append(prizes, prize)
	}
	return
}

// ListPoolPrizes return all the prizes for a pool
// Warning: expensive operation
func (k Keeper) ListPoolPrizes(ctx sdk.Context, poolID uint64) (prizes []types.Prize) {
	prizeStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(prizeStore, types.GetPoolPrizesKey(poolID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var prize types.Prize
		k.cdc.MustUnmarshal(iterator.Value(), &prize)
		prizes = append(prizes, prize)
	}
	return
}

// ListPrizes return all the prizes for an address
// Warning: expensive operation
func (k Keeper) ListPrizes(ctx sdk.Context) (prizes []types.Prize) {
	prizeStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(prizeStore, types.GetPrizesKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var prize types.Prize
		k.cdc.MustUnmarshal(iterator.Value(), &prize)
		prizes = append(prizes, prize)
	}
	return
}
