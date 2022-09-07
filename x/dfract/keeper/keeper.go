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

func (k Keeper) GetModuleAccount(ctx sdk.Context) sdk.AccAddress {
	return k.AuthKeeper.GetModuleAddress(types.ModuleName)
}

func (k Keeper) CreateModuleAccount(ctx sdk.Context, amount sdk.Coin) {
	moduleAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter)
	k.AuthKeeper.SetModuleAccount(ctx, moduleAcc)

	if err := k.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount)); err != nil {
		panic(err)
	}
}

func (k Keeper) GetModuleAccountBalance(ctx sdk.Context, denom string) sdk.Coin {
	moduleAcc := k.GetModuleAccount(ctx)
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	if denom != params.DepositDenom && denom != params.MintDenom {
		//TODO: handle that case
	}

	return k.BankKeeper.GetBalance(ctx, moduleAcc, denom)
}

func (k Keeper) GetDeposit(ctx sdk.Context, depositId string) (types.Deposit, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetDepositKey(depositId))
	if bz == nil {
		return types.Deposit{}, sdkerrors.Wrapf(types.ErrDepositNotFound, "ID %s does not exist", depositId)
	}

	var deposit types.Deposit
	err := k.cdc.Unmarshal(bz, &deposit)
	if err != nil {
		return types.Deposit{}, err
	}
	return deposit, nil
}

func (k Keeper) HasDeposit(ctx sdk.Context, depositId string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetDepositKey(depositId))
}

func (k Keeper) SetDeposit(ctx sdk.Context, depositId string, deposit *types.Deposit) {
	store := ctx.KVStore(k.storeKey)
	encodedDeposit := k.cdc.MustMarshal(deposit)
	store.Set(types.GetDepositKey(depositId), encodedDeposit)
}

func (k Keeper) CreateDeposit(ctx sdk.Context, msg types.MsgDeposit) error {
	// Make sure our ID does not exist yet
	if k.HasDeposit(ctx, msg.GetId()) {
		return types.ErrDepositAlreadyExists
	}

	// Acquire the parameters to get the denoms
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// Make sure the deposit is made of allowed denom
	if params.DepositDenom != msg.GetAmount().Denom {
		return types.ErrUnauthorizedDenom
	}

	// Create the deposit
	var deposit = &types.Deposit{
		DepositorAddress: msg.GetDepositorAddress(),
		Id:               msg.GetId(),
		Amount:           msg.GetAmount(),
		DepositedAt:      ctx.BlockTime(),
	}

	creatorAddress, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress())
	if err != nil {
		return sdkerrors.ErrInvalidAddress
	}

	// Move the funds
	if err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddress, types.ModuleName, sdk.NewCoins(msg.GetAmount())); err != nil {
		return err
	}

	// Store the deposit
	k.SetDeposit(ctx, deposit.GetId(), deposit)

	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeDeposit, sdk.NewAttribute(types.AttributeKeyDepositor, msg.GetDepositorAddress())),
	})
	return nil
}

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
	balance := k.GetModuleAccountBalance(ctx, params.GetDepositDenom())

	// Move funds to the destination address
	if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, destinationAddress, sdk.NewCoins(balance)); err != nil {
		return err
	}
	return nil
}

func (k Keeper) Mint(ctx sdk.Context, amount sdk.Coin) error {
	// Acquire the parameters to get the mint denom
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	if params.GetMintDenom() != amount.GetDenom() {
		return types.ErrUnauthorizedDenom
	}

	// Mint the coins
	if err := k.MintKeeper.MintCoins(ctx, sdk.NewCoins(amount)); err != nil {
		return err
	}

	// TODO: whom to transfer minted coins to
	return nil
}

func (k Keeper) DepositsIterator(ctx sdk.Context) sdk.Iterator {
	kvStore := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(kvStore, types.DepositsPrefix)
}

func (k Keeper) IterateDeposits(ctx sdk.Context, cb func(deposit types.Deposit) (stop bool)) {
	iterator := k.DepositsIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		err := k.cdc.Unmarshal(iterator.Value(), &deposit)
		if err != nil {
			panic(err)
		}
		if cb(deposit) {
			break
		}
	}
}
