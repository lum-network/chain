package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeUpdatePool = "UpdatePool"
)

var (
	_ govtypes.Content = &ProposalUpdatePool{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdatePool)
}

func NewUpdatePoolProposal(title, description string, poolId uint64, validators []string, minDepositAmount *math.Int, prizeStrategy *PrizeStrategy, drawSchedule *DrawSchedule) govtypes.Content {
	return &ProposalUpdatePool{
		Title:            title,
		Description:      description,
		PoolId:           poolId,
		Validators:       validators,
		MinDepositAmount: minDepositAmount,
		PrizeStrategy:    prizeStrategy,
		DrawSchedule:     drawSchedule,
	}
}

func (p *ProposalUpdatePool) ProposalRoute() string { return RouterKey }

func (p *ProposalUpdatePool) ProposalType() string {
	return ProposalTypeUpdatePool
}

func (p *ProposalUpdatePool) ValidateBasic() error {
	// Validate root proposal content
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	if p.MinDepositAmount != nil {
		if p.MinDepositAmount.IsNil() || p.MinDepositAmount.LT(sdk.NewInt(MinAcceptableDepositAmount)) {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "min deposit denom must be gte %d", MinAcceptableDepositAmount)
		}
	}
	if p.PrizeStrategy != nil {
		if len(p.PrizeStrategy.PrizeBatches) <= 0 {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "at least one prize strategy batch is required")
		}
	}
	if p.DrawSchedule != nil {
		if p.DrawSchedule.DrawDelta < MinAcceptableDrawDelta {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "draw delta cannot be lower than %s", MinAcceptableDrawDelta.String())
		}
	}
	return nil
}

func (p ProposalUpdatePool) String() string {
	return fmt.Sprintf(`Update Pool Proposal:
	Title:            	%s
	Description:      	%s
	Pool ID:			%d
	Validators:       	%+v
	Min Deposit Amount: %d
	======Draw Schedule======
	%s
	======Prize Strategy======
	%s
  `,
		p.Title, p.Description,
		p.PoolId,
		p.Validators,
		p.MinDepositAmount.Int64(),
		p.DrawSchedule.String(),
		p.PrizeStrategy.String(),
	)
}
