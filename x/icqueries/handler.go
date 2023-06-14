package icqueries

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/lum-network/chain/x/icqueries/keeper"
	"github.com/lum-network/chain/x/icqueries/types"
)

// NewHandler returns a handler for interchainquery module messages.
func NewHandler(_ keeper.Keeper) sdk.Handler {
	return func(_ sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	}
}
