package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"time"
)

func mintAndSend(ctx sdk.Context, keeper Keeper, id string, mintTime time.Time, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	accAddress, err := sdk.AccAddressFromBech32(id)
	if err != nil {
		return nil, err
	}

	msg := keeper.Mint(ctx, accAddress, mintTime)
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, msg)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
