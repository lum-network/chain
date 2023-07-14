package types

import (
	"fmt"
	"sort"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate prizeStrategy validation
// - validate strategy based on params
// - validate each batch individually
// - validate total number of prizes > 0
// - validate total percentage of prize pool = 100
func (ps PrizeStrategy) Validate(params Params) error {
	if uint64(len(ps.PrizeBatches)) > params.MaxPrizeStrategyBatches {
		return fmt.Errorf("prize strategy must have maximum %d batches", params.MaxPrizeStrategyBatches)
	}
	totalPrizes := uint64(0)
	totalPercent := uint64(0)
	for _, batch := range ps.PrizeBatches {
		totalPrizes += batch.Quantity
		totalPercent += batch.PoolPercent
		if err := batch.Validate(params); err != nil {
			return err
		}
	}
	if totalPrizes <= 0 {
		return fmt.Errorf("prize strategy must contain at least one prize")
	}
	if totalPercent != 100 {
		return fmt.Errorf("prize batches pool percentage must be equal to 100")
	}
	return nil
}

// ComputePrizesProbs computes final prizes probs list based on each prizeBatch configuration
// always sort the prizes by most valuable to least valuable
// error triggered if the prizeStrategy does not pass its internal validation function
func (ps PrizeStrategy) ComputePrizesProbs(prizePool sdk.Coin) (prizesProbs []PrizeProb, usedAmount math.Int, remainingAmount math.Int, err error) {
	usedAmount = sdk.ZeroInt()
	remainingAmount = prizePool.Amount
	sortedBatches := append([]PrizeBatch{}, ps.PrizeBatches...)
	sort.SliceStable(sortedBatches, func(i, j int) bool {
		return sortedBatches[i].GetPrizeAmount(prizePool).GT(sortedBatches[j].GetPrizeAmount(prizePool))
	})

	// Iterate over each prize batch
	for _, batch := range sortedBatches {
		// For each batch compute batch prizes
		bPrizesProbs, bUsed, _ := batch.ComputePrizesProbs(prizePool)
		prizesProbs = append(prizesProbs, bPrizesProbs...)
		usedAmount = usedAmount.Add(bUsed)
		remainingAmount = remainingAmount.Sub(bUsed)
	}

	return prizesProbs, usedAmount, remainingAmount, nil
}

// ContainsUniqueBatch returns whether or not at least one batch contains has the IsUnique property set to True
func (ps PrizeStrategy) ContainsUniqueBatch() bool {
	for _, b := range ps.PrizeBatches {
		if b.IsUnique {
			return true
		}
	}
	return false
}
