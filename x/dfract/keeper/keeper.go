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
	}
)

// NewKeeper Create a new keeper instance and return the pointer
func NewKeeper(cdc codec.BinaryCodec, storeKey, memKey sdk.StoreKey, auth authkeeper.AccountKeeper, bank bankkeeper.Keeper, gk govkeeper.Keeper) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		AuthKeeper: auth,
		BankKeeper: bank,
		GovKeeper:  gk,
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
	// Make sure our ID does not exists yet
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

	// Move the funds
	// TODO: implement

	// Store the deposit
	k.SetDeposit(ctx, deposit.GetId(), deposit)

	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeDeposit, sdk.NewAttribute(types.AttributeKeyDepositor, msg.GetDepositorAddress())),
	})
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
