package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/lum-network/chain/x/dfract/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) GetDepositsForAddress(c context.Context, req *types.QueryGetDepositsForAddressRequest) (*types.QueryGetDepositsForAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	waitingProposal, _ := k.GetFromWaitingProposalDeposits(ctx, req.Address)
	waitingMint, _ := k.GetFromWaitingMintDeposits(ctx, req.Address)
	minted, _ := k.GetFromMintedDeposits(ctx, req.Address)

	return &types.QueryGetDepositsForAddressResponse{
		WaitingProposalDeposits: &waitingProposal,
		WaitingMintDeposits:     &waitingMint,
		MintedDeposits:          &minted,
	}, nil
}

func (k Keeper) FetchDeposits(c context.Context, req *types.QueryFetchDepositsRequest) (*types.QueryFetchDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	var depositStore prefix.Store
	if req.Type == types.DepositsQueryType_TypeWaitingProposal {
		depositStore = prefix.NewStore(store, types.GetWaitingProposalDepositsKey(""))
	} else if req.Type == types.DepositsQueryType_TypeWaitingMint {
		depositStore = prefix.NewStore(store, types.GetWaitingMintDepositsKey(""))
	} else if req.Type == types.DepositsQueryType_TypeMinted {
		depositStore = prefix.NewStore(store, types.GetMintedDepositsKey(""))
	}

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

	return &types.QueryFetchDepositsResponse{Deposits: deposits, Pagination: pageRes}, nil
}
