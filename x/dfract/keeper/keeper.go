package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		AuthKeeper authkeeper.AccountKeeper
		BankKeeper bankkeeper.Keeper
		GovKeeper  govkeeper.Keeper
		MintKeeper mintkeeper.Keeper
	}
)

// NewKeeper Create a new keeper instance and return the pointer
func NewKeeper(cdc codec.BinaryCodec, storeKey, memKey sdk.StoreKey, auth authkeeper.AccountKeeper, bank bankkeeper.Keeper, mk mintkeeper.Keeper) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		AuthKeeper: auth,
		BankKeeper: bank,
		MintKeeper: mk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetModuleAccount Return the module account address
func (k Keeper) GetModuleAccount(ctx sdk.Context) sdk.AccAddress {
	return k.AuthKeeper.GetModuleAddress(types.ModuleName)
}

// CreateModuleAccount Initialize the module account and set the original amount of coins
func (k Keeper) CreateModuleAccount(ctx sdk.Context, amount sdk.Coins) {
	moduleAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter)
	k.AuthKeeper.SetModuleAccount(ctx, moduleAcc)

	if err := k.BankKeeper.MintCoins(ctx, types.ModuleName, amount); err != nil {
		panic(err)
	}
}

func (k Keeper) GetModuleAccountBalance(ctx sdk.Context) sdk.Coins {
	moduleAcc := k.GetModuleAccount(ctx)
	return k.BankKeeper.GetAllBalances(ctx, moduleAcc)
}

// GetModuleAccountBalanceForDenom Return the module account's balance
func (k Keeper) GetModuleAccountBalanceForDenom(ctx sdk.Context, denom string) sdk.Coin {
	moduleAcc := k.GetModuleAccount(ctx)
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	if denom != params.DepositDenom && denom != params.MintDenom {
		return sdk.Coin{}
	}

	return k.BankKeeper.GetBalance(ctx, moduleAcc, denom)
}

// CreateDeposit Process the deposit message
func (k Keeper) CreateDeposit(ctx sdk.Context, msg types.MsgDeposit) error {
	// Acquire the parameters to get the denoms
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// Make sure the deposit is made of allowed denom
	if params.DepositDenom != msg.GetAmount().Denom {
		return types.ErrUnauthorizedDenom
	}

	// Make sure we have an actual deposit to do
	if msg.GetAmount().IsNegative() || msg.GetAmount().IsZero() {
		return types.ErrEmptyDepositAmount
	}

	// Does the deposit exists ? If not, create it otherwise just append the amount
	deposit, found := k.GetFromWaitingProposalDeposits(ctx, msg.GetDepositorAddress())
	if !found {
		deposit = types.Deposit{
			DepositorAddress: msg.GetDepositorAddress(),
			Amount:           msg.GetAmount(),
			CreatedAt:        ctx.BlockTime(),
		}
	} else {
		deposit.Amount = deposit.Amount.Add(msg.GetAmount())
	}

	// Cast the depositor address
	depositorAddress, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress())
	if err != nil {
		return sdkerrors.ErrInvalidAddress
	}

	// Move the funds
	if err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, depositorAddress, types.ModuleName, sdk.NewCoins(msg.GetAmount())); err != nil {
		return err
	}

	// Insert into queue
	k.InsertIntoWaitingProposalDeposits(ctx, deposit.GetDepositorAddress(), deposit)

	// Trigger the events
	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeDeposit, sdk.NewAttribute(types.AttributeKeyDepositor, msg.GetDepositorAddress())),
	})
	return nil
}

// Spend Send the entire module account balance to a given address
func (k Keeper) Spend(ctx sdk.Context, destinationAddressStr string) error {
	// Acquire the parameters to get the denoms
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// Prepare destination address
	destinationAddress, err := sdk.AccAddressFromBech32(destinationAddressStr)
	if err != nil {
		return sdkerrors.ErrInvalidAddress
	}

	// Acquire balance
	balance := k.GetModuleAccountBalanceForDenom(ctx, params.GetDepositDenom())

	// Move funds to the destination address
	if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, destinationAddress, sdk.NewCoins(balance)); err != nil {
		return err
	}
	return nil
}

// Mint Emit a new amount of coins
func (k Keeper) Mint(ctx sdk.Context, mintRate int64) error {
	// Acquire the parameters to get the mint denom
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// Compute the total amount
	totalAmount := sdk.NewInt(0)
	deposits := k.ListWaitingMintDeposits(ctx)
	for _, deposit := range deposits {
		totalAmount = totalAmount.Add(deposit.GetAmount().Amount)
	}

	// Make sure we actually have something to mint
	if totalAmount.IsZero() {
		return types.ErrMintDontMatchTotal
	}

	// Mint the coins
	totalAmount.MulRaw(mintRate)
	if err := k.MintKeeper.MintCoins(ctx, sdk.NewCoins(sdk.NewCoin(params.MintDenom, totalAmount))); err != nil {
		return err
	}
	return nil
}

// Distribute Split the amount of emitted coins and distribute them to the depositors
func (k Keeper) Distribute(ctx sdk.Context, mintRate int64, deposits []*types.Deposit) error {
	// Acquire the parameters to get the mint denom
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// For each deposit, we will compute the amount of coins to send from the mint rate and process the distribution
	for _, deposit := range deposits {
		// Correctly format address
		destinationAddress, err := sdk.AccAddressFromBech32(deposit.GetDepositorAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}

		// Compute the coins to send from the rate
		mintedAmount := sdk.NewCoin(params.MintDenom, deposit.GetAmount().Amount.MulRaw(mintRate))

		// Transfer the coins
		if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, destinationAddress, sdk.NewCoins(mintedAmount)); err != nil {
			return err
		}

		// Remove from the waiting queue and move to minted
		k.RemoveFromWaitingMintDeposits(ctx, deposit.GetDepositorAddress())
		k.InsertIntoMintedDeposits(ctx, deposit.GetDepositorAddress(), *deposit)
	}
	return nil
}
