package types

import (
	"sort"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateBasic validates if a pool has a valid configuration
// meaning that it can be stored
func (pool *Pool) ValidateBasic(params Params) error {
	if pool.PoolId == UnknownID {
		return ErrInvalidID
	}
	if pool.State == PoolState_Unspecified {
		return errorsmod.Wrapf(ErrInvalidPoolParams, "no state specified")
	}
	if err := sdk.ValidateDenom(pool.Denom); err != nil {
		return errorsmod.Wrapf(ErrInvalidPoolParams, err.Error())
	}
	if err := sdk.ValidateDenom(pool.NativeDenom); err != nil {
		return errorsmod.Wrapf(ErrInvalidPoolParams, err.Error())
	}
	if strings.TrimSpace(pool.ChainId) == "" {
		return errorsmod.Wrapf(ErrInvalidPoolParams, "empty chain ID")
	}
	if strings.TrimSpace(pool.Bech32PrefixAccAddr) == "" {
		return errorsmod.Wrapf(ErrInvalidPoolParams, "empty bech32 prefix account address")
	}
	if strings.TrimSpace(pool.Bech32PrefixValAddr) == "" {
		return errorsmod.Wrapf(ErrInvalidPoolParams, "empty bech32 prefix validator address")
	}
	if len(pool.Validators) == 0 {
		return errorsmod.Wrapf(ErrInvalidPoolParams, "empty validators set")
	} else {
		for _, val := range pool.Validators {
			bz, err := sdk.GetFromBech32(val.OperatorAddress, pool.Bech32PrefixValAddr)
			if err != nil {
				return errorsmod.Wrapf(ErrInvalidPoolParams, "invalid validator address %s: %v", val.OperatorAddress, err)
			}
			err = sdk.VerifyAddressFormat(bz)
			if err != nil {
				return errorsmod.Wrapf(ErrInvalidPoolParams, "invalid validator address %s: %v", val.OperatorAddress, err)
			}
		}
	}
	if pool.MinDepositAmount.IsNil() || pool.MinDepositAmount.LT(params.MinDepositAmount) {
		return errorsmod.Wrapf(ErrInvalidPoolParams, "min deposit denom must be gte %d", params.MinDepositAmount.Int64())
	}
	if pool.AvailablePrizePool.IsNil() || pool.AvailablePrizePool.Denom != pool.Denom {
		return errorsmod.Wrapf(ErrInvalidPoolParams, "clawback prize pool must be initialized")
	}
	if err := pool.DrawSchedule.ValidateBasic(params); err != nil {
		return errorsmod.Wrapf(ErrInvalidPoolParams, err.Error())
	}
	if err := pool.PrizeStrategy.Validate(params); err != nil {
		return errorsmod.Wrapf(ErrInvalidPoolParams, err.Error())
	}
	return nil
}

func (p *Pool) IsLocalZone(ctx sdk.Context) bool {
	return p.ChainId == ctx.ChainID()
}

func (p *Pool) ShouldDraw(ctx sdk.Context) bool {
	if p.State == PoolState_Ready && (p.LastDrawState == DrawState_Unspecified || p.LastDrawState == DrawState_Success) {
		return p.DrawSchedule.ShouldDraw(ctx, p.LastDrawCreatedAt)
	}
	return false
}

// GetValidatorsMapIndex maps validators operator address to their index in the Pool.Validators array
// utility function to facilitate access to Validators
func (p *Pool) GetValidatorsMapIndex() map[string]int {
	valIdx := make(map[string]int)
	for i, v := range p.Validators {
		valIdx[v.OperatorAddress] = i
	}
	return valIdx
}

// ActiveValidators returns currently enabled validators
func (p *Pool) ActiveValidators() (vals []PoolValidator) {
	for _, v := range p.Validators {
		if v.IsEnabled {
			vals = append(vals, v)
		}
	}
	return
}

// BondedValidators returns active and inactive validators with a bonded amount > 0 for
func (p *Pool) BondedValidators() (activeVals, inactiveVals []PoolValidator) {
	for _, v := range p.Validators {
		if v.IsBonded() {
			if v.IsEnabled {
				activeVals = append(activeVals, v)
			} else {
				inactiveVals = append(inactiveVals, v)
			}
		}
	}
	return
}

// ComputeSplitDelegations computes the delegation split to enforce based on the active validators in the set
// amount is divided evenly to all active validators
func (p *Pool) ComputeSplitDelegations(ctx sdk.Context, amount math.Int) (splits []*SplitDelegation) {
	activeValidators := p.ActiveValidators()
	if len(activeValidators) <= 0 {
		return nil
	}
	used := math.ZeroInt()
	valShare := amount.QuoRaw(int64(len(activeValidators)))

	// Sort vals by bonded amount ascending to for the sake of consistency
	sort.SliceStable(activeValidators, func(i, j int) bool {
		return activeValidators[j].BondedAmount.GT(activeValidators[i].BondedAmount)
	})

	for i, v := range activeValidators {
		// Compute the amount to use
		var relativeAmount math.Int
		if i == len(activeValidators)-1 {
			relativeAmount = amount.Sub(used)
		} else {
			relativeAmount = valShare
		}

		// Append to the destination structure
		splits = append(splits, &SplitDelegation{
			ValidatorAddress: v.OperatorAddress,
			Amount:           relativeAmount,
		})
		used = used.Add(relativeAmount)
	}

	if !used.Equal(amount) {
		// Returns nil in case we did something sketchy when computing amount
		return nil
	}

	return
}

// ComputeSplitUndelegations compute the undelegation split to enforce based on the bonded validators in the set
// disabled validators are prioritized and remaining amount is divided evenly between all validators
func (p *Pool) ComputeSplitUndelegations(ctx sdk.Context, amount math.Int) (splits []*SplitDelegation) {
	bondedActiveVals, bondedInactiveVals := p.BondedValidators()
	if len(bondedActiveVals) <= 0 && len(bondedInactiveVals) <= 0 {
		return nil
	}
	used := math.ZeroInt()

	// Sort vals by bonded amount descending to ensure we can fulfill the request
	sort.SliceStable(bondedInactiveVals, func(i, j int) bool {
		return bondedInactiveVals[i].BondedAmount.GT(bondedInactiveVals[j].BondedAmount)
	})
	for _, v := range bondedInactiveVals {
		// Undelegate as much as we can
		relativeAmount := math.MinInt(v.BondedAmount, amount.Sub(used))
		if relativeAmount.LTE(sdk.ZeroInt()) {
			continue
		}

		// Append to the destination structure
		splits = append(splits, &SplitDelegation{
			ValidatorAddress: v.OperatorAddress,
			Amount:           relativeAmount,
		})
		used = used.Add(relativeAmount)

		if used.Equal(amount) {
			break
		}
	}

	if !used.Equal(amount) {
		// Sort vals by bonded amount ascending to ensure we can fulfill the request
		sort.SliceStable(bondedActiveVals, func(i, j int) bool {
			return bondedActiveVals[j].BondedAmount.GT(bondedActiveVals[i].BondedAmount)
		})

		// Undelegate from active validator set
		for i, v := range bondedActiveVals {
			// Compute the amount to use each round to ensure fairness
			valShare := amount.Sub(used).QuoRaw(int64(len(bondedActiveVals) - i))
			relativeAmount := math.MinInt(v.BondedAmount, valShare)
			if i == len(bondedActiveVals)-1 && v.BondedAmount.GTE(amount.Sub(used)) {
				// Take remaining amount
				relativeAmount = amount.Sub(used)
			}
			if relativeAmount.LTE(sdk.ZeroInt()) {
				continue
			}

			// Append to the destination structure
			splits = append(splits, &SplitDelegation{
				ValidatorAddress: v.OperatorAddress,
				Amount:           relativeAmount,
			})
			used = used.Add(relativeAmount)

			if used.Equal(amount) {
				break
			}
		}
	}

	if !used.Equal(amount) {
		// Returns nil in case we cannot consume the whole amount
		return nil
	}

	return
}

func (p *Pool) ApplySplitDelegate(ctx sdk.Context, splits []*SplitDelegation) {
	valIdx := p.GetValidatorsMapIndex()
	for _, split := range splits {
		p.Validators[valIdx[split.ValidatorAddress]].BondedAmount = p.Validators[valIdx[split.ValidatorAddress]].BondedAmount.Add(split.Amount)
	}
}

func (p *Pool) ApplySplitUndelegate(ctx sdk.Context, splits []*SplitDelegation) {
	valIdx := p.GetValidatorsMapIndex()
	for _, split := range splits {
		p.Validators[valIdx[split.ValidatorAddress]].BondedAmount = p.Validators[valIdx[split.ValidatorAddress]].BondedAmount.Sub(split.Amount)
		if p.Validators[valIdx[split.ValidatorAddress]].BondedAmount.LT(sdk.ZeroInt()) {
			panic(ErrPoolInvalidSplit)
		}
	}
}

// ApplySplitRedelegate serves as internal tracking to redelegate the bonded amount from the inactive to the active validators
func (p *Pool) ApplySplitRedelegate(ctx sdk.Context, splits []*SplitDelegation, operatorAddress string) {
	valIdx := p.GetValidatorsMapIndex()
	for _, split := range splits {
		// Add the split amount to the active validator's bonded amount
		p.Validators[valIdx[split.ValidatorAddress]].BondedAmount = p.Validators[valIdx[split.ValidatorAddress]].BondedAmount.Add(split.Amount)

		// Substract from the disabled validator
		p.Validators[valIdx[operatorAddress]].BondedAmount = p.Validators[valIdx[operatorAddress]].BondedAmount.Sub(split.Amount)
		if p.Validators[valIdx[operatorAddress]].BondedAmount.LT(sdk.ZeroInt()) {
			panic(ErrPoolInvalidSplit)
		}
	}
}

// GetDisabledValidators returns all disabled validators from the pool
func (p *Pool) GetDisabledValidators(ctx sdk.Context) ([]PoolValidator, error) {
	var disabledValidators []PoolValidator

	valIdx := p.GetValidatorsMapIndex()
	for _, index := range valIdx {
		if !p.Validators[index].IsEnabled {
			disabledValidators = append(disabledValidators, p.Validators[index])
		}
	}

	if len(disabledValidators) == 0 {
		return []PoolValidator{}, ErrNoDisabledValidator
	}

	return disabledValidators, nil
}
