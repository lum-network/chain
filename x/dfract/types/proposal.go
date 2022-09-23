package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeWithdrawAndMint = "WithdrawAndMint"
)

var _ govtypes.Content = &WithdrawAndMintProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeWithdrawAndMint)
	govtypes.RegisterProposalTypeCodec(&WithdrawAndMintProposal{}, "lum-network/WithdrawAndMintProposal")
}

func NewWithdrawAndMintProposal(title string, description string, withdrawalAddress string, microMintRate int64) *WithdrawAndMintProposal {
	return &WithdrawAndMintProposal{
		Title:             title,
		Description:       description,
		WithdrawalAddress: withdrawalAddress,
		MicroMintRate:     microMintRate,
	}
}

func (prop *WithdrawAndMintProposal) GetTitle() string {
	return prop.Title
}

func (prop *WithdrawAndMintProposal) GetDescription() string {
	return prop.Description
}

func (prop *WithdrawAndMintProposal) GetWithdrawalAddress() string {
	return prop.WithdrawalAddress
}

func (prop *WithdrawAndMintProposal) GetMicroMintRate() int64 {
	return prop.MicroMintRate
}

func (prop *WithdrawAndMintProposal) ProposalRoute() string {
	return RouterKey
}

func (prop *WithdrawAndMintProposal) ProposalType() string {
	return ProposalTypeWithdrawAndMint
}

func (prop *WithdrawAndMintProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(prop)
	if err != nil {
		return err
	}

	// Make sure we have an address
	if len(prop.GetWithdrawalAddress()) <= 0 {
		return ErrEmptyWithdrawalAddress
	}

	// Make sure it's actually an address
	_, err = sdk.AccAddressFromBech32(prop.GetWithdrawalAddress())
	if err != nil {
		return err
	}

	if prop.GetMicroMintRate() < 0 {
		return ErrEmptyMicroMintRate
	}
	return nil
}

func (prop *WithdrawAndMintProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Spend and adjust Proposal:
		Title:					%s
		Description:			%s
		Withdrawal Address: 	%s
		Micro Mint Rate:		%d
	`, prop.GetTitle(), prop.GetDescription(), prop.GetWithdrawalAddress(), prop.GetMicroMintRate()))
	return b.String()
}
