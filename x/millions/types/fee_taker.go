package types

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
