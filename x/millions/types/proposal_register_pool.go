package types

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeRegisterPool = "RegisterPool"
)

var (
	_ govtypes.Content = &ProposalRegisterPool{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeRegisterPool)
}

func NewRegisterPoolProposal(title, description, chainID string, denom string, nativeDenom string, connectionId string, bech32PrefixAccAddr string, bech32PrefixValAddr string, validators []string, minDepositAmount math.Int, prizeStrategy PrizeStrategy, drawSchedule DrawSchedule, unbondingFrequency math.Int) govtypes.Content {
	return &ProposalRegisterPool{
		Title:               title,
		Description:         description,
		ChainId:             chainID,
		Denom:               denom,
		NativeDenom:         nativeDenom,
		ConnectionId:        connectionId,
		Validators:          validators,
		MinDepositAmount:    minDepositAmount,
		Bech32PrefixAccAddr: bech32PrefixAccAddr,
		Bech32PrefixValAddr: bech32PrefixValAddr,
		PrizeStrategy:       prizeStrategy,
		DrawSchedule:        drawSchedule,
		UnbondingFrequency:  unbondingFrequency,
	}
}

func (p *ProposalRegisterPool) ProposalRoute() string { return RouterKey }

func (p *ProposalRegisterPool) ProposalType() string {
	return ProposalTypeRegisterPool
}

func (p *ProposalRegisterPool) ValidateBasic() error {
	// Validate root proposal content
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	// Validate payload
	if len(strings.TrimSpace(p.ChainId)) <= 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Chain ID is required")
	}
	if err := sdk.ValidateDenom(p.Denom); err != nil {
		return errorsmod.Wrapf(err, "a valid denom is required")
	}
	if err := sdk.ValidateDenom(p.NativeDenom); err != nil {
		return errorsmod.Wrapf(err, "a valid native_denom is required")
	}
	if len(p.Validators) <= 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "at least one validator is required")
	}
	if p.MinDepositAmount.IsNil() || p.MinDepositAmount.LT(sdk.NewInt(MinAcceptableDepositAmount)) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "min deposit denom must be gte %d", MinAcceptableDepositAmount)
	}
	if p.UnbondingFrequency.IsNil() || p.UnbondingFrequency.IsNegative() || p.UnbondingFrequency.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unbonding frequency must be gt 0")
	}
	if len(strings.TrimSpace(p.Bech32PrefixAccAddr)) <= 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "bech32 acc prefix is required")
	}
	if len(strings.TrimSpace(p.Bech32PrefixValAddr)) <= 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "bech32 val prefix is required")
	}
	if len(p.PrizeStrategy.PrizeBatches) <= 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "at least one prize strategy batch is required")
	}
	if p.DrawSchedule.DrawDelta < MinAcceptableDrawDelta {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "draw delta cannot be lower than %s", MinAcceptableDrawDelta)
	}
	return nil
}

func (p ProposalRegisterPool) String() string {
	return fmt.Sprintf(`Register Pool Proposal:
	Title:            		%s
	Description:      		%s
	ChainID:          		%s
	Denom:			  		%s
	Native Denom:     		%s
	Connection ID	  		%s
	Validators:       		%+v
	Min Deposit Amount: 	%d
	Unbonding Frequency:	%d
	Bech32 Acc Prefix: 		%s
	Bech32 Val Prefix: 		%s
	Transfer Channel ID:	%s
	======Draw Schedule======
	%s
	======Prize Strategy======
	%s
  `,
		p.Title, p.Description,
		p.ChainId, p.Denom, p.NativeDenom,
		p.ConnectionId,
		p.Validators,
		p.MinDepositAmount.Int64(),
		p.UnbondingFrequency.Int64(),
		p.Bech32PrefixAccAddr, p.Bech32PrefixValAddr,
		p.TransferChannelId,
		p.DrawSchedule.String(),
		p.PrizeStrategy.String(),
	)
}
