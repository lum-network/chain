package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/lum-network/chain/x/beam/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Beams(c context.Context, req *types.QueryFetchBeamsRequest) (*types.QueryFetchBeamsResponse, error) {
	// Is the payload valid ?
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var beams []*types.Beam

	// Acquire the context instance
	ctx := sdk.UnwrapSDKContext(c)

	// Acquire the store instance
	store := ctx.KVStore(k.storeKey)
	beamStore := prefix.NewStore(store, types.BeamKey(""))

	// Make the paginated query
	pageRes, err := query.Paginate(beamStore, req.Pagination, func(key []byte, value []byte) error {
		var beam types.Beam
		if err := k.cdc.Unmarshal(value, &beam); err != nil {
			return err
		}

		beams = append(beams, &beam)
		return nil
	})

	// Was there any error while acquiring the list of beams
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryFetchBeamsResponse{Beams: beams, Pagination: pageRes}, nil
}

func (k Keeper) Beam(c context.Context, req *types.QueryGetBeamRequest) (*types.QueryGetBeamResponse, error) {
	// Is the payload valid?
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var beam types.Beam

	// Acquire the context instance
	ctx := sdk.UnwrapSDKContext(c)

	// Acquire the store instance
	store := ctx.KVStore(k.storeKey)

	// Acquire the beam instance
	k.cdc.MustUnmarshal(store.Get(types.BeamKey(req.Id)), &beam)

	return &types.QueryGetBeamResponse{Beam: &beam}, nil
}
