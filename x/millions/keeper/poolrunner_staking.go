package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/lum-network/chain/x/millions/types"
)

type PoolRunnerStaking struct {
	PoolRunnerBase
}

func (runner *PoolRunnerStaking) String() string {
	return "PoolRunnerStaking"
}

func (runner *PoolRunnerStaking) OnUpdatePool(ctx sdk.Context, pool types.Pool) error {
	if pool.State == types.PoolState_Ready || pool.State == types.PoolState_Paused {
		if err := runner.keeper.RebalanceValidatorsBondings(ctx, pool.PoolId); err != nil {
			return err
		}
	}
	return nil
}

func (runner *PoolRunnerStaking) DelegateDepositOnRemoteZone(ctx sdk.Context, pool types.Pool, deposit types.Deposit) ([]*types.SplitDelegation, error) {
	logger := runner.Logger(ctx).With("ctx", "deposit_delegate")

	// Prepare delegations split
	splits := pool.ComputeSplitDelegations(ctx, deposit.Amount.Amount)
	if len(splits) == 0 {
		return splits, types.ErrPoolEmptySplitDelegations
	}

	// If pool is local, we just process operation in place
	// Otherwise we trigger IBC / ICA transactions
	if pool.IsLocalZone(ctx) {
		for _, split := range splits {
			valAddr, err := sdk.ValAddressFromBech32(split.ValidatorAddress)
			if err != nil {
				return splits, err
			}
			validator, found := runner.keeper.StakingKeeper.GetValidator(ctx, valAddr)
			if !found {
				return splits, errorsmod.Wrapf(stakingtypes.ErrNoValidatorFound, "%s", valAddr.String())
			}
			if _, err := runner.keeper.StakingKeeper.Delegate(
				ctx,
				sdk.MustAccAddressFromBech32(pool.IcaDepositAddress),
				split.Amount,
				stakingtypes.Unbonded,
				validator,
				true,
			); err != nil {
				return splits, errorsmod.Wrapf(err, "%s", valAddr.String())
			}
		}
		return splits, nil
	}

	// Construct our callback data
	callbackData := types.DelegateCallback{
		PoolId:           pool.PoolId,
		DepositId:        deposit.DepositId,
		SplitDelegations: splits,
	}
	marshalledCallbackData, err := runner.keeper.MarshalDelegateCallbackArgs(ctx, callbackData)
	if err != nil {
		return splits, err
	}

	// Build delegation tx
	var msgs []sdk.Msg
	for _, split := range splits {
		msgs = append(msgs, &stakingtypes.MsgDelegate{
			DelegatorAddress: pool.IcaDepositAddress,
			ValidatorAddress: split.ValidatorAddress,
			Amount:           sdk.NewCoin(pool.NativeDenom, split.Amount),
		})
	}

	// Dispatch our message with a timeout of 30 minutes in nanos
	sequence, err := runner.keeper.BroadcastICAMessages(ctx, pool, types.ICATypeDeposit, msgs, types.IBCTimeoutNanos, ICACallbackID_Delegate, marshalledCallbackData)
	if err != nil {
		// Save error state since we cannot simply recover from a failure at this stage
		// A subsequent call to DepositRetry will be made possible by setting an error state and not returning an error here
		logger.Error(
			fmt.Sprintf("failed to dispatch ICA delegation: %v", err),
			"pool_id", pool.PoolId,
			"deposit_id", deposit.DepositId,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
		return splits, err
	}
	logger.Debug(
		"ICA delegation dispatched",
		"pool_id", pool.PoolId,
		"deposit_id", deposit.DepositId,
		"chain_id", pool.GetChainId(),
		"sequence", sequence,
	)
	return splits, nil
}

func (runner *PoolRunnerStaking) UndelegateWithdrawalOnRemoteZone(ctx sdk.Context, pool types.Pool, withdrawal types.Withdrawal) ([]*types.SplitDelegation, *time.Time, error) {
	logger := runner.Logger(ctx).With("ctx", "withdrawal_undelegate")

	if pool.IsLocalZone(ctx) {
		modAddr := sdk.MustAccAddressFromBech32(pool.GetIcaDepositAddress())
		splits := pool.ComputeSplitUndelegations(ctx, withdrawal.GetAmount().Amount)
		if len(splits) == 0 {
			return nil, nil, types.ErrPoolEmptySplitDelegations
		}
		var unbondingEndsAt *time.Time
		for _, split := range splits {
			valAddr, err := sdk.ValAddressFromBech32(split.ValidatorAddress)
			if err != nil {
				return splits, unbondingEndsAt, err
			}
			shares, err := runner.keeper.StakingKeeper.ValidateUnbondAmount(
				ctx,
				modAddr,
				valAddr,
				split.Amount,
			)
			if err != nil {
				return splits, unbondingEndsAt, errorsmod.Wrapf(err, "%s", valAddr.String())
			}

			// Trigger undelegate
			endsAt, err := runner.keeper.StakingKeeper.Undelegate(ctx, modAddr, valAddr, shares)
			if err != nil {
				return splits, unbondingEndsAt, errorsmod.Wrapf(err, "%s", valAddr.String())
			}
			if unbondingEndsAt == nil || endsAt.After(*unbondingEndsAt) {
				unbondingEndsAt = &endsAt
			}
		}
		return splits, unbondingEndsAt, nil
	}

	// Prepare undelegate split
	splits := pool.ComputeSplitUndelegations(ctx, withdrawal.GetAmount().Amount)
	if len(splits) == 0 {
		return nil, nil, types.ErrPoolEmptySplitDelegations
	}

	// Construct our callback data
	callbackData := types.UndelegateCallback{
		PoolId:           pool.PoolId,
		WithdrawalId:     withdrawal.WithdrawalId,
		SplitDelegations: splits,
	}
	marshalledCallbackData, err := runner.keeper.MarshalUndelegateCallbackArgs(ctx, callbackData)
	if err != nil {
		return splits, nil, err
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

	// Dispatch our message with a timeout of 30 minutes in nanos
	sequence, err := runner.keeper.BroadcastICAMessages(ctx, pool, types.ICATypeDeposit, msgs, types.IBCTimeoutNanos, ICACallbackID_Undelegate, marshalledCallbackData)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to dispatch ICA undelegate: %v", err),
			"pool_id", pool.PoolId,
			"withdrawal_id", withdrawal.WithdrawalId,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
		return splits, nil, err
	}
	logger.Debug(
		"ICA undelegate dispatched",
		"pool_id", pool.PoolId,
		"withdrawal_id", withdrawal.WithdrawalId,
		"chain_id", pool.GetChainId(),
		"sequence", sequence,
	)
	return splits, nil, nil
}

func (runner *PoolRunnerStaking) ClaimYieldOnRemoteZone(ctx sdk.Context, pool types.Pool) error {
	// logger := runner.Logger(ctx).With("ctx", "claim_yield")
	// TODO: implementation
	return fmt.Errorf("not implemented yet")
}

func (runner *PoolRunnerStaking) QueryFreshPrizePoolCoinsOnRemoteZone(ctx sdk.Context, pool types.Pool) error {
	// logger := runner.Logger(ctx).With("ctx", "query_fresh_prize_pool")
	// TODO: implementation
	return fmt.Errorf("not implemented yet")
}
