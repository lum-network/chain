package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/lum-network/chain/x/dfract/types"
)

func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams Set the in-store params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// UpdateParams update the in-store params
// TODO - Method to use in the upgrade handler (remove comment once added)
func (k Keeper) UpdateParams(ctx sdk.Context, managementAddr string, isDepositEnabled *gogotypes.BoolValue) error {
	params := k.GetParams(ctx)

	if managementAddr != "" {
		params.ManagementAddress = managementAddr
	}

	if isDepositEnabled != nil {
		params.IsDepositEnabled = isDepositEnabled.Value
	}

	if err := params.ValidateBasics(); err != nil {
		return err
	}
	k.SetParams(ctx, params)
	return nil
}
