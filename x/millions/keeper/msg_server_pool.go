package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/millions/types"
)

func (k msgServer) RestoreInterchainAccounts(goCtx context.Context, msg *types.MsgRestoreInterchainAccounts) (*types.MsgRestoreInterchainAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Grab our pool instance, otherwise just return error
	pool, err := k.GetPool(ctx, msg.GetPoolId())
	if err != nil {
		return nil, types.ErrPoolNotFound
	}

	// We always want our pool to be ready
	if pool.State == types.PoolState_Created {
		return nil, types.ErrPoolNotReady
	}

	// Grab the app version to use
	appVersion, err := k.getPoolAppVersion(ctx, pool)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrFailedToRestorePool, err.Error())
	}

	// We don't want to run this against a local pool
	if pool.IsLocalZone(ctx) {
		return nil, errorsmod.Wrapf(types.ErrFailedToRestorePool, "Pool is local")
	}

	// Regenerate our deposit port only if required (open active channel not found)
	_, found := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, pool.GetConnectionId(), pool.GetIcaDepositPortId())
	if !found {
		icaDepositPortName := string(types.NewPoolName(pool.GetPoolId(), types.ICATypeDeposit))
		if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, pool.GetConnectionId(), icaDepositPortName, appVersion); err != nil {
			return nil, errorsmod.Wrapf(types.ErrFailedToRestorePool, fmt.Sprintf("Unable to trigger deposit account registration, err: %s", err.Error()))
		}
	}

	// Regenerate our prizepool port only if required (open active channel not found)
	_, found = k.ICAControllerKeeper.GetOpenActiveChannel(ctx, pool.GetConnectionId(), pool.GetIcaPrizepoolPortId())
	if !found {
		icaPrizePoolPortName := string(types.NewPoolName(pool.GetPoolId(), types.ICATypePrizePool))
		if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, pool.GetConnectionId(), icaPrizePoolPortName, appVersion); err != nil {
			return nil, errorsmod.Wrapf(types.ErrFailedToRestorePool, fmt.Sprintf("Unable to trigger prizepool account registration, err: %s", err.Error()))
		}
	}

	return &types.MsgRestoreInterchainAccountsResponse{}, nil
}
