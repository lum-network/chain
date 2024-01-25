package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeUpdateParams = "UpdateParams"
)

var (
	_ govtypes.Content = &ProposalUpdateParams{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdateParams)
}

func NewUpdateParamsProposal(title, description string, minDepositAmount *math.Int, prizeDelta *time.Duration, minDepDrawDelta *time.Duration, minDrawDelta *time.Duration, maxDrawDelta *time.Duration, maxBatchQuantity *math.Int, maxStrategyBatches *math.Int) govtypes.Content {
	return &ProposalUpdateParams{
		Title:                   title,
		Description:             description,
		MinDepositAmount:        minDepositAmount,
		PrizeExpirationDelta:    prizeDelta,
		MinDrawScheduleDelta:    minDrawDelta,
		MaxDrawScheduleDelta:    maxDrawDelta,
		MaxPrizeBatchQuantity:   maxBatchQuantity,
		MaxPrizeStrategyBatches: maxStrategyBatches,
		MinDepositDrawDelta:     minDepDrawDelta,
	}
}

func (p *ProposalUpdateParams) ProposalRoute() string { return RouterKey }

func (p *ProposalUpdateParams) ProposalType() string {
	return ProposalTypeUpdateParams
}

func (p *ProposalUpdateParams) ValidateBasic() error {
	// Validate root proposal content
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	// Validate payload
	params := DefaultParams()
	if p.MinDepositAmount != nil {
		params.MinDepositAmount = *p.MinDepositAmount
	}
	if p.MaxPrizeStrategyBatches != nil {
		params.MaxPrizeStrategyBatches = p.MaxPrizeStrategyBatches.Uint64()
	}
	if p.MaxPrizeBatchQuantity != nil {
		params.MaxPrizeBatchQuantity = p.MaxPrizeBatchQuantity.Uint64()
	}
	if p.MinDrawScheduleDelta != nil {
		params.MinDrawScheduleDelta = *p.MinDrawScheduleDelta
	}
	if p.MaxDrawScheduleDelta != nil {
		params.MaxDrawScheduleDelta = *p.MaxDrawScheduleDelta
	}
	if p.PrizeExpirationDelta != nil {
		params.PrizeExpirationDelta = *p.PrizeExpirationDelta
	}
	if p.MinDepositDrawDelta != nil {
		params.MinDepositDrawDelta = *p.MinDepositDrawDelta
	}
	return params.ValidateBasics()
}

func (p ProposalUpdateParams) String() string {
	return fmt.Sprintf(`Update Params Proposal:
	Title:            			%s
	Description:      			%s
	Min Deposit Amount:			%d
	Max Prize Strategy Batches	%d
	Max Prize Batch Quantity	%d
	Min Draw Schedule Delta		%s
	Max Draw Schedule Delta		%s
	Prize Expiration Delta		%s
	Min Deposit Draw Delta		%s
  `,
		p.Title, p.Description,
		p.MinDepositAmount.Int64(),
		p.MaxPrizeStrategyBatches.Int64(),
		p.MaxPrizeBatchQuantity.Int64(),
		p.MinDrawScheduleDelta.String(),
		p.MaxDrawScheduleDelta.String(),
		p.PrizeExpirationDelta.String(),
		p.MinDepositDrawDelta.String())
}
