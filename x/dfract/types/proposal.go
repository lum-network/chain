package types

import (
	"fmt"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"strings"
)

const (
	ProposalTypeSpendAndAdjust = "SpendAndAdjust"
)

var _ govtypes.Content = &SpendAndAdjustProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeSpendAndAdjust)
	govtypes.RegisterProposalTypeCodec(&SpendAndAdjustProposal{}, "lum-network/SpendAndAdjustProposal")
}

func NewSpendAndAdjustProposal(title string, description string, spendDestination string, mintRate int64) *SpendAndAdjustProposal {
	return &SpendAndAdjustProposal{
		Title:            title,
		Description:      description,
		SpendDestination: spendDestination,
		MintRate:         mintRate,
	}
}

func (prop *SpendAndAdjustProposal) GetTitle() string {
	return prop.Title
}

func (prop *SpendAndAdjustProposal) GetDescription() string {
	return prop.Description
}

func (prop *SpendAndAdjustProposal) GetSpendDestination() string {
	return prop.SpendDestination
}

func (prop *SpendAndAdjustProposal) GetMintRate() int64 {
	return prop.MintRate
}

func (prop *SpendAndAdjustProposal) ProposalRoute() string {
	return RouterKey
}

func (prop *SpendAndAdjustProposal) ProposalType() string {
	return ProposalTypeSpendAndAdjust
}

func (prop *SpendAndAdjustProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(prop)
	if err != nil {
		return err
	}

	if len(prop.GetSpendDestination()) <= 0 {
		return ErrEmptySpendDestination
	}
	return nil
}

func (prop *SpendAndAdjustProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Spend and adjust Proposal:
		Title:				%s
		Description:		%s
		Spend Destination: 	%s
		Mint Rate:		%d %s
	`, prop.GetTitle(), prop.GetDescription(), prop.GetSpendDestination(), prop.GetMintRate()))
	return b.String()
}
