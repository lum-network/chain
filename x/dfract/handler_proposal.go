package dfract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/lum-network/chain/x/dfract/keeper"
	"github.com/lum-network/chain/x/dfract/types"
)

func NewDFractProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.SpendAndAdjustProposal:
			return handleSpendAndAdjustProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized dfract proposal content type: %T", c)
		}
	}
}

func handleSpendAndAdjustProposal(ctx sdk.Context, k keeper.Keeper, p *types.SpendAndAdjustProposal) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	if err := k.Spend(ctx, p.GetSpendDestination()); err != nil {
		return err
	}

	if err := k.Mint(ctx, p.GetMintAmount()); err != nil {
		return err
	}

	if err := k.Distribute(ctx, p.GetDistribution()); err != nil {
		return err
	}

	return nil
}
