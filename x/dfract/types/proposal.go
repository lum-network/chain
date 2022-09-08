package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

func NewSpendAndAdjustProposal(title string, description string, spendDestination string, mintAmount sdk.Coin) *SpendAndAdjustProposal {
	return &SpendAndAdjustProposal{
		Title:            title,
		Description:      description,
		SpendDestination: spendDestination,
		MintAmount:       mintAmount,
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

func (prop *SpendAndAdjustProposal) GetMintAmount() sdk.Coin {
	return prop.MintAmount
}

func (prop *SpendAndAdjustProposal) GetDistribution() []*SpendAndAdjustDistribution {
	return prop.Distribution
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

	if prop.MintAmount.IsZero() {
		return ErrEmptyMintAmount
	}

	// Make sure the computed distribution equals the mint amount
	var total sdk.Coin
	for _, distribution := range prop.GetDistribution() {
		_, err := sdk.AccAddressFromBech32(distribution.GetAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}
		total.Add(distribution.GetAmount())
	}
	if !prop.MintAmount.IsEqual(total) {
		return ErrMintDontMatchTotal
	}
	return nil
}

func (prop *SpendAndAdjustProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Spend and adjust Proposal:
		Title:				%s
		Description:		%s
		Spend Destination: 	%s
		Mint Amount:		%d %s
	`, prop.GetTitle(), prop.GetDescription(), prop.GetSpendDestination(), prop.GetMintAmount().Amount.Int64(), prop.GetMintAmount().Denom))
	return b.String()
}
