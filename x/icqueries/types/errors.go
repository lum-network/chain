package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrAlreadyFulfilled    = errorsmod.Register(ModuleName, 1001, "query already fulfilled")
	ErrSucceededNoDelete   = errorsmod.Register(ModuleName, 1002, "query succeeded; do not not execute default behavior")
	ErrInvalidICQProof     = errorsmod.Register(ModuleName, 1003, "icq query response failed")
	ErrICQCallbackNotFound = errorsmod.Register(ModuleName, 1004, "icq callback id not found")
	ErrInvalidQuery        = errorsmod.Register(ModuleName, 1005, "Query has an invalid parameter")
)
