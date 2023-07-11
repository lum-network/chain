package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	apptesting "github.com/lum-network/chain/app/testing"
	epochstypes "github.com/lum-network/chain/x/epochs/types"
	millionstypes "github.com/lum-network/chain/x/millions/types"

	"github.com/lum-network/chain/app"
)

const testChainID = "lum-network-devnet-1"

var (
	cosmosPoolValidator       = "cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn"
	cosmosIcaDepositAddress   = "cosmos1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep4tgu9q"
	cosmosIcaPrizePoolAddress = "cosmos196ax4vc0lwpxndu9dyhvca7jhxp70rmcfhxsrt"
	cosmosAccAddress          = "cosmos1qggwr7vze9x7ustru9dctf6krhk2285lkf89g7"
	localPoolDenom            = "ulum"
	remotePoolDenom           = "uatom"
	remoteBech32PrefixValAddr = "cosmosvaloper"
	remoteBech32PrefixAccAddr = "cosmos"
	remoteChainId             = "cosmos"
	remoteConnectionId        = "connection-id"
	remoteTransferChannelId   = "transfer"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.App
	addrs       []sdk.AccAddress
	moduleAddrs []sdk.AccAddress
	valAddrs    []sdk.ValAddress
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app := app.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Setup the default application
	suite.app = app
	suite.ctx = ctx.WithChainID(testChainID).WithBlockTime(time.Now().UTC())

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

func TriggerEpochUpdate(suite *KeeperTestSuite) (epochInfo epochstypes.EpochInfo, err error) {
	app := suite.app
	ctx := suite.ctx
	epochsInfo, _ := app.EpochsKeeper.GetEpochInfo(ctx, epochstypes.DAY_EPOCH)
	epochsInfo.CurrentEpoch++
	epochsInfo.StartTime = ctx.BlockTime()
	epochsInfo.CurrentEpochStartHeight = ctx.BlockHeight()
	app.EpochsKeeper.SetEpochInfo(ctx, epochsInfo)

	return epochsInfo, nil
}

func TriggerEpochTrackerUpdate(suite *KeeperTestSuite, epochInfo epochstypes.EpochInfo) (epochTracker millionstypes.EpochTracker, err error) {
	app := suite.app
	ctx := suite.ctx
	epochTracker, err = app.MillionsKeeper.UpdateEpochTracker(ctx, epochInfo, "withdrawal")
	suite.Require().NoError(err)
	return epochTracker, nil
}

// floatToDec simple helper to create sdk.Dec with a precision of 6
func floatToDec(v float64) sdk.Dec {
	return sdk.NewDecWithPrec(int64(v*1_000_000), 6)
}

// newValidPool fills up missing params to the pool to make it valid in order to ease testing
func newValidPool(suite *KeeperTestSuite, pool millionstypes.Pool) *millionstypes.Pool {
	params := suite.app.MillionsKeeper.GetParams(suite.ctx)

	if pool.Denom == "" {
		pool.Denom = suite.app.StakingKeeper.BondDenom(suite.ctx)
	}
	if pool.NativeDenom == "" {
		pool.NativeDenom = suite.app.StakingKeeper.BondDenom(suite.ctx)
	}
	if pool.ChainId == "" {
		pool.ChainId = testChainID
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
	return &pool
}
