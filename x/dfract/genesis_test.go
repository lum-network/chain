package dfract_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	apptypes "github.com/lum-network/chain/app"
	"github.com/lum-network/chain/x/dfract"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
	"time"
)

var now = time.Now().UTC()
var acc1 = sdk.AccAddress([]byte("addr1---------------"))
var acc2 = sdk.AccAddress([]byte("addr2---------------"))
var acc3 = sdk.AccAddress([]byte("addr3---------------"))
var testGenesis = types.GenesisState{
	ModuleAccountBalance: []sdk.Coin{
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 15_000),
	},
	Params: types.Params{
		MintDenom:    sdk.DefaultBondDenom,
		DepositDenom: sdk.DefaultBondDenom,
	},
	MintedDeposits: []*types.Deposit{
		{
			Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			DepositorAddress: acc1.String(),
		},
		{
			Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200)),
			DepositorAddress: acc2.String(),
		},
		{
			Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(300)),
			DepositorAddress: acc3.String(),
		},
	},
	WaitingMintDeposits: []*types.Deposit{
		{
			Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			DepositorAddress: acc1.String(),
		},
		{
			Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200)),
			DepositorAddress: acc2.String(),
		},
		{
			Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(300)),
			DepositorAddress: acc3.String(),
		},
	},
	WaitingProposalDeposits: []*types.Deposit{
		{
			Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			DepositorAddress: acc1.String(),
		},
		{
			Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200)),
			DepositorAddress: acc2.String(),
		},
		{
			Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(300)),
			DepositorAddress: acc3.String(),
		},
	},
}

func TestInitGenesis(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	dfract.InitGenesis(ctx, *app.DFractKeeper, testGenesis)
	balance := app.DFractKeeper.GetModuleAccountBalanceForDenom(ctx, sdk.DefaultBondDenom)
	require.Equal(t, balance.Amount.Int64(), testGenesis.ModuleAccountBalance[0].Amount.Int64())

	params, err := app.DFractKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, params, testGenesis.Params)

	mintedDeposits := app.DFractKeeper.ListMintedDeposits(ctx)
	require.Equal(t, len(mintedDeposits), len(testGenesis.MintedDeposits))

	for i, dep := range mintedDeposits {
		require.Equal(t, dep, testGenesis.MintedDeposits[i])
	}

	waitingProposalDeposits := app.DFractKeeper.ListMintedDeposits(ctx)
	require.Equal(t, len(waitingProposalDeposits), len(testGenesis.WaitingProposalDeposits))

	for i, dep := range waitingProposalDeposits {
		require.Equal(t, dep, testGenesis.WaitingProposalDeposits[i])
	}

	waitingMintDeposits := app.DFractKeeper.ListMintedDeposits(ctx)
	require.Equal(t, len(waitingMintDeposits), len(testGenesis.WaitingMintDeposits))

	for i, dep := range waitingMintDeposits {
		require.Equal(t, dep, testGenesis.WaitingMintDeposits[i])
	}
}

func TestExportGenesis(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	dfract.InitGenesis(ctx, *app.DFractKeeper, testGenesis)
	exportGenesis := dfract.ExportGenesis(ctx, *app.DFractKeeper)

	require.Equal(t, len(exportGenesis.WaitingMintDeposits), len(testGenesis.WaitingMintDeposits))
	require.Equal(t, len(exportGenesis.WaitingProposalDeposits), len(testGenesis.WaitingProposalDeposits))
	require.Equal(t, len(exportGenesis.MintedDeposits), len(testGenesis.MintedDeposits))

	require.Equal(t, exportGenesis.ModuleAccountBalance, testGenesis.ModuleAccountBalance)

	require.Equal(t, exportGenesis.Params, testGenesis.Params)
}
