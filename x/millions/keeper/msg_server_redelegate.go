package keeper

import (
	"context"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

func (k msgServer) RedelegateRetry(goCtx context.Context, msg *types.MsgRedelegateRetry) (*types.MsgRedelegateRetryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Ensure msg received is valid
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	pool, err := k.GetPool(ctx, msg.GetPoolId())
	if err != nil {
		return nil, types.ErrPoolNotFound
	}

	_, err = sdk.AccAddressFromBech32(msg.RedelegateRetryAddress)
	if err != nil {
		return nil, types.ErrInvalidRedelegateRetryAddress
	}

	_, validator, err := pool.FindValidatorByOperatorAddress(ctx, msg.GetOperatetorAddress())
	if err != nil {
		return nil, err
	}

	if !validator.Redelegate.IsGovPropRedelegated {
		return nil, types.ErrValidatorNotRedelegated
	}

	if !validator.IsEnabled {
		return nil, types.ErrValidatorNotDisabled
	}

	// State should be set to failure in order to retry something
	if validator.Redelegate.State != types.RedelegateState_Failure {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidRedelegateState,
			"state is %s instead of %s",
			validator.Redelegate.State.String(), types.RedelegateState_Failure.String(),
		)
	}

	if validator.Redelegate.ErrorState == types.RedelegateState_IcaRedelegate {
		if err := k.UpdateRedelegateStatus(ctx, pool.PoolId, types.RedelegateState_IcaRedelegate, validator.GetOperatorAddress(), validator.Redelegate.RedelegationEndsAt, false); err != nil {
			return &types.MsgRedelegateRetryResponse{}, err
		}

		if err := k.Redelegate(ctx, msg.GetOperatetorAddress(), msg.GetPoolId()); err != nil {
			return nil, err
		}
	} else {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidRedelegateState,
			"error_state is %s instead of %s",
			validator.Redelegate.ErrorState.String(), types.RedelegateState_IcaRedelegate.String(),
		)
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeRedelegateRetry,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(pool.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, validator.GetOperatorAddress()),
		),
	})

	return &types.MsgRedelegateRetryResponse{}, nil
}
