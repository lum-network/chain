package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/lum-network/chain/x/millions/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k queryServer) Pools(goCtx context.Context, req *types.QueryPoolsRequest) (*types.QueryPoolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	poolStore := prefix.NewStore(store, types.PoolPrefix)

	// Make the paginated query
	var pools []types.Pool
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(key, value []byte) error {
		var pool types.Pool
		if err := k.cdc.Unmarshal(value, &pool); err != nil {
			return err
		}

		pools = append(pools, pool)
		return nil
	})
	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPoolsResponse{Pools: pools, Pagination: pageRes}, nil
}

func (k queryServer) Pool(goCtx context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	pool, err := k.GetPool(ctx, req.GetPoolId())
	if err != nil {
		return nil, status.Error(codes.NotFound, types.ErrPoolNotFound.Error())
	}
	return &types.QueryPoolResponse{Pool: pool}, nil
}
