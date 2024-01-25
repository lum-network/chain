package v162_test

import (
	"testing"
	"time"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	apptesting "github.com/lum-network/chain/app/testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	apptypes "github.com/lum-network/chain/app"
	epochstypes "github.com/lum-network/chain/x/epochs/types"
	v162 "github.com/lum-network/chain/x/millions/migrations/v162"
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

func TriggerEpochUpdate(suite *StoreMigrationTestSuite) (epochInfo epochstypes.EpochInfo, err error) {
	app := suite.app
	ctx := suite.ctx
	epochsInfo, _ := app.EpochsKeeper.GetEpochInfo(ctx, epochstypes.DAY_EPOCH)
	epochsInfo.CurrentEpoch++
	epochsInfo.StartTime = ctx.BlockTime()
	epochsInfo.CurrentEpochStartHeight = ctx.BlockHeight()
	app.EpochsKeeper.SetEpochInfo(ctx, epochsInfo)

	return epochsInfo, nil
}

func GetEpochInfo(suite *StoreMigrationTestSuite) (epochInfo epochstypes.EpochInfo, err error) {
	app := suite.app
	ctx := suite.ctx
	epochsInfo, _ := app.EpochsKeeper.GetEpochInfo(ctx, epochstypes.DAY_EPOCH)

	return epochsInfo, nil
}

func TriggerEpochTrackerUpdate(suite *StoreMigrationTestSuite, epochInfo epochstypes.EpochInfo) (epochTracker millionstypes.EpochTracker, err error) {
	app := suite.app
	ctx := suite.ctx
	epochTracker, err = app.MillionsKeeper.UpdateEpochTracker(ctx, epochInfo, "withdrawal")
	suite.Require().NoError(err)
	return epochTracker, nil
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

	if pool.UnbondingDuration == 0 {
		pool.UnbondingDuration = millionstypes.DefaultUnbondingDuration
	}
	if pool.MaxUnbondingEntries.IsNil() {
		pool.MaxUnbondingEntries = sdk.NewInt(millionstypes.DefaultMaxUnbondingEntries)
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
		MinDepositDrawDelta:     millionstypes.MinAcceptableDepositDrawDelta,
	})
}

func (suite *StoreMigrationTestSuite) TestMigratePendingWithdrawalsToEpochUnbonding() {
	poolID := suite.app.MillionsKeeper.GetNextPoolIDAndIncrement(suite.ctx)
	suite.app.MillionsKeeper.AddPool(suite.ctx, newValidPool(suite, millionstypes.Pool{PoolId: poolID}))

	// Grab our pool entity
	pool, err := suite.app.MillionsKeeper.GetPool(suite.ctx, poolID)
	suite.Require().NoError(err)

	for i := 0; i < 5; i++ {
		suite.app.MillionsKeeper.AddDeposit(suite.ctx, &millionstypes.Deposit{
			PoolId:           pool.PoolId,
			DepositorAddress: suite.addrs[i].String(),
			WinnerAddress:    suite.addrs[i].String(),
			State:            millionstypes.DepositState_IbcTransfer,
			Amount:           sdk.NewCoin("ulum", sdk.NewInt(1_000_0)),
		})
	}

	deposits := suite.app.MillionsKeeper.ListDeposits(suite.ctx)

	for i := 0; i < 5; i++ {
		suite.app.MillionsKeeper.AddWithdrawal(suite.ctx, millionstypes.Withdrawal{
			PoolId:           pool.PoolId,
			DepositId:        deposits[i].DepositId,
			DepositorAddress: suite.addrs[i].String(),
			ToAddress:        suite.addrs[i].String(),
			State:            millionstypes.WithdrawalState_Pending,
			ErrorState:       millionstypes.WithdrawalState_Unspecified,
			Amount:           sdk.NewCoin("ulum", sdk.NewInt(1_000_0)),
		})
	}

	// Delete deposits
	deposits = suite.app.MillionsKeeper.ListDeposits(suite.ctx)
	for i := 0; i < 5; i++ {
		suite.app.MillionsKeeper.RemoveDeposit(suite.ctx, &millionstypes.Deposit{
			PoolId:           deposits[i].PoolId,
			DepositId:        deposits[i].DepositId,
			DepositorAddress: suite.addrs[i].String(),
			WinnerAddress:    suite.addrs[i].String(),
			State:            millionstypes.DepositState_Success,
			Amount:           sdk.NewCoin("ulum", sdk.NewInt(1_000_0)),
		})
	}

	for epoch := int64(1); epoch <= 1; epoch++ {
		epochInfo, err := TriggerEpochUpdate(suite)
		suite.Require().NoError(err)
		suite.Require().Equal(epoch, epochInfo.CurrentEpoch)

		_, err = TriggerEpochTrackerUpdate(suite, epochInfo)
		suite.Require().NoError(err)
	}

	// Manually set the withdrawals on the first epoch unbonding to simulate the migration
	epochPoolUnbonding := millionstypes.EpochUnbonding{
		EpochIdentifier:    "day",
		EpochNumber:        1,
		PoolId:             1,
		WithdrawalIds:      []uint64{1, 2, 3, 4, 5},
		WithdrawalIdsCount: 5,
		TotalAmount:        sdk.NewCoin("ulum", sdk.NewInt(5_000_0)),
		CreatedAtHeight:    suite.ctx.BlockHeight(),
		UpdatedAtHeight:    suite.ctx.BlockHeight(),
		CreatedAt:          suite.ctx.BlockTime(),
		UpdatedAt:          suite.ctx.BlockTime(),
	}

	suite.app.MillionsKeeper.SetEpochPoolUnbonding(suite.ctx, epochPoolUnbonding)

	// Get the millions internal epoch tracker
	epochTracker, err := suite.app.MillionsKeeper.GetEpochTracker(suite.ctx, epochstypes.DAY_EPOCH, millionstypes.WithdrawalTrackerType)
	suite.Require().NoError(err)
	// Get epoch unbonding
	currentEpochUnbonding := suite.app.MillionsKeeper.GetEpochUnbondings(suite.ctx, epochTracker.EpochNumber)
	suite.Require().NoError(err)
	// Should be in first epoch
	suite.Require().Equal(uint64(1), currentEpochUnbonding[0].EpochNumber)
	// Should have 5 withdrawals
	suite.Require().Equal(uint64(5), currentEpochUnbonding[0].WithdrawalIdsCount)
	suite.Require().Equal(sdk.NewCoin("ulum", sdk.NewInt(5_000_0)), currentEpochUnbonding[0].TotalAmount)

	// Old withdrawals values
	oldWithdrawals := suite.app.MillionsKeeper.ListWithdrawals(suite.ctx)
	for _, oldWithdrawal := range oldWithdrawals {
		suite.Require().Equal(millionstypes.WithdrawalState_Pending, oldWithdrawal.State)
		suite.Require().Equal(millionstypes.WithdrawalState_Unspecified, oldWithdrawal.ErrorState)
	}

	// Run the migration operation
	err = v162.MigratePendingWithdrawalsToNewEpochUnbonding(suite.ctx, *suite.app.MillionsKeeper)
	suite.Require().NoError(err)

	// The state should remain unchanged
	newWithdrawals := suite.app.MillionsKeeper.ListWithdrawals(suite.ctx)
	for _, newWithdrawals := range newWithdrawals {
		suite.Require().Equal(millionstypes.WithdrawalState_Unspecified, newWithdrawals.ErrorState)
		suite.Require().Equal(millionstypes.WithdrawalState_Pending, newWithdrawals.State)
	}

	// The old epochUnbonding from epoch 1 should be cleared and contain no more withdrawals
	epochUnbonding := suite.app.MillionsKeeper.GetEpochUnbondings(suite.ctx, 1)
	suite.Require().Len(epochUnbonding, 0)

	// The withdrawals have been pushed to the 4th epochUnbonding - Epoch on which Hook will process them
	currentEpochUnbonding = suite.app.MillionsKeeper.GetEpochUnbondings(suite.ctx, epochTracker.EpochNumber+3)
	suite.Require().NoError(err)
	// Should have 1 entity for the current pool withdrawal
	suite.Require().Len(currentEpochUnbonding, 1)
	// Should be in 4th epoch
	suite.Require().Equal(uint64(4), currentEpochUnbonding[0].EpochNumber)
	// Should have 5 withdrawals
	suite.Require().Equal(uint64(5), currentEpochUnbonding[0].WithdrawalIdsCount)
	suite.Require().Equal(sdk.NewCoin("ulum", sdk.NewInt(5_000_0)), currentEpochUnbonding[0].TotalAmount)
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(StoreMigrationTestSuite))
}
