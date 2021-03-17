package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sandblockio/chain/x/faucet/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MintAndSend Mint a limited amount of coins and transfer them to an address
// This method is disabled in production mode, and should be used carefully
func (k Keeper) MintAndSend(c context.Context, req *types.QueryMintAndSendRequest) (*types.QueryMintAndSendResponse, error) {
	// Is the payload valid ?
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Acquire the context instance
	ctx := sdk.UnwrapSDKContext(c)

	// Extract account address
	accAddress, err := sdk.AccAddressFromBech32(req.Minter)
	if err != nil {
		return nil, err
	}

	// Try to proceed, if not return an error
	err = k.Mint(ctx, accAddress, req.MintTime)
	if err != nil {
		return nil, err
	}

	return &types.QueryMintAndSendResponse{
	}, nil
}
