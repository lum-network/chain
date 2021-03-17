package keeper

import (
	"fmt"
	bank "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	staking "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/tendermint/tendermint/libs/log"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/faucet/types"
)

type Keeper struct {
	BankKeeper    bank.Keeper
	StakingKeeper staking.Keeper

	defaultAmount int64
	defaultLimit  time.Duration

	cdc      codec.Marshaler
	storeKey sdk.StoreKey
	memKey   sdk.StoreKey
}

// NewKeeper Initialize the keeper instance with the required elements
func NewKeeper(bankKeeper bank.Keeper, stakingKeeper staking.Keeper, cdc codec.Marshaler, storeKey, memKey sdk.StoreKey, amount int64, rateLimit time.Duration) *Keeper {
	return &Keeper{
		BankKeeper:    bankKeeper,
		StakingKeeper: stakingKeeper,
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		defaultAmount: amount,
		defaultLimit:  rateLimit,
	}
}

// Logger Return a logger instance for the keeper
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
