package types

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
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

	if p.GetWithdrawalAddress() != "" {
		_, err = sdk.AccAddressFromBech32(p.GetWithdrawalAddress())
		if err != nil {
			return errorsmod.Wrapf(ErrInvalidWithdrawalAddress, err.Error())
		}
	}

	if p.MinDepositAmount != nil {
		if p.MinDepositAmount.IsNil() || p.MinDepositAmount.LT(sdk.NewInt(DefaultMinDepositAmount)) {
			return errorsmod.Wrapf(ErrInvalidParams, "min deposit amount cannot be lower than default minimum amount, got: %d", p.MinDepositAmount)
		}
	}

	if len(p.DepositDenoms) > 0 {
		for _, denom := range p.DepositDenoms {
			if strings.TrimSpace(denom) == "" {
				return ErrInvalidDepositDenom
			}
		}
	}

	return nil
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
