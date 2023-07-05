package keeper_test

import (
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	apptesting "github.com/lum-network/chain/app/testing"
	millionstypes "github.com/lum-network/chain/x/millions/types"

	"github.com/lum-network/chain/app"
)

const testChainID = "LUM-NETWORK"

var (
	cosmosPoolValidator       = "cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn"
	cosmosIcaDepositAddress   = "cosmos1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep4tgu9q"
	cosmosIcaPrizePoolAddress = "cosmos196ax4vc0lwpxndu9dyhvca7jhxp70rmcfhxsrt"
	cosmosAccAddress          = "cosmos1qggwr7vze9x7ustru9dctf6krhk2285lkf89g7"
	localPoolDenom            = "ulum"
	remotePoolDenom           = "uatom"
	remotePoolDenomIBC        = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	remoteBech32PrefixValAddr = "cosmosvaloper"
	remoteBech32PrefixAccAddr = "cosmos"
	remoteChainId             = "cosmos"
)

type KeeperTestSuite struct {
	app.TestPackage

	addrs       []sdk.AccAddress
	moduleAddrs []sdk.AccAddress
	valAddrs    []sdk.ValAddress
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	// Setup test account addresses
	coins := sdk.NewCoins(sdk.NewCoin(suite.App.StakingKeeper.BondDenom(suite.Ctx), sdk.NewInt(10_000_000_000)), sdk.NewCoin(remotePoolDenomIBC, sdk.NewInt(10_000_000_000)))
	suite.addrs = apptesting.AddTestAddrsWithDenom(suite.App, suite.Ctx, 6, coins)
	for i := 0; i < 6; i++ {
		poolAddress := millionstypes.NewPoolAddress(uint64(i+1), "unused-in-test")
		apptesting.AddTestModuleAccount(suite.App, suite.Ctx, poolAddress)
		suite.moduleAddrs = append(suite.moduleAddrs, poolAddress)
	}

	// Setup vals for pools
	vals := suite.App.StakingKeeper.GetAllValidators(suite.Ctx)
	suite.valAddrs = []sdk.ValAddress{}
	for _, v := range vals {
		addr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
		suite.Require().NoError(err)
		suite.valAddrs = append(suite.valAddrs, addr)
	}

	// Setup test params
	suite.App.MillionsKeeper.SetParams(suite.Ctx, millionstypes.Params{
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

// GetMsgServer Returns a fresh implementation of the message server
// This must be used everywhere, instead of static msg server acquisition
// It ensures we always have a working context (depending on IBC / Non IBC testing)
func (suite *KeeperTestSuite) GetMsgServer() millionstypes.MsgServer {
	return millionskeeper.NewMsgServerImpl(*suite.App.MillionsKeeper)
}

// floatToDec simple helper to create sdk.Dec with a precision of 6
func floatToDec(v float64) sdk.Dec {
	return sdk.NewDecWithPrec(int64(v*1_000_000), 6)
}

// newValidPool fills up missing params to the pool to make it valid in order to ease testing
func newValidPool(suite *KeeperTestSuite, pool millionstypes.Pool) *millionstypes.Pool {
	params := suite.App.MillionsKeeper.GetParams(suite.Ctx)

	if pool.Denom == "" {
		pool.Denom = suite.App.StakingKeeper.BondDenom(suite.Ctx)
	}
	if pool.NativeDenom == "" {
		pool.NativeDenom = suite.App.StakingKeeper.BondDenom(suite.Ctx)
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
		pool.CreatedAt = suite.Ctx.BlockTime()
	}
	if pool.UpdatedAt.IsZero() {
		pool.UpdatedAt = suite.Ctx.BlockTime()
	}
	return &pool
}
