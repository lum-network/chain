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
	r.HandleFunc("/bank/balances/{address}", QueryBalancesRequestHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/supply/total", totalSupplyHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/supply/total/{denom}", supplyOfHandlerFn(clientCtx)).Methods("GET")

	r.HandleFunc("/staking/delegators/{delegatorAddr}/delegations", delegatorDelegationsHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/staking/delegators/{delegatorAddr}/unbonding_delegations", delegatorUnbondingDelegationsHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/staking/pool", poolHandlerFn(clientCtx)).Methods("GET")

	r.HandleFunc("/distribution/delegators/{delegatorAddr}/rewards", delegatorRewardsHandlerFn(clientCtx)).Methods("GET")

	r.HandleFunc("/minting/inflation", queryInflationHandlerFn(clientCtx)).Methods("GET")
}

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	// this line is used by starport scaffolding # 3
}

func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
	// this line is used by starport scaffolding # 4
}
