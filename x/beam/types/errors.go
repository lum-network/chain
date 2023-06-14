package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrBeamNotFound                = errorsmod.Register(ModuleName, 1100, "Beam does not exists")
	ErrBeamNotAuthorized           = errorsmod.Register(ModuleName, 1101, "This beam does not belong to you")
	ErrBeamInvalidSecret           = errorsmod.Register(ModuleName, 1102, "Invalid secret provided")
	ErrBeamAlreadyExists           = errorsmod.Register(ModuleName, 1103, "This beam ID already exists")
	ErrBeamAutoCloseInThePast      = errorsmod.Register(ModuleName, 1104, "Cannot set autoclose height in the past")
	ErrBeamIDContainsForbiddenChar = errorsmod.Register(ModuleName, 1105, "This beam ID cannot contains comma")
)
