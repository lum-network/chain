package keeper

import (
	"fmt"

	icquerieskeeper "github.com/lum-network/chain/x/icqueries/keeper"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	account "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	icacallbackskeeper "github.com/lum-network/chain/x/icacallbacks/keeper"
	"github.com/lum-network/chain/x/millions/types"
)

type Keeper struct {
	cdc                 codec.BinaryCodec
	storeKey            storetypes.StoreKey
	paramSpace          paramtypes.Subspace
	scopedKeeper        capabilitykeeper.ScopedKeeper
	AccountKeeper       account.AccountKeeper
	IBCKeeper           ibckeeper.Keeper
	IBCTransferKeeper   ibctransferkeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICACallbacksKeeper  icacallbackskeeper.Keeper
	ICQueriesKeeper     icquerieskeeper.Keeper
	BankKeeper          bankkeeper.Keeper
	DistributionKeeper  *distributionkeeper.Keeper
	StakingKeeper       *stakingkeeper.Keeper

	poolRunners map[types.PoolType]PoolRunner
}

// NewKeeper Initialize the keeper with the base params
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, scopedKeeper capabilitykeeper.ScopedKeeper,
	accountKeeper account.AccountKeeper, ibcKeeper ibckeeper.Keeper, ibcTransferKeeper ibctransferkeeper.Keeper, icaKeeper icacontrollerkeeper.Keeper, icaCallbacksKeeper icacallbackskeeper.Keeper,
	icqueriesKeeper icquerieskeeper.Keeper, bank bankkeeper.Keeper, distribution *distributionkeeper.Keeper, stakingKeeper *stakingkeeper.Keeper,
) *Keeper {
	k := Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		paramSpace:          paramSpace,
		scopedKeeper:        scopedKeeper,
		AccountKeeper:       accountKeeper,
		IBCKeeper:           ibcKeeper,
		IBCTransferKeeper:   ibcTransferKeeper,
		ICAControllerKeeper: icaKeeper,
		ICACallbacksKeeper:  icaCallbacksKeeper,
		ICQueriesKeeper:     icqueriesKeeper,
		BankKeeper:          bank,
		DistributionKeeper:  distribution,
		StakingKeeper:       stakingKeeper,
	}
	k.RegisterPoolRunners()
	return &k
}

// Logger Return a keeper logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

// GetChainID Return the chain ID fetched from the ibc connection layer
func (k Keeper) GetChainID(ctx sdk.Context, connectionID string) (string, error) {
	conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return "", fmt.Errorf(fmt.Sprintf("invalid connection id, %s not found", connectionID))
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.ClientId)
	if !found {
		return "", fmt.Errorf(fmt.Sprintf("client id %s not found for connection %s", conn.ClientId, connectionID))
	}
	client, ok := clientState.(*tendermint.ClientState)
	if !ok {
		return "", fmt.Errorf(fmt.Sprintf("invalid client state for client %s on connection %s", conn.ClientId, connectionID))
	}

	return client.ChainId, nil
}

func (k Keeper) GetConnectionID(ctx sdk.Context, portId string) (string, error) {
	icas := k.ICAControllerKeeper.GetAllInterchainAccounts(ctx)
	for _, ica := range icas {
		if ica.PortId == portId {
			return ica.ConnectionId, nil
		}
	}
	return "", fmt.Errorf(fmt.Sprintf("portId %s has no associated connectionId", portId))
}

// RegisterPoolRunners register all know Pool Runners
func (k *Keeper) RegisterPoolRunners() {
	k.poolRunners = make(map[types.PoolType]PoolRunner)
	k.poolRunners[types.PoolType_Staking] = &PoolRunnerStaking{PoolRunnerBase{keeper: k}}
}

// GetPoolRunner returns the Pool Runner for the specified Pool Type
// returns an error if no runner registered for this Pool Type
func (k Keeper) GetPoolRunner(poolType types.PoolType) (PoolRunner, error) {
	if runner, found := k.poolRunners[poolType]; found {
		return runner, nil
	}
	return nil, fmt.Errorf("pool runner not registered for type %s", poolType.String())
}

// MustGetPoolRunner runs GetPoolRunner and panic in case of error
func (k Keeper) MustGetPoolRunner(poolType types.PoolType) PoolRunner {
	runner, err := k.GetPoolRunner(poolType)
	if err != nil {
		panic(err)
	}
	return runner
}
