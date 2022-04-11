package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrBeamNotFound                = sdkerrors.Register(ModuleName, 1100, "Beam does not exists")
	ErrBeamNotAuthorized           = sdkerrors.Register(ModuleName, 1101, "This beam does not belong to you")
	ErrBeamInvalidSecret           = sdkerrors.Register(ModuleName, 1102, "Invalid secret provided")
	ErrBeamAlreadyExists           = sdkerrors.Register(ModuleName, 1103, "This beam ID already exists")
	ErrBeamAutoCloseInThePast      = sdkerrors.Register(ModuleName, 1104, "Cannot set autoclose height in the past")
	ErrBeamIdContainsForbiddenChar = sdkerrors.Register(ModuleName, 1105, "This beam ID cannot contains comma")
)
