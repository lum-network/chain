package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsPrefix)
	if bz == nil {
		return types.DefaultParams()
	}
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := params.ValidateBasics(); err != nil {
		panic(err)
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ParamsPrefix, k.cdc.MustMarshal(&params))
}
