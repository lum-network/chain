package keeper

import (
	"fmt"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/lum-network/chain/utils"
	"github.com/tendermint/tendermint/libs/log"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/beam/types"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		storeKey      storetypes.StoreKey
		memKey        storetypes.StoreKey
		AuthKeeper    authkeeper.AccountKeeper
		BankKeeper    bankkeeper.Keeper
		StakingKeeper stakingkeeper.Keeper
	}
)

// NewKeeper Create a new keeper instance and return the pointer
func NewKeeper(cdc codec.BinaryCodec, storeKey, memKey storetypes.StoreKey, auth authkeeper.AccountKeeper, bank bankkeeper.Keeper, sk stakingkeeper.Keeper) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		AuthKeeper:    auth,
		BankKeeper:    bank,
		StakingKeeper: sk,
	}
}

// Logger Return a keeper logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetBeamAccount Return the beam module account interface
func (k Keeper) GetBeamAccount(ctx sdk.Context) sdk.AccAddress {
	return k.AuthKeeper.GetModuleAddress(types.ModuleName)
}

// CreateBeamModuleAccount create the module account
func (k Keeper) CreateBeamModuleAccount(ctx sdk.Context, amount sdk.Coin) {
	moduleAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter)
	k.AuthKeeper.SetModuleAccount(ctx, moduleAcc)

	if err := k.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount)); err != nil {
		panic(err)
	}
}

// GetBeamAccountBalance gets the airdrop coin balance of module account
func (k Keeper) GetBeamAccountBalance(ctx sdk.Context) sdk.Coin {
	moduleAccAddr := k.GetBeamAccount(ctx)
	params := k.StakingKeeper.GetParams(ctx)
	return k.BankKeeper.GetBalance(ctx, moduleAccAddr, params.GetBondDenom())
}

// moveCoinsToModuleAccount This moves coins from a given address to the beam module account
func (k Keeper) moveCoinsToModuleAccount(ctx sdk.Context, account sdk.AccAddress, amount sdk.Coin) error {
	if err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, sdk.NewCoins(amount)); err != nil {
		return err
	}
	return nil
}

// moveCoinsToAccount This moves coins from the beam module account to a end user account
func (k Keeper) moveCoinsToAccount(ctx sdk.Context, account sdk.AccAddress, amount sdk.Coin) error {
	if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, account, sdk.NewCoins(amount)); err != nil {
		return err
	}
	return nil
}

// InsertOpenBeamQueue Insert a beam ID inside the active beam queue
func (k Keeper) InsertOpenBeamQueue(ctx sdk.Context, beamID string) {
	store := ctx.KVStore(k.storeKey)
	bz := types.StringKeyToBytes(beamID)
	store.Set(types.GetOpenBeamQueueKey(beamID), bz)
}

// RemoveFromOpenBeamQueue Remove a beam ID from the active beam queue
func (k Keeper) RemoveFromOpenBeamQueue(ctx sdk.Context, beamID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetOpenBeamQueueKey(beamID))
}

// GetBeamIDsFromBlockQueue Return a slice of beam IDs for a given height
func (k Keeper) GetBeamIDsFromBlockQueue(ctx sdk.Context, height int) []string {
	// Acquire the store and key instance
	store := ctx.KVStore(k.storeKey)
	key := types.GetOpenBeamsByBlockQueueKey(height)

	// If key does not exists, return an empty array
	if !store.Has(key) {
		return []string{}
	}

	// Get the content
	content := store.Get(key)
	ids := strings.Split(types.BytesKeyToString(content), types.MemStoreQueueSeparator)
	return ids
}

// InsertOpenBeamByBlockQueue Insert a beam ID inside the by-block store entry
func (k Keeper) InsertOpenBeamByBlockQueue(ctx sdk.Context, height int, beamID string) {
	// Acquire the store and key instance
	store := ctx.KVStore(k.storeKey)
	key := types.GetOpenBeamsByBlockQueueKey(height)

	// Does it exist? If not, create the entry
	if exists := store.Has(key); !exists {
		dest := strings.Join([]string{beamID}, types.MemStoreQueueSeparator)
		store.Set(key, types.StringKeyToBytes(dest))
		return
	}

	// Otherwise, append the content
	content := store.Get(key)
	ids := strings.Split(types.BytesKeyToString(content), types.MemStoreQueueSeparator)
	ids = append(ids, beamID)

	// Put it back
	dest := strings.Join(ids, types.MemStoreQueueSeparator)
	store.Set(key, types.StringKeyToBytes(dest))
}

// RemoveFromOpenBeamByBlockQueue Remove a beam ID from the active beam queue by its height
func (k Keeper) RemoveFromOpenBeamByBlockQueue(ctx sdk.Context, height int, beamID string) {
	// Acquire the store and key instance
	store := ctx.KVStore(k.storeKey)
	key := types.GetOpenBeamsByBlockQueueKey(height)

	// Does it exist? If not, create the entry
	if exists := store.Has(key); !exists {
		return
	}

	// Remove from the content
	content := store.Get(key)
	ids := strings.Split(types.BytesKeyToString(content), types.MemStoreQueueSeparator)
	ids = utils.RemoveFromArray(ids, beamID)

	// If there is no more ID inside the slice, just delete the key
	if len(ids) <= 0 {
		store.Delete(key)
		return
	}

	// Put it back
	dest := strings.Join(ids, types.MemStoreQueueSeparator)
	store.Set(key, types.StringKeyToBytes(dest))
}

// InsertClosedBeamQueue Insert a beam ID inside the closed beam queue
func (k Keeper) InsertClosedBeamQueue(ctx sdk.Context, beamID string) {
	store := ctx.KVStore(k.storeKey)
	bz := types.StringKeyToBytes(beamID)
	store.Set(types.GetClosedBeamQueueKey(beamID), bz)
}

// RemoveFromClosedBeamQueue Remove a beam ID from the closed beam queue
func (k Keeper) RemoveFromClosedBeamQueue(ctx sdk.Context, beamID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetClosedBeamQueueKey(beamID))
}

// GetBeam Return a beam instance for the given key
func (k Keeper) GetBeam(ctx sdk.Context, key string) (types.Beam, error) {
	// Acquire the store instance
	store := ctx.KVStore(k.storeKey)

	// Acquire the data stream
	bz := store.Get(types.GetBeamKey(key))
	if bz == nil {
		return types.Beam{}, sdkerrors.Wrapf(types.ErrBeamNotFound, "beam not found: %s", key)
	}

	// Acquire the beam instance and return
	var beam types.Beam
	err := k.cdc.Unmarshal(bz, &beam)
	if err != nil {
		return types.Beam{}, err
	}
	return beam, nil
}

// ListBeams Return a list of in store beams
func (k Keeper) ListBeams(ctx sdk.Context) (beams []*types.Beam) {
	k.IterateBeams(ctx, func(beam types.Beam) bool {
		beams = append(beams, &beam)
		return false
	})
	return
}

// ListBeamsFromOldOpenQueue Return a list of in store queue beams
func (k Keeper) ListBeamsFromOldOpenQueue(ctx sdk.Context) (beams []*types.Beam) {
	k.IterateOpenBeamsQueue(ctx, func(beam types.Beam) bool {
		beams = append(beams, &beam)
		return false
	})
	return
}

// ListBeamsFromClosedQueue Return a list of in store queue beams
func (k Keeper) ListBeamsFromClosedQueue(ctx sdk.Context) (beams []*types.Beam) {
	k.IterateClosedBeamsQueue(ctx, func(beam types.Beam) bool {
		beams = append(beams, &beam)
		return false
	})
	return
}

func (k Keeper) ListBeamsFromOpenQueue(ctx sdk.Context) (beams []*types.Beam) {
	k.IterateOpenBeamsByBlockQueue(ctx, func(beam types.Beam) bool {
		beams = append(beams, &beam)
		return false
	})
	return
}

// HasBeam Check if a beam instance exists or not (by its key)
func (k Keeper) HasBeam(ctx sdk.Context, beamID string) bool {
	// Acquire the store instance
	store := ctx.KVStore(k.storeKey)

	// Return the presence boolean
	return store.Has(types.GetBeamKey(beamID))
}

// SetBeam Replace the beam at the specified "id" position
func (k Keeper) SetBeam(ctx sdk.Context, beamID string, beam *types.Beam) {
	// Acquire the store instance
	store := ctx.KVStore(k.storeKey)

	// Encode the beam
	encodedBeam := k.cdc.MustMarshal(beam)

	// Update in store
	store.Set(types.GetBeamKey(beamID), encodedBeam)
}

// OpenBeam Create a new beam instance
func (k Keeper) OpenBeam(ctx sdk.Context, msg types.MsgOpenBeam) error {
	// Make sure the ID is in the correct format
	if strings.Contains(msg.GetId(), types.MemStoreQueueSeparator) {
		return types.ErrBeamIdContainsForbiddenChar
	}

	// If the generated ID already exists, refuse the payload
	if k.HasBeam(ctx, msg.GetId()) {
		return types.ErrBeamAlreadyExists
	}

	// Acquire the staking params for default bond denom
	params := k.StakingKeeper.GetParams(ctx)

	var beam = &types.Beam{
		CreatorAddress: msg.GetCreatorAddress(),
		Id:             msg.GetId(),
		Secret:         msg.GetSecret(),
		Status:         types.BeamState_StateOpen,
		Amount:         sdk.NewCoin(params.GetBondDenom(), sdk.NewInt(0)),
		FundsWithdrawn: false,
		Claimed:        false,
		HideContent:    false,
		CancelReason:   "",
		Schema:         msg.GetSchema(),
		Data:           msg.GetData(),
		CreatedAt:      ctx.BlockTime(),
	}

	if msg.GetAmount() != nil && msg.GetAmount().IsPositive() {
		beam.Amount = *msg.GetAmount()
	}

	// If the payload includes an owner field, we auto claim it
	if len(msg.GetClaimAddress()) > 0 {
		beam.ClaimAddress = msg.GetClaimAddress()
		beam.Claimed = true
	}

	if msg.GetClosesAtBlock() > 0 {
		if int(msg.GetClosesAtBlock()) <= int(ctx.BlockHeight()) {
			return types.ErrBeamAutoCloseInThePast
		}
		beam.ClosesAtBlock = msg.GetClosesAtBlock()
	}

	if msg.GetClaimExpiresAtBlock() > 0 {
		beam.ClaimExpiresAtBlock = msg.GetClaimExpiresAtBlock()
	}

	// Only try to process coins move if present
	if msg.GetAmount() != nil && msg.GetAmount().IsPositive() {
		creatorAddress, err := sdk.AccAddressFromBech32(msg.GetCreatorAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}
		err = k.moveCoinsToModuleAccount(ctx, creatorAddress, *msg.GetAmount())
		if err != nil {
			return err
		}
	}

	k.SetBeam(ctx, beam.GetId(), beam)

	// If the beam is actually intended to auto close, we put it inside the by-block queue
	if beam.GetClosesAtBlock() > 0 {
		k.InsertOpenBeamByBlockQueue(ctx, int(msg.GetClosesAtBlock()), beam.GetId())
	}

	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeOpenBeam, sdk.NewAttribute(types.AttributeKeyOpener, msg.GetCreatorAddress())),
	})
	return nil
}

// UpdateBeamStatus This process a beam close request, but its also pass over checks.
// You should not use this directly but rather prefer the UpdateBeam method.
func (k Keeper) UpdateBeamStatus(ctx sdk.Context, beamID string, newStatus types.BeamState) error {
	beam, err := k.GetBeam(ctx, beamID)
	if err != nil {
		return err
	}

	switch newStatus {
	case types.BeamState_StateClosed:
		beam.Status = types.BeamState_StateClosed
		beam.ClosedAt = ctx.BlockTime()

		// Transfer funds only if the beam has been claimed already
		if beam.GetClaimed() && beam.GetFundsWithdrawn() == false {
			claimerAddress, err := sdk.AccAddressFromBech32(beam.GetClaimAddress())
			if err != nil {
				return sdkerrors.ErrInvalidAddress
			}

			if err = k.moveCoinsToAccount(ctx, claimerAddress, beam.GetAmount()); err != nil {
				return err
			}
			beam.FundsWithdrawn = true
		}

		// Update the queues
		if beam.GetClosesAtBlock() > 0 {
			k.RemoveFromOpenBeamByBlockQueue(ctx, int(beam.GetClosesAtBlock()), beam.GetId())
		}
		k.InsertClosedBeamQueue(ctx, beam.GetId())
		break

	case types.BeamState_StateCanceled:
		beam.Status = types.BeamState_StateCanceled
		beam.ClosedAt = ctx.BlockTime()

		// Refund every cent
		creatorAddress, err := sdk.AccAddressFromBech32(beam.GetCreatorAddress())
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Cannot acquire creator address")
		}

		if err = k.moveCoinsToAccount(ctx, creatorAddress, beam.GetAmount()); err != nil {
			return err
		}

		// Update the queues
		if beam.GetClosesAtBlock() > 0 {
			k.RemoveFromOpenBeamByBlockQueue(ctx, int(beam.GetClosesAtBlock()), beam.GetId())
		}
		k.InsertClosedBeamQueue(ctx, beam.GetId())
		break
	}

	k.SetBeam(ctx, beam.GetId(), &beam)

	return nil
}

// UpdateBeam Update a beam instance and proceeds any require state machine update
func (k Keeper) UpdateBeam(ctx sdk.Context, msg types.MsgUpdateBeam) error {
	// Does the beam exists?
	if !k.HasBeam(ctx, msg.Id) {
		return types.ErrBeamNotFound
	}

	// Acquire the beam instance
	beam, err := k.GetBeam(ctx, msg.Id)
	if err != nil {
		return err
	}

	// Is the beam still updatable
	if beam.GetStatus() != types.BeamState_StateOpen {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Beam is closed and thus cannot be updated")
	}

	// Make sure transaction signer is authorized
	if beam.GetCreatorAddress() != msg.GetUpdaterAddress() {
		return types.ErrBeamNotAuthorized
	}

	// First update the metadata before making change since we could want to f.e close but still update metadata
	if msg.GetData() != nil {
		beam.Data = msg.GetData()
	}

	if msg.GetAmount() != nil && msg.GetAmount().IsPositive() {
		updaterAddress, err := sdk.AccAddressFromBech32(msg.GetUpdaterAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}

		err = k.moveCoinsToModuleAccount(ctx, updaterAddress, *msg.GetAmount())
		if err != nil {
			return err
		}

		beam.Amount = beam.GetAmount().Add(*msg.GetAmount())
	}

	if len(msg.GetClaimAddress()) > 0 {
		beam.ClaimAddress = msg.GetClaimAddress()
		beam.Claimed = true
	}

	if msg.GetClosesAtBlock() > 0 {
		beam.ClosesAtBlock = msg.GetClosesAtBlock()
	}

	if msg.GetClaimExpiresAtBlock() > 0 {
		beam.ClaimExpiresAtBlock = msg.GetClaimExpiresAtBlock()
	}

	if msg.GetHideContent() != beam.GetHideContent() {
		beam.HideContent = msg.GetHideContent()
	}

	if msg.GetCancelReason() != beam.GetCancelReason() {
		beam.CancelReason = msg.GetCancelReason()
	}

	if msg.GetHideContent() != beam.GetHideContent() {
		beam.HideContent = msg.GetHideContent()
	}
	k.SetBeam(ctx, beam.GetId(), &beam)

	// We then check the status and return if required
	if msg.GetStatus() != types.BeamState_StateUnspecified {
		err = k.UpdateBeamStatus(ctx, beam.GetId(), msg.GetStatus())
		if err != nil {
			return err
		}
	}

	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeUpdateBeam, sdk.NewAttribute(types.AttributeKeyUpdater, msg.GetUpdaterAddress())),
	})
	return nil
}

// ClaimBeam Final user endpoint to claim and acquire the money
func (k Keeper) ClaimBeam(ctx sdk.Context, msg types.MsgClaimBeam) error {
	// Does the beam exists?
	if !k.HasBeam(ctx, msg.Id) {
		return types.ErrBeamNotFound
	}

	// Acquire the beam instance
	beam, err := k.GetBeam(ctx, msg.Id)
	if err != nil {
		return err
	}

	// If beam is already claimed, we should not be able to
	if beam.GetClaimed() {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "Beam is already claimed")
	}

	// Make sure transaction signer is authorized
	if utils.CompareHashAndString(beam.Secret, msg.Secret) == false {
		return types.ErrBeamInvalidSecret
	}

	// Acquire the claimer address
	claimerAddress, err := sdk.AccAddressFromBech32(msg.GetClaimerAddress())
	if err != nil {
		return sdkerrors.ErrInvalidAddress
	}

	// Transfer funds only if beam is already closed
	if beam.GetStatus() == types.BeamState_StateClosed && beam.GetFundsWithdrawn() == false {
		if beam.GetAmount().IsPositive() {
			if err = k.moveCoinsToAccount(ctx, claimerAddress, beam.GetAmount()); err != nil {
				return err
			}
			beam.FundsWithdrawn = true
		}
	}

	// Update beam status
	beam.Claimed = true
	beam.ClaimAddress = msg.GetClaimerAddress()
	k.SetBeam(ctx, msg.Id, &beam)

	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeClaimBeam, sdk.NewAttribute(types.AttributeKeyClaimer, msg.GetClaimerAddress())),
	})
	return nil
}

// IterateBeams Iterate over the whole beam queue
func (k Keeper) IterateBeams(ctx sdk.Context, cb func(beam types.Beam) (stop bool)) {
	iterator := k.BeamsIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var beam types.Beam
		err := k.cdc.Unmarshal(iterator.Value(), &beam)
		if err != nil {
			panic(err)
		}
		if cb(beam) {
			break
		}
	}
}

// IterateOpenBeamsQueue Iterate over the open only beams queue
func (k Keeper) IterateOpenBeamsQueue(ctx sdk.Context, cb func(beam types.Beam) (stop bool)) {
	iterator := k.OpenBeamsQueueIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		beam, error := k.GetBeam(ctx, types.BytesKeyToString(types.SplitBeamKey(iterator.Key())))
		if error != nil {
			panic(error)
		}

		if cb(beam) {
			break
		}
	}
}

// IterateOpenBeamsByBlockQueue Iterate over the open by block beams queue
func (k Keeper) IterateOpenBeamsByBlockQueue(ctx sdk.Context, cb func(beam types.Beam) (stop bool)) {
	iterator := k.OpenBeamsByBlockQueueIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		ids := strings.Split(types.BytesKeyToString(iterator.Value()), types.MemStoreQueueSeparator)
		for _, id := range ids {
			beam, error := k.GetBeam(ctx, id)
			if error != nil {
				panic(error)
			}

			if cb(beam) {
				break
			}
		}
	}
}

// IterateClosedBeamsQueue Iterate over the closed only beams queue
func (k Keeper) IterateClosedBeamsQueue(ctx sdk.Context, cb func(beam types.Beam) (stop bool)) {
	iterator := k.ClosedBeamsQueueIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		beam, error := k.GetBeam(ctx, types.BytesKeyToString(types.SplitBeamKey(iterator.Key())))
		if error != nil {
			panic(error)
		}

		if cb(beam) {
			break
		}
	}
}

// BeamsIterator Return a ready to use iterator for the whole beams queue
func (k Keeper) BeamsIterator(ctx sdk.Context) sdk.Iterator {
	kvStore := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(kvStore, types.BeamsPrefix)
}

// OpenBeamsQueueIterator Return a ready to use iterator for the open only beams queue
func (k Keeper) OpenBeamsQueueIterator(ctx sdk.Context) sdk.Iterator {
	kvStore := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(kvStore, types.OpenBeamsQueuePrefix)
}

// ClosedBeamsQueueIterator Return a ready to use iterator for the closed only beams queue
func (k Keeper) ClosedBeamsQueueIterator(ctx sdk.Context) sdk.Iterator {
	kvStore := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(kvStore, types.ClosedBeamsQueuePrefix)
}

// OpenBeamsByBlockQueueIterator Return a ready to use iterator for the open by block only beams queue
func (k Keeper) OpenBeamsByBlockQueueIterator(ctx sdk.Context) sdk.Iterator {
	kvStore := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(kvStore, types.OpenBeamsByBlockQueuePrefix)
}
