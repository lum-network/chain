package dfract

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/lum-network/chain/x/dfract/keeper"
)

func NewDFractProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized dfract proposal content type: %T", c)
		}
	}
}
