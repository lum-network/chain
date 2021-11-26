package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/lum-network/chain/simapp"
	"github.com/lum-network/chain/x/airdrop/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var now = time.Now().UTC()
var defaultClaimDenom = sdk.DefaultBondDenom
var defaultClaimBalance = int64(1_000_000)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	queryClient types.QueryClient
	app         *simapp.SimApp
}

// SetupTest Create our testing app, and make sure everything is correctly usable
func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.AirdropKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient

	airdropStartTime := time.Now()
	suite.app.AirdropKeeper.CreateModuleAccount(suite.ctx, sdk.NewCoin(defaultClaimDenom, sdk.NewInt(defaultClaimBalance)))
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

func (suite *KeeperTestSuite) TestHookOfUnclaimableAccount() {
	suite.SetupTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	claim, err := suite.app.AirdropKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.NoError(err)
	suite.Equal(types.ClaimRecord{}, claim)

	suite.app.AirdropKeeper.AfterProposalVote(suite.ctx, 12, addr1)

	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Equal(sdk.Coins{}, balances)
}

func (suite *KeeperTestSuite) TestHookBeforeAirdropStart() {
	suite.SetupTest()

	airdropStartTime := time.Now().Add(time.Hour)

	params := types.Params{
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: time.Hour,
		DurationOfDecay:    time.Hour * 4,
		ClaimDenom:         defaultClaimDenom,
	}
	err := suite.app.AirdropKeeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecords := []types.ClaimRecord{
		{
			Address: addr1.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(params.ClaimDenom, 1000),
				sdk.NewInt64Coin(params.ClaimDenom, 1000),
			},
			ActionCompleted: []bool{false, false},
		},
	}
	suite.app.AccountKeeper.SetAccount(
		suite.ctx,
		vestingtypes.NewContinuousVestingAccount(authtypes.NewBaseAccountWithAddress(addr1), sdk.Coins{}, now.Unix(), now.Add(24*time.Hour).Unix()),
	)

	err = suite.app.AirdropKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	cfc, cvc, err := suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	suite.True(cfc.IsZero())
	suite.True(cvc.IsZero())

	cfc, cvc, err = suite.app.AirdropKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionVote)
	suite.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	suite.True(cfc.IsZero())
	suite.True(cvc.IsZero())

	suite.app.AirdropKeeper.AfterProposalVote(suite.ctx, 12, addr1)
	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	// Now, it is before starting air drop, so claim module should not send the balances to the user after swap.
	suite.True(balances.Empty())

	// After the airdrop
	cfc, cvc, err = suite.app.AirdropKeeper.GetClaimableAmountForAction(suite.ctx.WithBlockTime(airdropStartTime), addr1, types.ActionVote)
	suite.NoError(err)
	suite.Equal(claimRecords[0].InitialClaimableAmount[0].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted))), cfc.Amount)
	suite.Equal(claimRecords[0].InitialClaimableAmount[1].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted))), cvc.Amount)

	suite.app.AirdropKeeper.AfterProposalVote(suite.ctx.WithBlockTime(airdropStartTime), 12, addr1)
	balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Equal(
		claimRecords[0].InitialClaimableAmount[0].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted))).Add(claimRecords[0].InitialClaimableAmount[1].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted)))),
		balances.AmountOf(params.ClaimDenom),
	)
	acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr1)
	suite.NotNil(acc)
	vestingAcc, ok := acc.(*vestingtypes.ContinuousVestingAccount)
	suite.True(ok)
	suite.Equal(claimRecords[0].InitialClaimableAmount[0].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted))), vestingAcc.OriginalVesting.AmountOf(defaultClaimDenom))
}

func (suite *KeeperTestSuite) TestHookAfterAirdropEnd() {
	suite.SetupTest()

	// airdrop recipient address
	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecords := []types.ClaimRecord{
		{
			Address: addr1.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(defaultClaimDenom, 1000),
				sdk.NewInt64Coin(defaultClaimDenom, 1000),
			},
			ActionCompleted: []bool{false, false},
		},
	}
	suite.app.AccountKeeper.SetAccount(
		suite.ctx,
		vestingtypes.NewContinuousVestingAccount(authtypes.NewBaseAccountWithAddress(addr1), sdk.Coins{}, now.Unix(), now.Add(24*time.Hour).Unix()),
	)
	err := suite.app.AirdropKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	params, err := suite.app.AirdropKeeper.GetParams(suite.ctx)
	suite.Require().NoError(err)
	suite.ctx = suite.ctx.WithBlockTime(params.AirdropStartTime.Add(params.DurationUntilDecay).Add(params.DurationOfDecay))

	suite.app.AirdropKeeper.EndAirdrop(suite.ctx)

	suite.Require().NotPanics(func() {
		suite.app.AirdropKeeper.AfterProposalVote(suite.ctx, 12, addr1)
	})
	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.True(balances.Empty())
	acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr1)
	suite.NotNil(acc)
	vestingAcc, ok := acc.(*vestingtypes.ContinuousVestingAccount)
	suite.True(ok)
	suite.True(vestingAcc.OriginalVesting.Empty())
}

func (suite *KeeperTestSuite) TestDuplicatedAction() {
	suite.SetupTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecords := []types.ClaimRecord{
		{
			Address: addr1.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(defaultClaimDenom, 750),
				sdk.NewInt64Coin(defaultClaimDenom, 800),
			},
			ActionCompleted: []bool{false, false},
		},
	}
	suite.app.AccountKeeper.SetAccount(
		suite.ctx,
		vestingtypes.NewContinuousVestingAccount(authtypes.NewBaseAccountWithAddress(addr1), sdk.Coins{}, now.Unix(), now.Add(24*time.Hour).Unix()),
	)

	err := suite.app.AirdropKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	cfc, cvc, err := suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().Equal(claimRecords[0].InitialClaimableAmount[0], cfc)
	suite.Require().Equal(claimRecords[0].InitialClaimableAmount[1], cvc)

	suite.app.AirdropKeeper.AfterProposalVote(suite.ctx, 12, addr1)
	claim, err := suite.app.AirdropKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.NoError(err)
	suite.True(claim.ActionCompleted[types.ActionVote])
	claimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Equal(claimedCoins.AmountOf(defaultClaimDenom), claimRecords[0].InitialClaimableAmount[0].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted))).Add(claimRecords[0].InitialClaimableAmount[1].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted)))))
	acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr1)
	suite.NotNil(acc)
	vestingAcc, ok := acc.(*vestingtypes.ContinuousVestingAccount)
	suite.True(ok)
	suite.Equal(claimRecords[0].InitialClaimableAmount[1].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted))), vestingAcc.OriginalVesting.AmountOf(defaultClaimDenom))

	suite.app.AirdropKeeper.AfterProposalVote(suite.ctx, 12, addr1)
	claim, err = suite.app.AirdropKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.NoError(err)
	suite.True(claim.ActionCompleted[types.ActionVote])
	claimedCoins = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Equal(claimedCoins.AmountOf(defaultClaimDenom), claimRecords[0].InitialClaimableAmount[0].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted))).Add(claimRecords[0].InitialClaimableAmount[1].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted)))))
	acc = suite.app.AccountKeeper.GetAccount(suite.ctx, addr1)
	suite.NotNil(acc)
	vestingAcc, ok = acc.(*vestingtypes.ContinuousVestingAccount)
	suite.True(ok)
	suite.Equal(claimRecords[0].InitialClaimableAmount[1].Amount.QuoRaw(int64(len(claimRecords[0].ActionCompleted))), vestingAcc.OriginalVesting.AmountOf(defaultClaimDenom))
}

func (suite *KeeperTestSuite) TestDelegationAutoWithdrawAndDelegateMore() {
	suite.SetupTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	pub2 := secp256k1.GenPrivKey().PubKey()
	addrs := []sdk.AccAddress{sdk.AccAddress(pub1.Address()), sdk.AccAddress(pub2.Address())}
	claimRecords := []types.ClaimRecord{
		{
			Address: addrs[0].String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(defaultClaimDenom, 750),
				sdk.NewInt64Coin(defaultClaimDenom, 800),
			},
			ActionCompleted: []bool{false, false},
		},
		{
			Address: addrs[1].String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(defaultClaimDenom, 900),
				sdk.NewInt64Coin(defaultClaimDenom, 1150),
			},
			ActionCompleted: []bool{false, false},
		},
	}

	// initialize accts
	for i := 0; i < len(addrs); i++ {
		suite.app.AccountKeeper.SetAccount(
			suite.ctx,
			vestingtypes.NewContinuousVestingAccount(authtypes.NewBaseAccountWithAddress(addrs[i]), sdk.Coins{}, now.Unix(), now.Add(24*time.Hour).Unix()),
		)
	}
	// initialize claim records
	err := suite.app.AirdropKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	// test claim records set
	for i := 0; i < len(addrs); i++ {
		cfc, cvc, err := suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addrs[i])
		suite.Require().NoError(err)
		suite.Require().Equal(cfc, claimRecords[i].InitialClaimableAmount[0])
		suite.Require().Equal(cvc, claimRecords[i].InitialClaimableAmount[1])
	}

	// set addr[0] as a validator
	validator, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[0]), pub1, stakingtypes.Description{})
	suite.Require().NoError(err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())

	validator, _ = validator.AddTokensFromDel(sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction))
	delAmount := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addrs[1], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, delAmount)))
	suite.Require().NoError(err)
	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addrs[1], delAmount, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	// delegation should automatically call claim and withdraw balance
	actualClaimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrs[1])
	actualClaimedCoin := actualClaimedCoins.AmountOf(defaultClaimDenom)
	expectedClaimedCoin := claimRecords[1].InitialClaimableAmount[0].Amount.Quo(sdk.NewInt(int64(len(claimRecords[1].ActionCompleted)))).Add(claimRecords[1].InitialClaimableAmount[1].Amount.Quo(sdk.NewInt(int64(len(claimRecords[1].ActionCompleted)))))
	suite.Require().Equal(expectedClaimedCoin.String(), actualClaimedCoin.String())

	// only free coins should be transferable
	err = suite.app.BankKeeper.SendCoins(suite.ctx, addrs[1], addrs[0], actualClaimedCoins)
	suite.Error(err)
	err = suite.app.BankKeeper.SendCoins(suite.ctx, addrs[1], addrs[0], sdk.NewCoins(sdk.NewCoin(defaultClaimDenom, claimRecords[1].InitialClaimableAmount[0].Amount.Quo(sdk.NewInt(int64(len(claimRecords[1].ActionCompleted)))))))
	suite.NoError(err)

	// vested claim coins should be delegatable
	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addrs[1], claimRecords[1].InitialClaimableAmount[1].Amount.Quo(sdk.NewInt(int64(len(claimRecords[1].ActionCompleted)))), stakingtypes.Unbonded, validator, true)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) TestAirdropFlow() {
	suite.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// initialize accts
	suite.app.AccountKeeper.SetAccount(
		suite.ctx,
		vestingtypes.NewContinuousVestingAccount(authtypes.NewBaseAccountWithAddress(addr1), sdk.Coins{}, now.Unix(), now.Add(24*time.Hour).Unix()),
	)
	suite.app.AccountKeeper.SetAccount(
		suite.ctx,
		vestingtypes.NewContinuousVestingAccount(authtypes.NewBaseAccountWithAddress(addr2), sdk.Coins{}, now.Unix(), now.Add(24*time.Hour).Unix()),
	)
	suite.app.AccountKeeper.SetAccount(
		suite.ctx,
		vestingtypes.NewContinuousVestingAccount(authtypes.NewBaseAccountWithAddress(addr3), sdk.Coins{}, now.Unix(), now.Add(24*time.Hour).Unix()),
	)

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
	}

	err := suite.app.AirdropKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	cfc, cvc, err := suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().Equal(cfc, claimRecords[0].InitialClaimableAmount[0], cfc.String())
	suite.Require().Equal(cvc, claimRecords[0].InitialClaimableAmount[1], cfc.String())

	cfc, cvc, err = suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr2)
	suite.Require().NoError(err)
	suite.Require().Equal(cfc, claimRecords[1].InitialClaimableAmount[0], cfc.String())
	suite.Require().Equal(cvc, claimRecords[1].InitialClaimableAmount[1], cfc.String())

	cfc, cvc, err = suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr3)
	suite.Require().NoError(err)
	suite.Require().True(cfc.IsZero())
	suite.Require().True(cvc.IsZero())

	// get rewards amount per action
	cfc, cvc, err = suite.app.AirdropKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionDelegateStake)
	suite.Require().NoError(err)
	suite.Require().Equal(cfc.Amount, claimRecords[0].InitialClaimableAmount[0].Amount.Quo(sdk.NewInt(int64(len(claimRecords[0].ActionCompleted)))), cfc.String())
	suite.Require().Equal(cvc.Amount, claimRecords[0].InitialClaimableAmount[1].Amount.Quo(sdk.NewInt(int64(len(claimRecords[0].ActionCompleted)))), cfc.String())

	// get completed activities
	claimRecord, err := suite.app.AirdropKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.Require().NoError(err)
	for i := range types.Action_name {
		suite.Require().False(claimRecord.ActionCompleted[i])
	}

	// do one action per account
	suite.app.AirdropKeeper.AfterProposalVote(suite.ctx, 12, addr1)
	suite.app.AirdropKeeper.AfterDelegationModified(suite.ctx, addr2, sdk.ValAddress(addr1))
	suite.app.AirdropKeeper.AfterDelegationModified(suite.ctx, addr3, sdk.ValAddress(addr1))

	// check that half are completed
	claimRecord, err = suite.app.AirdropKeeper.GetClaimRecord(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().True(claimRecord.ActionCompleted[types.ActionVote])
	claimRecord, err = suite.app.AirdropKeeper.GetClaimRecord(suite.ctx, addr2)
	suite.Require().NoError(err)
	suite.Require().True(claimRecord.ActionCompleted[types.ActionDelegateStake])
	claimRecord, err = suite.app.AirdropKeeper.GetClaimRecord(suite.ctx, addr3)
	suite.Require().NoError(err)
	suite.Require().Empty(claimRecord.ActionCompleted)

	// get balance after 1 action done
	coins1 := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(defaultClaimDenom, 150)).String())
	coins2 := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr2)
	suite.Require().Equal(coins2.String(), sdk.NewCoins(sdk.NewInt64Coin(defaultClaimDenom, 350)).String())
	coins3 := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr3)
	suite.Require().Equal(coins3.String(), sdk.Coins{}.String())

	// get vesting after 1 action done
	acc1 := suite.app.AccountKeeper.GetAccount(suite.ctx, addr1)
	suite.NotNil(acc1)
	vestingAcc1, ok := acc1.(*vestingtypes.ContinuousVestingAccount)
	suite.True(ok)
	suite.Equal(sdk.NewCoins(sdk.NewInt64Coin(defaultClaimDenom, 100)).String(), vestingAcc1.OriginalVesting.String())
	acc2 := suite.app.AccountKeeper.GetAccount(suite.ctx, addr2)
	suite.NotNil(acc2)
	vestingAcc2, ok := acc2.(*vestingtypes.ContinuousVestingAccount)
	suite.True(ok)
	suite.Equal(sdk.NewCoins(sdk.NewInt64Coin(defaultClaimDenom, 200)).String(), vestingAcc2.OriginalVesting.String())
	acc3 := suite.app.AccountKeeper.GetAccount(suite.ctx, addr3)
	suite.NotNil(acc3)
	vestingAcc3, ok := acc3.(*vestingtypes.ContinuousVestingAccount)
	suite.True(ok)
	suite.Equal(sdk.Coins{}.String(), vestingAcc3.OriginalVesting.String())

	// check that claimable for completed activity is 0
	cfc, cvc, err = suite.app.AirdropKeeper.GetClaimableAmountForAction(suite.ctx, addr1, types.ActionVote)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64Coin(defaultClaimDenom, 0).String(), cfc.String())
	suite.Require().Equal(sdk.NewInt64Coin(defaultClaimDenom, 0).String(), cvc.String())
	cfc, cvc, err = suite.app.AirdropKeeper.GetClaimableAmountForAction(suite.ctx, addr2, types.ActionDelegateStake)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64Coin(defaultClaimDenom, 0).String(), cfc.String())
	suite.Require().Equal(sdk.NewInt64Coin(defaultClaimDenom, 0).String(), cvc.String())
	cfc, cvc, err = suite.app.AirdropKeeper.GetClaimableAmountForAction(suite.ctx, addr3, types.ActionDelegateStake)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64Coin(defaultClaimDenom, 0).String(), cfc.String())
	suite.Require().Equal(sdk.NewInt64Coin(defaultClaimDenom, 0).String(), cvc.String())

	// do rest of actions
	suite.app.AirdropKeeper.AfterDelegationModified(suite.ctx, addr1, sdk.ValAddress(addr1))
	suite.app.AirdropKeeper.AfterProposalVote(suite.ctx, 12, addr2)
	suite.app.AirdropKeeper.AfterProposalVote(suite.ctx, 12, addr3)

	// get balance after rest actions done
	coins1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(defaultClaimDenom, 300)).String())
	coins2 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr2)
	suite.Require().Equal(coins2.String(), sdk.NewCoins(sdk.NewInt64Coin(defaultClaimDenom, 700)).String())
	coins3 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr3)
	suite.Require().Equal(coins3.String(), sdk.Coins{}.String())

	// get vesting after 1 action done
	acc1 = suite.app.AccountKeeper.GetAccount(suite.ctx, addr1)
	suite.NotNil(acc1)
	vestingAcc1, ok = acc1.(*vestingtypes.ContinuousVestingAccount)
	suite.True(ok)
	suite.Equal(sdk.NewCoins(sdk.NewInt64Coin(defaultClaimDenom, 200)).String(), vestingAcc1.OriginalVesting.String())
	acc2 = suite.app.AccountKeeper.GetAccount(suite.ctx, addr2)
	suite.NotNil(acc2)
	vestingAcc2, ok = acc2.(*vestingtypes.ContinuousVestingAccount)
	suite.True(ok)
	suite.Equal(sdk.NewCoins(sdk.NewInt64Coin(defaultClaimDenom, 400)).String(), vestingAcc2.OriginalVesting.String())
	acc3 = suite.app.AccountKeeper.GetAccount(suite.ctx, addr3)
	suite.NotNil(acc3)
	vestingAcc3, ok = acc3.(*vestingtypes.ContinuousVestingAccount)
	suite.True(ok)
	suite.Equal(sdk.Coins{}.String(), vestingAcc3.OriginalVesting.String())

	// get claimable after withdrawing all
	cfc, cvc, err = suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().True(cfc.IsZero())
	suite.Require().True(cvc.IsZero())
	cfc, cvc, err = suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr2)
	suite.Require().NoError(err)
	suite.Require().True(cfc.IsZero())
	suite.Require().True(cvc.IsZero())
	cfc, cvc, err = suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr3)
	suite.Require().NoError(err)
	suite.Require().True(cfc.IsZero())
	suite.Require().True(cvc.IsZero())

	err = suite.app.AirdropKeeper.EndAirdrop(suite.ctx)
	suite.Require().NoError(err)

	moduleAccAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	coins := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAccAddr, defaultClaimDenom)
	suite.Require().Equal(coins, sdk.NewInt64Coin(defaultClaimDenom, 0))

	cfc, cvc, err = suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().True(cfc.IsZero())
	suite.Require().True(cvc.IsZero())
	cfc, cvc, err = suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr2)
	suite.Require().NoError(err)
	suite.Require().True(cfc.IsZero())
	suite.Require().True(cvc.IsZero())
	cfc, cvc, err = suite.app.AirdropKeeper.GetUserTotalClaimable(suite.ctx, addr3)
	suite.Require().NoError(err)
	suite.Require().True(cfc.IsZero())
	suite.Require().True(cvc.IsZero())
}

func (suite *KeeperTestSuite) TestModuleBalance() {
	suite.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// initialize accts
	suite.app.AccountKeeper.SetAccount(
		suite.ctx,
		vestingtypes.NewContinuousVestingAccount(authtypes.NewBaseAccountWithAddress(addr1), sdk.Coins{}, now.Unix(), now.Add(24*time.Hour).Unix()),
	)

	// initialize claim records
	claimRecords := []types.ClaimRecord{
		{
			Address: addr1.String(),
			InitialClaimableAmount: sdk.Coins{
				sdk.NewInt64Coin(defaultClaimDenom, 400_000),
				sdk.NewInt64Coin(defaultClaimDenom, 500_000),
			},
			ActionCompleted: []bool{false, false},
		},
	}
	err := suite.app.AirdropKeeper.SetClaimRecords(suite.ctx, claimRecords)
	suite.Require().NoError(err)

	// Initial balance should be in the supply
	supply := suite.app.BankKeeper.GetSupply(suite.ctx, defaultClaimDenom)
	suite.Equal(sdk.NewInt64Coin(defaultClaimDenom, defaultClaimBalance), supply)
	moduleBalance := suite.app.AirdropKeeper.GetAirdropAccountBalance(suite.ctx)
	suite.Equal(sdk.NewInt64Coin(defaultClaimDenom, defaultClaimBalance), moduleBalance)

	// Claim should debit the module account while leaving the supply intact
	suite.app.AirdropKeeper.AfterProposalVote(suite.ctx, 12, addr1)
	supply = suite.app.BankKeeper.GetSupply(suite.ctx, defaultClaimDenom)
	suite.Equal(sdk.NewInt64Coin(defaultClaimDenom, defaultClaimBalance), supply)
	moduleBalance = suite.app.AirdropKeeper.GetAirdropAccountBalance(suite.ctx)
	suite.Equal(sdk.NewInt64Coin(defaultClaimDenom, defaultClaimBalance-450_000), moduleBalance)

	suite.app.AirdropKeeper.AfterDelegationModified(suite.ctx, addr1, sdk.ValAddress(addr1))
	supply = suite.app.BankKeeper.GetSupply(suite.ctx, defaultClaimDenom)
	suite.Equal(sdk.NewInt64Coin(defaultClaimDenom, defaultClaimBalance), supply)
	moduleBalance = suite.app.AirdropKeeper.GetAirdropAccountBalance(suite.ctx)
	suite.Equal(sdk.NewInt64Coin(defaultClaimDenom, defaultClaimBalance-900_000), moduleBalance)

	// Ending the airdrop should send the unclaimed amounts into the community pool
	communityBalance := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Equal(sdk.NewDec(0), communityBalance.AmountOf(defaultClaimDenom))

	err = suite.app.AirdropKeeper.EndAirdrop(suite.ctx)
	suite.Require().NoError(err)
	communityBalance = suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Equal(sdk.NewDec(100_000), communityBalance.AmountOf(defaultClaimDenom))
}

// TestKeeperSuite Main entry point for the testing suite
func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
