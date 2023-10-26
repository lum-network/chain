package keeper

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/lum-network/chain/x/dfract/types"
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
		StakingKeeper *stakingkeeper.Keeper
	}
)

// NewKeeper Create a new keeper instance and return the pointer
func NewKeeper(cdc codec.BinaryCodec, storeKey, memKey storetypes.StoreKey, paramSpace paramtypes.Subspace, auth authkeeper.AccountKeeper, bank bankkeeper.Keeper, sk *stakingkeeper.Keeper) *Keeper {
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
	return k.BankKeeper.GetBalance(ctx, moduleAcc, denom)
}

func (k Keeper) UnsafeGetAllCoinsBack(ctx sdk.Context) (sdk.Coin, error) {
	accounts := k.AuthKeeper.GetAllAccounts(ctx)
	for _, account := range accounts {
		supply := k.BankKeeper.GetBalance(ctx, account.GetAddress(), types.MintDenom)
		if supply.IsPositive() {
			if err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, account.GetAddress(), types.ModuleName, sdk.NewCoins(supply)); err != nil {
				return sdk.Coin{}, err
			}
			k.Logger(ctx).Info(fmt.Sprintf("Transfer %s from %s to %s", supply.String(), account.GetAddress(), types.ModuleName))
		}
	}

	// Get the exact balance of the module account for the mint denom, and return it
	moduleAccount := k.AuthKeeper.GetModuleAccount(ctx, types.ModuleName)
	supply := k.BankKeeper.GetBalance(ctx, moduleAccount.GetAddress(), types.MintDenom)

	return supply, nil
}
