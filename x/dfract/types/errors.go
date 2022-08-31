package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrUnauthorizedDenom    = sdkerrors.Register(ModuleName, 1200, "Unauthorized denom for deposit")
	ErrDepositNotFound      = sdkerrors.Register(ModuleName, 1201, "Deposit not found")
	ErrDepositAlreadyExists = sdkerrors.Register(ModuleName, 1202, "Deposit ID already exists")
)
