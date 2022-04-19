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
	var beamStore prefix.Store
	if req.State == types.BeamState_StateUnspecified {
		beamStore = prefix.NewStore(store, types.GetBeamKey(""))
	} else if req.State == types.BeamState_StateOpen {
		beamStore = prefix.NewStore(store, types.ClosedBeamsQueuePrefix)
	} else if req.State == types.BeamState_StateClosed {
		if req.Old {
			beamStore = prefix.NewStore(store, types.OpenBeamsQueuePrefix)
		} else {
			beamStore = prefix.NewStore(store, types.OpenBeamsByBlockQueuePrefix)
		}
	}

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
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	beam, err := k.GetBeam(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, types.ErrBeamNotFound.Error())
	}
	return &types.QueryGetBeamResponse{Beam: &beam}, nil
}
