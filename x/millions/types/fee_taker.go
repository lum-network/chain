package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type FeeTakers []FeeTaker

func (feetaker *FeeTaker) ValidateBasic() error {
	if feetaker.Destination == "" {
		return ErrInvalidFeeTakerDestination
	}

	if feetaker.Amount.IsZero() || feetaker.Amount.IsNegative() {
		return ErrInvalidFeeTakerAmount
	}

	if feetaker.Type == FeeTakerType_Unspecified {
		return ErrInvalidFeeTakerType
	}

	return nil
}

func (feetakers FeeTakers) ValidateBasic() error {
	totalFeePercentage := sdk.ZeroDec()
	for _, feetaker := range feetakers {
		if err := feetaker.ValidateBasic(); err != nil {
			return err
		}
		totalFeePercentage = totalFeePercentage.Add(feetaker.Amount)
	}

	if totalFeePercentage.GT(sdk.OneDec()) {
		return ErrInvalidFeeTakersTotalAmount
	}

	return nil
}

func ValidateFeeTakers(feeTakers []FeeTaker) error {
	return FeeTakers(feeTakers).ValidateBasic()
}
