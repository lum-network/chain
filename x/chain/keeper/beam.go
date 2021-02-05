package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sandblockio/chain/x/chain/types"
)

// GetBeam Return a beam instance for the given key
func (k Keeper) GetBeam(ctx sdk.Context, key string) types.Beam {
	// Acquire the store instance
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BeamKey))

	// Acquire the beam instance and return
	var beam types.Beam
	k.cdc.MustUnmarshalBinaryBare(store.Get(types.KeyPrefix(types.BeamKey+key)), &beam)
	return beam
}

// OpenBeam Create a new beam instance
func (k Keeper) OpenBeam(ctx sdk.Context, msg types.MsgOpenBeam) {
	// Generate the random id
	id := GenerateSecureToken(10)

	// If it already exists, call the same function again
	if k.HasBeam(ctx, id) {
		k.OpenBeam(ctx, msg)
		return
	}

	// Create the beam payload
	var beam = types.Beam{
		Creator: msg.Creator,
		Id:      id,
		Secret:  msg.Secret,
		Amount:  msg.Amount,
	}

	// Acquire the store instance
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BeamKey))

	// Construct our new beam key
	key := types.KeyPrefix(types.BeamKey + beam.Id)

	// Marshal the beam payload to be store-compatible
	value := k.cdc.MustMarshalBinaryBare(&beam)

	// Save the payload to the store
	store.Set(key, value)
}

// HasBeam Check if a beam instance exists or not (by its key)
func (k Keeper) HasBeam(ctx sdk.Context, id string) bool {
	// Acquire the store instance
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BeamKey))

	// Return the presence boolean
	return store.Has(types.KeyPrefix(types.BeamKey + id))
}
