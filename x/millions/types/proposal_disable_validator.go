package types

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeDisableValidator = "DisableValidator"
)

var (
	_ govtypes.Content = &ProposalDisableValidator{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeDisableValidator)
}

func NewDisableValidatorProposal(title, description string, operatorAddr string, poolID uint64) govtypes.Content {
	return &ProposalDisableValidator{
		Title:           title,
		Description:     description,
		OperatorAddress: operatorAddr,
		PoolId:          poolID,
	}
}

func (p *ProposalDisableValidator) ProposalRoute() string { return RouterKey }

func (p *ProposalDisableValidator) ProposalType() string {
	return ProposalTypeDisableValidator
}

func (p *ProposalDisableValidator) ValidateBasic() error {
	// Validate root proposal content
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	if len(strings.TrimSpace(p.OperatorAddress)) <= 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "operator address is required")
	}

	if p.GetPoolId() == UnknownID {
		return ErrInvalidID
	}

	return nil
}

func (p ProposalDisableValidator) String() string {
	return fmt.Sprintf(`Update Pool Proposal:
	Title:            	%s
	Description:      	%s
	OperatorAddress: 	%s
	Pool ID:			%d
  `,
		p.Title, p.Description,
		p.OperatorAddress,
		p.PoolId,
	)
}
