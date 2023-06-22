package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func NewUpdateParamsProposal(title string, description string, managementAddr string, isDepositEnabled *gogotypes.BoolValue) govtypes.Content {
	return &ProposalUpdateParams{
		Title:             title,
		Description:       description,
		ManagementAddress: managementAddr,
		IsDepositEnabled:  isDepositEnabled,
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

	if p.GetManagementAddress() != "" {
		_, err = sdk.AccAddressFromBech32(p.GetManagementAddress())
		if err != nil {
			return errorsmod.Wrapf(ErrInvalidManagementAddress, err.Error())
		}
	}

	return nil
}

func (p ProposalUpdateParams) String() string {
	return fmt.Sprintf(`Update Params Proposal:
	Title:            			%s
	Description:      			%s
	Management address:			%s
	Is deposit enabled:			%v
  `,
		p.Title, p.Description,
		p.ManagementAddress,
		p.IsDepositEnabled,
	)
}
