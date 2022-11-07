package keeper

import (
	"fmt"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/tendermint/tendermint/libs/log"
)

const MicroPrecision = 1000000

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
		cdc           codec.BinaryCodec
		storeKey      storetypes.StoreKey
		memKey        storetypes.StoreKey
		paramSpace    paramtypes.Subspace
		AuthKeeper    authkeeper.AccountKeeper
		BankKeeper    bankkeeper.Keeper
		GovKeeper     govkeeper.Keeper
		StakingKeeper stakingkeeper.Keeper
	}
)

// NewKeeper Create a new keeper instance and return the pointer
func NewKeeper(cdc codec.BinaryCodec, storeKey, memKey storetypes.StoreKey, paramSpace paramtypes.Subspace, auth authkeeper.AccountKeeper, bank bankkeeper.Keeper, sk stakingkeeper.Keeper) *Keeper {
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

	// Make sure we have the parameters table
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramSpace:    paramSpace,
		AuthKeeper:    auth,
		BankKeeper:    bank,
		StakingKeeper: sk,
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
	params := k.GetParams(ctx)

	if denom != params.DepositDenom && denom != types.MintDenom {
		return sdk.Coin{}
	}

	return k.BankKeeper.GetBalance(ctx, moduleAcc, denom)
}

// CreateDeposit Process the deposit message
func (k Keeper) CreateDeposit(ctx sdk.Context, msg types.MsgDeposit) error {
	// Acquire the parameters to get the denoms
	params := k.GetParams(ctx)

	// Make sure the deposit is made of allowed denom
	if params.DepositDenom != msg.GetAmount().Denom {
		return types.ErrUnauthorizedDepositDenom
	}

	// Make sure we have an actual deposit to do
	if msg.GetAmount().IsNegative() || msg.GetAmount().IsZero() {
		return types.ErrEmptyDepositAmount
	}

	// Make sure the deposit is sufficient
	if msg.GetAmount().Amount.LT(sdk.NewInt(int64(params.MinDepositAmount))) {
		return types.ErrInsufficientDepositAmount
	}

	// Cast the depositor address
	depositorAddress, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress())
	if err != nil {
		return sdkerrors.ErrInvalidAddress
	}

	// Does the deposit exists ? If not, create it otherwise just append the amount
	deposit, found := k.GetDepositPendingWithdrawal(ctx, depositorAddress)
	if !found {
		deposit = types.Deposit{
			DepositorAddress: msg.GetDepositorAddress(),
			Amount:           msg.GetAmount(),
			CreatedAt:        ctx.BlockTime(),
		}
	} else {
		deposit.Amount = deposit.Amount.Add(msg.GetAmount())
	}

	// Move the funds
	if err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, depositorAddress, types.ModuleName, sdk.NewCoins(msg.GetAmount())); err != nil {
		return err
	}

	// Insert into queue
	k.SetDepositPendingWithdrawal(ctx, depositorAddress, deposit)

	// Trigger the events
	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDeposit,
			sdk.NewAttribute(types.AttributeKeyDepositor, msg.GetDepositorAddress()),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.GetAmount().String()),
		),
	})
	return nil
}

func (k Keeper) ProcessWithdrawAndMintProposal(ctx sdk.Context, proposal *types.WithdrawAndMintProposal) error {
	// Acquire the parameters to get the denoms
	params := k.GetParams(ctx)

	// Acquire the module account balance
	balance := k.GetModuleAccountBalanceForDenom(ctx, params.GetDepositDenom())

	// Acquire the list of deposits pending withdrawal and mint
	depositsPendingWithdrawal := k.ListDepositsPendingWithdrawal(ctx)
	depositsPendingMint := k.ListDepositsPendingMint(ctx)

	// Withdrawal the coins from the pending withdrawal deposits using the module account
	if balance.Amount.IsPositive() {
		destinationAddress, err := sdk.AccAddressFromBech32(proposal.GetWithdrawalAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}
		if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, destinationAddress, sdk.NewCoins(balance)); err != nil {
			return err
		}
	}

	// Mint the coins for all the pending mint deposits and add them to the the minted deposits
	for _, deposit := range depositsPendingMint {
		depositorAddress, err := sdk.AccAddressFromBech32(deposit.GetDepositorAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}

		toMint := sdk.NewCoin(types.MintDenom, deposit.GetAmount().Amount.MulRaw(proposal.GetMicroMintRate()).QuoRaw(MicroPrecision))
		if err := k.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(toMint)); err != nil {
			return err
		}

		if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositorAddress, sdk.NewCoins(toMint)); err != nil {
			return err
		}

		k.RemoveDepositPendingMint(ctx, depositorAddress)
		k.AddDepositMinted(ctx, depositorAddress, *deposit)
		ctx.EventManager().Events().AppendEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeMint,
				sdk.NewAttribute(types.AttributeKeyDepositor, depositorAddress.String()),
				sdk.NewAttribute(types.AttributeKeyAmount, deposit.GetAmount().String()),
				sdk.NewAttribute(types.AttributeKeyMinted, toMint.String()),
				sdk.NewAttribute(types.AttributeKeyMicroMintRate, fmt.Sprintf("%d", proposal.GetMicroMintRate())),
			),
		})
	}

	// Process the pending withdrawal deposits
	for _, deposit := range depositsPendingWithdrawal {
		depositorAddress, err := sdk.AccAddressFromBech32(deposit.GetDepositorAddress())
		if err != nil {
			return sdkerrors.ErrInvalidAddress
		}
		k.RemoveDepositPendingWithdrawal(ctx, depositorAddress)
		k.SetDepositPendingMint(ctx, depositorAddress, *deposit)
		ctx.EventManager().Events().AppendEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeWithdraw,
				sdk.NewAttribute(types.AttributeKeyDepositor, depositorAddress.String()),
				sdk.NewAttribute(types.AttributeKeyAmount, deposit.GetAmount().String()),
			),
		})
	}

	// Trigger the events
	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeMint, sdk.NewAttribute(types.AttributeKeyMintBlock, strconv.FormatInt(ctx.BlockHeight(), 10))),
	})

	return nil
}
