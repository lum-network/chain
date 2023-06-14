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

func (k queryServer) Prizes(goCtx context.Context, req *types.QueryPrizesRequest) (*types.QueryPrizesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	prizeStore := prefix.NewStore(store, types.GetPrizesKey())

	// Make the paginated query
	var prizes []types.Prize
	pageRes, err := query.Paginate(prizeStore, req.Pagination, func(key []byte, value []byte) error {
		var prize types.Prize
		if err := k.cdc.Unmarshal(value, &prize); err != nil {
			return err
		}

		prizes = append(prizes, prize)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPrizesResponse{Prizes: prizes, Pagination: pageRes}, nil
}

func (k queryServer) PoolPrizes(goCtx context.Context, req *types.QueryPoolPrizesRequest) (*types.QueryPrizesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !k.HasPool(ctx, req.PoolId) {
		return nil, status.Error(codes.NotFound, types.ErrPoolNotFound.Error())
	}

	store := ctx.KVStore(k.storeKey)
	prizeStore := prefix.NewStore(store, types.GetPoolPrizesKey(req.GetPoolId()))

	// Make the paginated query
	var prizes []types.Prize
	pageRes, err := query.Paginate(prizeStore, req.Pagination, func(key []byte, value []byte) error {
		var prize types.Prize
		if err := k.cdc.Unmarshal(value, &prize); err != nil {
			return err
		}

		prizes = append(prizes, prize)
		return nil
	})
	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPrizesResponse{Prizes: prizes, Pagination: pageRes}, nil
}

func (k queryServer) PoolDrawPrizes(goCtx context.Context, req *types.QueryPoolDrawPrizesRequest) (*types.QueryPrizesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !k.HasPoolDraw(ctx, req.PoolId, req.GetDrawId()) {
		return nil, status.Error(codes.NotFound, types.ErrPoolDrawNotFound.Error())
	}

	store := ctx.KVStore(k.storeKey)
	prizeStore := prefix.NewStore(store, types.GetPoolDrawPrizesKey(req.GetPoolId(), req.GetDrawId()))

	// Make the paginated query
	var prizes []types.Prize
	pageRes, err := query.Paginate(prizeStore, req.Pagination, func(key []byte, value []byte) error {
		var prize types.Prize
		if err := k.cdc.Unmarshal(value, &prize); err != nil {
			return err
		}

		prizes = append(prizes, prize)
		return nil
	})
	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPrizesResponse{Prizes: prizes, Pagination: pageRes}, nil
}

func (k queryServer) PoolDrawPrize(goCtx context.Context, req *types.QueryPoolDrawPrizeRequest) (*types.QueryPrizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	prize, err := k.GetPoolDrawPrize(ctx, req.GetPoolId(), req.GetDrawId(), req.GetPrizeId())
	if err != nil {
		return &types.QueryPrizeResponse{}, err
	}

	return &types.QueryPrizeResponse{Prize: prize}, nil
}

func (k queryServer) AccountPrizes(goCtx context.Context, req *types.QueryAccountPrizesRequest) (*types.QueryPrizesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	addr, err := sdk.AccAddressFromBech32(req.GetWinnerAddress())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	prizeStore := prefix.NewStore(store, types.GetAccountPrizesKey(addr))

	// Make the paginated query
	var prizes []types.Prize
	pageRes, err := query.Paginate(prizeStore, req.Pagination, func(key []byte, value []byte) error {
		var prize types.Prize
		if err := k.cdc.Unmarshal(value, &prize); err != nil {
			return err
		}

		prizes = append(prizes, prize)
		return nil
	})
	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPrizesResponse{Prizes: prizes, Pagination: pageRes}, nil
}

func (k queryServer) AccountPoolPrizes(goCtx context.Context, req *types.QueryAccountPoolPrizesRequest) (*types.QueryPrizesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	addr, err := sdk.AccAddressFromBech32(req.GetWinnerAddress())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !k.HasPool(ctx, req.PoolId) {
		return nil, status.Error(codes.NotFound, types.ErrPoolNotFound.Error())
	}

	store := ctx.KVStore(k.storeKey)
	prizeStore := prefix.NewStore(store, types.GetAccountPoolPrizesKey(addr, req.GetPoolId()))

	// Make the paginated query
	var prizes []types.Prize
	pageRes, err := query.Paginate(prizeStore, req.Pagination, func(key []byte, value []byte) error {
		var prize types.Prize
		if err := k.cdc.Unmarshal(value, &prize); err != nil {
			return err
		}

		prizes = append(prizes, prize)
		return nil
	})
	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPrizesResponse{Prizes: prizes, Pagination: pageRes}, nil
}

func (k queryServer) AccountPoolDrawPrizes(goCtx context.Context, req *types.QueryAccountPoolDrawPrizesRequest) (*types.QueryPrizesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	addr, err := sdk.AccAddressFromBech32(req.GetWinnerAddress())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !k.HasPoolDraw(ctx, req.PoolId, req.DrawId) {
		return nil, status.Error(codes.NotFound, types.ErrPoolDrawNotFound.Error())
	}

	store := ctx.KVStore(k.storeKey)
	prizeStore := prefix.NewStore(store, types.GetAccountPoolDrawPrizesKey(addr, req.GetPoolId(), req.GetDrawId()))

	// Make the paginated query
	var prizes []types.Prize
	pageRes, err := query.Paginate(prizeStore, req.Pagination, func(key []byte, value []byte) error {
		var prize types.Prize
		if err := k.cdc.Unmarshal(value, &prize); err != nil {
			return err
		}

		prizes = append(prizes, prize)
		return nil
	})
	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPrizesResponse{Prizes: prizes, Pagination: pageRes}, nil
}
