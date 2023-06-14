package types

func NewGenesisState(params Params, nextPoolID, nextDepositID, nextPrizeID, nexWithdrawalID uint64) *GenesisState {
	return &GenesisState{
		Params:           params,
		NextPoolId:       nextPoolID,
		NextDepositId:    nextDepositID,
		NextPrizeId:      nextPrizeID,
		NextWithdrawalId: nexWithdrawalID,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:           DefaultParams(),
		NextPoolId:       1,
		NextDepositId:    1,
		NextPrizeId:      1,
		NextWithdrawalId: 1,
	}
}

func ValidateGenesis(data GenesisState) error {
	return nil
}
