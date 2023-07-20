package v152_test

import (
	"testing"
	"time"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	apptesting "github.com/lum-network/chain/app/testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	apptypes "github.com/lum-network/chain/app"
	v152 "github.com/lum-network/chain/x/millions/migrations/v152"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

type StoreMigrationTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *apptypes.App
	addrs       []sdk.AccAddress
	moduleAddrs []sdk.AccAddress
	valAddrs    []sdk.ValAddress
}

// newValidPool fills up missing params to the pool to make it valid in order to ease testing
func newValidPool(suite *StoreMigrationTestSuite, pool millionstypes.Pool) *millionstypes.Pool {
	params := suite.app.MillionsKeeper.GetParams(suite.ctx)

	if pool.Denom == "" {
		pool.Denom = suite.app.StakingKeeper.BondDenom(suite.ctx)
	}
	if pool.NativeDenom == "" {
		pool.NativeDenom = suite.app.StakingKeeper.BondDenom(suite.ctx)
	}
	if pool.ChainId == "" {
		pool.ChainId = "lum-network-devnet"
	}

	if pool.PoolType == millionstypes.PoolType_Unspecified {
		pool.PoolType = millionstypes.PoolType_Staking
	}

	if pool.Validators == nil {
		for _, addr := range suite.valAddrs {
			pool.Validators = append(pool.Validators, millionstypes.PoolValidator{
				OperatorAddress: addr.String(),
				BondedAmount:    sdk.ZeroInt(),
				IsEnabled:       true,
			})
		}
	}
	if pool.Bech32PrefixAccAddr == "" {
		pool.Bech32PrefixAccAddr = sdk.GetConfig().GetBech32AccountAddrPrefix()
	}
	if pool.Bech32PrefixValAddr == "" {
		pool.Bech32PrefixValAddr = sdk.GetConfig().GetBech32ValidatorAddrPrefix()
	}
	if pool.MinDepositAmount.IsNil() {
		pool.MinDepositAmount = params.MinDepositAmount
	}
	if err := pool.DrawSchedule.ValidateBasic(params); err != nil {
		pool.DrawSchedule = millionstypes.DrawSchedule{DrawDelta: 1 * time.Hour, InitialDrawAt: time.Now().UTC()}
	}
	if err := pool.PrizeStrategy.Validate(params); err != nil {
		pool.PrizeStrategy = millionstypes.PrizeStrategy{PrizeBatches: []millionstypes.PrizeBatch{{PoolPercent: 100, Quantity: 1, DrawProbability: sdk.NewDec(1)}}}
	}
	if pool.LocalAddress == "" {
		pool.LocalAddress = suite.moduleAddrs[0].String()
	}
	if pool.IcaDepositAddress == "" {
		pool.IcaDepositAddress = suite.moduleAddrs[1].String()
	}
	if pool.IcaPrizepoolAddress == "" {
		pool.IcaPrizepoolAddress = suite.moduleAddrs[2].String()
	}
	if pool.NextDrawId == millionstypes.UnknownID {
		pool.NextDrawId = millionstypes.UnknownID + 1
	}
	if pool.TvlAmount.IsNil() {
		pool.TvlAmount = sdk.ZeroInt()
	}
	if pool.SponsorshipAmount.IsNil() {
		pool.SponsorshipAmount = sdk.ZeroInt()
	}
	if pool.AvailablePrizePool.IsNil() {
		pool.AvailablePrizePool = sdk.NewCoin(pool.Denom, sdk.ZeroInt())
	}
	if pool.State == millionstypes.PoolState_Unspecified {
		pool.State = millionstypes.PoolState_Ready
	}
	if pool.CreatedAt.IsZero() {
		pool.CreatedAt = suite.ctx.BlockTime()
	}
	if pool.UpdatedAt.IsZero() {
		pool.UpdatedAt = suite.ctx.BlockTime()
	}
	if pool.IcaDepositPortId == "" {
		pool.IcaDepositPortId = icatypes.ControllerPortPrefix + string(millionstypes.NewPoolName(pool.GetPoolId(), millionstypes.ICATypeDeposit))
	}
	if pool.IcaPrizepoolPortId == "" {
		pool.IcaPrizepoolPortId = icatypes.ControllerPortPrefix + string(millionstypes.NewPoolName(pool.GetPoolId(), millionstypes.ICATypePrizePool))
	}
	return &pool
}

func (suite *StoreMigrationTestSuite) SetupTest() {
	app := apptypes.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Setup the default application
	suite.app = app
	suite.ctx = ctx.WithChainID("lum-network-devnet").WithBlockTime(time.Now().UTC())

	// Setup test account addresses
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 6, sdk.NewInt(10_000_000_000), app.StakingKeeper.BondDenom(ctx))
	for i := 0; i < 6; i++ {
		poolAddress := millionstypes.NewPoolAddress(uint64(i+1), "unused-in-test")
		apptesting.AddTestModuleAccount(app, ctx, poolAddress)
		suite.moduleAddrs = append(suite.moduleAddrs, poolAddress)
	}

	// Setup vals for pools
	vals := app.StakingKeeper.GetAllValidators(ctx)
	suite.valAddrs = []sdk.ValAddress{}
	for _, v := range vals {
		addr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
		suite.Require().NoError(err)
		suite.valAddrs = append(suite.valAddrs, addr)
	}

	// Setup test params
	suite.app.MillionsKeeper.SetParams(ctx, millionstypes.Params{
		MinDepositAmount:        sdk.NewInt(millionstypes.MinAcceptableDepositAmount),
		MaxPrizeStrategyBatches: 1_000,
		MaxPrizeBatchQuantity:   1_000_000,
		MinDrawScheduleDelta:    1 * time.Hour,
		MaxDrawScheduleDelta:    366 * 24 * time.Hour,
		PrizeExpirationDelta:    30 * 24 * time.Hour,
		FeesStakers:             sdk.ZeroDec(),
		MinDepositDrawDelta:     millionstypes.MinAcceptableDepositDrawDelta,
	})
}

func (suite *StoreMigrationTestSuite) TestUpdatePoolType() {
	poolID := suite.app.MillionsKeeper.GetNextPoolIDAndIncrement(suite.ctx)
	suite.app.MillionsKeeper.AddPool(suite.ctx, newValidPool(suite, millionstypes.Pool{PoolId: poolID}))

	// Run the migration operation
	err := v152.MigratePoolType(suite.ctx, *suite.app.MillionsKeeper)
	suite.Require().NoError(err)

	// Grab our pool
	pool, err := suite.app.MillionsKeeper.GetPool(suite.ctx, poolID)
	suite.Require().NoError(err)

	// Ensure our pool has the new poolType
	suite.Require().Equal(millionstypes.PoolType_Staking, pool.PoolType)
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(StoreMigrationTestSuite))
}
