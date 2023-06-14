package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gogotypes "github.com/gogo/protobuf/types"
	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	"github.com/lum-network/chain/x/millions/types"
)

// TransferDepositToNativeChain Transfer a deposit to a native chain
// - wait for the ICA callback to move to OnTransferDepositToNativeChainCompleted
// - or go to OnTransferDepositToNativeChainCompleted directly if local zone.
func (k Keeper) TransferDepositToNativeChain(ctx sdk.Context, poolID, depositID uint64) error {
	logger := k.Logger(ctx).With("ctx", "deposit_transfer")

	// Acquire pool config
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}
	if pool.IsLocalZone(ctx) {
		return k.OnTransferDepositToNativeChainCompleted(ctx, poolID, depositID, false)
	}

	// Acquire the deposit
	deposit, err := k.GetPoolDeposit(ctx, poolID, depositID)
	if err != nil {
		return err
	}
	if deposit.State != types.DepositState_IbcTransfer {
		return errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DepositState_IbcTransfer.String(), deposit.State.String())
	}

	// Construct our callback
	transferToNativeCallback := types.TransferToNativeCallback{
		PoolId:    poolID,
		DepositId: depositID,
	}
	marshalledCallbackData, err := k.MarshalTransferToNativeCallbackArgs(ctx, transferToNativeCallback)
	if err != nil {
		return err
	}

	// Request funds transfer
	msg, msgResponse, err := k.TransferAmountFromPoolToNativeChain(ctx, poolID, deposit.GetAmount())
	if err != nil {
		// Return with error here since it is the first operation and nothing needs to be saved to state
		logger.Error(
			fmt.Sprintf("failed to dispatch IBC transfer: %v", err),
			"pool_id", poolID,
			"deposit_id", depositID,
			"chain_id", pool.GetChainId(),
			"sequence", msgResponse.GetSequence(),
		)
		return err
	}
	logger.Debug(
		"IBC transfer dispatched",
		"pool_id", poolID,
		"deposit_id", depositID,
		"chain_id", pool.GetChainId(),
		"sequence", msgResponse.GetSequence(),
	)

	// Store the callback data
	callback := icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(msg.SourcePort, msg.SourceChannel, msgResponse.GetSequence()),
		PortId:       msg.SourcePort,
		ChannelId:    msg.SourceChannel,
		Sequence:     msgResponse.GetSequence(),
		CallbackId:   ICACallbackID_TransferToNative,
		CallbackArgs: marshalledCallbackData,
	}
	k.ICACallbacksKeeper.SetCallbackData(ctx, callback)

	return nil
}

// OnTransferDepositToNativeChainCompleted Acknowledge the IBC transfer to the native chain response
// then moves to DelegateDepositOnNativeChain in case of success.
func (k Keeper) OnTransferDepositToNativeChainCompleted(ctx sdk.Context, poolID, depositID uint64, isError bool) error {
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
	return k.DelegateDepositOnNativeChain(ctx, poolID, depositID)
}

// DelegateDepositOnNativeChain Delegates a deposit to the native chain validators
// - wait for the ICA callback to move to OnDelegateDepositOnNativeChainCompleted
// - or go to OnDelegateDepositOnNativeChainCompleted directly if local zone.
func (k Keeper) DelegateDepositOnNativeChain(ctx sdk.Context, poolID, depositID uint64) error {
	logger := k.Logger(ctx).With("ctx", "deposit_delegate")

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

	// Prepare delegations split
	splits := pool.ComputeSplitDelegations(ctx, deposit.GetAmount().Amount)
	if len(splits) == 0 {
		return types.ErrPoolEmptySplitDelegations
	}

	// If pool is local, we just process operation in place
	// Otherwise we trigger IBC / ICA transactions
	if pool.IsLocalZone(ctx) {
		for _, split := range splits {
			valAddr, err := sdk.ValAddressFromBech32(split.ValidatorAddress)
			if err != nil {
				return err
			}
			validator, found := k.StakingKeeper.GetValidator(ctx, valAddr)
			if !found {
				return errorsmod.Wrapf(stakingtypes.ErrNoValidatorFound, "%s", valAddr.String())
			}
			if _, err := k.StakingKeeper.Delegate(
				ctx,
				sdk.MustAccAddressFromBech32(pool.GetIcaDepositAddress()),
				split.Amount,
				stakingtypes.Unbonded,
				validator,
				true,
			); err != nil {
				return errorsmod.Wrapf(err, "%s", valAddr.String())
			}
		}
		return k.OnDelegateDepositOnNativeChainCompleted(ctx, poolID, depositID, splits, false)
	}

	// Construct our callback data
	callbackData := types.DelegateCallback{
		PoolId:           poolID,
		DepositId:        depositID,
		SplitDelegations: splits,
	}
	marshalledCallbackData, err := k.MarshalDelegateCallbackArgs(ctx, callbackData)
	if err != nil {
		return err
	}

	// Build delegation tx
	var msgs []sdk.Msg
	for _, split := range splits {
		msgs = append(msgs, &stakingtypes.MsgDelegate{
			DelegatorAddress: pool.GetIcaDepositAddress(),
			ValidatorAddress: split.ValidatorAddress,
			Amount:           sdk.NewCoin(pool.NativeDenom, split.Amount),
		})
	}

	// Dispatch our message with a timeout of 30 minutes in nanos
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + types.IBCTransferTimeoutNanos
	sequence, err := k.BroadcastICAMessages(ctx, poolID, types.ICATypeDeposit, msgs, timeoutTimestamp, ICACallbackID_Delegate, marshalledCallbackData)
	if err != nil {
		// Save error state since we cannot simply recover from a failure at this stage
		// A subsequent call to DepositRetry will be made possible by setting an error state and not returning an error here
		logger.Error(
			fmt.Sprintf("failed to dispatch ICA delegation: %v", err),
			"pool_id", poolID,
			"deposit_id", depositID,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
		return k.OnDelegateDepositOnNativeChainCompleted(ctx, poolID, depositID, splits, true)
	}
	logger.Debug(
		"ICA delegation dispatched",
		"pool_id", poolID,
		"deposit_id", depositID,
		"chain_id", pool.GetChainId(),
		"sequence", sequence,
	)
	return nil
}

// OnDelegateDepositOnNativeChainCompleted Acknowledge the ICA delegate to the native chain validators response.
func (k Keeper) OnDelegateDepositOnNativeChainCompleted(ctx sdk.Context, poolID, depositID uint64, splits []*types.SplitDelegation, isError bool) error {
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

// GetNextDepositID gets the next deposit ID.
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

// GetNextDepositIdAndIncrement gets the next deposit ID and store the incremented ID.
func (k Keeper) GetNextDepositIdAndIncrement(ctx sdk.Context) uint64 {
	nextDepositId := k.GetNextDepositID(ctx)
	k.SetNextDepositID(ctx, nextDepositId+1)
	return nextDepositId
}

// SetNextDepositID sets next deposit ID.
func (k Keeper) SetNextDepositID(ctx sdk.Context, depositId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: depositId})
	store.Set(types.NextDepositPrefix, bz)
}

// GetPoolDeposit returns a deposit by ID for a given poolID.
func (k Keeper) GetPoolDeposit(ctx sdk.Context, poolID, depositID uint64) (types.Deposit, error) {
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
// - adds it to the account {address, pool_id, deposit_id} deposits.
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
// - removes it from the {address, pool_id, deposit_id} deposits.
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

// UpdateDepositStatus Update a given deposit status by its ID.
func (k Keeper) UpdateDepositStatus(ctx sdk.Context, poolID, depositID uint64, status types.DepositState, isError bool) {
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
	k.setAccountDeposit(ctx, &deposit)
	k.setPoolDeposit(ctx, &deposit)
}

// setPoolDeposit sets a deposit to the pool deposit key.
func (k Keeper) setPoolDeposit(ctx sdk.Context, deposit *types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPoolDepositKey(deposit.PoolId, deposit.DepositId)
	encodedDeposit := k.cdc.MustMarshal(deposit)
	store.Set(key, encodedDeposit)
}

// setAccountDeposit set the deposit for the Account deposit key.
func (k Keeper) setAccountDeposit(ctx sdk.Context, deposit *types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	depositorAddress := sdk.MustAccAddressFromBech32(deposit.DepositorAddress)
	key := types.GetAccountPoolDepositKey(depositorAddress, deposit.PoolId, deposit.DepositId)
	encodedDeposit := k.cdc.MustMarshal(deposit)
	store.Set(key, encodedDeposit)
}

// hasPoolDeposit returns whether or not an account has at least one deposit for a given pool.
func (k Keeper) hasPoolDeposit(ctx sdk.Context, address string, poolID uint64) bool {
	addr := sdk.MustAccAddressFromBech32(address)

	kvStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(kvStore, types.GetAccountPoolDepositsKey(addr, poolID))

	defer iterator.Close()
	return iterator.Valid()
}

// ListPoolDeposits returns all deposits for a given poolID
// Warning: expensive operation.
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
// Warning: expensive operation.
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
// Warning: expensive operation.
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
// Warning: expensive operation.
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
