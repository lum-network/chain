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
	DepositsMinted: []*types.Deposit{
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
	DepositsPendingMint: []*types.Deposit{
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
	DepositsPendingWithdrawal: []*types.Deposit{
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

	mintedDeposits := app.DFractKeeper.ListDepositsMinted(ctx)
	require.Equal(t, len(mintedDeposits), len(testGenesis.DepositsMinted))

	for i, dep := range mintedDeposits {
		require.Equal(t, dep, testGenesis.DepositsMinted[i])
	}

	waitingProposalDeposits := app.DFractKeeper.ListDepositsPendingWithdrawal(ctx)
	require.Equal(t, len(waitingProposalDeposits), len(testGenesis.DepositsPendingWithdrawal))

	for i, dep := range waitingProposalDeposits {
		require.Equal(t, dep, testGenesis.DepositsPendingWithdrawal[i])
	}

	waitingMintDeposits := app.DFractKeeper.ListDepositsPendingMint(ctx)
	require.Equal(t, len(waitingMintDeposits), len(testGenesis.DepositsPendingMint))

	for i, dep := range waitingMintDeposits {
		require.Equal(t, dep, testGenesis.DepositsPendingMint[i])
	}
}

func TestExportGenesis(t *testing.T) {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	dfract.InitGenesis(ctx, *app.DFractKeeper, testGenesis)
	exportGenesis := dfract.ExportGenesis(ctx, *app.DFractKeeper)

	require.Equal(t, len(exportGenesis.DepositsPendingMint), len(testGenesis.DepositsPendingMint))
	require.Equal(t, len(exportGenesis.DepositsPendingWithdrawal), len(testGenesis.DepositsPendingWithdrawal))
	require.Equal(t, len(exportGenesis.DepositsMinted), len(testGenesis.DepositsMinted))

	require.Equal(t, exportGenesis.ModuleAccountBalance, testGenesis.ModuleAccountBalance)

	require.Equal(t, exportGenesis.Params, testGenesis.Params)
}
