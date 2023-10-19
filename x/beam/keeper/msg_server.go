package keeper

import (
	"context"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/lum-network/chain/utils"
	"github.com/lum-network/chain/x/beam/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// OpenBeam Create a new beam instance
func (k Keeper) OpenBeam(goCtx context.Context, msg *types.MsgOpenBeam) (*types.MsgOpenBeamResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Make sure the ID is in the correct format
	if strings.Contains(msg.GetId(), types.MemStoreQueueSeparator) {
		return nil, types.ErrBeamIdContainsForbiddenChar
	}

	// If the generated ID already exists, refuse the payload
	if k.HasBeam(ctx, msg.GetId()) {
		return nil, types.ErrBeamAlreadyExists
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
			return nil, types.ErrBeamAutoCloseInThePast
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
			return nil, sdkerrors.ErrInvalidAddress
		}
		err = k.moveCoinsToModuleAccount(ctx, creatorAddress, *msg.GetAmount())
		if err != nil {
			return nil, err
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
	return &types.MsgOpenBeamResponse{}, nil
}

// UpdateBeam Update a beam instance and proceeds any require state machine update
func (k Keeper) UpdateBeam(goCtx context.Context, msg *types.MsgUpdateBeam) (*types.MsgUpdateBeamResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Does the beam exists?
	if !k.HasBeam(ctx, msg.Id) {
		return nil, types.ErrBeamNotFound
	}

	// Acquire the beam instance
	beam, err := k.GetBeam(ctx, msg.Id)
	if err != nil {
		return nil, err
	}

	// Is the beam still updatable
	if beam.GetStatus() != types.BeamState_StateOpen {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Beam is closed and thus cannot be updated")
	}

	// Make sure transaction signer is authorized
	if beam.GetCreatorAddress() != msg.GetUpdaterAddress() {
		return nil, types.ErrBeamNotAuthorized
	}

	// First update the metadata before making change since we could want to f.e close but still update metadata
	if msg.GetData() != nil {
		beam.Data = msg.GetData()
	}

	if msg.GetAmount() != nil && msg.GetAmount().IsPositive() {
		updaterAddress, err := sdk.AccAddressFromBech32(msg.GetUpdaterAddress())
		if err != nil {
			return nil, sdkerrors.ErrInvalidAddress
		}

		err = k.moveCoinsToModuleAccount(ctx, updaterAddress, *msg.GetAmount())
		if err != nil {
			return nil, err
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
			return nil, err
		}
	}

	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeUpdateBeam, sdk.NewAttribute(types.AttributeKeyUpdater, msg.GetUpdaterAddress())),
	})
	return &types.MsgUpdateBeamResponse{}, nil
}

// ClaimBeam Final user endpoint to claim and acquire the money
func (k Keeper) ClaimBeam(goCtx context.Context, msg *types.MsgClaimBeam) (*types.MsgClaimBeamResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Does the beam exists?
	if !k.HasBeam(ctx, msg.Id) {
		return nil, types.ErrBeamNotFound
	}

	// Acquire the beam instance
	beam, err := k.GetBeam(ctx, msg.Id)
	if err != nil {
		return nil, err
	}

	// If beam is already claimed, we should not be able to
	if beam.GetClaimed() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "Beam is already claimed")
	}

	// Make sure transaction signer is authorized
	if !utils.CompareHashAndString(beam.Secret, msg.Secret) {
		return nil, types.ErrBeamInvalidSecret
	}

	// Acquire the claimer address
	claimerAddress, err := sdk.AccAddressFromBech32(msg.GetClaimerAddress())
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}

	// Transfer funds only if beam is already closed
	if beam.GetStatus() == types.BeamState_StateClosed && !beam.GetFundsWithdrawn() {
		if beam.GetAmount().IsPositive() {
			if err = k.moveCoinsToAccount(ctx, claimerAddress, beam.GetAmount()); err != nil {
				return nil, err
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
	return &types.MsgClaimBeamResponse{}, nil
}
