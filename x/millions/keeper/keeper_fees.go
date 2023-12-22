package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	"github.com/lum-network/chain/x/millions/types"
)

type feeDestination struct {
	taker  types.FeeTaker
	amount math.Int
}

type feeManager struct {
	keeper                Keeper
	pool                  types.Pool
	takers                []types.FeeTaker
	totalPercentToCollect math.LegacyDec
	collectedAmount       sdk.Coin
}

// CollectedAmount returns the collected and not sent amount
func (fm *feeManager) CollectedAmount() sdk.Coin {
	return fm.collectedAmount
}

func (k Keeper) NewFeeManager(ctx sdk.Context, pool types.Pool) *feeManager {
	total := math.LegacyZeroDec()

	// Sanitize the amount of each fee taker
	// Compute the total fees percent to collect
	for _, ft := range pool.FeeTakers {
		if ft.Amount.IsNil() {
			ft.Amount = sdk.ZeroDec()
		}
		total = total.Add(ft.Amount)
	}

	return &feeManager{
		keeper:                k,
		pool:                  pool,
		takers:                pool.FeeTakers,
		totalPercentToCollect: total,
		collectedAmount:       sdk.NewCoin(pool.Denom, math.ZeroInt()),
	}
}

// CollectPrizeFees computes and collects the fees for a prize and updates its final amount
// It loops through the fee takers and collect the fees for each one of them
func (fm *feeManager) CollectPrizeFees(ctx sdk.Context, prize *types.Prize) (newAmount, fees math.Int) {
	if fm.totalPercentToCollect.IsZero() {
		return prize.Amount.Amount, math.ZeroInt()
	}

	// Take the fees
	fees = fm.totalPercentToCollect.MulInt(prize.Amount.Amount).RoundInt()

	// Update the collected amount
	fm.collectedAmount = fm.collectedAmount.AddAmount(fees)

	// Update the prize amount
	prize.Amount = prize.Amount.SubAmount(fees)

	return prize.Amount.Amount, fees
}

// SendCollectedFees effectively sends the collected fees (if any) to their destination
// For each type, it handles specific logic
func (fm *feeManager) SendCollectedFees(ctx sdk.Context) (err error) {
	// If nothing was collected, there is nothing to do
	if fm.totalPercentToCollect.IsZero() || fm.collectedAmount.Amount.IsZero() {
		return nil
	}

	// Compute the destinations
	for _, ft := range fm.takers {
		am := ft.Amount.MulInt(fm.collectedAmount.Amount).RoundInt()
		amount := sdk.NewCoin(fm.pool.Denom, am)

		// Process the fee taker depending on its type
		switch ft.Type {
		case types.FeeTakerType_LocalAddr:
			err = fm.sendCollectedFeesToLocalAddr(ctx, ft, amount)
		case types.FeeTakerType_LocalModuleAccount:
			err = fm.sendCollectedFeesToLocalModuleAccount(ctx, ft, amount)
		case types.FeeTakerType_RemoteAddr:
			err = fm.sendCollectedFeesToRemoteAddr(ctx, ft, amount)
		}
		if err != nil {
			return err
		}
	}

	if err == nil {
		fm.collectedAmount.Amount = math.ZeroInt()
	}

	return nil
}

// sendCollectedFeesToLocalAddr sends the collected fees to a local address
func (fm *feeManager) sendCollectedFeesToLocalAddr(ctx sdk.Context, ft types.FeeTaker, amount sdk.Coin) (err error) {
	return fm.keeper.BankKeeper.SendCoins(
		ctx,
		sdk.MustAccAddressFromBech32(fm.pool.GetLocalAddress()),
		sdk.MustAccAddressFromBech32(ft.Destination),
		sdk.NewCoins(amount),
	)
}

// sendCollectedFeesToLocalModuleAccount sends the collected fees to a local module account
func (fm *feeManager) sendCollectedFeesToLocalModuleAccount(ctx sdk.Context, ft types.FeeTaker, amount sdk.Coin) (err error) {
	return fm.keeper.BankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sdk.MustAccAddressFromBech32(fm.pool.GetLocalAddress()),
		ft.Destination,
		sdk.NewCoins(amount),
	)
}

// sendCollectedFeesToRemoteAddr sends the collected fees to a remote address through IBC
func (fm *feeManager) sendCollectedFeesToRemoteAddr(ctx sdk.Context, ft types.FeeTaker, amount sdk.Coin) (err error) {
	// Build the timeout timestamp
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + types.IBCTimeoutNanos

	// Build the transfer message
	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		fm.pool.GetTransferChannelId(),
		amount,
		fm.pool.GetLocalAddress(),
		ft.Destination,
		clienttypes.Height{},
		timeoutTimestamp,
		"Cosmos Millions Revenue Sharing",
	)

	res, err := fm.keeper.IBCTransferKeeper.Transfer(ctx, msg)
	if err != nil {
		return err
	}

	fm.keeper.Logger(ctx).Debug(
		fmt.Sprintf("Broadcasted IBC fees transfer with sequence %d", res.Sequence),
		"pool_id", fm.pool.GetPoolId(),
	)

	return nil
}
