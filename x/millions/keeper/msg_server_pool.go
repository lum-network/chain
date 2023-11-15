package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
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
	icaDepositPortName := string(types.NewPoolName(pool.GetPoolId(), types.ICATypeDeposit))
	if pool.GetIcaDepositPortId() == "" {
		// Unusual case - no port ID was set at Pool registration time (ex: genesis import without portID)
		pool.IcaDepositPortId = icaDepositPortName
		k.updatePool(ctx, &pool)
	}
	_, found := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, pool.GetConnectionId(), pool.GetIcaDepositPortId())
	if !found {
		if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, pool.GetConnectionId(), icaDepositPortName, appVersion); err != nil {
			return nil, errorsmod.Wrapf(types.ErrFailedToRestorePool, fmt.Sprintf("Unable to trigger deposit account registration, err: %s", err.Error()))
		}
		k.restoreICADepositEntities(ctx, pool.GetPoolId())
		// Exit to prevent a double channel registration which might create unpredictable behaviours
		// The registration of the ICA PrizePool will either be automatically triggered once the ICA Deposit gets ACK (see keeper.OnSetupPoolICACompleted)
		// or can be restored by calling this method again
		return &types.MsgRestoreInterchainAccountsResponse{}, nil
	}

	// Regenerate our prizepool port only if required (open active channel not found)
	icaPrizePoolPortName := string(types.NewPoolName(pool.GetPoolId(), types.ICATypePrizePool))
	if pool.GetIcaPrizepoolPortId() == "" {
		// Unusual case - no port ID was set at Pool ICA Deposit ACK time
		pool.IcaPrizepoolPortId = icaPrizePoolPortName
		k.updatePool(ctx, &pool)
	}
	_, found = k.ICAControllerKeeper.GetOpenActiveChannel(ctx, pool.GetConnectionId(), pool.GetIcaPrizepoolPortId())
	if !found {
		if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, pool.GetConnectionId(), icaPrizePoolPortName, appVersion); err != nil {
			return nil, errorsmod.Wrapf(types.ErrFailedToRestorePool, fmt.Sprintf("Unable to trigger prizepool account registration, err: %s", err.Error()))
		}
		k.restoreICAPrizePoolEntities(ctx, pool.GetPoolId())
	}

	return &types.MsgRestoreInterchainAccountsResponse{}, nil
}

// restoreICADepositEntities restores all blocked entities by putting them in error state instead of locked state
// thus leaving users the opportunity to retry the operation
func (k msgServer) restoreICADepositEntities(ctx sdk.Context, poolID uint64) {
	// Restore deposits ICA locked operations on ICADeposit account
	deposits := k.Keeper.ListPoolDeposits(ctx, poolID)
	for _, d := range deposits {
		if d.State == types.DepositState_IcaDelegate {
			d.ErrorState = d.State
			d.State = types.DepositState_Failure
			k.Keeper.setPoolDeposit(ctx, &d)
		}
	}
	// Restore withdrawals ICA locked operations on ICADeposit account
	withdrawals := k.Keeper.ListPoolWithdrawals(ctx, poolID)
	for _, w := range withdrawals {
		if w.State == types.WithdrawalState_IcaUndelegate {
			w.ErrorState = w.State
			w.State = types.WithdrawalState_Failure
			k.Keeper.setPoolWithdrawal(ctx, w)
		}
	}
	// Restore draws ICA locked operations on ICADeposit account
	draws := k.Keeper.ListPoolDraws(ctx, poolID)
	for _, d := range draws {
		if d.State == types.DrawState_IcaWithdrawRewards {
			d.ErrorState = d.State
			d.State = types.DrawState_Failure
			k.Keeper.SetPoolDraw(ctx, d)
		}
	}
}

// restoreICAPrizePoolEntities restores all blocked entities by putting them in error state instead of locked state
// thus leaving users the opportunity to retry the operation
func (k msgServer) restoreICAPrizePoolEntities(ctx sdk.Context, poolID uint64) {
	// Restore draws ICQ locked operations on ICAPrizePool account
	draws := k.Keeper.ListPoolDraws(ctx, poolID)
	for _, d := range draws {
		if d.State == types.DrawState_IcqBalance {
			d.ErrorState = d.State
			d.State = types.DrawState_Failure
			k.Keeper.SetPoolDraw(ctx, d)
		}
	}
}
