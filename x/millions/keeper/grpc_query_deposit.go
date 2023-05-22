package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/lum-network/chain/x/millions/types"
)

func (k queryServer) Deposits(goCtx context.Context, req *types.QueryDepositsRequest) (*types.QueryDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	depositStore := prefix.NewStore(store, types.GetDepositsKey())

	// Make the paginated query
	var deposits []types.Deposit
	pageRes, err := query.Paginate(depositStore, req.Pagination, func(key []byte, value []byte) error {
		var deposit types.Deposit
		if err := k.cdc.Unmarshal(value, &deposit); err != nil {
			return err
		}

		deposits = append(deposits, deposit)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDepositsResponse{Deposits: deposits, Pagination: pageRes}, nil
}

func (k queryServer) PoolDeposits(goCtx context.Context, req *types.QueryPoolDepositsRequest) (*types.QueryDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	depositStore := prefix.NewStore(store, types.GetPoolDepositsKey(req.GetPoolId()))

	if !k.HasPool(ctx, req.PoolId) {
		return nil, status.Error(codes.NotFound, types.ErrPoolNotFound.Error())
	}

	// Make the paginated query
	var deposits []types.Deposit
	pageRes, err := query.Paginate(depositStore, req.Pagination, func(key []byte, value []byte) error {
		var deposit types.Deposit
		if err := k.cdc.Unmarshal(value, &deposit); err != nil {
			return err
		}

		deposits = append(deposits, deposit)
		return nil
	})

	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDepositsResponse{Deposits: deposits, Pagination: pageRes}, nil
}

func (k queryServer) PoolDeposit(goCtx context.Context, req *types.QueryPoolDepositRequest) (*types.QueryDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	deposit, err := k.GetPoolDeposit(ctx, req.GetPoolId(), req.GetDepositId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryDepositResponse{Deposit: deposit}, nil
}

func (k queryServer) AccountDeposits(goCtx context.Context, req *types.QueryAccountDepositsRequest) (*types.QueryDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	addr, err := sdk.AccAddressFromBech32(req.GetDepositorAddress())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	depositStore := prefix.NewStore(store, types.GetAccountDepositsKey(addr))

	// Make the paginated query
	var deposits []types.Deposit
	pageRes, err := query.Paginate(depositStore, req.Pagination, func(key []byte, value []byte) error {
		var deposit types.Deposit
		if err := k.cdc.Unmarshal(value, &deposit); err != nil {
			return err
		}

		deposits = append(deposits, deposit)
		return nil
	})

	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDepositsResponse{Deposits: deposits, Pagination: pageRes}, nil
}

func (k queryServer) AccountPoolDeposits(goCtx context.Context, req *types.QueryAccountPoolDepositsRequest) (*types.QueryDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	addr, err := sdk.AccAddressFromBech32(req.GetDepositorAddress())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !k.HasPool(ctx, req.PoolId) {
		return nil, status.Error(codes.NotFound, types.ErrPoolNotFound.Error())
	}

	store := ctx.KVStore(k.storeKey)
	depositStore := prefix.NewStore(store, types.GetAccountPoolDepositsKey(addr, req.PoolId))

	// Make the paginated query
	var deposits []types.Deposit
	pageRes, err := query.Paginate(depositStore, req.Pagination, func(key []byte, value []byte) error {
		var deposit types.Deposit
		if err := k.cdc.Unmarshal(value, &deposit); err != nil {
			return err
		}

		deposits = append(deposits, deposit)
		return nil
	})

	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDepositsResponse{Deposits: deposits, Pagination: pageRes}, nil
}
