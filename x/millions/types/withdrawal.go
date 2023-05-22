package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateBasic validates if a withdrawal is complete and valid
func (withdrawal *Withdrawal) ValidateBasic() error {
	if withdrawal.PoolId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "pool ID")
	}
	if withdrawal.WithdrawalId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "withdrawal ID")
	}
	if withdrawal.State == WithdrawalState_Unspecified {
		return errorsmod.Wrapf(ErrInvalidWithdrawalState, "no state specified")
	}
	if _, err := sdk.AccAddressFromBech32(withdrawal.DepositorAddress); err != nil {
		return errorsmod.Wrapf(ErrInvalidWithdrawalDepositorAddress, err.Error())
	}
	if withdrawal.Amount.Amount.LTE(sdk.ZeroInt()) {
		return ErrDepositAlreadyWithdrawn
	}

	return nil
}

// ValidateWithdrawDepositRetryBasic validates the incoming message for a withdraw deposit retry
func (msg *MsgWithdrawDepositRetry) ValidateWithdrawDepositRetryBasic() error {
	if msg.PoolId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "pool ID")
	}
	if msg.WithdrawalId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "withdrawal ID")
	}

	return nil
}
