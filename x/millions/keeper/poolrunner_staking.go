package keeper

import (
	"fmt"

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
	sequence, err := runner.keeper.BroadcastICAMessages(ctx, pool.PoolId, types.ICATypeDeposit, msgs, types.IBCTimeoutNanos, ICACallbackID_Delegate, marshalledCallbackData)
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
