package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/icacallbacks module sentinel errors.
var (
	ErrCallbackHandlerNotFound = errorsmod.Register(ModuleName, 1502, "icacallback handler not found")
	ErrCallbackFailed          = errorsmod.Register(ModuleName, 1504, "icacallback failed")
	ErrInvalidAcknowledgement  = errorsmod.Register(ModuleName, 1507, "invalid acknowledgement")
)
