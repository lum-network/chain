package keeper

import (
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/airdrop/types"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		storeKey      storetypes.StoreKey
		memKey        storetypes.StoreKey
		AuthKeeper    authkeeper.AccountKeeper
		BankKeeper    bankkeeper.Keeper
		StakingKeeper stakingkeeper.Keeper
		DistrKeeper   distrkeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	sk stakingkeeper.Keeper,
	dk distrkeeper.Keeper,

) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		AuthKeeper:    ak,
		BankKeeper:    bk,
		StakingKeeper: sk,
		DistrKeeper:   dk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetAirdropAccount(ctx sdk.Context) sdk.AccAddress {
	return k.AuthKeeper.GetModuleAddress(types.ModuleName)
}

// CreateModuleAccount create the module account
func (k Keeper) CreateModuleAccount(ctx sdk.Context, amount sdk.Coin) {
	moduleAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter)
	k.AuthKeeper.SetModuleAccount(ctx, moduleAcc)

	if err := k.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount)); err != nil {
		panic(err)
	}
}

// GetAirdropAccountBalance gets the airdrop coin balance of module account
func (k Keeper) GetAirdropAccountBalance(ctx sdk.Context) sdk.Coin {
	moduleAccAddr := k.GetAirdropAccount(ctx)
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return k.BankKeeper.GetBalance(ctx, moduleAccAddr, params.ClaimDenom)
}

func (k Keeper) EndAirdrop(ctx sdk.Context) error {
	err := k.fundRemainingsToCommunity(ctx)
	if err != nil {
		return err
	}
	k.clearInitialClaimables(ctx)
	return nil
}

// clearInitialClaimables clear claimable amounts
func (k Keeper) clearInitialClaimables(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.ClaimRecordsStorePrefix))
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}
	iterator.Close()
}

// SetClaimRecords set claimable amount from balances object
func (k Keeper) SetClaimRecords(ctx sdk.Context, claimRecords []types.ClaimRecord) error {
	for _, claimRecord := range claimRecords {
		err := k.SetClaimRecord(ctx, claimRecord)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetClaimRecords get claimables for genesis export
func (k Keeper) GetClaimRecords(ctx sdk.Context) []types.ClaimRecord {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimRecordsStorePrefix))

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	claimRecords := []types.ClaimRecord{}
	for ; iterator.Valid(); iterator.Next() {

		claimRecord := types.ClaimRecord{}

		err := proto.Unmarshal(iterator.Value(), &claimRecord)
		if err != nil {
			panic(err)
		}

		claimRecords = append(claimRecords, claimRecord)
	}
	return claimRecords
}

// GetClaimRecord returns the claim record for a specific address
func (k Keeper) GetClaimRecord(ctx sdk.Context, addr sdk.AccAddress) (types.ClaimRecord, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimRecordsStorePrefix))
	if !prefixStore.Has(addr) {
		return types.ClaimRecord{}, nil
	}
	bz := prefixStore.Get(addr)

	claimRecord := types.ClaimRecord{}
	err := proto.Unmarshal(bz, &claimRecord)
	if err != nil {
		return types.ClaimRecord{}, err
	}

	return claimRecord, nil
}

// SetClaimRecord sets a claim record for an address in store
func (k Keeper) SetClaimRecord(ctx sdk.Context, claimRecord types.ClaimRecord) error {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimRecordsStorePrefix))

	bz, err := proto.Marshal(&claimRecord)
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(claimRecord.Address)
	if err != nil {
		return err
	}

	prefixStore.Set(addr, bz)
	return nil
}

// GetClaimableAmountForAction returns claimable amount (free, vested) for a specific action done by an address
func (k Keeper) GetClaimableAmountForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Coin, sdk.Coin, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	zeroCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0))
	claimRecord, err := k.GetClaimRecord(ctx, addr)
	if err != nil {
		return zeroCoin, zeroCoin, err
	}

	if claimRecord.Address == "" {
		return zeroCoin, zeroCoin, nil
	}

	// if action already completed, nothing is claimable
	if len(claimRecord.ActionCompleted) > int(action) && claimRecord.ActionCompleted[action] {
		return zeroCoin, zeroCoin, nil
	}

	// If we are before the start time, do nothing.
	// This case _shouldn't_ occur on chain, since the
	// start time ought to be chain start time.
	if ctx.BlockTime().Before(params.AirdropStartTime) {
		return zeroCoin, zeroCoin, nil
	}

	InitialClaimablePerAction := sdk.Coins{}
	for _, coin := range claimRecord.InitialClaimableAmount {
		// Volountary do not use coins.Add method to avoid merging coins
		InitialClaimablePerAction = append(
			InitialClaimablePerAction, sdk.NewCoin(coin.Denom, coin.Amount.QuoRaw(int64(len(types.Action_name)))),
		)
	}

	if len(InitialClaimablePerAction) != 2 {
		// claimable amount should always contain 2 entries
		// first entry for free upon claim
		// second entry for vested upon claim
		return zeroCoin, zeroCoin, errors.New("invalid claimable entry")
	}

	elapsedAirdropTime := ctx.BlockTime().Sub(params.AirdropStartTime)
	// Are we early enough in the airdrop s.t. theres no decay?
	if elapsedAirdropTime <= params.DurationUntilDecay {
		return InitialClaimablePerAction[0], InitialClaimablePerAction[1], nil
	}

	// The entire airdrop has completed
	if elapsedAirdropTime > params.DurationUntilDecay+params.DurationOfDecay {
		return zeroCoin, zeroCoin, nil
	}

	// Positive, since goneTime > params.DurationUntilDecay
	decayTime := elapsedAirdropTime - params.DurationUntilDecay
	decayPercent := sdk.NewDec(decayTime.Nanoseconds()).QuoInt64(params.DurationOfDecay.Nanoseconds())
	claimablePercent := sdk.OneDec().Sub(decayPercent)

	claimableCoins := sdk.Coins{}
	for _, coin := range InitialClaimablePerAction {
		claimableCoins = claimableCoins.Add(sdk.NewCoin(coin.Denom, coin.Amount.Mul(claimablePercent.RoundInt())))
	}

	return claimableCoins[0], claimableCoins[1], nil
}

// GetUserTotalClaimable returns total claimable amounts for an address
func (k Keeper) GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, sdk.Coin, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	claimableFreeCoin := sdk.NewCoin(params.ClaimDenom, sdk.NewInt(0))
	claimableVestedCoin := sdk.NewCoin(params.ClaimDenom, sdk.NewInt(0))

	claimRecord, err := k.GetClaimRecord(ctx, addr)
	if err != nil {
		return claimableFreeCoin, claimableVestedCoin, err
	}
	if claimRecord.Address == "" {
		return claimableFreeCoin, claimableVestedCoin, nil
	}

	for action := range types.Action_name {
		cfc, cvc, err := k.GetClaimableAmountForAction(ctx, addr, types.Action(action))
		if err != nil {
			return cfc, cvc, err
		}
		if cfc.Denom != claimableFreeCoin.Denom || cvc.Denom != claimableVestedCoin.Denom {
			return cfc, cvc, errors.New("inconsistent claimable entry")
		}
		claimableFreeCoin.Amount = claimableFreeCoin.Amount.Add(cfc.Amount)
		claimableVestedCoin.Amount = claimableVestedCoin.Amount.Add(cvc.Amount)
	}
	return claimableFreeCoin, claimableVestedCoin, nil
}

// ClaimCoinsForAction remove claimable amount entry and transfer it to user's account
func (k Keeper) ClaimCoinsForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Coin, sdk.Coin, error) {
	claimableFreeCoin, claimableVestedCoin, err := k.GetClaimableAmountForAction(ctx, addr, action)
	if err != nil {
		return claimableFreeCoin, claimableVestedCoin, err
	}

	if claimableFreeCoin.Amount.IsZero() && claimableVestedCoin.IsZero() {
		return claimableFreeCoin, claimableVestedCoin, nil
	}

	claimRecord, err := k.GetClaimRecord(ctx, addr)
	if err != nil {
		return claimableFreeCoin, claimableVestedCoin, err
	}

	// First send coins to bank
	err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.Coins{claimableFreeCoin}.Add(claimableVestedCoin))
	if err != nil {
		return claimableFreeCoin, claimableVestedCoin, err
	}

	// If some coins are vested updated the vesting account
	if !claimableVestedCoin.Amount.IsZero() {
		acc := k.AuthKeeper.GetAccount(ctx, addr)
		if acc == nil {
			return claimableFreeCoin, claimableVestedCoin, errors.New("account not found")
		}
		vestingAcc, ok := acc.(*vestingtypes.ContinuousVestingAccount)
		if !ok {
			return claimableFreeCoin, claimableVestedCoin, errors.New("account should have continuous vesting type")
		}
		vestingAcc.OriginalVesting = vestingAcc.OriginalVesting.Add(claimableVestedCoin)
		k.AuthKeeper.SetAccount(ctx, vestingAcc)
	}

	claimRecord.ActionCompleted[action] = true

	err = k.SetClaimRecord(ctx, claimRecord)
	if err != nil {
		return claimableFreeCoin, claimableVestedCoin, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(sdk.AttributeKeySender, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, claimableFreeCoin.String()),
			sdk.NewAttribute("amount_vested", claimableVestedCoin.String()),
		),
	})

	return claimableFreeCoin, claimableVestedCoin, nil
}

// fundRemainingsToCommunity fund remainings to the community when airdrop period end
func (k Keeper) fundRemainingsToCommunity(ctx sdk.Context) error {
	moduleAccAddr := k.GetAirdropAccount(ctx)
	amt := k.GetAirdropAccountBalance(ctx)
	return k.DistrKeeper.FundCommunityPool(ctx, sdk.NewCoins(amt), moduleAccAddr)
}
