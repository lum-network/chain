package types

import (
	"fmt"

	"cosmossdk.io/math"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	gogotypes "github.com/cosmos/gogoproto/types"
)

const (
	ProposalTypeUpdateParams = "UpdateParamsProposal"
)

var (
	_ govtypes.Content = &ProposalUpdateParams{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdateParams)
}

func NewUpdateParamsProposal(title string, description string, withdrawalAddr string, isDepositEnabled *gogotypes.BoolValue, depositDenoms []string, minDepositAmount *math.Int) govtypes.Content {
	return &ProposalUpdateParams{
		Title:             title,
		Description:       description,
		WithdrawalAddress: withdrawalAddr,
		IsDepositEnabled:  isDepositEnabled,
		DepositDenoms:     depositDenoms,
		MinDepositAmount:  minDepositAmount,
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

	params := DefaultParams()
	if p.GetWithdrawalAddress() != "" {
		params.WithdrawalAddress = p.WithdrawalAddress
	}

	if p.IsDepositEnabled != nil {
		params.IsDepositEnabled = p.IsDepositEnabled.Value
	}

	if len(p.DepositDenoms) > 0 {
		params.DepositDenoms = p.DepositDenoms
	}

	if p.MinDepositAmount != nil {
		params.MinDepositAmount = uint32(p.MinDepositAmount.Int64())
	}

	return params.ValidateBasics()
}

func (p ProposalUpdateParams) String() string {
	return fmt.Sprintf(`Update Params Proposal:
	Title:            			%s
	Description:      			%s
	Withdrawal address:			%s
	Is deposit enabled:			%v
	Deposit denoms:				%+v
	Min deposit amount:			%d
  `,
		p.Title, p.Description,
		p.WithdrawalAddress,
		p.IsDepositEnabled,
		p.DepositDenoms,
		uint32(p.MinDepositAmount.Int64()),
	)
}
