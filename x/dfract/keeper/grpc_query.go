package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/lum-network/chain/x/dfract/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) GetDepositsForAddress(c context.Context, req *types.QueryGetDepositsForAddressRequest) (*types.QueryGetDepositsForAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	accAddr, err := sdk.AccAddressFromBech32(req.Address)
	if req.Address == "" || err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}

	waitingProposal, _ := k.GetDepositPendingWithdrawal(ctx, accAddr)
	waitingMint, _ := k.GetDepositPendingMint(ctx, accAddr)
	minted, _ := k.GetMintedDeposit(ctx, accAddr)

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
		depositStore = prefix.NewStore(store, types.DepositsPendingWithdrawalPrefix)
	} else if req.Type == types.DepositsQueryType_TypeWaitingMint {
		depositStore = prefix.NewStore(store, types.DepositsPendingMintPrefix)
	} else if req.Type == types.DepositsQueryType_TypeMinted {
		depositStore = prefix.NewStore(store, types.DepositsMintedPrefix)
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
