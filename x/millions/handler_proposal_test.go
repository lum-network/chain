package millions_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/x/millions"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

type HandlerTestSuite struct {
	suite.Suite

	app   *app.App
	ctx   sdk.Context
	addrs []sdk.AccAddress
	pool  millionstypes.Pool

	handler govtypesv1beta1.Handler
}

func (suite *HandlerTestSuite) SetupTest() {
	app := app.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "lum-network-devnet"})

	suite.app = app
	suite.ctx = ctx
	suite.handler = millions.NewMillionsProposalHandler(*app.MillionsKeeper)
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 2, sdk.NewInt(300000000), "ulum")

	// Initialize dummy pool
	poolID := app.MillionsKeeper.GetNextPoolIDAndIncrement(ctx)
	drawDelta1 := 1 * time.Hour
	app.MillionsKeeper.AddPool(ctx, &millionstypes.Pool{
		Denom:               "ulum",
		NativeDenom:         "ulum",
		ChainId:             "lum-network-devnet",
		Bech32PrefixValAddr: "lumvaloper",
		Bech32PrefixAccAddr: "lum",
		Validators: []millionstypes.PoolValidator{{
			OperatorAddress: "lumvaloper16rlynj5wvzwts5lqep0je5q4m3eaepn5cqj38s",
			IsEnabled:       true,
			BondedAmount:    sdk.NewInt(10),
		}},
		MinDepositAmount: sdk.NewInt(1000000),
		State:            millionstypes.PoolState_Ready,
		PoolId:           poolID,
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
	suite.pool, _ = app.MillionsKeeper.GetPool(suite.ctx, poolID)
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

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
