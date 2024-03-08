package keeper

import (
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	icquerieskeeper "github.com/lum-network/chain/x/icqueries/keeper"
	icqueriestypes "github.com/lum-network/chain/x/icqueries/types"
	"github.com/lum-network/chain/x/millions/types"
)

func BalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqueriestypes.Query, status icqueriestypes.QueryResponseStatus) error {
	// Extract our keys from the extra id
	keys := types.DecombineStringKeys(query.ExtraId)
	if len(keys) != 2 {
		return errorsmod.Wrapf(types.ErrPoolNotFound, "keys is invalid, must contain two keys (pool ID and draw ID)")
	}
	poolID, err := strconv.ParseUint(keys[0], 10, 64)
	if err != nil {
		return err
	}
	drawID, err := strconv.ParseUint(keys[1], 10, 64)
	if err != nil {
		return err
	}

	// Confirm host exists
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return errorsmod.Wrapf(types.ErrPoolNotFound, "no registered pool for queried chain ID (%s) or combined keys (%s)", query.ChainId, query.ExtraId)
	}

	// Confirm draw exists
	draw, err := k.GetPoolDraw(ctx, poolID, drawID)
	if err != nil {
		return errorsmod.Wrapf(types.ErrPoolNotFound, "no registered draw %d for pool %d", drawID, poolID)
	}

	// Scenarios:
	// - Timeout: Treated as an error (see below)
	// - Error: Put entity in error state to allow users to retry
	if status == icqueriestypes.QueryResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(fmt.Sprintf("QUERY TIMEOUT - QueryId: %s, TTL: %d, BlockTime: %d", query.Id, query.TimeoutTimestamp, ctx.BlockHeader().Time.UnixNano()))
		_, err := k.OnQueryFreshPrizePoolCoinsOnRemoteZoneCompleted(ctx, pool.GetPoolId(), draw.GetDrawId(), sdk.NewCoins(), true)
		if err != nil {
			return err
		}
	} else if status == icqueriestypes.QueryResponseStatus_FAILURE {
		k.Logger(ctx).Error(fmt.Sprintf("QUERY FAILURE - QueryId: %s, TTL: %d, BlockTime: %d", query.Id, query.TimeoutTimestamp, ctx.BlockHeader().Time.UnixNano()))
		_, err := k.OnQueryFreshPrizePoolCoinsOnRemoteZoneCompleted(ctx, pool.GetPoolId(), draw.GetDrawId(), sdk.NewCoins(), true)
		if err != nil {
			return err
		}
	} else if status == icqueriestypes.QueryResponseStatus_SUCCESS {
		// Unmarshal query response
		balanceAmount, err := icquerieskeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
		if err != nil {
			return errorsmod.Wrap(err, "unable to determine balance from query response")
		}

		k.Logger(ctx).Info(fmt.Sprintf("Query response - Rewards Balance: %d %s", balanceAmount, pool.GetNativeDenom()))

		// Notify the internal systems
		_, err = k.OnQueryFreshPrizePoolCoinsOnRemoteZoneCompleted(ctx, pool.GetPoolId(), draw.GetDrawId(), sdk.NewCoins(sdk.NewCoin(pool.GetNativeDenom(), balanceAmount)), false)
		if err != nil {
			return err
		}
	}
	return nil
}
