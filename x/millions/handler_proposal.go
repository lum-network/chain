package millions

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/lum-network/chain/x/millions/keeper"
	"github.com/lum-network/chain/x/millions/types"
)

func NewMillionsProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.ProposalUpdatePool:
			{
				return k.UpdatePool(ctx, c.PoolId, c.Validators, c.MinDepositAmount, c.DrawSchedule, c.PrizeStrategy)
			}
		case *types.ProposalRegisterPool:
			{
				_, err := k.RegisterPool(
					ctx,
					c.GetPoolType(),
					c.GetDenom(),
					c.GetNativeDenom(),
					c.GetChainId(),
					c.GetConnectionId(),
					c.GetTransferChannelId(),
					c.GetValidators(),
					c.GetBech32PrefixAccAddr(),
					c.GetBech32PrefixValAddr(),
					c.MinDepositAmount,
					c.GetDrawSchedule(),
					c.GetPrizeStrategy(),
				)
				return err
			}
		case *types.ProposalUpdateParams:
			{
				params := k.GetParams(ctx)
				if c.MinDepositAmount != nil {
					params.MinDepositAmount = *c.MinDepositAmount
				}
				if c.MaxPrizeStrategyBatches != nil {
					params.MaxPrizeStrategyBatches = c.MaxPrizeStrategyBatches.Uint64()
				}
				if c.MaxPrizeBatchQuantity != nil {
					params.MaxPrizeBatchQuantity = c.MaxPrizeBatchQuantity.Uint64()
				}
				if c.MinDrawScheduleDelta != nil {
					params.MinDrawScheduleDelta = *c.MinDrawScheduleDelta
				}
				if c.MaxDrawScheduleDelta != nil {
					params.MaxDrawScheduleDelta = *c.MaxDrawScheduleDelta
				}
				if c.PrizeExpirationDelta != nil {
					params.PrizeExpirationDelta = *c.PrizeExpirationDelta
				}
				if c.FeesStakers != nil {
					params.FeesStakers = *c.FeesStakers
				}
				if c.MinDepositDrawDelta != nil {
					params.MinDepositDrawDelta = *c.MinDepositDrawDelta
				}
				if err := params.ValidateBasics(); err != nil {
					// Prevent panic in case of invalid params
					return err
				}
				k.SetParams(ctx, params)
				return nil
			}
		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized millions proposal content type: %T", c)
		}
	}
}
