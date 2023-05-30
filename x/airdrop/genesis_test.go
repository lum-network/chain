package airdrop_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	apptypes "github.com/lum-network/chain/app"
	"github.com/lum-network/chain/x/airdrop"
	"github.com/lum-network/chain/x/airdrop/types"
	"github.com/stretchr/testify/require"
)

var now = time.Now().UTC()
var acc1 = sdk.AccAddress([]byte("addr1---------------"))
var acc2 = sdk.AccAddress([]byte("addr2---------------"))
var acc3 = sdk.AccAddress([]byte("addr3---------------"))
var testGenesis = types.GenesisState{
	ModuleAccountBalance: sdk.NewInt64Coin(types.DefaultClaimDenom, 15_000),
	Params: types.Params{
		AirdropStartTime:   now,
		DurationUntilDecay: types.DefaultDurationUntilDecay,
		DurationOfDecay:    types.DefaultDurationOfDecay,
		ClaimDenom:         types.DefaultClaimDenom, // ulum
	},
	ClaimRecords: []types.ClaimRecord{
		{
			Address: acc1.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(types.DefaultClaimDenom, 1_000),
				sdk.NewInt64Coin(types.DefaultClaimDenom, 5_000),
			},
			ActionCompleted: []bool{false, false},
		},
		{
			Address: acc2.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(types.DefaultClaimDenom, 3_000),
				sdk.NewInt64Coin(types.DefaultClaimDenom, 2_000),
			},
			ActionCompleted: []bool{false, false},
		},
		{
			Address: acc3.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(types.DefaultClaimDenom, 1_000),
				sdk.NewInt64Coin(types.DefaultClaimDenom, 3_000),
			},
			ActionCompleted: []bool{false, false},
		},
	},
}

func TestAirdropInitGenesis(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	airdrop.InitGenesis(ctx, *app.AirdropKeeper, genesis)

	coin := app.AirdropKeeper.GetAirdropAccountBalance(ctx)
	require.Equal(t, coin.String(), genesis.ModuleAccountBalance.String())

	params, err := app.AirdropKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, params, genesis.Params)

	claimRecords := app.AirdropKeeper.GetClaimRecords(ctx)
	require.Equal(t, len(claimRecords), len(genesis.ClaimRecords))
	for i, rec := range claimRecords {
		require.Equal(t, rec, genesis.ClaimRecords[i])
	}
}

func TestAirdropExportGenesis(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	airdrop.InitGenesis(ctx, *app.AirdropKeeper, genesis)

	for _, addr := range []sdk.AccAddress{acc1, acc2, acc3} {
		app.AccountKeeper.SetAccount(
			ctx,
			vestingtypes.NewContinuousVestingAccount(authtypes.NewBaseAccountWithAddress(addr), sdk.Coins{}, now.Unix(), now.Add(24*time.Hour).Unix()),
		)
	}

	claimRecord, err := app.AirdropKeeper.GetClaimRecord(ctx, acc1)
	require.NoError(t, err)
	require.Equal(t, claimRecord, types.ClaimRecord{
		Address: acc1.String(),
		InitialClaimableAmount: sdk.Coins{
			sdk.NewInt64Coin(types.DefaultClaimDenom, 1_000),
			sdk.NewInt64Coin(types.DefaultClaimDenom, 5_000),
		},
		ActionCompleted: []bool{false, false},
	})

	claimableFreeAmount, claimableVestedAmount, err := app.AirdropKeeper.GetClaimableAmountForAction(ctx, acc1, types.ActionVote)
	require.NoError(t, err)
	require.Equal(t, claimableFreeAmount, sdk.NewInt64Coin(types.DefaultClaimDenom, 500))
	require.Equal(t, claimableVestedAmount, sdk.NewInt64Coin(types.DefaultClaimDenom, 2_500))

	app.AirdropKeeper.AfterProposalVote(ctx, 12, acc1)

	claimableFreeAmount, claimableVestedAmount, err = app.AirdropKeeper.GetClaimableAmountForAction(ctx, acc1, types.ActionVote)
	require.NoError(t, err)
	require.Equal(t, claimableFreeAmount, sdk.NewInt64Coin(types.DefaultClaimDenom, 0))
	require.Equal(t, claimableVestedAmount, sdk.NewInt64Coin(types.DefaultClaimDenom, 0))

	claimableFreeAmount, claimableVestedAmount, err = app.AirdropKeeper.GetClaimableAmountForAction(ctx, acc1, types.ActionDelegateStake)
	require.NoError(t, err)
	require.Equal(t, claimableFreeAmount, sdk.NewInt64Coin(types.DefaultClaimDenom, 500))
	require.Equal(t, claimableVestedAmount, sdk.NewInt64Coin(types.DefaultClaimDenom, 2_500))

	genesisExported := airdrop.ExportGenesis(ctx, *app.AirdropKeeper)
	require.Equal(t, genesisExported.ModuleAccountBalance, genesis.ModuleAccountBalance.Sub(claimableFreeAmount).Sub(claimableVestedAmount))
	require.Equal(t, genesisExported.Params, genesis.Params)
	require.Equal(t, genesisExported.ClaimRecords, []types.ClaimRecord{
		{
			Address: acc1.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(types.DefaultClaimDenom, 1_000),
				sdk.NewInt64Coin(types.DefaultClaimDenom, 5_000),
			},
			ActionCompleted: []bool{true, false},
		},
		{
			Address: acc2.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(types.DefaultClaimDenom, 3_000),
				sdk.NewInt64Coin(types.DefaultClaimDenom, 2_000),
			},
			ActionCompleted: []bool{false, false},
		},
		{
			Address: acc3.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(types.DefaultClaimDenom, 1_000),
				sdk.NewInt64Coin(types.DefaultClaimDenom, 3_000),
			},
			ActionCompleted: []bool{false, false},
		},
	})
}

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := apptypes.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := airdrop.NewAppModule(appCodec, *app.AirdropKeeper)

	genesis := testGenesis
	airdrop.InitGenesis(ctx, *app.AirdropKeeper, genesis)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	require.NotPanics(t, func() {
		app := apptypes.SetupForTesting(false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := airdrop.NewAppModule(appCodec, *app.AirdropKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}
