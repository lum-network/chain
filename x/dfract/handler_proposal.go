package dfract

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/lum-network/chain/x/dfract/keeper"
	"github.com/lum-network/chain/x/dfract/types"
)

func NewDFractProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.ProposalUpdateParams:
			{
				params := k.GetParams(ctx)
				if c.GetWithdrawalAddress() != "" {
					params.WithdrawalAddress = c.GetWithdrawalAddress()
				}
				if c.GetIsDepositEnabled() != nil {
					params.IsDepositEnabled = c.IsDepositEnabled.Value
				}
				if len(c.GetDepositDenoms()) > 0 {
					params.DepositDenoms = c.GetDepositDenoms()
				}
				if c.MinDepositAmount != nil {
					params.MinDepositAmount = *c.MinDepositAmount
				}
				if err := params.ValidateBasics(); err != nil {
					return err
				}
				k.SetParams(ctx, params)
				return nil
			}

		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized dfract proposal content type: %T", c)
		}
	}
}
