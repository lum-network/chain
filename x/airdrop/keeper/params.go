package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/airdrop/types"
)

// GetParams Get the in-store params
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.ParamsKey))
	params := types.Params{}
	err := k.cdc.Unmarshal(bz, &params)
	return params, err
}

// SetParams Set the in-store params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set([]byte(types.ParamsKey), bz)
	return nil
}
