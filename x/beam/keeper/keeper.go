package keeper

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/beam/types"
)

type (
	Keeper struct {
		cdc        codec.Codec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		AuthKeeper authkeeper.AccountKeeper
		BankKeeper bankkeeper.Keeper
	}
)

// NewKeeper Create a new keeper instance and return the pointer
func NewKeeper(cdc codec.Codec, storeKey, memKey sdk.StoreKey, auth authkeeper.AccountKeeper, bank bankkeeper.Keeper) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		AuthKeeper: auth,
		BankKeeper: bank,
	}
}

// Logger Return a keeper logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetBeamAccount Return the beam module account interface
func (k Keeper) GetBeamAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.AuthKeeper.GetModuleAccount(ctx, types.ModuleName)
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

// GetBeam Return a beam instance for the given key
func (k Keeper) GetBeam(ctx sdk.Context, key string) (*types.Beam, error) {
	// Acquire the store instance
	store := ctx.KVStore(k.storeKey)

	// Acquire the data stream
	bz := store.Get(types.KeyBeam(key))
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrBeamNotFound, "beam not found: %s", key)
	}

	// Acquire the beam instance and return
	var beam *types.Beam
	k.cdc.MustUnmarshal(bz, beam)
	return beam, nil
}

// ListBeams Return a list of in store beams
func (k Keeper) ListBeams(ctx sdk.Context) (msgs []*types.Beam) {
	// Acquire the store instance
	store := ctx.KVStore(k.storeKey)

	// Define the iterator
	iterator := sdk.KVStorePrefixIterator(store, types.KeyBeam(""))

	// Defer the iterator shutdown
	defer iterator.Close()

	// For each beam, unmarshal and append to return structure
	for ; iterator.Valid(); iterator.Next() {
		var msg *types.Beam
		k.cdc.MustUnmarshal(iterator.Value(), msg)
		msgs = append(msgs, msg)
	}

	return
}

// HasBeam Check if a beam instance exists or not (by its key)
func (k Keeper) HasBeam(ctx sdk.Context, id string) bool {
	// Acquire the store instance
	store := ctx.KVStore(k.storeKey)

	// Return the presence boolean
	return store.Has(types.KeyBeam(id))
}

// SetBeam Replace the beam at the specified "id" position
func (k Keeper) SetBeam(ctx sdk.Context, key string, beam *types.Beam) {
	// Acquire the store instance
	store := ctx.KVStore(k.storeKey)

	// Encode the beam
	encodedBeam := k.cdc.MustMarshal(beam)

	// Update in store
	store.Set(types.KeyBeam(key), encodedBeam)
}

// OpenBeam Create a new beam instance
func (k Keeper) OpenBeam(ctx sdk.Context, msg types.MsgOpenBeam) error {
	// If the generated ID already exists, refuse the payload
	if k.HasBeam(ctx, msg.GetId()) {
		return types.ErrBeamAlreadyExists
	}

	var beam = &types.Beam{
		CreatorAddress: msg.GetCreatorAddress(),
		Id:             msg.GetId(),
		Secret:         msg.GetSecret(),
		Status:         types.BeamState_StateOpen,
		Amount:         sdk.NewCoin("ulum", sdk.NewInt(0)),
		FundsWithdrawn: false,
		Claimed:        false,
		HideContent:    false,
		CancelReason:   "",
		Schema:         msg.GetSchema(),
		Data:           msg.GetData(),
	}

	if msg.GetAmount() != nil && msg.GetAmount().IsPositive() {
		beam.Amount = *msg.GetAmount()
	}

	// If the payload includes a owner field, we auto claim it
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

	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeOpenBeam, sdk.NewAttribute(types.AttributeKeyOpener, msg.GetCreatorAddress())),
	})
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

	// We then check the status and return if required
	if msg.GetStatus() != types.BeamState_StateUnspecified {
		switch msg.GetStatus() {
		case types.BeamState_StateClosed:
			beam.Status = types.BeamState_StateClosed

			if msg.GetHideContent() != beam.GetHideContent() {
				beam.HideContent = msg.GetHideContent()
			}

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
			break

		case types.BeamState_StateCanceled:
			beam.Status = types.BeamState_StateCanceled

			if msg.GetCancelReason() != beam.GetCancelReason() {
				beam.CancelReason = msg.GetCancelReason()
			}

			if msg.GetHideContent() != beam.GetHideContent() {
				beam.HideContent = msg.GetHideContent()
			}

			// Refund every cent
			creatorAddress, err := sdk.AccAddressFromBech32(beam.GetCreatorAddress())
			if err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Cannot acquire creator address")
			}

			if err = k.moveCoinsToAccount(ctx, creatorAddress, beam.GetAmount()); err != nil {
				return err
			}
			break
		}
	}

	k.SetBeam(ctx, beam.GetId(), beam)

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
	if types.CompareHashAndString(beam.Secret, msg.Secret) == false {
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
	k.SetBeam(ctx, msg.Id, beam)

	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeClaimBeam, sdk.NewAttribute(types.AttributeKeyClaimer, msg.GetClaimerAddress())),
	})
	return nil
}
