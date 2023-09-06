package keeper

import (
	"context"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

// ClaimPrize claim a prize from the transaction message
func (k msgServer) ClaimPrize(goCtx context.Context, msg *types.MsgClaimPrize) (*types.MsgClaimPrizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify address
	winnerAddr, err := sdk.AccAddressFromBech32(msg.GetWinnerAddress())
	if err != nil {
		return nil, types.ErrInvalidWinnerAddress
	}

	prize, err := k.GetPoolDrawPrize(ctx, msg.PoolId, msg.DrawId, msg.PrizeId)
	if err != nil {
		return nil, types.ErrPrizeNotFound
	}

	// msg.WinnerAddress should always match the prize.WinnerAddress
	if winnerAddr.String() != prize.WinnerAddress {
		return nil, types.ErrInvalidWinnerAddress
	}

	if prize.State != types.PrizeState_Pending {
		return nil, types.ErrInvalidPrizeState
	}

	// Get pool
	pool, err := k.GetPool(ctx, msg.PoolId)
	if err != nil {
		return nil, types.ErrPoolNotFound
	}

	if msg.IsAutoCompound {
		// Move claim prize fund to winner address
		if err := k.BankKeeper.SendCoins(
			ctx,
			sdk.MustAccAddressFromBech32(pool.GetLocalAddress()),
			winnerAddr,
			sdk.NewCoins(prize.Amount),
		); err != nil {
			return nil, err
		}

		// New deposit instance
		deposit := types.Deposit{
			PoolId:           pool.PoolId,
			State:            types.DepositState_IbcTransfer,
			DepositorAddress: winnerAddr.String(),
			Amount:           prize.Amount,
			WinnerAddress:    winnerAddr.String(),
			IsSponsor:        msg.GetIsSponsor(),
			DepositOrigin:    types.DepositOrigin_Autocompound,
			CreatedAtHeight:  ctx.BlockHeight(),
			UpdatedAtHeight:  ctx.BlockHeight(),
			CreatedAt:        ctx.BlockTime(),
			UpdatedAt:        ctx.BlockTime(),
		}

		// Remove prize entity
		if err := k.RemovePrize(ctx, prize); err != nil {
			return nil, err
		}

		// Check pool state before being able to deposit
		if pool.State != types.PoolState_Ready && pool.State != types.PoolState_Paused {
			return nil, errorsmod.Wrapf(
				types.ErrInvalidPoolState, "cannot deposit in pool during state %s", pool.State.String(),
			)
		}

		// Double check if deposit denom issued from prize is suitable for pool
		if deposit.Amount.Denom != pool.Denom {
			return nil, types.ErrInvalidDepositDenom
		}

		// Move funds to pool
		poolRunner, err := k.GetPoolRunner(pool.PoolType)
		if err != nil {
			return nil, errorsmod.Wrapf(types.ErrInvalidPoolType, err.Error())
		}
		if err := poolRunner.SendDepositToPool(ctx, pool, deposit); err != nil {
			return nil, err
		}

		// Store deposit
		k.AddDeposit(ctx, &deposit)

		// Emit event
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			),
			sdk.NewEvent(
				types.EventTypeDeposit,
				sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(deposit.PoolId, 10)),
				sdk.NewAttribute(types.AttributeKeyDepositID, strconv.FormatUint(deposit.DepositId, 10)),
				sdk.NewAttribute(types.AttributeKeyDepositor, deposit.DepositorAddress),
				sdk.NewAttribute(types.AttributeKeyWinner, deposit.WinnerAddress),
				sdk.NewAttribute(sdk.AttributeKeyAmount, deposit.Amount.String()),
				sdk.NewAttribute(types.AttributeKeySponsor, strconv.FormatBool(msg.IsSponsor)),
				sdk.NewAttribute(types.AttributeKeyDepositOrigin, deposit.DepositOrigin.String()),
			),
		})

		// Transfer to appropriate zone
		if err := k.TransferDepositToRemoteZone(ctx, deposit.GetPoolId(), deposit.GetDepositId()); err != nil {
			return nil, err
		}

		return nil, nil
	}

	// Move funds for a simple claim
	if err := k.BankKeeper.SendCoins(
		ctx,
		sdk.MustAccAddressFromBech32(pool.GetLocalAddress()),
		winnerAddr,
		sdk.NewCoins(prize.Amount),
	); err != nil {
		return nil, err
	}

	if err := k.RemovePrize(ctx, prize); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeClaimPrize,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(prize.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyDrawID, strconv.FormatUint(prize.DrawId, 10)),
			sdk.NewAttribute(types.AttributeKeyPrizeID, strconv.FormatUint(prize.PrizeId, 10)),
			sdk.NewAttribute(types.AttributeKeyWinner, prize.WinnerAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, prize.Amount.String()),
		),
	})

	return &types.MsgClaimPrizeResponse{}, nil
}
