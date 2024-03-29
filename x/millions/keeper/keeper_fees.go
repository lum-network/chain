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
	keeper          Keeper
	pool            types.Pool
	collectedAmount sdk.Coin
	destinations    []*feeDestination
}

// CollectedAmount returns the collected and not sent amount
func (fm *feeManager) CollectedAmount() sdk.Coin {
	return fm.collectedAmount
}

func (k Keeper) NewFeeManager(ctx sdk.Context, pool types.Pool) *feeManager {
	destinations := []*feeDestination{}

	// Prepare takers destinations collectors
	for _, ft := range pool.FeeTakers {
		if !ft.Amount.IsNil() && !ft.Amount.LTE(sdk.ZeroDec()) {
			destinations = append(destinations, &feeDestination{
				taker:  ft,
				amount: sdk.ZeroInt(),
			})
		}
	}

	return &feeManager{
		keeper:          k,
		pool:            pool,
		collectedAmount: sdk.NewCoin(pool.Denom, math.ZeroInt()),
		destinations:    destinations,
	}
}

// CollectPrizeFees computes and collects the fees for a prize and updates its final amount
// It loops through the fee takers and collect the fees for each one of them
// returns the new prize amount alongside the total collected fees
func (fm *feeManager) CollectPrizeFees(ctx sdk.Context, prize *types.Prize) (newAmount, fees math.Int) {
	fees = sdk.ZeroInt()

	// Collect fees
	for _, fd := range fm.destinations {
		// Compute taker fees
		takerFees := fd.taker.Amount.MulInt(prize.Amount.Amount).RoundInt()
		fees = fees.Add(takerFees)
		// Collect it
		fd.amount = fd.amount.Add(takerFees)
		fm.collectedAmount = fm.collectedAmount.AddAmount(takerFees)
	}

	// Deduce from prize amount
	prize.Amount = prize.Amount.SubAmount(fees)

	return prize.Amount.Amount, fees
}

// SendCollectedFees effectively sends the collected fees (if any) to their destination
// For each destination type, it handles specific logic
// Reset the fee manager to 0 collected fees upon succesful completion
func (fm *feeManager) SendCollectedFees(ctx sdk.Context) (err error) {
	// If nothing was collected, there is nothing to do
	if fm.collectedAmount.Amount.IsZero() {
		return nil
	}

	// Otherwise, handle each fee taker by calling specific logic with the taker and amount to send
	for _, destination := range fm.destinations {
		if destination.amount.LTE(sdk.ZeroInt()) {
			continue
		}

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

	// Reset fee manager
	fm.collectedAmount.Amount = math.ZeroInt()
	for _, fd := range fm.destinations {
		fd.amount = sdk.ZeroInt()
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
