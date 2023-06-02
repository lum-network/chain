package millions_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/x/millions"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type HandlerTestSuite struct {
	suite.Suite

	app         *app.App
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	valAddrs    []sdk.ValAddress
	moduleAddrs []sdk.AccAddress
	pool        millionstypes.Pool

	handler govtypesv1beta1.Handler
}

func (suite *HandlerTestSuite) SetupTest() {
	app := app.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "lum-network-devnet"})
	msgServer := millionskeeper.NewMsgServerImpl(*app.MillionsKeeper)
	goCtx := sdk.WrapSDKContext(ctx)

	suite.app = app
	suite.ctx = ctx
	suite.handler = millions.NewMillionsProposalHandler(*app.MillionsKeeper)
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 2, sdk.NewInt(300000000), "ulum")

	// Initialize module accounts for IcaDeposits
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 3, sdk.NewInt(1_000_0000_000), app.StakingKeeper.BondDenom(ctx))
	for i := 0; i < 3; i++ {
		poolAddress := millionstypes.NewPoolAddress(uint64(i+1), "unused-in-test")
		apptesting.AddTestModuleAccount(app, ctx, poolAddress)
		suite.moduleAddrs = append(suite.moduleAddrs, poolAddress)
	}

	pub1 := ed25519.GenPrivKey().PubKey()
	pub2 := ed25519.GenPrivKey().PubKey()
	addrs := []sdk.AccAddress{sdk.AccAddress(pub1.Address()), sdk.AccAddress(pub2.Address())}

	validator2, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[0]), pub1, stakingtypes.Description{})
	suite.Require().NoError(err)
	validator2 = stakingkeeper.TestingUpdateValidator(*suite.app.StakingKeeper, suite.ctx, validator2, true)
	err = app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator2.GetOperator())
	suite.Require().NoError(err)

	validator3, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[1]), pub2, stakingtypes.Description{})
	suite.Require().NoError(err)
	validator2 = stakingkeeper.TestingUpdateValidator(*suite.app.StakingKeeper, suite.ctx, validator3, true)
	err = app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator3.GetOperator())
	suite.Require().NoError(err)

	vals := app.StakingKeeper.GetAllValidators(ctx)
	suite.Require().Len(vals, 3)
	suite.valAddrs = []sdk.ValAddress{}
	for _, v := range vals {
		addr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
		suite.Require().NoError(err)
		suite.valAddrs = append(suite.valAddrs, addr)
	}

	valAddrs := []string{
		suite.valAddrs[0].String(),
		suite.valAddrs[1].String(),
		suite.valAddrs[2].String(),
	}

	valSet := map[string]*millionstypes.PoolValidator{
		valAddrs[0]: {
			OperatorAddress: valAddrs[0],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
			Redelegate: &millionstypes.Redelegate{
				IsGovPropRedelegated: false,
				ErrorState:           millionstypes.RedelegateState_Unspecified,
			},
		},
		valAddrs[1]: {
			OperatorAddress: valAddrs[1],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
			Redelegate: &millionstypes.Redelegate{
				IsGovPropRedelegated: false,
				ErrorState:           millionstypes.RedelegateState_Unspecified,
			},
		},
		valAddrs[2]: {
			OperatorAddress: valAddrs[2],
			BondedAmount:    sdk.NewInt(0),
			IsEnabled:       true,
			Redelegate: &millionstypes.Redelegate{
				IsGovPropRedelegated: false,
				ErrorState:           millionstypes.RedelegateState_Unspecified,
			},
		},
	}

	// Initialize dummy pool
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta1 := 1 * time.Hour
	app.MillionsKeeper.AddPool(ctx, &millionstypes.Pool{
		Denom:               "ulum",
		NativeDenom:         "ulum",
		ChainId:             "lum-network-devnet",
		Bech32PrefixValAddr: "lumvaloper",
		Bech32PrefixAccAddr: "lum",
		IcaDepositAddress:   suite.moduleAddrs[0].String(),
		Validators:          valSet,
		MinDepositAmount:    sdk.NewInt(1000000),
		State:               millionstypes.PoolState_Ready,
		PoolId:              poolID,
		PrizeStrategy: millionstypes.PrizeStrategy{
			PrizeBatches: []millionstypes.PrizeBatch{
				{PoolPercent: 100, Quantity: 1, DrawProbability: sdk.NewDecWithPrec(int64(0.00*1_000_000), 6)},
			},
		},
		DrawSchedule: millionstypes.DrawSchedule{
			InitialDrawAt: ctx.BlockTime().Add(drawDelta1),
			DrawDelta:     drawDelta1,
		},
		AvailablePrizePool: sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), math.NewInt(1000)),
	})
	suite.pool, _ = app.MillionsKeeper.GetPool(ctx, poolID)
	err = app.BankKeeper.SendCoins(ctx, suite.addrs[0], sdk.MustAccAddressFromBech32(suite.pool.IcaDepositAddress), sdk.Coins{sdk.NewCoin("ulum", sdk.NewInt(8_000_000))})
	suite.Require().NoError(err)

	validator2, _ = validator2.AddTokensFromDel(sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction))
	apptesting.InitAccountWithCoins(suite.app, suite.ctx, sdk.MustAccAddressFromBech32(suite.pool.IcaDepositAddress), sdk.NewCoins(sdk.NewCoin("ulum", sdk.NewInt(1_000_000))))
	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, sdk.MustAccAddressFromBech32(suite.pool.IcaDepositAddress), sdk.NewInt(1_000_000), stakingtypes.Unbonded, validator2, true)
	validator2.UpdateStatus(stakingtypes.Bonded)
	suite.Require().NoError(err)

	validator3, _ = validator3.AddTokensFromDel(sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction))
	apptesting.InitAccountWithCoins(suite.app, suite.ctx, sdk.MustAccAddressFromBech32(suite.pool.IcaDepositAddress), sdk.NewCoins(sdk.NewCoin("ulum", sdk.NewInt(1_000_000))))
	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, sdk.MustAccAddressFromBech32(suite.pool.IcaDepositAddress), sdk.NewInt(1_000_000), stakingtypes.Unbonded, validator3, true)
	validator3.UpdateStatus(stakingtypes.Bonded)
	suite.Require().NoError(err)

	// Initialize a deposit for redelegation to validators
	_, err = msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		DepositorAddress: suite.addrs[0].String(),
		PoolId:           suite.pool.PoolId,
		Amount:           sdk.NewCoin("ulum", sdk.NewInt(6_000_000)),
	})
	suite.Require().NoError(err)
}

func (suite *HandlerTestSuite) TestProposal_RegisterPool() {
	validPrizeStrategy := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 90, DrawProbability: sdk.NewDec(1), Quantity: 10},
			{PoolPercent: 10, DrawProbability: sdk.NewDec(1), Quantity: 10},
		},
	}

	invalidPrizeStrategy := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 90, DrawProbability: sdk.NewDec(1), Quantity: 10},
		},
	}

	emptyPrizeStrategy := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{},
	}

	validDrawSchedule := millionstypes.DrawSchedule{
		DrawDelta:     1 * time.Hour,
		InitialDrawAt: time.Now().Add(1 * time.Hour),
	}

	invalidDrawSchedule := millionstypes.DrawSchedule{
		DrawDelta:     millionstypes.MinAcceptableDrawDelta - 1,
		InitialDrawAt: time.Time{},
	}

	validValidatorSet := []string{"lumvaloper16rlynj5wvzwts5lqep0je5q4m3eaepn5cqj38s"}
	invalidValidatorSet := []string{"lumvaloper16rlynj5wvzw"}
	var emptyValidatorSet []string

	cases := []struct {
		name            string
		proposal        govtypesv1beta1.Content
		expectPreError  bool
		expectPostError bool
	}{
		{
			"Title cannot be empty",
			millionstypes.NewRegisterPoolProposal("", "Test", "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule),
			true,
			false,
		},
		{
			"Description cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "", "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule),
			true,
			false,
		},
		{
			"Chain ID cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "Test", "", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule),
			true,
			true,
		},
		{
			"Bech 32 acc prefix cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "Test", "lum-network-devnet", "ulum", "ulum", "connection-0", "", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule),
			true,
			true,
		},
		{
			"Bech 32 val prefix cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "Test", "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule),
			true,
			true,
		},
		{
			"Validators list cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "Test", "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", emptyValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule),
			true,
			true,
		},
		{
			"Validators list cannot be invalid",
			millionstypes.NewRegisterPoolProposal("Test", "Test", "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", invalidValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule),
			false,
			true,
		},
		{
			"Min deposit amount cannot be less than min acceptable amount",
			millionstypes.NewRegisterPoolProposal("Test", "Test", "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(millionstypes.MinAcceptableDepositAmount-1), validPrizeStrategy, validDrawSchedule),
			true,
			true,
		},
		{
			"Prize strategy cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "Test", "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), emptyPrizeStrategy, validDrawSchedule),
			true,
			true,
		},
		{
			"Prize strategy cannot be invalid",
			millionstypes.NewRegisterPoolProposal("Test", "Test", "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), invalidPrizeStrategy, validDrawSchedule),
			false,
			true,
		},
		{
			"Draw Schedule cannot be invalid",
			millionstypes.NewRegisterPoolProposal("Test", "Test", "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, invalidDrawSchedule),
			true,
			true,
		},
		{
			"Fine should be fine",
			millionstypes.NewRegisterPoolProposal("Test", "Test", "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule),
			false,
			false,
		},
	}

	for _, tc := range cases {
		tc := tc
		suite.Run(tc.name, func() {
			preError := tc.proposal.ValidateBasic()
			if tc.expectPreError {
				suite.Require().Error(preError)
			} else {
				suite.Require().NoError(preError)
			}
			err := suite.handler(suite.ctx, tc.proposal)
			if tc.expectPostError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestProposal_UpdatePool() {
	validPrizeStrategy := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 90, DrawProbability: sdk.NewDec(1), Quantity: 10},
			{PoolPercent: 10, DrawProbability: sdk.NewDec(1), Quantity: 10},
		},
	}

	invalidPrizeStrategy := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{
			{PoolPercent: 90, DrawProbability: sdk.NewDec(1), Quantity: 10},
		},
	}

	emptyPrizeStrategy := millionstypes.PrizeStrategy{
		PrizeBatches: []millionstypes.PrizeBatch{},
	}

	validDrawSchedule := millionstypes.DrawSchedule{
		DrawDelta:     1 * time.Hour,
		InitialDrawAt: time.Now().Add(1 * time.Hour),
	}

	invalidDrawSchedule := millionstypes.DrawSchedule{
		DrawDelta:     0,
		InitialDrawAt: time.Time{},
	}

	validValidatorSet := []string{"lumvaloper16rlynj5wvzwts5lqep0je5q4m3eaepn5cqj38s"}
	invalidValidatorSet := []string{"lumvaloper16rlynj5wvzw"}

	validMinDepositAmount := millionstypes.DefaultParams().MinDepositAmount
	invalidMinDepositAmount := sdk.NewInt(millionstypes.MinAcceptableDepositAmount - 1)

	cases := []struct {
		name            string
		proposal        govtypesv1beta1.Content
		expectPreError  bool
		expectPostError bool
	}{
		{
			"Title cannot be empty",
			millionstypes.NewUpdatePoolProposal("", "Test", suite.pool.GetPoolId(), validValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule),
			true,
			false,
		},
		{
			"Description cannot be empty",
			millionstypes.NewUpdatePoolProposal("Test", "", suite.pool.GetPoolId(), validValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule),
			true,
			false,
		},
		{
			"Validators list can be empty",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule),
			false,
			false,
		},
		{
			"Validators list cannot be invalid",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), invalidValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule),
			false,
			true,
		},
		{
			"Min deposit amount cannot be less than 1000000 (default params)",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), validValidatorSet, &invalidMinDepositAmount, &validPrizeStrategy, &validDrawSchedule),
			true,
			true,
		},
		{
			"Prize strategy cannot be empty",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), validValidatorSet, &validMinDepositAmount, &emptyPrizeStrategy, &validDrawSchedule),
			true,
			true,
		},
		{
			"Prize strategy cannot be invalid",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), validValidatorSet, &validMinDepositAmount, &invalidPrizeStrategy, &validDrawSchedule),
			false,
			true,
		},
		{
			"Draw Schedule cannot be invalid",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), validValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &invalidDrawSchedule),
			true,
			true,
		},
		{
			"Fine should be fine",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), validValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule),
			false,
			false,
		},
		{
			"Partial should be fine",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), validValidatorSet, nil, nil, nil),
			false,
			false,
		},
	}

	for _, tc := range cases {
		tc := tc
		suite.Run(tc.name, func() {
			preError := tc.proposal.ValidateBasic()
			if tc.expectPreError {
				suite.Require().Error(preError)
			} else {
				suite.Require().NoError(preError)
			}
			err := suite.handler(suite.ctx, tc.proposal)
			if tc.expectPostError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestProposal_UpdateParams() {
	validMinDepositAmount := sdk.NewInt(millionstypes.MinAcceptableDepositAmount)
	validFees := sdk.NewDecWithPrec(int64(0.5*1_000_000), 6)

	prizeDelta := 5 * time.Hour
	depositDelta := 2 * time.Hour
	minDrawDelta := 1 * time.Hour
	maxDrawDelta := 10 * time.Hour
	maxBatchQuantity := sdk.NewInt(100)
	maxStrategyBatches := sdk.NewInt(50)

	cases := []struct {
		name            string
		proposal        govtypesv1beta1.Content
		expectPreError  bool
		expectPostError bool
	}{
		{
			"Full update should be fine",
			millionstypes.NewUpdateParamsProposal("Test", "Test", &validMinDepositAmount, &validFees, &prizeDelta, &depositDelta, &minDrawDelta, &maxDrawDelta, &maxBatchQuantity, &maxStrategyBatches),
			false,
			false,
		},
		{
			"Partial update should be fine",
			millionstypes.NewUpdateParamsProposal("Test", "Test", nil, nil, nil, nil, nil, nil, nil, nil),
			false,
			false,
		},
	}

	for _, tc := range cases {
		tc := tc
		suite.Run(tc.name, func() {
			preError := tc.proposal.ValidateBasic()
			if tc.expectPreError {
				suite.Require().Error(preError)
			} else {
				suite.Require().NoError(preError)
			}
			err := suite.handler(suite.ctx, tc.proposal)
			if tc.expectPostError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestProposal_DisableValidator() {
	cases := []struct {
		name            string
		proposal        govtypesv1beta1.Content
		expectPreError  bool
		expectPostError bool
	}{
		{
			"Should fail with wrong operator address",
			millionstypes.NewDisableValidatorProposal("Test title", "test description", "", suite.pool.GetPoolId()),
			true,
			true,
		},
		{
			"Should fail with wrong poolID",
			millionstypes.NewDisableValidatorProposal("Test title", "test description", suite.valAddrs[0].String(), uint64(0)),
			true,
			true,
		},
		{
			"Should pass with sufficient delegation shares",
			millionstypes.NewDisableValidatorProposal("Test title", "test description", suite.valAddrs[2].String(), suite.pool.GetPoolId()),
			false,
			false,
		},
	}

	for _, tc := range cases {
		tc := tc
		suite.Run(tc.name, func() {
			preError := tc.proposal.ValidateBasic()
			if tc.expectPreError {
				suite.Require().Error(preError)
			} else {
				suite.Require().NoError(preError)
			}
			err := suite.handler(suite.ctx, tc.proposal)
			if tc.expectPostError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
