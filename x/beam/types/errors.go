package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/chain module sentinel errors
var (
	ErrBeamNotFound = sdkerrors.Register(ModuleName, 110, "Beam does not exists")
)
