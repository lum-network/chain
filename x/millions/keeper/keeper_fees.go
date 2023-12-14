package keeper

import (
	"cosmossdk.io/math"
	"fmt"
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
	keeper          Keeper
	pool            types.Pool
	takers          []types.FeeTaker
	collectedAmount sdk.Coin
	destinations    []feeDestination
}

// CollectedAmount returns the collected and not sent amount
func (fm *feeManager) CollectedAmount() sdk.Coin {
	return fm.collectedAmount
}

func (k Keeper) NewFeeManager(ctx sdk.Context, pool types.Pool) *feeManager {
	// Sanitize the amount of each fee taker
	for _, ft := range pool.FeeTakers {
		if ft.Amount.IsNil() {
			ft.Amount = sdk.ZeroDec()
		}
	}

	return &feeManager{
		keeper:          k,
		pool:            pool,
		takers:          pool.FeeTakers,
		collectedAmount: sdk.NewCoin(pool.Denom, math.ZeroInt()),
		destinations:    make([]feeDestination, 0),
	}
}

// CollectPrizeFees computes and collects the fees for a prize and updates its final amount
// It loops through the fee takers and collect the fees for each one of them
func (fm *feeManager) CollectPrizeFees(ctx sdk.Context, prize *types.Prize) (newAmount, fees math.Int) {
	// Compute the fees
	fees = math.ZeroInt()
	for _, ft := range fm.takers {
		am := ft.Amount.MulInt(prize.Amount.Amount).RoundInt()
		fees = fees.Add(am)

		// Update the collected amount for this destination
		fm.destinations = append(fm.destinations, feeDestination{
			taker:  ft,
			amount: am,
		})
	}

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
	if fm.collectedAmount.Amount.IsZero() {
		return nil
	}

	// Otherwise, handle each fee taker by calling specific logic with the taker and amount to send
	for _, destination := range fm.destinations {
		// Compute the amount to send for each, depending on their amount (which is a percentage)
		amount := sdk.NewCoin(fm.pool.Denom, destination.amount)

		// Process the fee taker depending on its type
		switch destination.taker.Type {
		case types.FeeTakerType_LocalAddr:
			err = fm.sendCollectedFeesToLocalAddr(ctx, destination.taker, amount)
		case types.FeeTakerType_LocalModuleAccount:
			err = fm.sendCollectedFeesToLocalModuleAccount(ctx, destination.taker, amount)
		case types.FeeTakerType_RemoteAddr:
			err = fm.sendCollectedFeesToRemoteAddr(ctx, destination.taker, amount)
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
