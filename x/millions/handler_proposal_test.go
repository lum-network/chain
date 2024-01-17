package millions_test

import (
	"testing"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

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
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 2, sdk.NewInt(300000000), "ulum")

	// Initialize module accounts for IcaDeposits
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 3, sdk.NewInt(1_000_0000_000), app.StakingKeeper.BondDenom(ctx))
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
		PoolType: millionstypes.PoolType_Staking,
		FeeTakers: []millionstypes.FeeTaker{
			{Destination: authtypes.FeeCollectorName, Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		},
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
		IcaDepositAddress:   suite.moduleAddrs[0].String(),
		MinDepositAmount:    sdk.NewInt(1000000),
		UnbondingDuration:   time.Duration(millionstypes.DefaultUnbondingDuration),
		MaxUnbondingEntries: sdk.NewInt(millionstypes.DefaultMaxUnbondingEntries),
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

	UnbondingDuration := time.Duration(millionstypes.DefaultUnbondingDuration)
	maxUnbondingEntries := sdk.NewInt(millionstypes.DefaultMaxUnbondingEntries)
	validFees := []millionstypes.FeeTaker{
		{Destination: authtypes.FeeCollectorName, Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
	}

	invalidFees := []millionstypes.FeeTaker{
		{Destination: "", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
	}
	invalidExcessiveFees := []millionstypes.FeeTaker{
		{Destination: "test_1", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_2", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_3", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_4", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_5", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_6", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_7", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_8", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_9", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_10", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_11", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
	}

	cases := []struct {
		name            string
		proposal        govtypesv1beta1.Content
		expectPreError  bool
		expectPostError bool
	}{
		{
			"Title cannot be empty",
			millionstypes.NewRegisterPoolProposal("", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			true,
			false,
		},
		{
			"Description cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			true,
			false,
		},
		{
			"Pool Type cannot be unspecified",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Unspecified, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"Chain ID cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"Bech 32 acc prefix cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"Bech 32 val prefix cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"Validators list cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", emptyValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"Validators list cannot be invalid",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", invalidValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			false,
			true,
		},
		{
			"Min deposit amount cannot be less than min acceptable amount",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(millionstypes.MinAcceptableDepositAmount-1), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"Prize strategy cannot be empty",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), emptyPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"Prize strategy cannot be invalid",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), invalidPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			false,
			true,
		},
		{
			"Draw Schedule cannot be invalid",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, invalidDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"ZoneUnbonding duration cannot be below the min unbonding",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, time.Duration(6*24*time.Hour), maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"MaxUnbondingEntries cannot exceed 7 parallel unbondings",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, sdk.NewInt(8), validFees),
			true,
			true,
		},
		{
			"MaxUnbondingEntries cannot exceed cannot be negative",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, sdk.NewInt(-2), validFees),
			true,
			true,
		},
		{
			"FeesStakers cannot be invalid",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, invalidFees),
			true,
			true,
		},
		{
			"FeesStakers cannot exceed maximum allowed amount",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, invalidExcessiveFees),
			true,
			true,
		},
		{
			"Fine should be fine",
			millionstypes.NewRegisterPoolProposal("Test", "Test", millionstypes.PoolType_Staking, "lum-network-devnet", "ulum", "ulum", "connection-0", "lum", "lumvaloper", validValidatorSet, sdk.NewInt(1000000), validPrizeStrategy, validDrawSchedule, UnbondingDuration, maxUnbondingEntries, validFees),
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

	UnbondingDuration := time.Duration(millionstypes.DefaultUnbondingDuration)
	maxUnbondingEntries := sdk.NewInt(millionstypes.DefaultMaxUnbondingEntries)
	validFees := []millionstypes.FeeTaker{
		{Destination: authtypes.FeeCollectorName, Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
	}
	invalidFees := []millionstypes.FeeTaker{
		{Destination: "", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
	}
	invalidExcessiveFees := []millionstypes.FeeTaker{
		{Destination: "test_1", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_2", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_3", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_4", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_5", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_6", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_7", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_8", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_9", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_10", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
		{Destination: "test_11", Amount: sdk.NewDecWithPrec(millionstypes.DefaultFeeTakerAmount, 2), Type: millionstypes.FeeTakerType_LocalModuleAccount},
	}

	cases := []struct {
		name            string
		proposal        govtypesv1beta1.Content
		expectPreError  bool
		expectPostError bool
	}{
		{
			"Title cannot be empty",
			millionstypes.NewUpdatePoolProposal("", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStateUnspecified, &UnbondingDuration, &maxUnbondingEntries, validFees),
			true,
			false,
		},
		{
			"Description cannot be empty",
			millionstypes.NewUpdatePoolProposal("Test", "", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStatePaused, &UnbondingDuration, &maxUnbondingEntries, validFees),
			true,
			false,
		},
		{
			"Validators list can be empty",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStateReady, &UnbondingDuration, &maxUnbondingEntries, validFees),
			false,
			false,
		},
		{
			"Validators list cannot be invalid",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), invalidValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStatePaused, &UnbondingDuration, &maxUnbondingEntries, validFees),
			false,
			true,
		},
		{
			"Min deposit amount cannot be less than 1000000 (default params)",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &invalidMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStateReady, &UnbondingDuration, &maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"Prize strategy cannot be empty",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &emptyPrizeStrategy, &validDrawSchedule, poolStatePaused, &UnbondingDuration, &maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"Prize strategy cannot be invalid",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &invalidPrizeStrategy, &validDrawSchedule, poolStateReady, &UnbondingDuration, &maxUnbondingEntries, validFees),
			false,
			true,
		},
		{
			"Draw Schedule cannot be invalid",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, &validMinDepositAmount, &validPrizeStrategy, &invalidDrawSchedule, poolStatePaused, &UnbondingDuration, &maxUnbondingEntries, validFees),
			true,
			true,
		},
		{
			"Fees Stakers cannot be invalid",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), validValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStatePaused, &UnbondingDuration, &maxUnbondingEntries, invalidFees),
			true,
			true,
		},
		{
			"Fees Stakers cannot exceed maximum allowed amount",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), validValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStatePaused, &UnbondingDuration, &maxUnbondingEntries, invalidExcessiveFees),
			true,
			true,
		},
		{
			"Partial should be fine",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), nil, nil, nil, nil, poolStateUnspecified, nil, nil, nil),
			false,
			false,
		},
		{
			"Fine should be fine",
			millionstypes.NewUpdatePoolProposal("Test", "Test", suite.pool.GetPoolId(), validValidatorSet, &validMinDepositAmount, &validPrizeStrategy, &validDrawSchedule, poolStateUnspecified, &UnbondingDuration, &maxUnbondingEntries, validFees),
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
			millionstypes.NewUpdateParamsProposal("Test", "Test", &validMinDepositAmount, &prizeDelta, &depositDelta, &minDrawDelta, &maxDrawDelta, &maxBatchQuantity, &maxStrategyBatches),
			false,
			false,
		},
		{
			"Partial update should be fine",
			millionstypes.NewUpdateParamsProposal("Test", "Test", nil, nil, nil, nil, nil, nil, nil),
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
