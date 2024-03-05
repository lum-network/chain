package types

import (
	"fmt"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeClosePool = "ClosePool"
)

var (
	_ govtypes.Content = &ProposalClosePool{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeClosePool)
}

func NewClosePoolProposal(title string, description string, poolID uint64) govtypes.Content {
	return &ProposalClosePool{
		Title:       title,
		Description: description,
		PoolId:      poolID,
	}
}

func (p *ProposalClosePool) ProposalRoute() string { return RouterKey }

func (p *ProposalClosePool) ProposalType() string {
	return ProposalTypeClosePool
}

func (p *ProposalClosePool) ValidateBasic() error {
	// Validate root proposal content
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	return nil
}

func (p ProposalClosePool) String() string {
	// Return the string
	return fmt.Sprintf(`Register Pool Proposal:
	Title:            			%s
	Description:      			%s
	PoolID:          			%d
  `,
		p.Title, p.Description, p.PoolId,
	)
}
