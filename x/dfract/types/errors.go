package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrUnauthorizedDepositDenom  = errorsmod.Register(ModuleName, 1200, "Unauthorized denom for deposit")
	ErrEmptyWithdrawalAddress    = errorsmod.Register(ModuleName, 1201, "Empty withdrawal address")
	ErrEmptyDepositAmount        = errorsmod.Register(ModuleName, 1202, "Empty deposit amount")
	ErrInsufficientDepositAmount = errorsmod.Register(ModuleName, 1203, "Insufficient deposit amount")
	ErrEmptyMicroMintRate        = errorsmod.Register(ModuleName, 1204, "Empty micro mint rate")
	ErrInvalidMinDepositAmount   = errorsmod.Register(ModuleName, 1205, "Min deposit amount should be greater than 0")
	ErrInvalidMintDenom          = errorsmod.Register(ModuleName, 1206, "Invalid mint denom")
	ErrInvalidDepositDenom       = errorsmod.Register(ModuleName, 1207, "Invalid deposit denom")
	ErrIllegalMintDenom          = errorsmod.Register(ModuleName, 1208, "Mint denom cannot be the bond denom")
	ErrInvalidManagementAddress  = errorsmod.Register(ModuleName, 1209, "Invalid management address")
	ErrInvalidSignerAddress      = errorsmod.Register(ModuleName, 1210, "Invalid signer address")
	ErrDepositNotEnabled         = errorsmod.Register(ModuleName, 1211, "Deposit not enabled")
)
