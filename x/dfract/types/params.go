package types

// DefaultParams return the default dfract module params
func DefaultParams() Params {
	return Params{
		DepositDenom:     "ulum",
		MintDenom:        "udfr",
		MinDepositAmount: 1000000,
	}
}
