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
		case *types.WithdrawAndMintProposal:
			return handleWithdrawAndMintProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized dfract proposal content type: %T", c)
		}
	}
}

func handleWithdrawAndMintProposal(ctx sdk.Context, k keeper.Keeper, p *types.WithdrawAndMintProposal) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	if err := k.ProcessWithdrawAndMintProposal(ctx, p); err != nil {
		return err
	}
	return nil
}
