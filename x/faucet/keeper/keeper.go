package keeper

import (
	"fmt"
	bank "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	mint "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	staking "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/tendermint/tendermint/libs/log"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sandblockio/chain/x/faucet/types"
)

type Keeper struct {
	BankKeeper    bank.Keeper
	MintKeeper    mint.Keeper
	StakingKeeper staking.Keeper

	defaultAmount int64
	defaultLimit  time.Duration

	cdc      codec.Marshaler
	storeKey sdk.StoreKey
	memKey   sdk.StoreKey
}

func NewKeeper(cdc codec.Marshaler, storeKey, memKey sdk.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		memKey:   memKey,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
