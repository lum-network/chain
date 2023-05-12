package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateBasic validates if a deposit has a valid poolID, depositID and depositorAddress
// meaning that it can be stored
func (deposit *Deposit) ValidateBasic() error {
	if deposit.PoolId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "pool ID")
	}
	if deposit.DepositId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "deposit ID")
	}
	if deposit.State == DepositState_Unspecified {
		return errorsmod.Wrapf(ErrInvalidDepositState, "no state specified")
	}
	if _, err := sdk.AccAddressFromBech32(deposit.DepositorAddress); err != nil {
		return errorsmod.Wrapf(ErrInvalidDepositorAddress, err.Error())
	}
	if _, err := sdk.AccAddressFromBech32(deposit.WinnerAddress); err != nil {
		return errorsmod.Wrapf(ErrInvalidWinnerAddress, err.Error())
	}
	return nil
}

// ValidateDepositRetryBasic validates the incoming message for a deposit retry
func (msg *MsgDepositRetry) ValidateDepositRetryBasic() error {
	if msg.PoolId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "pool ID")
	}
	if msg.DepositId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "deposit ID")
	}

	return nil
}
