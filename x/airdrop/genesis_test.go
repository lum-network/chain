package airdrop_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/simapp"
	"github.com/lum-network/chain/x/airdrop"
	"github.com/lum-network/chain/x/airdrop/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
	"time"
)

var now = time.Now().UTC()
var acc1 = sdk.AccAddress([]byte("addr1---------------"))
var acc2 = sdk.AccAddress([]byte("addr2---------------"))
var testGenesis = types.GenesisState{
	ModuleAccountBalance: sdk.NewInt64Coin(types.DefaultClaimDenom, 750000000),
	Params: types.Params{
		AirdropStartTime:   now,
		DurationUntilDecay: types.DefaultDurationUntilDecay,
		DurationOfDecay:    types.DefaultDurationOfDecay,
		ClaimDenom:         types.DefaultClaimDenom, // ulum
	},
	ClaimRecords: []types.ClaimRecord{
		{
			Address:                acc1.String(),
			InitialClaimableAmount: sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 1000000000)},
			ActionCompleted:        []bool{true, false},
		},
		{
			Address:                acc2.String(),
			InitialClaimableAmount: sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 500000000)},
			ActionCompleted:        []bool{false, false},
		},
	},
}

func TestAirdropInitGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	airdrop.InitGenesis(ctx, *app.AirdropKeeper, genesis)
	app.AirdropKeeper.CreateModuleAccount(ctx, sdk.NewInt64Coin(types.DefaultClaimDenom, 750000000))

	coin := app.AirdropKeeper.GetAirdropAccountBalance(ctx)
	require.Equal(t, coin.String(), genesis.ModuleAccountBalance.String())

	params, err := app.AirdropKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, params, genesis.Params)

	claimRecords := app.AirdropKeeper.GetClaimRecords(ctx)
	require.Equal(t, claimRecords, genesis.ClaimRecords)
}

func TestAirdropExportGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	airdrop.InitGenesis(ctx, *app.AirdropKeeper, genesis)
	app.AirdropKeeper.CreateModuleAccount(ctx, sdk.NewInt64Coin(types.DefaultClaimDenom, 750000000))

	claimRecord, err := app.AirdropKeeper.GetClaimRecord(ctx, acc2)
	require.NoError(t, err)
	require.Equal(t, claimRecord, types.ClaimRecord{
		Address:                acc2.String(),
		InitialClaimableAmount: sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 500000000)},
		ActionCompleted:        []bool{false, false},
	})

	claimableAmount, err := app.AirdropKeeper.GetClaimableAmountForAction(ctx, acc2, types.ActionVote)
	require.NoError(t, err)
	require.Equal(t, claimableAmount, sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 250000000)})

	app.AirdropKeeper.AfterProposalVote(ctx, 12, acc2)

	genesisExported := airdrop.ExportGenesis(ctx, *app.AirdropKeeper)
	require.Equal(t, genesisExported.ModuleAccountBalance, genesis.ModuleAccountBalance.Sub(claimableAmount[0]))
	require.Equal(t, genesisExported.Params, genesis.Params)
	require.Equal(t, genesisExported.ClaimRecords, []types.ClaimRecord{
		{
			Address:                acc1.String(),
			InitialClaimableAmount: sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 1000000000)},
			ActionCompleted:        []bool{true, false},
		},
		{
			Address:                acc2.String(),
			InitialClaimableAmount: sdk.Coins{sdk.NewInt64Coin(types.DefaultClaimDenom, 500000000)},
			ActionCompleted:        []bool{true, false},
		},
	})
}

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := simapp.MakeTestEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := airdrop.NewAppModule(appCodec, *app.AirdropKeeper)

	genesis := testGenesis
	airdrop.InitGenesis(ctx, *app.AirdropKeeper, genesis)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	assert.NotPanics(t, func() {
		app := simapp.Setup(t, false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := airdrop.NewAppModule(appCodec, *app.AirdropKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}