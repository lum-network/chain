package keeper

import (
	"fmt"
	"strconv"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distribtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	icqueriestypes "github.com/lum-network/chain/x/icqueries/types"

	gogotypes "github.com/gogo/protobuf/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"
	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"

	"github.com/lum-network/chain/x/millions/types"
)

// SetupPoolICA registers the ICA account on the native chain
// - waits for the ICA callback to move to OnSetupPoolICACompleted
// - or go to OnSetupPoolICACompleted directly if local zone
func (k Keeper) SetupPoolICA(ctx sdk.Context, poolID uint64) (*types.Pool, error) {
	// Acquire and deserialize our pool entity
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	if pool.IsLocalZone(ctx) {
		// Skip ICA setup for local pools
		return k.SetupPoolWithdrawalAddress(ctx, poolID)
	}

	// Get the chain ID from the connection
	chainID, err := k.GetChainID(ctx, pool.GetConnectionId())
	if err != nil {
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, "unable to obtain chain id from connection %s, err: %v", pool.GetConnectionId(), err)
	}
	if chainID != pool.ChainId {
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, "provided chain id %s differs from the connection chain id %s", pool.ChainId, chainID)
	}

	// Compute the app version structure for ICA registration
	appVersion, err := k.getPoolAppVersion(ctx, pool)
	if err != nil {
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
	}

	// Register the accounts deposit account first
	// Wait for this account to be setup to register the prize pool account
	// This is done to avoid race conditions for the last setup step (SetWithdrawAddress)
	icaDepositPortName := string(types.NewPoolName(pool.GetPoolId(), types.ICATypeDeposit))
	pool.IcaDepositPortId, err = icatypes.NewControllerPortID(icaDepositPortName)
	if err != nil {
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, fmt.Sprintf("Unable to create deposit account port id, err: %s", err.Error()))
	}
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, pool.GetConnectionId(), icaDepositPortName, appVersion); err != nil {
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, fmt.Sprintf("Unable to trigger deposit account registration, err: %s", err.Error()))
	}

	k.updatePool(ctx, &pool)
	return &pool, nil
}

// OnPoolICASetupCompleted Acknowledge the ICA account creation on the native chain
// then moves to SetupPoolWithdrawalAddress once all ICA accounts have been created
// TODO: error management based on the callback response
func (k Keeper) OnSetupPoolICACompleted(ctx sdk.Context, poolID uint64, icaType string, icaAddress string) (*types.Pool, error) {
	logger := k.Logger(ctx).With("ctx", "pool_on_setup_ica_completed")

	// Grab our local pool instance
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	// Make sure our pool is in state created, otherwise just continue without error
	if pool.GetState() != types.PoolState_Created {
		return &pool, nil
	}

	// If it's a local pool, not further processing required
	if pool.IsLocalZone(ctx) {
		// Ignore this step for local pools
		return k.SetupPoolWithdrawalAddress(ctx, poolID)
	}

	if pool.IcaDepositAddress == "" && icaType == types.ICATypeDeposit && len(icaAddress) > 0 {
		// Assign the ICA deposit address
		pool.IcaDepositAddress = icaAddress

		if pool.IcaPrizepoolPortId == "" {
			// First time registering ICA Deposit
			// Initialize ICA PrizePool
			appVersion, err := k.getPoolAppVersion(ctx, pool)
			if err != nil {
				return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
			}
			icaPrizePoolPortName := string(types.NewPoolName(pool.GetPoolId(), types.ICATypePrizePool))
			pool.IcaPrizepoolPortId, err = icatypes.NewControllerPortID(icaPrizePoolPortName)
			if err != nil {
				return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, fmt.Sprintf("Unable to create prizepool account port id, err: %s", err.Error()))
			}
			if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, pool.GetConnectionId(), icaPrizePoolPortName, appVersion); err != nil {
				logger.Error("Unable to trigger prizepool account registration, err: %s", err.Error())
			}
		}

		// Save pool state
		k.updatePool(ctx, &pool)
	} else if pool.IcaPrizepoolAddress == "" && icaType == types.ICATypePrizePool && len(icaAddress) > 0 {
		// Assign the ICA prize pool address
		pool.IcaPrizepoolAddress = icaAddress
		k.updatePool(ctx, &pool)
	}

	if len(pool.IcaDepositAddress) > 0 && len(pool.IcaPrizepoolAddress) > 0 {
		// Move to next step since both accounts have been created
		return k.SetupPoolWithdrawalAddress(ctx, poolID)
	}

	return &pool, nil
}

// SetupPoolWithdrawalAddress sets the PrizePoolAddress as the Staking withdrawal address for the DepositAddress
// - waits for the ICA callback to move to OnSetupPoolWithdrawalAddressCompleted
// - or go to OnSetupPoolWithdrawalAddressCompleted directly upon setting up the withdrawal address if local zone
func (k Keeper) SetupPoolWithdrawalAddress(ctx sdk.Context, poolID uint64) (*types.Pool, error) {
	logger := k.Logger(ctx).With("ctx", "pool_setup_withdrawal")

	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	if pool.GetState() != types.PoolState_Created {
		return &pool, nil
	}

	if pool.IsLocalZone(ctx) {
		if err := k.DistributionKeeper.SetWithdrawAddr(
			ctx,
			sdk.MustAccAddressFromBech32(pool.IcaDepositAddress),
			sdk.MustAccAddressFromBech32(pool.IcaPrizepoolAddress),
		); err != nil {
			logger.Error(
				fmt.Sprintf("failed to dispatch set withdrawal address for local pool: %v", err),
				"pool_id", poolID,
				"chain_id", pool.GetChainId(),
			)
			return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
		}
		return k.OnSetupPoolWithdrawalAddressCompleted(ctx, poolID)
	}

	callbackData := types.SetWithdrawAddressCallback{
		PoolId: poolID,
	}
	marshalledCallbackData, err := k.MarshalSetWithdrawAddressCallbackArgs(ctx, callbackData)
	if err != nil {
		return &pool, err
	}
	msgs := []sdk.Msg{&distribtypes.MsgSetWithdrawAddress{
		DelegatorAddress: pool.GetIcaDepositAddress(),
		WithdrawAddress:  pool.GetIcaPrizepoolAddress(),
	}}
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + types.IBCTransferTimeoutNanos
	sequence, err := k.BroadcastICAMessages(ctx, poolID, types.ICATypeDeposit, msgs, timeoutTimestamp, ICACallbackID_SetWithdrawAddress, marshalledCallbackData)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to dispatch ICA set withdraw address: %v", err),
			"pool_id", poolID,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
		return &pool, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
	} else {
		logger.Debug(
			"ICA set withdraw address dispatched",
			"pool_id", poolID,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
	}

	return &pool, nil
}

// OnSetupPoolWithdrawalAddressCompleted Acknowledge the withdrawal address configuration on the native chain
// then sets the pool to status ready in case of success
func (k Keeper) OnSetupPoolWithdrawalAddressCompleted(ctx sdk.Context, poolID uint64) (*types.Pool, error) {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	if pool.GetState() != types.PoolState_Created {
		return &pool, nil
	}
	pool.State = types.PoolState_Ready
	k.updatePool(ctx, &pool)
	return &pool, nil
}

// GetNextPoolID Return the next pool id to be used
func (k Keeper) GetNextPoolID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextPoolId := gogotypes.UInt64Value{}

	b := store.Get(types.NextPoolIdPrefix)
	if b == nil {
		panic(fmt.Errorf("getting at key (%v) should not have been nil", types.NextPoolIdPrefix))
	}
	k.cdc.MustUnmarshal(b, &nextPoolId)
	return nextPoolId.GetValue()
}

// SetNextPoolID sets next pool ID
func (k Keeper) SetNextPoolID(ctx sdk.Context, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: poolId})
	store.Set(types.NextPoolIdPrefix, bz)
}

func (k Keeper) GetNextPoolIDAndIncrement(ctx sdk.Context) uint64 {
	nextPoolId := k.GetNextPoolID(ctx)
	k.SetNextPoolID(ctx, nextPoolId+1)
	return nextPoolId
}

// HasPool Returns a boolean that indicates if the given poolID exists in the KVStore or not
func (k Keeper) HasPool(ctx sdk.Context, poolID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetPoolKey(poolID))
}

// AddPool Set a pool structure in the KVStore for a given pool ID
func (k Keeper) AddPool(ctx sdk.Context, pool *types.Pool) {
	// Automatically affect ID if missing
	if pool.GetPoolId() == types.UnknownID {
		pool.PoolId = k.GetNextPoolIDAndIncrement(ctx)
	}
	// Ensure payload is valid
	if err := pool.ValidateBasic(k.GetParams(ctx)); err != nil {
		panic(err)
	}
	// Ensure we never override an existing entity
	if k.HasPool(ctx, pool.GetPoolId()) {
		panic(errorsmod.Wrapf(types.ErrEntityOverride, "ID %d", pool.GetPoolId()))
	}
	store := ctx.KVStore(k.storeKey)
	encodedPool := k.cdc.MustMarshal(pool)
	store.Set(types.GetPoolKey(pool.GetPoolId()), encodedPool)
}

// RegisterPool Register a given pool from the transaction message
func (k Keeper) RegisterPool(
	ctx sdk.Context,
	denom, nativeDenom, chainId, connectionId, transferChannelId string,
	vals []string,
	bech32Acc, bech32Val string,
	minDepositAmount math.Int,
	drawSchedule types.DrawSchedule,
	prizeStrategy types.PrizeStrategy,
) (uint64, error) {

	// Acquire new pool ID
	poolID := k.GetNextPoolIDAndIncrement(ctx)

	// Initialize validators
	var validators []types.PoolValidator
	for _, addr := range vals {
		validators = append(validators, types.PoolValidator{
			OperatorAddress: addr,
			IsEnabled:       true,
			BondedAmount:    sdk.ZeroInt(),
		})
	}

	// Initialize our local deposit address
	localAddress := types.NewPoolAddress(poolID, types.ICATypeDeposit)
	poolAccount := k.AccountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(authtypes.NewBaseAccountWithAddress(localAddress), localAddress.String()))
	k.AccountKeeper.SetAccount(ctx, poolAccount)

	// Prepare new pool
	var pool = types.Pool{
		PoolId:              poolID,
		Denom:               denom,
		NativeDenom:         nativeDenom,
		ChainId:             chainId,
		ConnectionId:        connectionId,
		Validators:          validators,
		Bech32PrefixAccAddr: bech32Acc,
		Bech32PrefixValAddr: bech32Val,
		MinDepositAmount:    minDepositAmount,
		DrawSchedule:        drawSchedule.Sanitized(),
		PrizeStrategy:       prizeStrategy,
		LocalAddress:        localAddress.String(),
		NextDrawId:          1,
		TvlAmount:           sdk.ZeroInt(),
		DepositorsCount:     0,
		SponsorshipAmount:   sdk.ZeroInt(),
		AvailablePrizePool:  sdk.NewCoin(denom, sdk.ZeroInt()),
		State:               types.PoolState_Created,
		TransferChannelId:   transferChannelId,
		CreatedAtHeight:     ctx.BlockHeight(),
		UpdatedAtHeight:     ctx.BlockHeight(),
		CreatedAt:           ctx.BlockTime(),
		UpdatedAt:           ctx.BlockTime(),
	}

	// Validate pool configuration
	if err := pool.ValidateBasic(k.GetParams(ctx)); err != nil {
		return 0, errorsmod.Wrapf(types.ErrFailedToRegisterPool, err.Error())
	}

	// If it's a local zone, we have more setup steps for module accounts
	if pool.IsLocalZone(ctx) {
		// Set the deposit address to the local module account address
		pool.IcaDepositAddress = localAddress.String()

		// Initialize our local prizepool address
		icaPrizePoolAddress := types.NewPoolAddress(poolID, types.ICATypePrizePool)
		icaPrizePoolAccount := k.AccountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(authtypes.NewBaseAccountWithAddress(icaPrizePoolAddress), icaPrizePoolAddress.String()))
		k.AccountKeeper.SetAccount(ctx, icaPrizePoolAccount)

		pool.IcaPrizepoolAddress = icaPrizePoolAddress.String()
	}

	// Commit the pool to the KVStore
	k.AddPool(ctx, &pool)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeRegisterPool,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(pool.PoolId, 10)),
		),
	})

	// Remote zone pool
	_, err := k.SetupPoolICA(ctx, poolID)
	if err != nil {
		return poolID, err
	}

	return poolID, nil
}

// UpdatePool Update the updatable properties of a pool from the transaction message
func (k Keeper) UpdatePool(
	ctx sdk.Context,
	poolID uint64,
	vals []string,
	minDepositAmount *math.Int,
	drawSchedule *types.DrawSchedule,
	prizeStrategy *types.PrizeStrategy,
) error {
	// Acquire and deserialize our pool entity
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	// Update enabled validators
	if len(vals) > 0 {
		for i := range pool.Validators {
			pool.Validators[i].IsEnabled = false
		}
		valIdx := pool.GetValidatorsMapIndex()
		for _, addr := range vals {
			if _, exists := valIdx[addr]; exists {
				pool.Validators[valIdx[addr]].IsEnabled = true
			} else {
				pool.Validators = append(pool.Validators, types.PoolValidator{
					OperatorAddress: addr,
					IsEnabled:       true,
					BondedAmount:    sdk.ZeroInt(),
				})
				valIdx[addr] = len(pool.Validators) - 1
			}
		}
	}

	// Only a few properties can be updated
	if minDepositAmount != nil {
		pool.MinDepositAmount = *minDepositAmount
	}
	if drawSchedule != nil {
		pool.DrawSchedule = *drawSchedule
		if pool.DrawSchedule.InitialDrawAt.After(ctx.BlockTime()) {
			// Specifying a new valid InitialDrawAt resets the Pool draw timing to this date
			// Also useful for governance to change the timing of Draws in case of time drift
			pool.LastDrawCreatedAt = nil
		}
	}
	if prizeStrategy != nil {
		pool.PrizeStrategy = *prizeStrategy
	}

	// Validate pool configuration
	if err := pool.ValidateBasic(k.GetParams(ctx)); err != nil {
		return errorsmod.Wrapf(types.ErrFailedToUpdatePool, err.Error())
	}

	// Commit the pool to the KVStore
	k.updatePool(ctx, &pool)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeUpdatePool,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(pool.PoolId, 10)),
		),
	})

	return nil
}

// UnsafeKillPool This method switches the provided pool state but does not handle any withdrawal or deposit.
// It shouldn't be used and is very specific to UNUSED and EMPTY pools
func (k Keeper) UnsafeKillPool(ctx sdk.Context, poolID uint64) (types.Pool, error) {
	// Grab our pool instance
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return types.Pool{}, err
	}

	// Make sure the pool isn't killed yet
	if pool.GetState() == types.PoolState_Killed {
		return pool, errorsmod.Wrapf(types.ErrPoolKilled, "%d", poolID)
	}

	// Kill the pool
	pool.State = types.PoolState_Killed
	k.updatePool(ctx, &pool)
	return pool, nil
}

func (k Keeper) updatePool(ctx sdk.Context, pool *types.Pool) {
	pool.UpdatedAt = ctx.BlockTime()
	pool.UpdatedAtHeight = ctx.BlockHeight()
	// Ensure payload is valid
	if err := pool.ValidateBasic(k.GetParams(ctx)); err != nil {
		panic(err)
	}
	store := ctx.KVStore(k.storeKey)
	encodedPool := k.cdc.MustMarshal(pool)
	store.Set(types.GetPoolKey(pool.GetPoolId()), encodedPool)
}

// GetPool Returns a pool instance for the given poolID
func (k Keeper) GetPool(ctx sdk.Context, poolID uint64) (types.Pool, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolKey(poolID))
	if bz == nil {
		return types.Pool{}, errorsmod.Wrapf(types.ErrPoolNotFound, "%d", poolID)
	}

	var pool types.Pool
	if err := k.cdc.Unmarshal(bz, &pool); err != nil {
		return types.Pool{}, err
	}

	return pool, nil
}

func (k Keeper) GetPoolForChainID(ctx sdk.Context, chainID string) (types.Pool, bool) {
	var pool = types.Pool{}
	found := false
	k.IteratePools(ctx, func(p types.Pool) bool {
		if p.GetChainId() == chainID {
			pool = p
			found = true
			return true
		}
		return false
	})

	return pool, found
}

func (k Keeper) GetPoolForConnectionID(ctx sdk.Context, connectionID string) (types.Pool, bool) {
	var pool = types.Pool{}
	found := false
	k.IteratePools(ctx, func(p types.Pool) bool {
		if p.GetConnectionId() == connectionID {
			pool = p
			found = true
			return true
		}
		return false
	})

	return pool, found
}

func (k Keeper) GetPoolForControllerPortID(ctx sdk.Context, controllerPortID string) (types.Pool, bool) {
	var pool = types.Pool{}
	found := false
	k.IteratePools(ctx, func(p types.Pool) bool {
		if p.GetIcaDepositPortId() == controllerPortID || p.GetIcaPrizepoolPortId() == controllerPortID {
			pool = p
			found = true
			return true
		}
		return false
	})

	return pool, found
}

// IteratePools Iterate over the pools, and for each entry call the callback
func (k Keeper) IteratePools(ctx sdk.Context, callback func(pool types.Pool) (stop bool)) {
	iterator := k.PoolsIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var pool types.Pool
		k.cdc.MustUnmarshal(iterator.Value(), &pool)
		if callback(pool) {
			break
		}
	}
}

// PoolsIterator Return a ready to use iterator for the pools store
func (k Keeper) PoolsIterator(ctx sdk.Context) sdk.Iterator {
	kvStore := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(kvStore, types.PoolPrefix)
}

// ListPools Return the pools
func (k Keeper) ListPools(ctx sdk.Context) (pools []types.Pool) {
	k.IteratePools(ctx, func(pool types.Pool) bool {
		pools = append(pools, pool)
		return false
	})
	return pools
}

// ListPoolsToDraw Returns the pools which should be launching a Draw
func (k Keeper) ListPoolsToDraw(ctx sdk.Context) (pools []types.Pool) {
	allPools := k.ListPools(ctx)
	for _, p := range allPools {
		if p.ShouldDraw(ctx) {
			pools = append(pools, p)
		}
	}
	return pools
}

// TransferAmountFromPoolToNativeChain Transfer a given amount to the native chain ICA account from the local module account
// amount denom must be based on pool.Denom
func (k Keeper) TransferAmountFromPoolToNativeChain(ctx sdk.Context, poolID uint64, amount sdk.Coin) (*ibctransfertypes.MsgTransfer, *ibctransfertypes.MsgTransferResponse, error) {
	// Acquire our pool instance
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, nil, err
	}

	// Pool must be ready to process those kind of operations
	if pool.GetState() != types.PoolState_Ready {
		return nil, nil, types.ErrPoolNotReady
	}

	// Timeout is now plus 30 minutes in nanoseconds
	// We use the standard transfer port ID and not the one opened for ICA
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + types.IBCTransferTimeoutNanos
	// From Local to Remote - use transfer channel ID
	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		pool.GetTransferChannelId(),
		amount,
		pool.GetLocalAddress(),
		pool.GetIcaDepositAddress(),
		clienttypes.Height{},
		timeoutTimestamp,
	)

	// Broadcast the transfer
	msgResponse, err := k.IBCTransferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, nil, err
	}

	return msg, msgResponse, nil
}

func (k Keeper) BroadcastICAMessages(ctx sdk.Context, poolID uint64, accountType string, msgs []sdk.Msg, timeoutTimestamp uint64, callbackId string, callbackArgs []byte) (uint64, error) {
	// Acquire our pool instance
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return 0, err
	}

	// Compute the port ID
	var portID string
	if accountType == types.ICATypeDeposit {
		portID = pool.GetIcaDepositPortId()
	} else if accountType == types.ICATypePrizePool {
		portID = pool.GetIcaPrizepoolPortId()
	}

	// Acquire the channel capacities
	channelID, found := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, pool.GetConnectionId(), portID)
	if !found {
		return 0, errorsmod.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve open active channel for port %s", portID)
	}
	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		return 0, errorsmod.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	// Serialize the data and construct the packet to send
	data, err := icatypes.SerializeCosmosTx(k.cdc, msgs)
	if err != nil {
		return 0, err
	}
	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: "Cosmos Millions ICA",
	}

	// Broadcast the messages
	sequence, err := k.ICAControllerKeeper.SendTx(ctx, chanCap, pool.GetConnectionId(), portID, packetData, timeoutTimestamp)
	if err != nil {
		return 0, err
	}

	// Store the callback data
	if callbackId != "" && callbackArgs != nil {
		callback := icacallbackstypes.CallbackData{
			CallbackKey:  icacallbackstypes.PacketID(portID, channelID, sequence),
			PortId:       portID,
			ChannelId:    channelID,
			Sequence:     sequence,
			CallbackId:   callbackId,
			CallbackArgs: callbackArgs,
		}
		k.ICACallbacksKeeper.SetCallbackData(ctx, callback)
	}

	k.Logger(ctx).Debug(fmt.Sprintf("Broadcasted ICA messages with sequence %d", sequence))
	return sequence, nil
}

func (k Keeper) QueryBalance(ctx sdk.Context, poolID uint64, drawID uint64) (*types.Draw, error) {
	logger := k.Logger(ctx).With("ctx", "pool_query_balance")

	// Acquire our pool instance
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	// Pool must be ready to process those kind of operations
	if pool.GetState() != types.PoolState_Ready {
		return nil, types.ErrPoolNotReady
	}

	draw, err := k.GetPoolDraw(ctx, poolID, drawID)
	if err != nil {
		return nil, err
	}

	// If it's a local pool, proceed with local balance fetch and synchronously return
	if pool.IsLocalZone(ctx) {
		moduleAccAddress := sdk.MustAccAddressFromBech32(pool.GetIcaPrizepoolAddress())
		balance := k.BankKeeper.GetBalance(ctx, moduleAccAddress, pool.GetNativeDenom())
		return k.OnQueryRewardsOnNativeChainCompleted(ctx, poolID, drawID, sdk.NewCoins(balance), false)
	}

	// Encode the ica address for query
	_, icaAddressBz, err := bech32.DecodeAndConvert(pool.GetIcaPrizepoolAddress())
	if err != nil {
		panic(err)
	}

	// Construct the query data and timeout timestamp (now + 30 minutes)
	queryData := append(banktypes.CreateAccountBalancesPrefix(icaAddressBz), []byte(pool.GetNativeDenom())...)
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + types.IBCTransferTimeoutNanos

	// Submit the ICQ
	extraId := types.CombineStringKeys(strconv.FormatUint(poolID, 10), strconv.FormatUint(drawID, 10))
	err = k.ICQueriesKeeper.MakeRequest(ctx, types.ModuleName, ICQCallbackID_Balance, pool.GetChainId(), pool.GetConnectionId(), extraId, icqueriestypes.BANK_STORE_QUERY_WITH_PROOF, queryData, timeoutTimestamp)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to dispatch icq query to fetch prize pool balance: %v", err),
			"pool_id", poolID,
			"draw_id", drawID,
		)
		return k.OnQueryRewardsOnNativeChainCompleted(ctx, poolID, drawID, sdk.NewCoins(), true)
	}

	return &draw, nil
}

// getPoolAppVersion returns the ICA app version for the pool connection
func (k Keeper) getPoolAppVersion(ctx sdk.Context, pool types.Pool) (string, error) {
	connectionEnd, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, pool.GetConnectionId())
	if !found {
		return "", fmt.Errorf("connection with id %s not found", pool.GetConnectionId())
	}
	return string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: pool.GetConnectionId(),
		HostConnectionId:       connectionEnd.Counterparty.GetConnectionID(),
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	})), nil
}
