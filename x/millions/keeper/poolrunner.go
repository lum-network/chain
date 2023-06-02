package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

// PoolRunner interface to implement for all Pool Runners
// runners are responsible for all coins operations (local, ICA, IBC)
// runners are NOT responsible and should never update any entity state
type PoolRunner interface {
	String() string
	Logger(ctx sdk.Context) log.Logger
	// SendDepositToPool sends the deposit amount to a local Pool owned address
	SendDepositToPool(ctx sdk.Context, pool types.Pool, deposit types.Deposit) error
	// TransferDepositToRemoteZone transfers the deposit amount from a local Pool owned address to a remote Pool owned address
	TransferDepositToRemoteZone(ctx sdk.Context, pool types.Pool, deposit types.Deposit) error
	// DelegateDepositOnRemoteZone launches an ICA action on the remote Pool owned address (such as delegate coins for native staking Pools)
	DelegateDepositOnRemoteZone(ctx sdk.Context, pool types.Pool, deposit types.Deposit) ([]*types.SplitDelegation, error)
}

// PoolRunnerBase common base implementation for Pool Runners
type PoolRunnerBase struct {
	keeper *Keeper
}

func (runner *PoolRunnerBase) String() string {
	return "PoolRunnerBase"
}

func (runner *PoolRunnerBase) Logger(ctx sdk.Context) log.Logger {
	return runner.keeper.Logger(ctx).With("pool_runner", runner.String())
}

func (runner *PoolRunnerBase) SendDepositToPool(ctx sdk.Context, pool types.Pool, deposit types.Deposit) error {
	if pool.IsLocalZone(ctx) {
		// Directly send funds to the local deposit address in case of local pool
		return runner.keeper.BankKeeper.SendCoins(
			ctx,
			sdk.MustAccAddressFromBech32(deposit.DepositorAddress),
			sdk.MustAccAddressFromBech32(pool.IcaDepositAddress),
			sdk.NewCoins(deposit.Amount),
		)
	}
	return runner.keeper.BankKeeper.SendCoins(
		ctx,
		sdk.MustAccAddressFromBech32(deposit.DepositorAddress),
		sdk.MustAccAddressFromBech32(pool.LocalAddress),
		sdk.NewCoins(deposit.Amount),
	)
}

func (runner *PoolRunnerBase) TransferDepositToRemoteZone(ctx sdk.Context, pool types.Pool, deposit types.Deposit) error {
	logger := runner.Logger(ctx).With("ctx", "deposit_transfer")

	if pool.IsLocalZone(ctx) {
		// no-op
		return nil
	}

	// Construct our callback
	transferToNativeCallback := types.TransferToNativeCallback{
		PoolId:    pool.PoolId,
		DepositId: deposit.DepositId,
	}
	marshalledCallbackData, err := runner.keeper.MarshalTransferToNativeCallbackArgs(ctx, transferToNativeCallback)
	if err != nil {
		return err
	}

	// Dispatch our transfer with a timeout of 30 minutes in nanos
	sequence, err := runner.keeper.BroadcastIBCTransfer(ctx, pool, deposit.Amount, types.IBCTimeoutNanos, ICACallbackID_TransferToNative, marshalledCallbackData)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to dispatch IBC transfer: %v", err),
			"pool_id", pool.PoolId,
			"deposit_id", deposit.DepositId,
			"chain_id", pool.ChainId,
			"sequence", sequence,
		)
		return err
	}
	logger.Debug(
		"IBC transfer dispatched",
		"pool_id", pool.PoolId,
		"deposit_id", deposit.DepositId,
		"chain_id", pool.ChainId,
		"sequence", sequence,
	)
	return nil
}
