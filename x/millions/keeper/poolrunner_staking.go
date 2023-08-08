package keeper

import (
	"fmt"
	"strconv"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	icqueriestypes "github.com/lum-network/chain/x/icqueries/types"
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

func (runner *PoolRunnerStaking) UndelegateWithdrawalsOnRemoteZone(ctx sdk.Context, epochUnbonding types.EpochUnbonding) error {
	logger := runner.Logger(ctx).With("ctx", "withdrawal_undelegate")

	pool, err := runner.keeper.GetPool(ctx, epochUnbonding.PoolId)
	if err != nil {
		return err
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
			shares, err := runner.keeper.StakingKeeper.ValidateUnbondAmount(
				ctx,
				modAddr,
				valAddr,
				split.Amount,
			)
			if err != nil {
				return errorsmod.Wrapf(err, "%s", valAddr.String())
			}

			// Trigger undelegate
			endsAt, err := runner.keeper.StakingKeeper.Undelegate(ctx, modAddr, valAddr, shares)
			if err != nil {
				return errorsmod.Wrapf(err, "%s", valAddr.String())
			}
			if unbondingEndsAt == nil || endsAt.After(*unbondingEndsAt) {
				unbondingEndsAt = &endsAt
			}
		}
		// Apply undelegate pool validators update
		pool.ApplySplitUndelegate(ctx, splits)
		runner.keeper.updatePool(ctx, &pool)
		return runner.keeper.OnUndelegateWithdrawalsOnRemoteZoneCompleted(ctx, epochUnbonding.PoolId, epochUnbonding.WithdrawalIds, unbondingEndsAt, false)
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
	marshalledCallbackData, err := runner.keeper.MarshalUndelegateCallbackArgs(ctx, callbackData)
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
	runner.keeper.updatePool(ctx, &pool)

	// Dispatch our message with a timeout of 30 minutes in nanos
	sequence, err := runner.keeper.BroadcastICAMessages(ctx, pool, types.ICATypeDeposit, msgs, types.IBCTimeoutNanos, ICACallbackID_Undelegate, marshalledCallbackData)
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

func (runner *PoolRunnerStaking) RedelegateToActiveValidatorsOnRemoteZone(ctx sdk.Context, pool types.Pool, inactiveVal types.PoolValidator, splits []*types.SplitDelegation) error {
	logger := runner.Logger(ctx).With("ctx", "pool_redelegate")

	// If pool is local, we just process operation in place
	// Otherwise we trigger ICA transactions
	if pool.IsLocalZone(ctx) {
		delAddr := sdk.MustAccAddressFromBech32(pool.GetIcaDepositAddress())
		valSrcAddr, err := sdk.ValAddressFromBech32(inactiveVal.GetOperatorAddress())
		if err != nil {
			return err
		}

		for _, split := range splits {
			valDstAddr, err := sdk.ValAddressFromBech32(split.GetValidatorAddress())
			if err != nil {
				return err
			}

			// Validate the redelegation sharesAmount
			sharesAmount, err := runner.keeper.StakingKeeper.ValidateUnbondAmount(
				ctx,
				delAddr,
				valSrcAddr,
				split.Amount,
			)
			if err != nil {
				return errorsmod.Wrapf(err, "%s", valDstAddr.String())
			}

			_, err = runner.keeper.StakingKeeper.BeginRedelegation(ctx, delAddr, valSrcAddr, valDstAddr, sharesAmount)
			if err != nil {
				return errorsmod.Wrapf(err, "%s", valDstAddr.String())
			}
		}

		// ApplySplitRedelegate to pool validator set
		pool.ApplySplitRedelegate(ctx, inactiveVal.GetOperatorAddress(), splits)
		runner.keeper.updatePool(ctx, &pool)

		return nil
	}

	// Construct our callback data
	callbackData := types.RedelegateCallback{
		PoolId:           pool.PoolId,
		OperatorAddress:  inactiveVal.GetOperatorAddress(),
		SplitDelegations: splits,
	}
	marshalledCallbackData, err := runner.keeper.MarshalRedelegateCallbackArgs(ctx, callbackData)
	if err != nil {
		return err
	}

	// Build the MsgBeginRedelegate
	var msgs []sdk.Msg
	for _, split := range splits {
		msgs = append(msgs, &stakingtypes.MsgBeginRedelegate{
			DelegatorAddress:    pool.GetIcaDepositAddress(),
			ValidatorSrcAddress: inactiveVal.GetOperatorAddress(),
			ValidatorDstAddress: split.GetValidatorAddress(),
			Amount:              sdk.NewCoin(pool.NativeDenom, split.Amount),
		})
	}

	// ApplySplitRedelegate to pool validator set
	pool.ApplySplitRedelegate(ctx, inactiveVal.GetOperatorAddress(), splits)
	runner.keeper.updatePool(ctx, &pool)

	// Dispatch our message with a timeout of 30 minutes in nanos
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + types.IBCTimeoutNanos
	sequence, err := runner.keeper.BroadcastICAMessages(ctx, pool, types.ICATypeDeposit, msgs, timeoutTimestamp, ICACallbackID_Redelegate, marshalledCallbackData)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to dispatch ICA redelegation: %v", err),
			"pool_id", pool.PoolId,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
		return err
	}
	logger.Debug(
		"ICA redelegation dispatched",
		"pool_id", pool.PoolId,
		"chain_id", pool.GetChainId(),
		"sequence", sequence,
	)

	return nil
}

func (runner *PoolRunnerStaking) ClaimYieldOnRemoteZone(ctx sdk.Context, pool types.Pool, draw types.Draw) error {
	logger := runner.Logger(ctx).With("ctx", "claim_yield")

	if pool.IsLocalZone(ctx) {
		coins := sdk.Coins{}
		for _, validator := range pool.GetValidators() {
			if validator.IsBonded() {
				rewardCoins, err := runner.keeper.DistributionKeeper.WithdrawDelegationRewards(
					ctx,
					sdk.MustAccAddressFromBech32(pool.GetIcaDepositAddress()),
					validator.MustValAddressFromBech32(),
				)
				if err != nil {
					// Return with error here since it is the first operation and nothing needs to be saved to state
					return errorsmod.Wrapf(err, "%s", validator.OperatorAddress)
				}
				coins = coins.Add(rewardCoins...)
			}
		}
	} else {
		var msgs []sdk.Msg
		for _, validator := range pool.GetValidators() {
			if validator.IsBonded() {
				msgs = append(msgs, &distributiontypes.MsgWithdrawDelegatorReward{
					DelegatorAddress: pool.GetIcaDepositAddress(),
					ValidatorAddress: validator.OperatorAddress,
				})
			}
		}

		if len(msgs) == 0 {
			// Special case - no bonded validator
			// Does not need to do any ICA call
			return nil
		}

		callbackData := types.ClaimRewardsCallback{
			PoolId: pool.PoolId,
			DrawId: draw.DrawId,
		}

		marshalledCallbackData, err := runner.keeper.MarshalClaimCallbackArgs(ctx, callbackData)
		if err != nil {
			return err
		}

		// Dispatch our message with a timeout of 30 minutes in nanos
		sequence, err := runner.keeper.BroadcastICAMessages(ctx, pool, types.ICATypeDeposit, msgs, types.IBCTimeoutNanos, ICACallbackID_Claim, marshalledCallbackData)
		if err != nil {
			// Return with error here since it is the first operation and nothing needs to be saved to state
			logger.Error(
				fmt.Sprintf("failed to dispatch ICA claim delegator rewards: %v", err),
				"pool_id", pool.PoolId,
				"draw_id", draw.DrawId,
				"chain_id", pool.GetChainId(),
				"sequence", sequence,
			)
			return err
		}
		logger.Debug(
			"ICA claim delegator rewards dispatched",
			"pool_id", pool.PoolId,
			"draw_id", draw.DrawId,
			"chain_id", pool.GetChainId(),
			"sequence", sequence,
		)
	}
	return nil
}

func (runner *PoolRunnerStaking) QueryFreshPrizePoolCoinsOnRemoteZone(ctx sdk.Context, pool types.Pool, draw types.Draw) error {
	logger := runner.Logger(ctx).With("ctx", "query_fresh_prize_pool")

	// Encode the ica address for query
	_, icaAddressBz, err := bech32.DecodeAndConvert(pool.GetIcaPrizepoolAddress())
	if err != nil {
		panic(err)
	}

	// Construct the query data and timeout timestamp (now + 30 minutes)
	extraID := types.CombineStringKeys(strconv.FormatUint(pool.PoolId, 10), strconv.FormatUint(draw.DrawId, 10))
	queryData := append(banktypes.CreateAccountBalancesPrefix(icaAddressBz), []byte(pool.GetNativeDenom())...)

	err = runner.keeper.BroadcastICQuery(
		ctx,
		pool,
		ICQCallbackID_Balance,
		extraID,
		icqueriestypes.BANK_STORE_QUERY_WITH_PROOF,
		queryData,
		types.IBCTimeoutNanos,
	)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to dispatch ICQ query balance: %v", err),
			"pool_id", pool.PoolId,
			"draw_id", draw.DrawId,
			"chain_id", pool.GetChainId(),
		)
		return err
	}
	logger.Debug(
		"ICQ query balance dispatched",
		"pool_id", pool.PoolId,
		"draw_id", draw.DrawId,
		"chain_id", pool.GetChainId(),
	)
	return nil
}
