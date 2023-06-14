package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lum-network/chain/x/dfract/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

func (k queryServer) ModuleAccountBalance(c context.Context, _ *types.QueryModuleAccountBalanceRequest) (*types.QueryModuleAccountBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	moduleAccBal := k.GetModuleAccountBalance(ctx)

	return &types.QueryModuleAccountBalanceResponse{ModuleAccountBalance: moduleAccBal}, nil
}

func (k queryServer) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (k queryServer) GetDepositsForAddress(c context.Context, req *types.QueryGetDepositsForAddressRequest) (*types.QueryGetDepositsForAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	accAddr, err := sdk.AccAddressFromBech32(req.Address)
	if req.Address == "" || err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}

	waitingProposal, _ := k.GetDepositPendingWithdrawal(ctx, accAddr)
	waitingMint, _ := k.GetDepositPendingMint(ctx, accAddr)
	minted, _ := k.GetDepositMinted(ctx, accAddr)

	return &types.QueryGetDepositsForAddressResponse{
		DepositsPendingWithdrawal: &waitingProposal,
		DepositsPendingMint:       &waitingMint,
		DepositsMinted:            &minted,
	}, nil
}

func (k queryServer) FetchDeposits(c context.Context, req *types.QueryFetchDepositsRequest) (*types.QueryFetchDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	var depositStore prefix.Store
	if req.Type == types.DepositsQueryType_TypePendingWithdrawal {
		depositStore = prefix.NewStore(store, types.DepositsPendingWithdrawalPrefix)
	} else if req.Type == types.DepositsQueryType_TypePendingMint {
		depositStore = prefix.NewStore(store, types.DepositsPendingMintPrefix)
	} else if req.Type == types.DepositsQueryType_TypeMinted {
		depositStore = prefix.NewStore(store, types.DepositsMintedPrefix)
	}

	var deposits []types.Deposit
	pageRes, err := query.Paginate(depositStore, req.Pagination, func(key, value []byte) error {
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
