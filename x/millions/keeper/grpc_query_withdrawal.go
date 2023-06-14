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

func (k queryServer) Withdrawals(goCtx context.Context, req *types.QueryWithdrawalsRequest) (*types.QueryWithdrawalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	withdrawalStore := prefix.NewStore(store, types.GetWithdrawalsKey())

	// Make the paginated query
	var withdrawals []types.Withdrawal
	pageRes, err := query.Paginate(withdrawalStore, req.Pagination, func(key, value []byte) error {
		var withdrawal types.Withdrawal
		if err := k.cdc.Unmarshal(value, &withdrawal); err != nil {
			return err
		}

		withdrawals = append(withdrawals, withdrawal)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryWithdrawalsResponse{Withdrawals: withdrawals, Pagination: pageRes}, nil
}

func (k queryServer) PoolWithdrawals(goCtx context.Context, req *types.QueryPoolWithdrawalsRequest) (*types.QueryWithdrawalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !k.HasPool(ctx, req.PoolId) {
		return nil, status.Error(codes.NotFound, types.ErrPoolNotFound.Error())
	}

	store := ctx.KVStore(k.storeKey)
	withdrawalStore := prefix.NewStore(store, types.GetPoolWithdrawalsKey(req.GetPoolId()))

	// Make the paginated query
	var withdrawals []types.Withdrawal
	pageRes, err := query.Paginate(withdrawalStore, req.Pagination, func(key, value []byte) error {
		var withdrawal types.Withdrawal
		if err := k.cdc.Unmarshal(value, &withdrawal); err != nil {
			return err
		}

		withdrawals = append(withdrawals, withdrawal)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryWithdrawalsResponse{Withdrawals: withdrawals, Pagination: pageRes}, nil
}

func (k queryServer) PoolWithdrawal(goCtx context.Context, req *types.QueryPoolWithdrawalRequest) (*types.QueryWithdrawalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !k.HasPool(ctx, req.PoolId) {
		return nil, status.Error(codes.NotFound, types.ErrPoolNotFound.Error())
	}

	withdrawal, err := k.GetPoolWithdrawal(ctx, req.GetPoolId(), req.GetWithdrawalId())
	if err != nil {
		return &types.QueryWithdrawalResponse{}, err
	}

	return &types.QueryWithdrawalResponse{Withdrawal: withdrawal}, nil
}

func (k queryServer) AccountWithdrawals(goCtx context.Context, req *types.QueryAccountWithdrawalsRequest) (*types.QueryWithdrawalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	addr, err := sdk.AccAddressFromBech32(req.GetDepositorAddress())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	withdrawalStore := prefix.NewStore(store, types.GetAccountWithdrawalsKey(addr))

	// Make the paginated query
	var withdrawals []types.Withdrawal
	pageRes, err := query.Paginate(withdrawalStore, req.Pagination, func(key, value []byte) error {
		var withdrawal types.Withdrawal
		if err := k.cdc.Unmarshal(value, &withdrawal); err != nil {
			return err
		}

		withdrawals = append(withdrawals, withdrawal)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryWithdrawalsResponse{Withdrawals: withdrawals, Pagination: pageRes}, nil
}

func (k queryServer) AccountPoolWithdrawals(goCtx context.Context, req *types.QueryAccountPoolWithdrawalsRequest) (*types.QueryWithdrawalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

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
	withdrawalStore := prefix.NewStore(store, types.GetAccountPoolWithdrawalsKey(addr, req.GetPoolId()))

	// Make the paginated query
	var withdrawals []types.Withdrawal
	pageRes, err := query.Paginate(withdrawalStore, req.Pagination, func(key, value []byte) error {
		var withdrawal types.Withdrawal
		if err := k.cdc.Unmarshal(value, &withdrawal); err != nil {
			return err
		}

		withdrawals = append(withdrawals, withdrawal)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryWithdrawalsResponse{Withdrawals: withdrawals, Pagination: pageRes}, nil
}
