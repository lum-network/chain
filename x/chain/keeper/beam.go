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

func (k Keeper) ListBeams(ctx sdk.Context) (msgs []types.Beam) {
	// Acquire the store instance
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BeamKey))

	// Define the iterator
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefix(types.BeamKey))

	// Defer the iterator shutdown
	defer iterator.Close()

	// For each beam, unmarshal and append to return structure
	for ; iterator.Valid(); iterator.Next() {
		var msg types.Beam
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &msg)
		msgs = append(msgs, msg)
	}

	return
}

// HasBeam Check if a beam instance exists or not (by its key)
func (k Keeper) HasBeam(ctx sdk.Context, id string) bool {
	// Acquire the store instance
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BeamKey))

	// Return the presence boolean
	return store.Has(types.KeyPrefix(types.BeamKey + id))
}

// UpdateBeam Replace the beam at the specified "id" position
func (k Keeper) UpdateBeam(ctx sdk.Context, key string, beam types.Beam) {
	// Acquire the store instance
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BeamKey))

	// Encode the beam
	encodedBeam := k.cdc.MustMarshalBinaryBare(&beam)

	// Update in store
	store.Set(types.KeyPrefix(types.BeamKey+key), encodedBeam)
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

func (k Keeper) IncreaseBeam(ctx sdk.Context, msg types.MsgIncreaseBeam) {
	// Does the beam exists?
	if !k.HasBeam(ctx, msg.Id) {
		return
	}

	// Acquire the beam instance
	beam := k.GetBeam(ctx, msg.Id)

	// Make sure transaction signer is authorized
	//TODO: implement

	// Update the value
	beam.Amount += msg.Amount

	// Append to beam logs
	//TODO: implement

	// Update the in-store beam
	k.UpdateBeam(ctx, msg.Id, beam)
}

func (k Keeper) CloseBeam(ctx sdk.Context, msg types.MsgCloseBeam) {
	// Does the beam exists?
	if !k.HasBeam(ctx, msg.Id) {
		return
	}

	// Acquire the beam instance
	// beam := k.GetBeam(ctx, msg.Id)

	// Make sure transaction signer is authorized
	//TODO: implement

	// Proceed money transfer from module account
	//TODO: implement

	// Update the beam status
	//TODO: implement
}

func (k Keeper) CancelBeam(ctx sdk.Context, msg types.MsgCancelBeam) {
	// Does the beam exists?
	if !k.HasBeam(ctx, msg.Id) {
		return
	}

	// Acquire the beam instance
	// beam := k.GetBeam(ctx, msg.Id)

	// Make sure transaction signer is authorized
	//TODO: implement

	// Refund creator
	//TODO: implement

	// Update beam status
	//TODO: implement
}

func (k Keeper) ClaimBeam(ctx sdk.Context, msg types.MsgClaimBeam) {
	// Does the beam exists?
	if !k.HasBeam(ctx, msg.Id) {
		return
	}

	// Acquire the beam instance
	// beam := k.GetBeam(ctx, msg.Id)

	// Make sure transaction signer is authorized
	//TODO: implement

	// Transfer funds
	//TODO: implement

	// Update beam status
	//TODO: implement
}
