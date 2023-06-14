package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/millions/types"
)

// DrawRetry allows to retry a failed draw.
func (k msgServer) DrawRetry(goCtx context.Context, msg *types.MsgDrawRetry) (*types.MsgDrawRetryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Ensure msg received is valid
	if err := msg.ValidateDrawRetryBasic(); err != nil {
		return nil, err
	}

	if !k.HasPool(ctx, msg.PoolId) {
		return nil, types.ErrPoolNotFound
	}

	// Acquire Draw
	draw, err := k.GetPoolDraw(ctx, msg.PoolId, msg.DrawId)
	if err != nil {
		return nil, err
	}

	// State should be set to failure in order to retry something
	if draw.State != types.DrawState_Failure {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidDrawState,
			"state is %s instead of %s",
			draw.State.String(), types.DrawState_Failure.String(),
		)
	}

	// DrawState_IcaWithdrawRewards refers to the failed ica callback if OnClaimRewardsOnNativeChainCompleted fails
	if draw.ErrorState == types.DrawState_IcaWithdrawRewards {
		draw.UpdatedAtHeight = ctx.BlockHeight()
		draw.UpdatedAt = ctx.BlockTime()
		draw.State = types.DrawState_IcaWithdrawRewards
		draw.ErrorState = types.DrawState_Unspecified
		k.SetPoolDraw(ctx, draw)
		if _, err := k.ClaimRewardsOnNativeChain(ctx, draw.PoolId, draw.DrawId); err != nil {
			return nil, err
		}
		// DrawState_IcqRewards refers to the failed icq callback
	} else if draw.ErrorState == types.DrawState_IcqBalance {
		draw.UpdatedAtHeight = ctx.BlockHeight()
		draw.UpdatedAt = ctx.BlockTime()
		draw.State = types.DrawState_IcqBalance
		draw.ErrorState = types.DrawState_Unspecified
		k.SetPoolDraw(ctx, draw)
		if _, err := k.QueryBalance(ctx, draw.GetPoolId(), draw.GetDrawId()); err != nil {
			return nil, err
		}
		// DrawState_IbcTransfer refers to the failed ibc call if OnTransferRewardsToLocalChainCompleted fails
	} else if draw.ErrorState == types.DrawState_IbcTransfer {
		draw.UpdatedAtHeight = ctx.BlockHeight()
		draw.UpdatedAt = ctx.BlockTime()
		draw.State = types.DrawState_IbcTransfer
		draw.ErrorState = types.DrawState_Unspecified
		k.SetPoolDraw(ctx, draw)
		if _, err := k.TransferRewardsToLocalChain(ctx, draw.PoolId, draw.DrawId); err != nil {
			return nil, err
		}
	} else if draw.ErrorState == types.DrawState_Drawing {
		draw.UpdatedAtHeight = ctx.BlockHeight()
		draw.UpdatedAt = ctx.BlockTime()
		draw.State = types.DrawState_Drawing
		draw.ErrorState = types.DrawState_Unspecified
		k.SetPoolDraw(ctx, draw)
		if _, err := k.ExecuteDraw(ctx, draw.PoolId, draw.DrawId); err != nil {
			return nil, err
		}
	} else {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidDrawState,
			"error_state is %s instead of %s or %s",
			draw.ErrorState.String(), types.DrawState_IcaWithdrawRewards.String(), types.DrawState_IbcTransfer.String(),
		)
	}

	return &types.MsgDrawRetryResponse{}, nil
}
