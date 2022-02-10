package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	// this line is used by starport scaffolding # 1
)

const (
	MethodGet = "GET"
)

// RegisterRoutes registers chain-related REST handlers to a router
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/cosmos/base/tendermint/v1beta1/node_info", NodeInfoRequestHandlerFn(clientCtx)).Methods(MethodGet)
	r.HandleFunc("/node_info", NodeInfoRequestHandlerFn(clientCtx)).Methods(MethodGet)
	r.HandleFunc("/cosmos/base/tendermint/v1beta1/syncing", NodeSyncingRequestHandlerFn(clientCtx)).Methods(MethodGet)
	r.HandleFunc("/syncing", NodeSyncingRequestHandlerFn(clientCtx)).Methods(MethodGet)

	r.HandleFunc("/bank/balances/{address}", QueryBalancesRequestHandlerFn(clientCtx)).Methods(MethodGet)
	r.HandleFunc("/supply/total", totalSupplyHandlerFn(clientCtx)).Methods(MethodGet)
	r.HandleFunc("/supply/total/{denom}", supplyOfHandlerFn(clientCtx)).Methods(MethodGet)

	r.HandleFunc("/staking/delegators/{delegatorAddr}/delegations", delegatorDelegationsHandlerFn(clientCtx)).Methods(MethodGet)
	r.HandleFunc("/staking/delegators/{delegatorAddr}/unbonding_delegations", delegatorUnbondingDelegationsHandlerFn(clientCtx)).Methods(MethodGet)
	r.HandleFunc("/staking/pool", poolHandlerFn(clientCtx)).Methods(MethodGet)

	r.HandleFunc("/distribution/delegators/{delegatorAddr}/rewards", delegatorRewardsHandlerFn(clientCtx)).Methods(MethodGet)

	r.HandleFunc("/minting/inflation", queryInflationHandlerFn(clientCtx)).Methods(MethodGet)
}

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	// this line is used by starport scaffolding # 3
}

func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
	// this line is used by starport scaffolding # 4
}
