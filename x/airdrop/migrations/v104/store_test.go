package v104_test

import (
	"testing"
	"time"

	apptypes "github.com/lum-network/chain/app"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"

	v104 "github.com/lum-network/chain/x/airdrop/migrations/v104"
	"github.com/lum-network/chain/x/airdrop/types"
)

var (
	defaultClaimDenom = sdk.DefaultBondDenom
	now               = time.Now().UTC()
)

type StoreMigrationTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	queryClient types.QueryClient
	app         *apptypes.App
}

// SetupTest Create our testing app, and make sure everything is correctly usable.
func (suite *StoreMigrationTestSuite) SetupTest() {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.AirdropKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient

	airdropStartTime := time.Now()
	suite.app.AirdropKeeper.CreateModuleAccount(suite.ctx, sdk.NewCoin(defaultClaimDenom, sdk.NewInt(0)))
	err := suite.app.AirdropKeeper.SetParams(suite.ctx, types.Params{
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: types.DefaultDurationUntilDecay,
		DurationOfDecay:    types.DefaultDurationOfDecay,
		ClaimDenom:         defaultClaimDenom,
	})
	if err != nil {
		panic(err)
	}
	suite.ctx = suite.ctx.WithBlockTime(airdropStartTime)
}

func (suite *StoreMigrationTestSuite) TestMigration() {
	// Simulate legacy faulty state
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// initialize accts
	suite.app.AccountKeeper.SetAccount(
		suite.ctx,
		vestingtypes.NewContinuousVestingAccount(
			authtypes.NewBaseAccountWithAddress(addr1),
			sdk.Coins{sdk.NewInt64Coin(defaultClaimDenom, 0)},
			now.Unix(),
			now.Add(24*time.Hour).Unix(),
		),
	)
	suite.app.AccountKeeper.SetAccount(
		suite.ctx,
		vestingtypes.NewContinuousVestingAccount(
			authtypes.NewBaseAccountWithAddress(addr2),
			sdk.Coins{sdk.NewInt64Coin(defaultClaimDenom, 0)},
			now.Unix(),
			now.Add(24*time.Hour).Unix(),
		),
	)
	suite.app.AccountKeeper.SetAccount(
		suite.ctx,
		vestingtypes.NewContinuousVestingAccount(
			authtypes.NewBaseAccountWithAddress(addr3),
			sdk.Coins{sdk.NewInt64Coin(defaultClaimDenom, 0)},
			now.Unix(),
			now.Add(24*time.Hour).Unix(),
		),
	)

	// Give free coins to accounts
	suite.app.AccountKeeper.SetModuleAccount(suite.ctx, authtypes.NewEmptyModuleAccount(minttypes.ModuleName, authtypes.Minter))
	err := suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(defaultClaimDenom, 60)})
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr1, sdk.Coins{sdk.NewInt64Coin(defaultClaimDenom, 10)})
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr2, sdk.Coins{sdk.NewInt64Coin(defaultClaimDenom, 20)})
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr3, sdk.Coins{sdk.NewInt64Coin(defaultClaimDenom, 30)})
	suite.Require().NoError(err)

	// initialize claim records
	claimRecords := []types.ClaimRecord{
		{
			Address: addr1.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(defaultClaimDenom, 100),
				sdk.NewInt64Coin(defaultClaimDenom, 200),
			},
			ActionCompleted: []bool{false, false},
		},
		{
			Address: addr2.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(defaultClaimDenom, 300),
				sdk.NewInt64Coin(defaultClaimDenom, 400),
			},
			ActionCompleted: []bool{false, false},
		},
		{
			Address: addr3.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(defaultClaimDenom, 0),
				sdk.NewInt64Coin(defaultClaimDenom, 500),
			},
			ActionCompleted: []bool{false, false},
		},
	}
	err = suite.app.AirdropKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)
	supply := suite.app.BankKeeper.GetSupply(suite.ctx, defaultClaimDenom)
	suite.Equal(sdk.NewInt64Coin(defaultClaimDenom, 60), supply)
	moduleBalance := suite.app.AirdropKeeper.GetAirdropAccountBalance(suite.ctx)
	suite.Equal(sdk.NewInt64Coin(defaultClaimDenom, 0), moduleBalance)

	// Run migration
	err = v104.MigrateModuleBalance(suite.ctx, *suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.AirdropKeeper.GetClaimRecords(suite.ctx))
	suite.Require().NoError(err)

	//// Check state post migration

	// Module balance and supply should have been updated
	err = suite.app.AirdropKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)
	moduleBalance = suite.app.AirdropKeeper.GetAirdropAccountBalance(suite.ctx)
	suite.Equal(sdk.NewInt64Coin(defaultClaimDenom, 1500), moduleBalance)
	supply = suite.app.BankKeeper.GetSupply(suite.ctx, defaultClaimDenom)
	suite.Equal(sdk.NewInt64Coin(defaultClaimDenom, 60+1500), supply)
}

// TestKeeperSuite Main entry point for the testing suite.
func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(StoreMigrationTestSuite))
}
