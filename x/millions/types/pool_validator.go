package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (val *PoolValidator) MustValAddressFromBech32() sdk.ValAddress {
	valAddr, err := sdk.ValAddressFromBech32(val.OperatorAddress)
	if err != nil {
		panic(err)
	}
	return valAddr
}

func (val *PoolValidator) IsBonded() bool {
	return !val.BondedAmount.IsZero()
}
