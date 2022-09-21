package types

func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{
			MintDenom:        "udfr",
			DepositDenom:     "ulum",
			MinDepositAmount: 1000000,
		},
	}
}

func ValidateGenesis(data GenesisState) error {
	return nil
}
