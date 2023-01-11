package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrUnauthorizedDepositDenom          = sdkerrors.Register(ModuleName, 1200, "Unauthorized denom for deposit")
	ErrEmptyWithdrawalAddress            = sdkerrors.Register(ModuleName, 1201, "Empty withdrawal address")
	ErrEmptyDepositAmount                = sdkerrors.Register(ModuleName, 1202, "Empty deposit amount")
	ErrInsufficientDepositAmount         = sdkerrors.Register(ModuleName, 1203, "Insufficient deposit amount")
	ErrEmptyMicroMintRate                = sdkerrors.Register(ModuleName, 1204, "Empty micro mint rate")
	ErrInvalidMinDepositAmount           = sdkerrors.Register(ModuleName, 1205, "Min deposit amount should be greater than 0")
	ErrInvalidMintDenom                  = sdkerrors.Register(ModuleName, 1206, "Invalid mint denom")
	ErrInvalidDepositDenom               = sdkerrors.Register(ModuleName, 1207, "Invalid deposit denom")
	ErrIllegalMintDenom                  = sdkerrors.Register(ModuleName, 1208, "Mint denom cannot be the bond denom")
	ErrInvalidStakingDenom               = sdkerrors.Register(ModuleName, 1209, "Invalid staking denom")
	ErrInvalidUnbondingTime              = sdkerrors.Register(ModuleName, 1210, "Unbonding time must be positive")
	ErrInvalidWithdrawalDelegatorAddress = sdkerrors.Register(ModuleName, 1211, "Invalid withdrawal status")
	ErrInvalidUnbondAmount               = sdkerrors.Register(ModuleName, 1212, "Invalid unbond amount")
)
