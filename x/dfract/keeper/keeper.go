package keeper

import (
	"fmt"
	"strconv"

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

func permContains(perms []string, perm string) bool {
	for _, v := range perms {
		if v == perm {
			return true
		}
	}

	return false
}

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
	moduleAddr, perms := auth.GetModuleAddressAndPermissions(types.ModuleName)
	if moduleAddr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// Ensure our module account have the required permissions
	if !permContains(perms, authtypes.Minter) {
		panic(fmt.Sprintf("%s module account should have the minter permission", types.ModuleName))
	}
	if !permContains(perms, authtypes.Burner) {
		panic(fmt.Sprintf("%s module account should have the burner permission", types.ModuleName))
	}

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

func (k Keeper) ProcessSpendAndAdjustProposal(ctx sdk.Context, proposal *types.SpendAndAdjustProposal) error {
	// Acquire the parameters to get the denoms
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// Acquire the module account balance
	balance := k.GetModuleAccountBalanceForDenom(ctx, params.GetDepositDenom())

	// Acquire the list of waiting proposal & waiting mint deposits
	waitingProposalDeposits := k.ListWaitingProposalDeposits(ctx)
	waitingMintDeposits := k.ListWaitingMintDeposits(ctx)

	// Spend the coins from the module account if we have one
	if balance.Amount.IsPositive() {
		destinationAddress, err := sdk.AccAddressFromBech32(proposal.GetSpendDestination())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}
		if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, destinationAddress, sdk.NewCoins(balance)); err != nil {
			return err
		}
	}

	// Process the waiting mint deposits
	for _, deposit := range waitingMintDeposits {
		depositorAddress, err := sdk.AccAddressFromBech32(deposit.GetDepositorAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}

		toMint := sdk.NewCoin(params.MintDenom, deposit.GetAmount().Amount.MulRaw(proposal.GetMintRate()))
		if err := k.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(toMint)); err != nil {
			return err
		}

		if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositorAddress, sdk.NewCoins(toMint)); err != nil {
			return err
		}

		k.RemoveFromWaitingMintDeposits(ctx, deposit.GetDepositorAddress())
		k.InsertIntoMintedDeposits(ctx, deposit.GetDepositorAddress(), *deposit)
	}

	// Process the waiting proposal deposits
	for _, deposit := range waitingProposalDeposits {
		k.RemoveFromWaitingProposalDeposits(ctx, deposit.GetDepositorAddress())
		k.InsertIntoWaitingMintDeposits(ctx, deposit.GetDepositorAddress(), *deposit)
	}

	// Trigger the events
	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeMint, sdk.NewAttribute(types.AttributeKeyMintBlock, strconv.FormatInt(ctx.BlockHeight(), 10))),
	})

	return nil
}
