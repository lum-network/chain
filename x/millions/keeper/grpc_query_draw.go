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

func (k queryServer) Draws(goCtx context.Context, req *types.QueryDrawsRequest) (*types.QueryDrawsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	drawStore := prefix.NewStore(store, types.DrawPrefix)

	// Make the paginated query
	var draws []types.Draw
	pageRes, err := query.Paginate(drawStore, req.Pagination, func(key, value []byte) error {
		var draw types.Draw
		if err := k.cdc.Unmarshal(value, &draw); err != nil {
			return err
		}

		draws = append(draws, draw)
		return nil
	})
	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDrawsResponse{Draws: draws, Pagination: pageRes}, nil
}

func (k queryServer) PoolDraws(goCtx context.Context, req *types.QueryPoolDrawsRequest) (*types.QueryDrawsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !k.HasPool(ctx, req.PoolId) {
		return nil, status.Error(codes.NotFound, types.ErrPoolNotFound.Error())
	}

	store := ctx.KVStore(k.storeKey)
	drawStore := prefix.NewStore(store, types.GetPoolDrawsKey(req.GetPoolId()))

	// Make the paginated query
	var draws []types.Draw
	pageRes, err := query.Paginate(drawStore, req.Pagination, func(key, value []byte) error {
		var draw types.Draw
		if err := k.cdc.Unmarshal(value, &draw); err != nil {
			return err
		}

		draws = append(draws, draw)
		return nil
	})
	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDrawsResponse{Draws: draws, Pagination: pageRes}, nil
}

func (k queryServer) PoolDraw(goCtx context.Context, req *types.QueryPoolDrawRequest) (*types.QueryDrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	draw, err := k.GetPoolDraw(ctx, req.GetPoolId(), req.GetDrawId())
	if err != nil {
		return &types.QueryDrawResponse{}, err
	}

	return &types.QueryDrawResponse{Draw: &draw}, nil
}
