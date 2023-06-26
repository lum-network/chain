package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrUnauthorizedDepositDenom  = errorsmod.Register(ModuleName, 1200, "Unauthorized denom for deposit")
	ErrEmptyDepositAmount        = errorsmod.Register(ModuleName, 1201, "Empty deposit amount")
	ErrInsufficientDepositAmount = errorsmod.Register(ModuleName, 1202, "Insufficient deposit amount")
	ErrEmptyMicroMintRate        = errorsmod.Register(ModuleName, 1203, "Empty micro mint rate")
	ErrInvalidMinDepositAmount   = errorsmod.Register(ModuleName, 1204, "Min deposit amount should be greater than 0")
	ErrInvalidDepositDenom       = errorsmod.Register(ModuleName, 1205, "Invalid deposit denom")
	ErrInvalidManagementAddress  = errorsmod.Register(ModuleName, 1206, "Invalid management address")
	ErrInvalidSignerAddress      = errorsmod.Register(ModuleName, 1207, "Invalid signer address")
	ErrDepositNotEnabled         = errorsmod.Register(ModuleName, 1208, "Deposit not enabled")
)
