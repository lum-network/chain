package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PrizeProb struct {
	Amount          math.Int
	DrawProbability sdk.Dec
	IsUnique        bool
}

// Validate prizeBatch validation
// - poolPercent must be [0, 100]
// - quantity must be >= 0
// - drawProbability must be [0, 1]
// - if unique, quantity must not be superior to 100
func (pb PrizeBatch) Validate(params Params) error {
	if pb.PoolPercent == 0 || pb.PoolPercent > 100 {
		return fmt.Errorf("prize batch pool percentrage must be gt 0 and lte 100")
	}
	if pb.Quantity == 0 {
		return fmt.Errorf("prize batch quantity must be gt 0")
	} else if pb.Quantity > params.MaxPrizeBatchQuantity {
		return fmt.Errorf("prize batch quantity must be lte %d", params.MaxPrizeBatchQuantity)
	}
	if pb.DrawProbability.LT(sdk.NewDec(0)) || pb.DrawProbability.GT(sdk.NewDec(1)) {
		return fmt.Errorf("prize batch draw probability must be gte 0 and lte 1")
	}
	if pb.GetIsUnique() {
		if pb.Quantity > 100 {
			return fmt.Errorf("prize batch quantity must be lte 100 when unique mode is enabled")
		}
	}
	return nil
}

func (pb PrizeBatch) GetTotalPrizesAmount(prizePool sdk.Coin) math.Int {
	return math.LegacyNewDec(int64(pb.PoolPercent)).QuoInt64(int64(100)).MulInt(prizePool.Amount).TruncateInt()
}

func (pb PrizeBatch) GetPrizeAmount(prizePool sdk.Coin) math.Int {
	return pb.GetTotalPrizesAmount(prizePool).QuoRaw(int64(pb.Quantity))
}

// ComputePrizesProbs computes batch prizes probs list based on the batch percentage of the prizePool
// Returns as much prize up to quantity as possible, until the total batch amount has been fully consumed or is
// not enough to create another prize
// Each prize is identical and the usedAmount and remainingAmount are returned upon computation
func (pb PrizeBatch) ComputePrizesProbs(prizePool sdk.Coin) (prizesProbs []PrizeProb, usedAmount math.Int, remainingAmount math.Int) {
	usedAmount = math.ZeroInt()
	remainingAmount = pb.GetTotalPrizesAmount(prizePool)
	prizeAmount := pb.GetPrizeAmount(prizePool)

	if prizeAmount.LTE(math.ZeroInt()) {
		return nil, math.ZeroInt(), remainingAmount
	}

	for q := uint64(0); q < pb.Quantity; q++ {
		if remainingAmount.LT(prizeAmount) {
			break
		}
		prizesProbs = append(prizesProbs, PrizeProb{
			Amount:          prizeAmount,
			DrawProbability: pb.DrawProbability,
			IsUnique:        pb.IsUnique,
		})
		remainingAmount = remainingAmount.Sub(prizeAmount)
		usedAmount = usedAmount.Add(prizeAmount)
	}

	return prizesProbs, usedAmount, remainingAmount
}
