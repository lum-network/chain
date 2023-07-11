package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/lum-network/chain/x/epochs/types"
)

const (
	MaxAcceptableWithdrawalIDsCount = 10_000
	WithdrawalTrackerType           = "withdrawal"
)

func (e *EpochUnbonding) ValidateBasic(params Params) error {
	if e.PoolId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "pool ID")
	}

	if e.EpochIdentifier != epochstypes.DAY_EPOCH {
		return errorsmod.Wrapf(ErrInvalidEpochField, "should be DAY_EPOCH")
	}

	if len(e.WithdrawalIds) == 0 {
		return errorsmod.Wrapf(ErrInvalidEpochField, "empty withdrawalIDs set")
	}

	if e.TotalAmount.Amount.LTE(sdk.ZeroInt()) {
		return errorsmod.Wrapf(ErrInvalidEpochField, "withdrawal amount must be gte than 0")
	}

	if e.WithdrawalIdsCount == 0 {
		return errorsmod.Wrapf(ErrInvalidEpochField, "withdrawalIDs count should be gte than 0")
	}

	return nil
}

func (e *EpochUnbonding) WithdrawalIDExists(withdrawalID uint64) bool {
	for _, id := range e.WithdrawalIds {
		if id == withdrawalID {
			return true
		}
	}
	return false
}

func (e *EpochUnbonding) WithdrawalIDsLimitReached(count uint64) bool {
	return count >= MaxAcceptableWithdrawalIDsCount
}
