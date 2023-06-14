package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/icqueries/types"
)

var _ types.QueryServiceServer = Keeper{}

// PendingQueries Query all queries that have been requested but have not received a response.
func (k Keeper) PendingQueries(c context.Context, req *types.QueryPendingQueriesRequest) (*types.QueryPendingQueriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	pendingQueries := []types.Query{}
	for _, query := range k.AllQueries(ctx) {
		if query.RequestSent {
			pendingQueries = append(pendingQueries, query)
		}
	}

	return &types.QueryPendingQueriesResponse{PendingQueries: pendingQueries}, nil
}

// Queries Query all queries that have been requested but have not received a response.
func (k Keeper) Queries(c context.Context, req *types.QueryQueriesRequest) (*types.QueryQueriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	queries := k.AllQueries(ctx)
	return &types.QueryQueriesResponse{Queries: queries}, nil
}
