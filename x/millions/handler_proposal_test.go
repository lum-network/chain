package millions_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/x/millions"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

type HandlerTestSuite struct {
	suite.Suite

	app         *app.App
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	moduleAddrs []sdk.AccAddress
	pool        millionstypes.Pool

	handler govtypesv1beta1.Handler
}

func (suite *HandlerTestSuite) SetupTest() {
	app := app.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "lum-network-devnet"})

	suite.app = app
	suite.ctx = ctx
	suite.handler = millions.NewMillionsProposalHandler(*app.MillionsKeeper)
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 2, sdk.NewCoins(sdk.NewCoin("ulum", sdk.NewInt(300000000))))

	// Initialize module accounts for IcaDeposits
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 3, sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1_000_0000_000))))
	for i := 0; i < 3; i++ {
		poolAddress := millionstypes.NewPoolAddress(uint64(i+1), "unused-in-test")
		apptesting.AddTestModuleAccount(app, ctx, poolAddress)
		suite.moduleAddrs = append(suite.moduleAddrs, poolAddress)
	}

	// Generate validators
	privKeys := make([]cryptotypes.PubKey, 4)
	addrs := make([]sdk.AccAddress, 4)
	validators := make([]stakingtypes.Validator, 4)
	for i := 0; i < 4; i++ {
		privKey := ed25519.GenPrivKey().PubKey()
		privKeys[i] = privKey
		addrs[i] = sdk.AccAddress(privKey.Address())

		validator, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[i]), privKeys[i], stakingtypes.Description{})
		suite.Require().NoError(err)
		validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, false)
		err = app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, validator.GetOperator())
		suite.Require().NoError(err)

		validators[i] = validator
		validators = append(validators, validator)

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
		Validators: []millionstypes.PoolValidator{
			{
				OperatorAddress: validators[0].OperatorAddress,
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: validators[1].OperatorAddress,
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: validators[2].OperatorAddress,
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
			{
				OperatorAddress: validators[3].OperatorAddress,
				BondedAmount:    sdk.NewInt(0),
				IsEnabled:       true,
			},
		},
		IcaDepositAddress: suite.moduleAddrs[0].String(),
		MinDepositAmount:  sdk.NewInt(1000000),
		State:             millionstypes.PoolState_Ready,
		PoolId:            poolID,
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
	msgServer := millionskeeper.NewMsgServerImpl(*suite.app.MillionsKeeper)
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Initialize a deposit to trigger delegation to validators
	_, err := msgServer.Deposit(goCtx, &millionstypes.MsgDeposit{
		DepositorAddress: suite.addrs[0].String(),
		PoolId:           suite.pool.PoolId,
		Amount:           sdk.NewCoin("ulum", sdk.NewInt(10_000_000)),
	})
	suite.Require().NoError(err)

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

	validValidatorSet := []string{suite.pool.Validators[0].OperatorAddress, suite.pool.Validators[2].OperatorAddress}
	invalidValidatorSet := []string{"lumvaloper16rlynj5wvzw"}

	poolStatePaused := millionstypes.PoolState_Paused
	poolStateReady := millionstypes.PoolState_Ready
	poolStateUnspecified := millionstypes.PoolState_Unspecified

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
			millionstypes.NewUpdatePoolProposal("", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStateUnspecified),
			true,
			false,
		},
		{
			"Description cannot be empty",
			millionstypes.NewUpdatePoolProposal("Test", "", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStatePaused),
			true,
			false,
		},
		{
			"Validators list can be empty",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStateReady),
			false,
			false,
		},
		{
			"Validators list cannot be invalid",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), invalidValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStatePaused),
			false,
			true,
		},
		{
			"Min deposit amount cannot be less than 1000000 (default params)",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &invalidMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStateReady),
			true,
			true,
		},
		{
			"Prize strategy cannot be empty",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &emptyPrizeStrategy, &validDrawSchedule, poolStatePaused),
			true,
			true,
		},
		{
			"Prize strategy cannot be invalid",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &invalidPrizeStrategy, &validDrawSchedule, poolStateReady),
			false,
			true,
		},
		{
			"Draw Schedule cannot be invalid",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &validPrizeStrategy, &invalidDrawSchedule, poolStatePaused),
			true,
			true,
		},
		{
			"Partial should be fine",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, nil, nil, nil, poolStateUnspecified),
			false,
			false,
		},
		{
			"Fine should be fine",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), validValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStatePaused),
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

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
