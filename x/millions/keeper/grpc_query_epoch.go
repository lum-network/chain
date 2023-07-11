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

// EpochUnbondings allows to retrive all epoch unbondings per epoch number
func (k queryServer) EpochUnbondings(goCtx context.Context, req *types.QueryEpochUnbondingsRequest) (*types.QueryEpochUnbondingsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	epochUnbondingStore := prefix.NewStore(store, types.EpochUnbondingPrefix)

	// Make the paginated query
	var epochUnbondings []types.EpochUnbonding
	pageRes, err := query.Paginate(epochUnbondingStore, req.Pagination, func(key []byte, value []byte) error {
		var epochUnbonding types.EpochUnbonding
		if err := k.cdc.Unmarshal(value, &epochUnbonding); err != nil {
			return err
		}

		epochUnbondings = append(epochUnbondings, epochUnbonding)
		return nil
	})

	// Was there any error while acquiring the list of pools
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryEpochUnbondingsResponse{EpochUnbondings: epochUnbondings, Pagination: pageRes}, nil
}

// EpochPoolUnbonding allows to retrive the epoch unbonding per epoch number and poolID
func (k queryServer) EpochPoolUnbonding(goCtx context.Context, req *types.QueryEpochPoolUnbondingRequest) (*types.QueryEpochPoolUnbondingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Make sure context request is valid
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !k.HasPool(ctx, req.PoolId) {
		return nil, status.Error(codes.NotFound, types.ErrPoolNotFound.Error())
	}

	poolEpochUnbonding, err := k.GetEpochPoolUnbonding(ctx, req.GetEpochNumber(), req.GetPoolId())
	if err != nil {
		return &types.QueryEpochPoolUnbondingResponse{}, err
	}

	return &types.QueryEpochPoolUnbondingResponse{EpochUnbonding: poolEpochUnbonding}, nil
}
