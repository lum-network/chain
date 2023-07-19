package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default parameter constants
const (
	DefaultDenom            = "ibc/05554A9BFDD28894D7F18F4C707AA0930D778751A437A9FE1F4684A3E1199728" // USDC ibc denom from Osmosis to Lum Network mainnet
	DefaultMinDepositAmount = 1000000
)

// DefaultParams return the default dfract module params
func DefaultParams() Params {
	return Params{
		DepositDenoms:    []string{DefaultDenom},
		MinDepositAmount: math.NewInt(DefaultMinDepositAmount),
		IsDepositEnabled: true,
	}
}

func (p *Params) ValidateBasics() error {
	if p.GetWithdrawalAddress() != "" {
		if _, err := sdk.AccAddressFromBech32(p.GetWithdrawalAddress()); err != nil {
			return ErrInvalidWithdrawalAddress
		}
	}

	if _, ok := interface{}(p.IsDepositEnabled).(bool); !ok {
		return errorsmod.Wrapf(ErrInvalidParams, "IsDepositEnabled must be a bool, got: %v", p.IsDepositEnabled)
	}

	if len(p.DepositDenoms) > 0 {
		for _, denom := range p.DepositDenoms {
			if err := sdk.ValidateDenom(denom); err != nil {
				return ErrInvalidDepositDenom
			}
		}
	}

	if p.MinDepositAmount.LT(math.NewInt(DefaultMinDepositAmount)) {
		return errorsmod.Wrapf(ErrInvalidParams, "min deposit amount must be gte %d, got: %d", DefaultMinDepositAmount, p.MinDepositAmount.Int64())
	}

	return nil
}
