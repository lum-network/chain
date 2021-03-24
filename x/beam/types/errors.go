package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/chain module sentinel errors
var (
	ErrBeamNotFound = sdkerrors.Register(ModuleName, 1100, "Beam does not exists")
	ErrBeamNotAuthorized = sdkerrors.Register(ModuleName, 1101, "This beam does not belong to you")
	ErrBeamAlreadyExists = sdkerrors.Register(ModuleName, 1102, "This beam ID already exists")
)
