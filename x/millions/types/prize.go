package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateBasic validates if a prize is complete and valid.
func (prize *Prize) ValidateBasic() error {
	if prize.PoolId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "pool ID")
	}
	if prize.DrawId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "draw ID")
	}
	if prize.PrizeId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "prize ID")
	}
	if prize.State == PrizeState_Unspecified {
		return errorsmod.Wrapf(ErrInvalidPrizeState, "no state specified")
	}
	if _, err := sdk.AccAddressFromBech32(prize.WinnerAddress); err != nil {
		return errorsmod.Wrapf(ErrInvalidWinnerAddress, err.Error())
	}
	if prize.Amount.IsNil() || prize.Amount.Amount.LTE(sdk.ZeroInt()) {
		return ErrInvalidPrizeAmount
	}

	return nil
}
