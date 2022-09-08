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

	// Acquire the list of waiting proposal & waiting mint deposits
	waitingProposalDeposits := k.ListWaitingProposalDeposits(ctx)
	waitingMintDeposits := k.ListWaitingMintDeposits(ctx)

	// Move the funds from module account to destination address
	if err := k.Spend(ctx, p.GetSpendDestination()); err != nil {
		return err
	}

	// Mint the funds at the provided rate
	if err := k.Mint(ctx, p.GetMintRate()); err != nil {
		return err
	}

	// Distribute the funds to the waiting proposal deposits, and move them to minted deposits
	if err := k.Distribute(ctx, p.GetMintRate(), waitingMintDeposits); err != nil {
		return err
	}

	// Remove the waiting proposal deposits and move them to waiting mint deposits
	for _, deposit := range waitingProposalDeposits {
		k.RemoveFromWaitingProposalQueue(ctx, deposit.GetId())
		k.InsertIntoWaitingMintQueue(ctx, deposit.GetId())
	}

	return nil
}
