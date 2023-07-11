package keeper

import (
	"fmt"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/lum-network/chain/x/millions/types"
)

// UndelegateWithdrawalsOnRemoteZone Undelegates an epoch unbonding containing withdrawalIDs from the native chain validators
// - go to OnUndelegateWithdrawalsOnRemoteZoneCompleted directly upon undelegate success if local zone
// - or wait for the ICA callback to move to OnUndelegateWithdrawalsOnRemoteZoneCompleted
func (k Keeper) UndelegateWithdrawalsOnRemoteZone(ctx sdk.Context, epochUnbonding types.EpochUnbonding) error {
	logger := k.Logger(ctx).With("ctx", "withdrawal_undelegate")

	pool, err := k.GetPool(ctx, epochUnbonding.PoolId)
	if err != nil {
		return err
	}

	// Update withdrawal status to Ica_undelegate
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

	if pool.IsLocalZone(ctx) {
		modAddr := sdk.MustAccAddressFromBech32(pool.GetIcaDepositAddress())
		splits := pool.ComputeSplitUndelegations(ctx, epochUnbonding.GetTotalAmount().Amount)
		if len(splits) == 0 {
			return types.ErrPoolEmptySplitDelegations
		}

		var unbondingEndsAt *time.Time
		for _, split := range splits {
			valAddr, err := sdk.ValAddressFromBech32(split.ValidatorAddress)
			if err != nil {
				return err
			}
			shares, err := k.StakingKeeper.ValidateUnbondAmount(
				ctx,
				modAddr,
				valAddr,
				split.Amount,
			)
			if err != nil {
				return errorsmod.Wrapf(err, "%s", valAddr.String())
			}

			// Trigger undelegate
			endsAt, err := k.StakingKeeper.Undelegate(ctx, modAddr, valAddr, shares)
			if err != nil {
				return errorsmod.Wrapf(err, "%s", valAddr.String())
			}
			if unbondingEndsAt == nil || endsAt.After(*unbondingEndsAt) {
				unbondingEndsAt = &endsAt
			}
		}
		// Apply undelegate pool validators update
		pool.ApplySplitUndelegate(ctx, splits)
		k.updatePool(ctx, &pool)
		return k.OnUndelegateWithdrawalsOnRemoteZoneCompleted(ctx, epochUnbonding.PoolId, epochUnbonding.WithdrawalIds, unbondingEndsAt, false)
	}

	// Prepare undelegate split
	splits := pool.ComputeSplitUndelegations(ctx, epochUnbonding.GetTotalAmount().Amount)
	if len(splits) == 0 {
		return types.ErrPoolEmptySplitDelegations
	}

	// Construct our callback data
	callbackData := types.UndelegateCallback{
		PoolId:        epochUnbonding.PoolId,
		WithdrawalIds: epochUnbonding.WithdrawalIds,
	}
	marshalledCallbackData, err := k.MarshalUndelegateCallbackArgs(ctx, callbackData)
	if err != nil {
		return err
	}

	// Build undelegate tx
	var msgs []sdk.Msg
	for _, split := range splits {
		msgs = append(msgs, &stakingtypes.MsgUndelegate{
			DelegatorAddress: pool.GetIcaDepositAddress(),
			ValidatorAddress: split.ValidatorAddress,
			Amount:           sdk.NewCoin(pool.NativeDenom, split.Amount),
		})
	}

	// Apply undelegate pool validators update
	pool.ApplySplitUndelegate(ctx, splits)
	k.updatePool(ctx, &pool)

	// Dispatch our message with a timeout of 30 minutes in nanos
	sequence, err := k.BroadcastICAMessages(ctx, epochUnbonding.PoolId, types.ICATypeDeposit, msgs, types.IBCTimeoutNanos, ICACallbackID_Undelegate, marshalledCallbackData)
	if err != nil {
		// Return with error here since it is the first operation and nothing needs to be saved to state
		logger.Error(
			fmt.Sprintf("failed to dispatch ICA undelegate: %v", err),
			"pool_id", epochUnbonding.PoolId,
			"epoch_number", epochUnbonding.EpochNumber,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
		return err
	}
	logger.Debug(
		"ICA undelegate dispatched",
		"pool_id", epochUnbonding.PoolId,
		"epoch_number", epochUnbonding.EpochNumber,
		"chain_id", pool.GetChainId(),
		"sequence", sequence,
	)
	return nil
}

// OnUndelegateWithdrawalsOnRemoteZoneCompleted aknowledged the undelegate callback
// If error occurs send the withdrawal to a new epoch unbonding
func (k Keeper) OnUndelegateWithdrawalsOnRemoteZoneCompleted(ctx sdk.Context, poolID uint64, withdrawalIDs []uint64, unbondingEndsAt *time.Time, isError bool) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

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
			// Revert undelegate pool validators update
			splits := pool.ComputeSplitDelegations(ctx, withdrawal.GetAmount().Amount)
			pool.ApplySplitDelegate(ctx, splits)
			k.updatePool(ctx, &pool)
			k.UpdateWithdrawalStatus(ctx, withdrawal.PoolId, withdrawal.WithdrawalId, types.WithdrawalState_IcaUndelegate, nil, true)
			// Add failed withdrawals to a fresh epoch unbonding for a retry
			k.AddEpochUnbonding(ctx, withdrawal, true)
			continue
		}

		// Set the unbondingEndsAt and add withdrawal to matured queue
		withdrawal.UnbondingEndsAt = unbondingEndsAt
		k.UpdateWithdrawalStatus(ctx, withdrawal.PoolId, withdrawal.WithdrawalId, types.WithdrawalState_IcaUnbonding, unbondingEndsAt, false)
		k.addWithdrawalToMaturedQueue(ctx, withdrawal)
	}

	return nil
}

// TransferWithdrawalToDestAddr transfers a withdrawal to its destination address
// - If local zone and local toAddress: BankSend with instant call to OnTransferWithdrawalToDestAddrCompleted
// - If remote zone and remote toAddress: ICA BankSend and wait for ICA callback
// - If remote zone and local toAddress: IBC Transfer and wait for ICA callback
func (k Keeper) TransferWithdrawalToDestAddr(ctx sdk.Context, poolID uint64, withdrawalID uint64) error {
	logger := k.Logger(ctx).With("ctx", "withdrawal_transfer")

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

	isLocalToAddress, toAddr, err := pool.AccAddressFromBech32(withdrawal.ToAddress)
	if err != nil {
		return err
	}

	if pool.IsLocalZone(ctx) {
		// Move funds
		if err := k.BankKeeper.SendCoins(ctx,
			sdk.MustAccAddressFromBech32(pool.GetIcaDepositAddress()),
			toAddr,
			sdk.NewCoins(withdrawal.Amount),
		); err != nil {
			// Return with error here and let the caller manage the state changes if needed
			return err
		}
		return k.OnTransferWithdrawalToDestAddrCompleted(ctx, poolID, withdrawalID, false)
	}

	var msgs []sdk.Msg
	var msgLog string
	var marshalledCallbackData []byte
	var callbackID string

	// We start by acquiring the counterparty channel id
	transferChannel, found := k.IBCKeeper.ChannelKeeper.GetChannel(ctx, ibctransfertypes.PortID, pool.GetTransferChannelId())
	if !found {
		return errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "transfer channel %s not found", pool.GetTransferChannelId())
	}
	counterpartyChannelId := transferChannel.Counterparty.ChannelId

	// Converts the local ibc Denom into the native chain Denom
	amount := sdk.NewCoin(pool.NativeDenom, withdrawal.Amount.Amount)

	if isLocalToAddress {
		// ICA transfer from remote zone to local zone
		msgLog = "ICA transfer"
		callbackID = ICACallbackID_TransferFromNative
		callbackData := types.TransferFromNativeCallback{
			Type:         types.TransferType_Withdraw,
			PoolId:       poolID,
			WithdrawalId: withdrawalID,
		}
		marshalledCallbackData, err = k.MarshalTransferFromNativeCallbackArgs(ctx, callbackData)
		if err != nil {
			return err
		}

		// From Remote to Local - use counterparty transfer channel ID
		msgs = append(msgs, ibctransfertypes.NewMsgTransfer(
			ibctransfertypes.PortID,
			counterpartyChannelId,
			amount,
			pool.GetIcaDepositAddress(),
			withdrawal.GetToAddress(),
			clienttypes.Height{},
			uint64(ctx.BlockTime().UnixNano())+types.IBCTimeoutNanos,
			"Cosmos Millions",
		))
	} else {
		// ICA bank send from remote to remote
		msgLog = "ICA bank send"
		callbackID = ICACallbackID_BankSend
		callbackData := types.BankSendCallback{
			PoolId:       poolID,
			WithdrawalId: withdrawalID,
		}
		marshalledCallbackData, err = k.MarshalBankSendCallbackArgs(ctx, callbackData)
		if err != nil {
			return err
		}

		msgs = append(msgs, &banktypes.MsgSend{
			FromAddress: pool.GetIcaDepositAddress(),
			ToAddress:   toAddr.String(),
			Amount:      sdk.NewCoins(amount),
		})
	}

	// Dispatch message
	sequence, err := k.BroadcastICAMessages(ctx, poolID, types.ICATypeDeposit, msgs, types.IBCTimeoutNanos, callbackID, marshalledCallbackData)
	if err != nil {
		// Return with error here and let the caller manage the state changes if needed
		logger.Error(
			fmt.Sprintf("failed to dispatch %s: %v", msgLog, err),
			"pool_id", poolID,
			"withdrawal_id", withdrawalID,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
		return err
	}
	logger.Debug(
		fmt.Sprintf("%s dispatched", msgLog),
		"pool_id", poolID,
		"withdrawal_id", withdrawalID,
		"chain_id", pool.GetChainId(),
		"sequence", sequence,
	)
	return nil
}

// OnTransferWithdrawalToDestAddrCompleted Acknowledge the withdraw IBC transfer
// - To to the local chain response if it's a transfer to local chain
// - To the native chain if it's BankSend for a native pool with a native destination address
func (k Keeper) OnTransferWithdrawalToDestAddrCompleted(ctx sdk.Context, poolID uint64, withdrawalID uint64, isError bool) error {
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
