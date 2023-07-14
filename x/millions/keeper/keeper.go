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
	AccountKeeper       account.AccountKeeper
	IBCKeeper           ibckeeper.Keeper
	IBCTransferKeeper   ibctransferkeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICACallbacksKeeper  icacallbackskeeper.Keeper
	ICQueriesKeeper     icquerieskeeper.Keeper
	BankKeeper          bankkeeper.Keeper
	DistributionKeeper  *distributionkeeper.Keeper
	StakingKeeper       *stakingkeeper.Keeper
}

// NewKeeper Initialize the keeper with the base params
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace,
	accountKeeper account.AccountKeeper, ibcKeeper ibckeeper.Keeper, ibcTransferKeeper ibctransferkeeper.Keeper, icaKeeper icacontrollerkeeper.Keeper, icaCallbacksKeeper icacallbackskeeper.Keeper,
	icqueriesKeeper icquerieskeeper.Keeper, bank bankkeeper.Keeper, distribution *distributionkeeper.Keeper, stakingKeeper *stakingkeeper.Keeper,
) *Keeper {
	return &Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		paramSpace:          paramSpace,
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
}

// Logger Return a keeper logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
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
