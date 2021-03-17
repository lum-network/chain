package keeper

import (
	// this line is used by starport scaffolding # 1

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/lum-network/chain/x/chain/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier Return a querier instance
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var (
			res []byte
			err error
		)

		switch path[0] {
		case types.QueryFetchBeams:
			return fetchBeams(ctx, k, legacyQuerierCdc)

		case types.QueryGetBeam:
			return getBeam(ctx, path[1], k, legacyQuerierCdc)

		default:
			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}

		return res, err
	}
}
