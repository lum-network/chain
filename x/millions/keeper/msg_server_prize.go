package keeper

import (
	"context"
	"strconv"

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

	// Move funds to winner address
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

	if msg.IsAutoCompound {
		// Construct the deposit msg
		msg := types.MsgDeposit{
			PoolId:           pool.PoolId,
			Amount:           prize.Amount,
			DepositorAddress: winnerAddr.String(),
			WinnerAddress:    winnerAddr.String(),
			IsSponsor:        msg.GetIsSponsor(),
		}

		// Check internal validation before creating a deposit
		if _, err := k.CreateDeposit(ctx, &msg, types.DepositOrigin_Autocompound); err != nil {
			return nil, err
		}
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
			sdk.NewAttribute(types.AttributeKeyAutoCompound, strconv.FormatBool(msg.IsAutoCompound)),
		),
	})

	return &types.MsgClaimPrizeResponse{}, nil
}
