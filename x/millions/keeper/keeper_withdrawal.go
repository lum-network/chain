package keeper

import (
	"fmt"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

// UndelegateWithdrawalsOnRemoteZone Undelegates an epoch unbonding containing withdrawalIDs from the native chain validators
// - go to OnUndelegateWithdrawalsOnRemoteZoneCompleted directly upon undelegate success if local zone
// - or wait for the ICA callback to move to OnUndelegateWithdrawalsOnRemoteZoneCompleted
func (k Keeper) UndelegateWithdrawalsOnRemoteZone(ctx sdk.Context, epochUnbonding types.EpochUnbonding) error {
	pool, err := k.GetPool(ctx, epochUnbonding.PoolId)
	if err != nil {
		return err
	}

	// Update withdrawal status to WithdrawalState_IcaUndelegate
	for _, wid := range epochUnbonding.WithdrawalIds {
		withdrawal, err := k.GetPoolWithdrawal(ctx, pool.PoolId, wid)
		if err != nil {
			return err
		}
		if withdrawal.State != types.WithdrawalState_Pending {
			return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.WithdrawalState_Pending.String(), withdrawal.State.String())
		}
		k.UpdateWithdrawalStatus(ctx, withdrawal.PoolId, withdrawal.WithdrawalId, types.WithdrawalState_IcaUndelegate, nil, false)
	}

	poolRunner := k.MustGetPoolRunner(pool.PoolType)
	splits, unbondingEndsAt, err := poolRunner.UndelegateWithdrawalsOnRemoteZone(ctx, epochUnbonding)

	// Apply undelegate pool validators update
	pool.ApplySplitUndelegate(ctx, splits)
	k.updatePool(ctx, &pool)

	if err != nil {
		// Return with error here and let the caller manage the state changes if needed
		return err
	}

	if pool.IsLocalZone(ctx) {
		return k.OnUndelegateWithdrawalsOnRemoteZoneCompleted(ctx, epochUnbonding.PoolId, epochUnbonding.WithdrawalIds, unbondingEndsAt, false)
	}

	return nil
}

// OnUndelegateWithdrawalsOnRemoteZoneCompleted aknowledged the undelegate callback
// If error occurs send the withdrawal to a new epoch unbonding
func (k Keeper) OnUndelegateWithdrawalsOnRemoteZoneCompleted(ctx sdk.Context, poolID uint64, withdrawalIDs []uint64, unbondingEndsAt *time.Time, isError bool) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	// Track total amount in case of error to revert pool to appropriate state
	totalAmount := sdk.ZeroInt()

	// Update withdrawals
	for _, wid := range withdrawalIDs {
		withdrawal, err := k.GetPoolWithdrawal(ctx, pool.PoolId, wid)
		if err != nil {
			return err
		}

		// Check withdrawal state
		if withdrawal.State != types.WithdrawalState_IcaUndelegate {
			return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.WithdrawalState_IcaUndelegate.String(), withdrawal.State.String())
		}

		if isError {
			totalAmount = totalAmount.Add(withdrawal.Amount.Amount)
			k.UpdateWithdrawalStatus(ctx, withdrawal.PoolId, withdrawal.WithdrawalId, types.WithdrawalState_IcaUndelegate, nil, true)
			// Add failed withdrawals to a fresh epoch unbonding for a retry
			if err := k.AddEpochUnbonding(ctx, withdrawal, true); err != nil {
				return err
			}
		} else {
			// Set the unbondingEndsAt and add withdrawal to matured queue
			withdrawal.UnbondingEndsAt = unbondingEndsAt
			k.UpdateWithdrawalStatus(ctx, withdrawal.PoolId, withdrawal.WithdrawalId, types.WithdrawalState_IcaUnbonding, unbondingEndsAt, false)
			k.addWithdrawalToMaturedQueue(ctx, withdrawal)
		}
	}

	if isError {
		// Revert undelegate pool validators update
		splits := pool.ComputeSplitUndelegations(ctx, totalAmount)
		pool.ApplySplitDelegate(ctx, splits)
		k.updatePool(ctx, &pool)
	}

	return nil
}

// TransferWithdrawalToRecipient transfers a withdrawal to its destination address
// - If local zone and local toAddress: BankSend with instant call to OnTransferWithdrawalToRecipientCompleted
// - If remote zone and remote toAddress: ICA BankSend and wait for ICA callback
// - If remote zone and local toAddress: IBC Transfer and wait for ICA callback
func (k Keeper) TransferWithdrawalToRecipient(ctx sdk.Context, poolID uint64, withdrawalID uint64) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}
	withdrawal, err := k.GetPoolWithdrawal(ctx, poolID, withdrawalID)
	if err != nil {
		return err
	}

	// Check state ibc transfer and unbonding condition
	if withdrawal.State != types.WithdrawalState_IbcTransfer {
		return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.WithdrawalState_IbcTransfer.String(), withdrawal.State.String())
	} else if withdrawal.UnbondingEndsAt == nil || withdrawal.UnbondingEndsAt.After(ctx.BlockTime()) {
		return errorsmod.Wrapf(types.ErrIllegalStateOperation, "unbonding in progress")
	}

	poolRunner := k.MustGetPoolRunner(pool.PoolType)
	if err := poolRunner.TransferWithdrawalToRecipient(ctx, pool, withdrawal); err != nil {
		// Return with error here and let the caller manage the state changes if needed
		return err
	}
	if pool.IsLocalZone(ctx) {
		// Move instantly to next stage since local ops don't wait for callbacks
		return k.OnTransferWithdrawalToRecipientCompleted(ctx, poolID, withdrawalID, false)
	}
	return nil
}

// OnTransferWithdrawalToRecipientCompleted Acknowledge the withdrawal IBC transfer
// - To the local chain response if it's a transfer to local chain
// - To the native chain if it's BankSend for a native pool with a native destination address
func (k Keeper) OnTransferWithdrawalToRecipientCompleted(ctx sdk.Context, poolID uint64, withdrawalID uint64, isError bool) error {
	withdrawal, err := k.GetPoolWithdrawal(ctx, poolID, withdrawalID)
	if err != nil {
		return err
	}

	if withdrawal.State != types.WithdrawalState_IbcTransfer {
		return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.WithdrawalState_IbcTransfer.String(), withdrawal.State.String())
	}

	if isError {
		k.UpdateWithdrawalStatus(ctx, poolID, withdrawalID, types.WithdrawalState_IbcTransfer, withdrawal.UnbondingEndsAt, true)
		return nil
	}

	if err := k.RemoveWithdrawal(ctx, withdrawal); err != nil {
		return err
	}

	// Notify the closing method to check if the step can continue
	// Discard any error here to avoid blocking the process on relaying side
	if err := k.ClosePool(ctx, poolID, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("ClosePool %v", err.Error()))
	}
	return nil
}

// GetNextWithdrawalID gets the next withdrawal deposit ID
func (k Keeper) GetNextWithdrawalID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextWithdrawalId := gogotypes.UInt64Value{}

	b := store.Get(types.NextWithdrawalPrefix)
	if b == nil {
		panic(fmt.Errorf("getting at key (%v) should not have been nil", types.NextWithdrawalPrefix))
	}
	k.cdc.MustUnmarshal(b, &nextWithdrawalId)
	return nextWithdrawalId.GetValue()
}

// GetNextWithdrawalIdAndIncrement gets the next withdrawal ID and store the incremented ID
func (k Keeper) GetNextWithdrawalIdAndIncrement(ctx sdk.Context) uint64 {
	nextWithdrawlId := k.GetNextWithdrawalID(ctx)
	k.SetNextWithdrawalID(ctx, nextWithdrawlId+1)
	return nextWithdrawlId
}

// SetNextWithdrawalID sets next withdrawal ID
func (k Keeper) SetNextWithdrawalID(ctx sdk.Context, withdrawalID uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: withdrawalID})
	store.Set(types.NextWithdrawalPrefix, bz)
}

// GetPoolWithdrawal returns a withdrawal by poolID, withdrawalID
func (k Keeper) GetPoolWithdrawal(ctx sdk.Context, poolID uint64, withdrawalID uint64) (types.Withdrawal, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolWithdrawalKey(poolID, withdrawalID))
	if bz == nil {
		return types.Withdrawal{}, types.ErrWithdrawalNotFound
	}

	var withdrawal types.Withdrawal
	if err := k.cdc.Unmarshal(bz, &withdrawal); err != nil {
		return types.Withdrawal{}, err
	}

	return withdrawal, nil
}

// RemoveWithdrawal removes a successful withdrawal for a given account and pool
func (k Keeper) RemoveWithdrawal(ctx sdk.Context, withdrawal types.Withdrawal) error {
	// Ensure payload is valid
	if err := withdrawal.ValidateBasic(); err != nil {
		return err
	}

	// Ensure withdrawal entity exists
	withdrawal, err := k.GetPoolWithdrawal(ctx, withdrawal.PoolId, withdrawal.WithdrawalId)
	if err != nil {
		return err
	}

	if withdrawal.UnbondingEndsAt != nil {
		k.removeWithdrawalFromMaturedQueue(ctx, withdrawal)
	}

	store := ctx.KVStore(k.storeKey)
	addr := sdk.MustAccAddressFromBech32(withdrawal.DepositorAddress)
	store.Delete(types.GetPoolWithdrawalKey(withdrawal.PoolId, withdrawal.WithdrawalId))
	store.Delete(types.GetAccountPoolWithdrawalKey(addr, withdrawal.PoolId, withdrawal.WithdrawalId))

	return nil
}

// UpdateWithdrawalStatus Update a given withdrawal status by its ID
func (k Keeper) UpdateWithdrawalStatus(ctx sdk.Context, poolID uint64, withdrawalID uint64, status types.WithdrawalState, unbondingEndsAt *time.Time, isError bool) {
	withdrawal, err := k.GetPoolWithdrawal(ctx, poolID, withdrawalID)
	if err != nil {
		panic(err)
	}

	if isError {
		withdrawal.State = types.WithdrawalState_Failure
		withdrawal.ErrorState = status
	} else {
		withdrawal.State = status
		withdrawal.ErrorState = types.WithdrawalState_Unspecified
	}

	withdrawal.UpdatedAtHeight = ctx.BlockHeight()
	withdrawal.UpdatedAt = ctx.BlockTime()
	if unbondingEndsAt != nil {
		withdrawal.UnbondingEndsAt = unbondingEndsAt
	}
	// Update pool withdraw deposit
	k.setPoolWithdrawal(ctx, withdrawal)
	// Update account withdraw deposit
	k.setAccountWithdrawal(ctx, withdrawal)
}

// AddWithdrawal adds a withdrawDeposit to a pool and account
// - adds it to the pool {pool_id, withdrawal_id}
// - adds it to the account {depositor_address, pool_id, withdrawal_id} withdrawDeposits
func (k Keeper) AddWithdrawal(ctx sdk.Context, withdrawal types.Withdrawal) {
	// Automatically affect ID if missing
	if withdrawal.GetWithdrawalId() == types.UnknownID {
		withdrawal.WithdrawalId = k.GetNextWithdrawalIdAndIncrement(ctx)
	}
	// Ensure payload is valid
	if err := withdrawal.ValidateBasic(); err != nil {
		panic(err)
	}

	// Ensure we never override an existing entity
	if _, err := k.GetPoolWithdrawal(ctx, withdrawal.GetPoolId(), withdrawal.GetWithdrawalId()); err == nil {
		panic(errorsmod.Wrapf(types.ErrEntityOverride, "ID %d", withdrawal.GetWithdrawalId()))
	}

	// Update pool withdraw deposit
	k.setPoolWithdrawal(ctx, withdrawal)
	// Update account withdraw deposit
	k.setAccountWithdrawal(ctx, withdrawal)

	// Add withdrawal to unbonding queue if needed
	if withdrawal.State == types.WithdrawalState_IcaUnbonding {
		k.addWithdrawalToMaturedQueue(ctx, withdrawal)
	}
}

// setPoolWithdrawal sets a withdraw deposit to the pool withdrawal key
func (k Keeper) setPoolWithdrawal(ctx sdk.Context, withdrawal types.Withdrawal) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPoolWithdrawalKey(withdrawal.PoolId, withdrawal.WithdrawalId)
	encodedWithDrawal := k.cdc.MustMarshal(&withdrawal)
	store.Set(key, encodedWithDrawal)
}

// setAccountWithdrawal sets a withdraw deposit to the account withdrawal key
func (k Keeper) setAccountWithdrawal(ctx sdk.Context, withdrawal types.Withdrawal) {
	store := ctx.KVStore(k.storeKey)
	withdrawAddress := sdk.MustAccAddressFromBech32(withdrawal.DepositorAddress)
	key := types.GetAccountPoolWithdrawalKey(withdrawAddress, withdrawal.PoolId, withdrawal.WithdrawalId)
	encodedWithDrawal := k.cdc.MustMarshal(&withdrawal)
	store.Set(key, encodedWithDrawal)
}

// addWithdrawalToMaturedQueue adds a withdrawal to the WithrawalMatured queue
func (k Keeper) addWithdrawalToMaturedQueue(ctx sdk.Context, withdrawal types.Withdrawal) {
	withdrawalStore := ctx.KVStore(k.storeKey)
	col := k.GetWithdrawalIDsMaturedQueue(ctx, *withdrawal.UnbondingEndsAt)
	col.WithdrawalsIds = append(col.WithdrawalsIds, types.WithdrawalIDs{
		PoolId:       withdrawal.PoolId,
		WithdrawalId: withdrawal.WithdrawalId,
	})
	withdrawalStore.Set(types.GetMaturedWithdrawalTimeKey(*withdrawal.UnbondingEndsAt), k.cdc.MustMarshal(&col))
}

// removeWithdrawalFromMaturedQueue removed a withdrawal from the WithrawalMatured queue
func (k Keeper) removeWithdrawalFromMaturedQueue(ctx sdk.Context, withdrawal types.Withdrawal) {
	withdrawalStore := ctx.KVStore(k.storeKey)
	col := k.GetWithdrawalIDsMaturedQueue(ctx, *withdrawal.UnbondingEndsAt)
	for i, wid := range col.WithdrawalsIds {
		if wid.PoolId == withdrawal.PoolId && wid.WithdrawalId == withdrawal.WithdrawalId {
			col.WithdrawalsIds = append(col.WithdrawalsIds[:i], col.WithdrawalsIds[i+1:]...)
			withdrawalStore.Set(types.GetMaturedWithdrawalTimeKey(*withdrawal.UnbondingEndsAt), k.cdc.MustMarshal(&col))
			break
		}
	}
	withdrawalStore.Set(types.GetMaturedWithdrawalTimeKey(*withdrawal.UnbondingEndsAt), k.cdc.MustMarshal(&col))
}

// GetWithdrawalIDsMaturedQueue gets a withdrawal IDs collection for the matured unbonding timestamp
func (k Keeper) GetWithdrawalIDsMaturedQueue(ctx sdk.Context, timestamp time.Time) (col types.WithdrawalIDsCollection) {
	withdrawalStore := ctx.KVStore(k.storeKey)
	bz := withdrawalStore.Get(types.GetMaturedWithdrawalTimeKey(timestamp))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &col)
	return
}

// MaturedWithdrawalQueueIterator returns an iterator for the Withdrawal Matured queue up to the specified endTime
func (k Keeper) MaturedWithdrawalQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	withdrawalStore := ctx.KVStore(k.storeKey)
	// Add end bytes to ensure the last item gets included in the iterator
	return withdrawalStore.Iterator(types.WithdrawalMaturationTimePrefix, append(types.GetMaturedWithdrawalTimeKey(endTime), byte(0x00)))
}

// DequeueMaturedWithdrawalQueue return all the Matured Withdrawals that can be transfered and can be removed from the queue
func (k Keeper) DequeueMaturedWithdrawalQueue(ctx sdk.Context, endTime time.Time) (withdrawalsIDs []types.WithdrawalIDs) {
	withdrawalStore := ctx.KVStore(k.storeKey)

	iterator := k.MaturedWithdrawalQueueIterator(ctx, endTime)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var col types.WithdrawalIDsCollection
		k.cdc.MustUnmarshal(iterator.Value(), &col)
		withdrawalsIDs = append(withdrawalsIDs, col.WithdrawalsIds...)
		withdrawalStore.Delete(iterator.Key())
	}

	return
}

// ListWithdrawals return all the withdrawals
// Warning: expensive operation
func (k Keeper) ListWithdrawals(ctx sdk.Context) (withdrawals []types.Withdrawal) {
	withdrawDepositStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(withdrawDepositStore, types.GetWithdrawalsKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var withdrawal types.Withdrawal
		k.cdc.MustUnmarshal(iterator.Value(), &withdrawal)
		withdrawals = append(withdrawals, withdrawal)
	}
	return
}

// ListAccountWithdrawals return all the withdraw deposits account
// Warning: expensive operation
func (k Keeper) ListAccountWithdrawals(ctx sdk.Context, addr sdk.Address) (withdrawals []types.Withdrawal) {
	withdrawDepositStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(withdrawDepositStore, types.GetAccountWithdrawalsKey(addr))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var withdrawal types.Withdrawal
		k.cdc.MustUnmarshal(iterator.Value(), &withdrawal)
		withdrawals = append(withdrawals, withdrawal)
	}
	return
}

// ListAccountPoolWithdrawals return all the withdrawals for and address and a poolID
// Warning: expensive operation
func (k Keeper) ListAccountPoolWithdrawals(ctx sdk.Context, addr sdk.Address, poolID uint64) (withdrawals []types.Withdrawal) {
	withdrawalStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(withdrawalStore, types.GetAccountPoolWithdrawalsKey(addr, poolID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var withdrawal types.Withdrawal
		k.cdc.MustUnmarshal(iterator.Value(), &withdrawal)
		withdrawals = append(withdrawals, withdrawal)
	}
	return
}

// ListPoolWithdrawals returns all withdrawals for a given poolID
// Warning: expensive operation
func (k Keeper) ListPoolWithdrawals(ctx sdk.Context, poolID uint64) (withdrawals []types.Withdrawal) {
	withdrawalStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(withdrawalStore, types.GetPoolWithdrawalsKey(poolID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var withdrawal types.Withdrawal
		k.cdc.MustUnmarshal(iterator.Value(), &withdrawal)
		withdrawals = append(withdrawals, withdrawal)
	}
	return
}

func (k Keeper) UnsafeSetUnpersistedWithdrawals(ctx sdk.Context) int {
	i := 0
	for _, withdrawal := range k.ListWithdrawals(ctx) {
		k.setPoolWithdrawal(ctx, withdrawal)
		k.setAccountWithdrawal(ctx, withdrawal)
		i++
	}

	return i
}
