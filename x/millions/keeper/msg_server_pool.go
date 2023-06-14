package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

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
	icaDepositPortName := string(types.NewPoolName(pool.GetPoolId(), types.ICATypeDeposit))
	if pool.GetIcaDepositPortId() == "" {
		// Unusual case - no port ID was set at Pool registration time (ex: genesis import without portID)
		// Basically create a new one
		pool.IcaDepositPortId, err = icatypes.NewControllerPortID(icaDepositPortName)
		if err != nil {
			return nil, errorsmod.Wrapf(types.ErrFailedToRegisterPool, fmt.Sprintf("Unable to create deposit account port id, err: %s", err.Error()))
		}
		k.updatePool(ctx, &pool)
	}
	_, found := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, pool.GetConnectionId(), pool.GetIcaDepositPortId())
	if !found {
		if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, pool.GetConnectionId(), icaDepositPortName, appVersion); err != nil {
			return nil, errorsmod.Wrapf(types.ErrFailedToRestorePool, fmt.Sprintf("Unable to trigger deposit account registration, err: %s", err.Error()))
		}
		// Exit to prevent a double channel registration which might create unpredictable behaviours
		// The registration of the ICA PrizePool will either be automatically triggered once the ICA Deposit gets ACK (see keeper.OnSetupPoolICACompleted)
		// or can be restored by calling this method again
		return &types.MsgRestoreInterchainAccountsResponse{}, nil
	}

	// Regenerate our prizepool port only if required (open active channel not found)
	icaPrizePoolPortName := string(types.NewPoolName(pool.GetPoolId(), types.ICATypePrizePool))
	if pool.GetIcaPrizepoolPortId() == "" {
		// Unusual case - no port ID was set at Pool ICA Deposit ACK time
		// Basically create a new one
		pool.IcaPrizepoolPortId, err = icatypes.NewControllerPortID(icaPrizePoolPortName)
		if err != nil {
			return nil, errorsmod.Wrapf(types.ErrFailedToRegisterPool, fmt.Sprintf("Unable to create prizepool account port id, err: %s", err.Error()))
		}
		k.updatePool(ctx, &pool)
	}
	_, found = k.ICAControllerKeeper.GetOpenActiveChannel(ctx, pool.GetConnectionId(), pool.GetIcaPrizepoolPortId())
	if !found {
		if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, pool.GetConnectionId(), icaPrizePoolPortName, appVersion); err != nil {
			return nil, errorsmod.Wrapf(types.ErrFailedToRestorePool, fmt.Sprintf("Unable to trigger prizepool account registration, err: %s", err.Error()))
		}
	}

	return &types.MsgRestoreInterchainAccountsResponse{}, nil
}
