package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/lum-network/chain/x/millions/types"
)

// PoolRunner interface to implement for all Pool Runners
// runners are responsible for all coins operations (local, ICA, IBC)
// runners are NOT responsible and should never update any entity state
type PoolRunner interface {
	String() string
	Logger(ctx sdk.Context) log.Logger
	// OnUpdatePool triggered upon pool update proposal
	OnUpdatePool(ctx sdk.Context, pool types.Pool) error
	// SendDepositToPool sends the deposit amount to a local Pool owned address
	SendDepositToPool(ctx sdk.Context, pool types.Pool, deposit types.Deposit) error
	// TransferDepositToRemoteZone transfers the deposit amount from a local Pool owned address to a remote Pool owned address
	TransferDepositToRemoteZone(ctx sdk.Context, pool types.Pool, deposit types.Deposit) error
	// DelegateDepositOnRemoteZone launches an ICA action on the remote Pool owned address (such as delegate coins for native staking Pools)
	DelegateDepositOnRemoteZone(ctx sdk.Context, pool types.Pool, deposit types.Deposit) ([]*types.SplitDelegation, error)
	// UndelegateWithdrawalOnRemoteZone launches and ICA action on the remote Pool owned address (such as undelegate coins for native staking Pools)
	UndelegateWithdrawalOnRemoteZone(ctx sdk.Context, pool types.Pool, withdrawal types.Withdrawal) ([]*types.SplitDelegation, *time.Time, error)
	// TransferWithdrawalToRecipient transfers the withdrawal amount from a remove Pool owned address to a user owned address
	TransferWithdrawalToRecipient(ctx sdk.Context, pool types.Pool, withdrawal types.Withdrawal) error
	// ClaimYieldOnRemoteZone launches an ICA action on the remote Pool owned address (such as claim rewards for native staking Pools)
	ClaimYieldOnRemoteZone(ctx sdk.Context, pool types.Pool) error
	// QueryPrizePoolOnRemoteZone launches an ICA action on the remote Pool owned address (such as query balance for native staking Pools)
	QueryFreshPrizePoolCoinsOnRemoteZone(ctx sdk.Context, pool types.Pool) error
	// TransferPrizePoolToLocalZone launches an IBC transfer (via ICA) from a remote Pool owned address to a local Pool owned address
	TransferFreshPrizePoolCoinsToLocalZone(ctx sdk.Context, pool types.Pool, draw types.Draw) error
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

func (runner *PoolRunnerBase) TransferWithdrawalToRecipient(ctx sdk.Context, pool types.Pool, withdrawal types.Withdrawal) error {
	logger := runner.Logger(ctx).With("ctx", "withdrawal_transfer")

	isLocalToAddress, toAddr, err := pool.AccAddressFromBech32(withdrawal.ToAddress)
	if err != nil {
		return err
	}

	if pool.IsLocalZone(ctx) {
		// Move funds
		if err := runner.keeper.BankKeeper.SendCoins(ctx,
			sdk.MustAccAddressFromBech32(pool.GetIcaDepositAddress()),
			toAddr,
			sdk.NewCoins(withdrawal.Amount),
		); err != nil {
			return err
		}
		return nil
	}

	var msgs []sdk.Msg
	var msgLog string
	var marshalledCallbackData []byte
	var callbackID string

	// We start by acquiring the counterparty channel id
	transferChannel, found := runner.keeper.IBCKeeper.ChannelKeeper.GetChannel(ctx, ibctransfertypes.PortID, pool.GetTransferChannelId())
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
			PoolId:       pool.PoolId,
			WithdrawalId: withdrawal.WithdrawalId,
		}
		marshalledCallbackData, err = runner.keeper.MarshalTransferFromNativeCallbackArgs(ctx, callbackData)
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
			PoolId:       pool.PoolId,
			WithdrawalId: withdrawal.WithdrawalId,
		}
		marshalledCallbackData, err = runner.keeper.MarshalBankSendCallbackArgs(ctx, callbackData)
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
	sequence, err := runner.keeper.BroadcastICAMessages(ctx, pool, types.ICATypeDeposit, msgs, types.IBCTimeoutNanos, callbackID, marshalledCallbackData)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to dispatch %s: %v", msgLog, err),
			"pool_id", pool.PoolId,
			"withdrawal_id", withdrawal.WithdrawalId,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
		return err
	}
	logger.Debug(
		fmt.Sprintf("%s dispatched", msgLog),
		"pool_id", pool.PoolId,
		"withdrawal_id", withdrawal.WithdrawalId,
		"chain_id", pool.GetChainId(),
		"sequence", sequence,
	)
	return nil
}

func (runner *PoolRunnerBase) TransferFreshPrizePoolCoinsToLocalZone(ctx sdk.Context, pool types.Pool, draw types.Draw) error {
	// logger := runner.Logger(ctx).With("ctx", "fresh_prizepool_transfer")
	// TODO: implementation
	return fmt.Errorf("not implemented yet")
}
