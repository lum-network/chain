package types

import (
	"fmt"
	"strings"
	"time"

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

func NewRegisterPoolProposal(title string, description string, poolType PoolType, chainID string, denom string, nativeDenom string, connectionId string, bech32PrefixAccAddr string, bech32PrefixValAddr string, validators []string, minDepositAmount math.Int, prizeStrategy PrizeStrategy, drawSchedule DrawSchedule, UnbondingDuration time.Duration, maxUnbondingEntries math.Int, fees []FeeTaker) govtypes.Content {
	return &ProposalRegisterPool{
		Title:               title,
		Description:         description,
		PoolType:            poolType,
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
		UnbondingDuration:   UnbondingDuration,
		MaxUnbondingEntries: maxUnbondingEntries,
		FeeTakers:           fees,
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

	if p.PoolType == PoolType_Unspecified {
		return errorsmod.Wrapf(ErrInvalidPoolType, "%s not allowed", p.PoolType.String())
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
	if p.UnbondingDuration < MinUnbondingDuration {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unbonding duration cannot be lower than %s", MinUnbondingDuration)
	}
	if p.MaxUnbondingEntries.IsNil() || p.MaxUnbondingEntries.IsNegative() || p.MaxUnbondingEntries.GT(sdk.NewInt(DefaultMaxUnbondingEntries)) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Unbonding entries cannot be negative or greated than %d", DefaultMaxUnbondingEntries)
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
	for _, fee := range p.FeeTakers {
		if err := fee.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

func (p ProposalRegisterPool) String() string {
	// Prepare the fee takers string by iterating over the array
	feeTakers := ""
	for _, fee := range p.FeeTakers {
		feeTakers += fee.String() + "\n"
	}

	// Return the string
	return fmt.Sprintf(`Register Pool Proposal:
	Title:            			%s
	Description:      			%s
	ChainID:          			%s
	Denom:			  			%s
	Native Denom:     			%s
	Connection ID	  			%s
	Validators:       			%+v
	Min Deposit Amount: 		%d
	Zone Unbonding Duration:	%s
	Max Unbonding Entries:		%d
	Bech32 Acc Prefix: 			%s
	Bech32 Val Prefix: 			%s
	Transfer Channel ID:		%s
	======Draw Schedule======
	%s
	======Prize Strategy======
	%s
	======Fee Takers======
	%s
  `,
		p.Title, p.Description,
		p.ChainId, p.Denom, p.NativeDenom,
		p.ConnectionId,
		p.Validators,
		p.MinDepositAmount.Int64(),
		p.UnbondingDuration.String(),
		p.MaxUnbondingEntries.Int64(),
		p.Bech32PrefixAccAddr, p.Bech32PrefixValAddr,
		p.TransferChannelId,
		p.DrawSchedule.String(),
		p.PrizeStrategy.String(),
		feeTakers,
	)
}
