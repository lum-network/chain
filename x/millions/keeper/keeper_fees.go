package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/lum-network/chain/x/millions/types"
)

type feeCollector struct {
	keeper        Keeper
	pool          types.Pool
	feesStakers   sdk.Dec
	stakersAmount sdk.Coin
}

// NewFeeCollector creates a new fee collector for the specified pool.
func (k Keeper) NewFeeCollector(ctx sdk.Context, pool types.Pool) feeCollector {
	params := k.GetParams(ctx)
	feesStaker := params.FeesStakers
	if feesStaker.IsNil() {
		feesStaker = sdk.ZeroDec()
	}
	return feeCollector{
		keeper:        k,
		pool:          pool,
		feesStakers:   feesStaker,
		stakersAmount: sdk.NewCoin(pool.Denom, math.ZeroInt()),
	}
}

// CollectedAmount returns the collected and not sent amount.
func (fc *feeCollector) CollectedAmount() sdk.Coin {
	return fc.stakersAmount
}

// CollectPrizeFees computes and collects the fees for a prize and updates its final amount.
func (fc *feeCollector) CollectPrizeFees(ctx sdk.Context, prize *types.Prize) (newAmount, fees math.Int) {
	fees = fc.feesStakers.MulInt(prize.Amount.Amount).RoundInt()
	fc.stakersAmount = fc.stakersAmount.AddAmount(fees)
	prize.Amount = prize.Amount.SubAmount(fees)
	return prize.Amount.Amount, fees
}

// SendCollectedFees effectively sends the collected fees (if any) to their destination.
func (fc *feeCollector) SendCollectedFees(ctx sdk.Context) (err error) {
	if fc.stakersAmount.Amount.GT(math.ZeroInt()) {
		err = fc.keeper.BankKeeper.SendCoinsFromAccountToModule(
			ctx,
			sdk.MustAccAddressFromBech32(fc.pool.GetLocalAddress()),
			authtypes.FeeCollectorName,
			sdk.NewCoins(fc.stakersAmount),
		)
	}
	if err == nil {
		fc.stakersAmount.Amount = math.ZeroInt()
	}
	return err
}
