package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/lum-network/chain/x/millions/types"
)

// TransferDepositToRemoteZone Transfer a deposit to a remote zone
// - wait for the ICA callback to move to OnTransferDepositToRemoteZoneCompleted
// - or go to OnTransferDepositToRemoteZoneCompleted directly if local zone (no-op)
func (k Keeper) TransferDepositToRemoteZone(ctx sdk.Context, poolID uint64, depositID uint64) error {
	// Acquire pool config
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	// Acquire the deposit
	deposit, err := k.GetPoolDeposit(ctx, poolID, depositID)
	if err != nil {
		return err
	}
	if deposit.State != types.DepositState_IbcTransfer {
		return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DepositState_IbcTransfer.String(), deposit.State.String())
	}

	poolRunner := k.MustGetPoolRunner(pool.PoolType)
	if err := poolRunner.TransferDepositToRemoteZone(ctx, pool, deposit); err != nil {
		// Return with error (if any) here since it is the first operation and nothing needs to be saved to state
		return err
	}
	if pool.IsLocalZone(ctx) {
		// Move instantly to next stage since local ops don't wait for callbacks
		return k.OnTransferDepositToRemoteZoneCompleted(ctx, poolID, depositID, false)
	}
	return nil
}

// OnTransferDepositToRemoteZoneCompleted Acknowledge the IBC transfer to a remote zone response
// then moves to DelegateDepositOnRemoteZone in case of success
func (k Keeper) OnTransferDepositToRemoteZoneCompleted(ctx sdk.Context, poolID uint64, depositID uint64, isError bool) error {
	// Acquire the deposit
	deposit, err := k.GetPoolDeposit(ctx, poolID, depositID)
	if err != nil {
		return err
	}

	// Verify deposit state
	if deposit.State != types.DepositState_IbcTransfer {
		return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DepositState_IbcTransfer.String(), deposit.State.String())
	}

	if isError {
		k.UpdateDepositStatus(ctx, poolID, depositID, types.DepositState_IbcTransfer, true)
		return nil
	}

	k.UpdateDepositStatus(ctx, poolID, depositID, types.DepositState_IcaDelegate, false)
	return k.DelegateDepositOnRemoteZone(ctx, poolID, depositID)
}

// DelegateDepositOnRemoteZone Delegates a deposit on the remote zone (put it to work)
// - wait for the ICA callback to move to OnDelegateDepositOnRemoteZoneCompleted
// - or go to OnDelegateDepositOnRemoteZoneCompleted directly if local zone
func (k Keeper) DelegateDepositOnRemoteZone(ctx sdk.Context, poolID uint64, depositID uint64) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	deposit, err := k.GetPoolDeposit(ctx, poolID, depositID)
	if err != nil {
		return err
	}
	if deposit.State != types.DepositState_IcaDelegate {
		return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DepositState_IcaDelegate.String(), deposit.State.String())
	}

	// Return with error (if any) here since it is the first operation and nothing needs to be saved to state
	poolRunner := k.MustGetPoolRunner(pool.PoolType)
	splits, err := poolRunner.DelegateDepositOnRemoteZone(ctx, pool, deposit)
	if pool.IsLocalZone(ctx) {
		// Always save state in case of local zone
		// TODO: we should probably return and error here since it is atomic with the deposit transaction
		return k.OnDelegateDepositOnRemoteZoneCompleted(ctx, poolID, depositID, splits, err != nil)
	} else if err != nil {
		// Only save state in case of error for remote zones (otherwise wait for the callback)
		return k.OnDelegateDepositOnRemoteZoneCompleted(ctx, poolID, depositID, splits, true)
	}
	return nil
}

// OnDelegateDepositOnRemoteZoneCompleted Acknowledge the ICA operation to delegate a deposit on the remote zone (deposit has been put to work)
func (k Keeper) OnDelegateDepositOnRemoteZoneCompleted(ctx sdk.Context, poolID uint64, depositID uint64, splits []*types.SplitDelegation, isError bool) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}
	deposit, err := k.GetPoolDeposit(ctx, poolID, depositID)
	if err != nil {
		return err
	}

	// Verify deposit state
	if deposit.State != types.DepositState_IcaDelegate {
		return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DepositState_IcaDelegate.String(), deposit.State.String())
	}

	if isError {
		k.UpdateDepositStatus(ctx, poolID, depositID, types.DepositState_IcaDelegate, true)
		return nil
	}

	// Update validators
	pool.ApplySplitDelegate(ctx, splits)
	k.updatePool(ctx, &pool)
	k.UpdateDepositStatus(ctx, poolID, depositID, types.DepositState_Success, false)
	return nil
}

// GetNextDepositID gets the next deposit ID
func (k Keeper) GetNextDepositID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextDepositId := gogotypes.UInt64Value{}

	b := store.Get(types.NextDepositPrefix)
	if b == nil {
		panic(fmt.Errorf("getting at key (%v) should not have been nil", types.NextDepositPrefix))
	}
	k.cdc.MustUnmarshal(b, &nextDepositId)
	return nextDepositId.GetValue()
}

// GetNextDepositIdAndIncrement gets the next deposit ID and store the incremented ID
func (k Keeper) GetNextDepositIdAndIncrement(ctx sdk.Context) uint64 {
	nextDepositId := k.GetNextDepositID(ctx)
	k.SetNextDepositID(ctx, nextDepositId+1)
	return nextDepositId
}

// SetNextDepositID sets next deposit ID
func (k Keeper) SetNextDepositID(ctx sdk.Context, depositId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: depositId})
	store.Set(types.NextDepositPrefix, bz)
}

// GetPoolDeposit returns a deposit by ID for a given poolID
func (k Keeper) GetPoolDeposit(ctx sdk.Context, poolID uint64, depositID uint64) (types.Deposit, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolDepositKey(poolID, depositID))
	if bz == nil {
		return types.Deposit{}, types.ErrDepositNotFound
	}

	var deposit types.Deposit
	if err := k.cdc.Unmarshal(bz, &deposit); err != nil {
		return types.Deposit{}, err
	}

	return deposit, nil
}

// AddDeposit adds a deposit to a pool and account
// A new depositID is generated if not provided
// - adds it to the pool {pool_id, deposit_id}
// - adds it to the account {address, pool_id, deposit_id} deposits
func (k Keeper) AddDeposit(ctx sdk.Context, deposit *types.Deposit) {
	// Automatically affect ID if missing
	if deposit.GetDepositId() == types.UnknownID {
		deposit.DepositId = k.GetNextDepositIdAndIncrement(ctx)
	}
	// Ensure payload is valid
	if err := deposit.ValidateBasic(); err != nil {
		panic(err)
	}
	// Ensure we never override an existing entity
	if _, err := k.GetPoolDeposit(ctx, deposit.GetPoolId(), deposit.GetDepositId()); err == nil {
		panic(errorsmod.Wrapf(types.ErrEntityOverride, "ID %d", deposit.GetDepositId()))
	}

	isFirstAccountDeposit := !k.hasPoolDeposit(ctx, deposit.DepositorAddress, deposit.PoolId)

	// Update pool deposits
	k.setPoolDeposit(ctx, deposit)

	// Update account deposits
	k.setAccountDeposit(ctx, deposit)

	// Update pool TVL and Depositors Count
	pool, err := k.GetPool(ctx, deposit.PoolId)
	if err != nil {
		panic(err)
	}
	pool.TvlAmount = pool.TvlAmount.Add(deposit.GetAmount().Amount)
	if deposit.IsSponsor {
		if pool.SponsorshipAmount.IsNil() {
			pool.SponsorshipAmount = sdk.ZeroInt()
		}
		pool.SponsorshipAmount = pool.SponsorshipAmount.Add(deposit.GetAmount().Amount)
	}

	if isFirstAccountDeposit {
		// First deposit for account added
		pool.DepositorsCount++
	}

	k.updatePool(ctx, &pool)
}

// RemoveDeposit removes a deposit from a pool
// - removes it from the {pool_id, deposit_id}
// - removes it from the {address, pool_id, deposit_id} deposits
func (k Keeper) RemoveDeposit(ctx sdk.Context, deposit *types.Deposit) {
	// Ensure payload is valid
	if err := deposit.ValidateBasic(); err != nil {
		panic(err)
	}
	store := ctx.KVStore(k.storeKey)

	addr := sdk.MustAccAddressFromBech32(deposit.DepositorAddress)
	store.Delete(types.GetPoolDepositKey(deposit.PoolId, deposit.DepositId))
	store.Delete(types.GetAccountPoolDepositKey(addr, deposit.PoolId, deposit.DepositId))

	// Update pool TVL and Depositors Count
	pool, err := k.GetPool(ctx, deposit.PoolId)
	if err != nil {
		panic(err)
	}
	pool.TvlAmount = pool.TvlAmount.Sub(deposit.GetAmount().Amount)
	if deposit.IsSponsor {
		pool.SponsorshipAmount = pool.SponsorshipAmount.Sub(deposit.GetAmount().Amount)
	}

	if !k.hasPoolDeposit(ctx, deposit.DepositorAddress, deposit.PoolId) {
		// Last deposit for account removed
		pool.DepositorsCount--
	}

	k.updatePool(ctx, &pool)
}

// UpdateDepositStatus Update a given deposit status by its ID
func (k Keeper) UpdateDepositStatus(ctx sdk.Context, poolID uint64, depositID uint64, status types.DepositState, isError bool) {
	deposit, err := k.GetPoolDeposit(ctx, poolID, depositID)
	if err != nil {
		panic(err)
	}

	if isError {
		deposit.State = types.DepositState_Failure
		deposit.ErrorState = status
	} else {
		deposit.State = status
		deposit.ErrorState = types.DepositState_Unspecified
	}

	deposit.UpdatedAtHeight = ctx.BlockHeight()
	deposit.UpdatedAt = ctx.BlockTime()
	deposit.StateChangedLastAt = ctx.BlockTime()
	k.setAccountDeposit(ctx, &deposit)
	k.setPoolDeposit(ctx, &deposit)
}

// EditDeposit edits a deposit winnerAddr and sponsor mode
func (k Keeper) EditDeposit(ctx sdk.Context, poolID uint64, depositID uint64, winnerAddr sdk.AccAddress, isSponsor bool) error {
	deposit, err := k.GetPoolDeposit(ctx, poolID, depositID)
	if err != nil {
		return err
	}

	// Get pool to grab SponsorshipAmount
	pool, err := k.GetPool(ctx, deposit.PoolId)
	if err != nil {
		return err
	}

	// Check incoming sponsor mode against the deposit sponsor
	if isSponsor != deposit.IsSponsor {
		if isSponsor {
			pool.SponsorshipAmount = pool.SponsorshipAmount.Add(deposit.GetAmount().Amount)
			k.updatePool(ctx, &pool)
		} else {
			pool.SponsorshipAmount = pool.SponsorshipAmount.Sub(deposit.GetAmount().Amount)
			k.updatePool(ctx, &pool)
		}
	}

	deposit.WinnerAddress = winnerAddr.String()
	deposit.IsSponsor = isSponsor
	deposit.UpdatedAtHeight = ctx.BlockHeight()
	deposit.UpdatedAt = ctx.BlockTime()

	k.setAccountDeposit(ctx, &deposit)
	k.setPoolDeposit(ctx, &deposit)

	return nil
}

// setPoolDeposit sets a deposit to the pool deposit key
func (k Keeper) setPoolDeposit(ctx sdk.Context, deposit *types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPoolDepositKey(deposit.PoolId, deposit.DepositId)
	encodedDeposit := k.cdc.MustMarshal(deposit)
	store.Set(key, encodedDeposit)
}

// setAccountDeposit set the deposit for the Account deposit key
func (k Keeper) setAccountDeposit(ctx sdk.Context, deposit *types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	depositorAddress := sdk.MustAccAddressFromBech32(deposit.DepositorAddress)
	key := types.GetAccountPoolDepositKey(depositorAddress, deposit.PoolId, deposit.DepositId)
	encodedDeposit := k.cdc.MustMarshal(deposit)
	store.Set(key, encodedDeposit)
}

// hasPoolDeposit returns whether or not an account has at least one deposit for a given pool
func (k Keeper) hasPoolDeposit(ctx sdk.Context, address string, poolID uint64) bool {
	addr := sdk.MustAccAddressFromBech32(address)

	kvStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(kvStore, types.GetAccountPoolDepositsKey(addr, poolID))

	defer iterator.Close()
	return iterator.Valid()
}

// ListPoolDeposits returns all deposits for a given poolID
// Warning: expensive operation
func (k Keeper) ListPoolDeposits(ctx sdk.Context, poolID uint64) (deposits []types.Deposit) {
	depositStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(depositStore, types.GetPoolDepositsKey(poolID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}
	return
}

// ListAccountPoolDeposits return all the deposits for and address and a poolID
// Warning: expensive operation
func (k Keeper) ListAccountPoolDeposits(ctx sdk.Context, addr sdk.Address, poolID uint64) (deposits []types.Deposit) {
	depositStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(depositStore, types.GetAccountPoolDepositsKey(addr, poolID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}
	return
}

// ListAccountDeposits return deposits all the deposits for an address
// Warning: expensive operation
func (k Keeper) ListAccountDeposits(ctx sdk.Context, addr sdk.Address) (deposits []types.Deposit) {
	depositStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(depositStore, types.GetAccountDepositsKey(addr))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}
	return
}

// ListDeposits return all the deposits for and address
// Warning: expensive operation
func (k Keeper) ListDeposits(ctx sdk.Context) (deposits []types.Deposit) {
	depositStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(depositStore, types.GetDepositsKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}
	return
}
