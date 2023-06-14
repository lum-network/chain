package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	icqueriestypes "github.com/lum-network/chain/x/icqueries/types"
)

const (
	ICQCallbackIDBalance = "balance"
)

// ICQCallback wrapper struct for millions keeper.
type ICQCallback func(Keeper, sdk.Context, []byte, icqueriestypes.Query, icqueriestypes.QueryResponseStatus) error

type ICQCallbacks struct {
	k         Keeper
	callbacks map[string]ICQCallback
}

var _ icqueriestypes.QueryCallbacks = ICQCallbacks{}

func (k Keeper) ICQCallbackHandler() ICQCallbacks {
	return ICQCallbacks{k, make(map[string]ICQCallback)}
}

func (c ICQCallbacks) CallICQCallback(ctx sdk.Context, id string, args []byte, query icqueriestypes.Query, status icqueriestypes.QueryResponseStatus) error {
	return c.callbacks[id](c.k, ctx, args, query, status)
}

func (c ICQCallbacks) HasICQCallback(id string) bool {
	_, found := c.callbacks[id]
	return found
}

func (c ICQCallbacks) AddICQCallback(id string, fn interface{}) icqueriestypes.QueryCallbacks {
	c.callbacks[id] = fn.(ICQCallback)
	return c
}

func (c ICQCallbacks) RegisterICQCallbacks() icqueriestypes.QueryCallbacks {
	return c.AddICQCallback(ICQCallbackIDBalance, ICQCallback(BalanceCallback))
}
