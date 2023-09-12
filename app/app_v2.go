package app

import (
	"cosmossdk.io/depinject"
	errorsmod "cosmossdk.io/errors"
	"fmt"
	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontrollermigrations "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/migrations/v6"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctmmigrations "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/migrations"
	beamtypes "github.com/lum-network/chain/x/beam/types"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
	epochstypes "github.com/lum-network/chain/x/epochs/types"
	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	icqueriestypes "github.com/lum-network/chain/x/icqueries/types"
	millionstypes "github.com/lum-network/chain/x/millions/types"
	"io"
	"os"
	"path/filepath"
)

type AppV2 struct {
	*runtime.App
	AppKeepers

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	DefaultNodeHome = filepath.Join(userHomeDir, ".lumd")
}

func NewAppV2(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp)) *AppV2 {
	var (
		app        = &AppV2{}
		appBuilder *runtime.AppBuilder

		appConfig = depinject.Configs(
			AppConfig,
			depinject.Supply(
				appOpts,
				logger,
			),
		)
	)

	if err := depinject.Inject(appConfig,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.AccountKeeper,
		&app.BankKeeper,
		&app.StakingKeeper,
		&app.SlashingKeeper,
		&app.MintKeeper,
		&app.DistrKeeper,
		&app.GovKeeper,
		&app.CrisisKeeper,
		&app.UpgradeKeeper,
		&app.ParamsKeeper,
		&app.AuthzKeeper,
		&app.EvidenceKeeper,
		&app.FeeGrantKeeper,
		&app.ConsensusParamsKeeper,
		&app.IBCKeeper,
		&app.ICAHostKeeper,
		&app.ICAControllerKeeper,
		&app.TransferKeeper,
		&app.ICACallbacksKeeper,
		&app.ICQueriesKeeper,
		&app.EpochsKeeper,
		&app.AirdropKeeper,
		&app.BeamKeeper,
		&app.DFractKeeper,
		&app.MillionsKeeper,
	); err != nil {
		panic(err)
	}

	// Create our application
	app.App = appBuilder.Build(logger, db, traceStore, baseAppOptions...)

	// Register streaming services
	/*if err := app.RegisterStreamingServices(appOpts, app.kvStoreKeys()); err != nil {
		panic(err)
	}*/

	// Register invariants modules
	app.ModuleManager.RegisterInvariants(app.CrisisKeeper)

	// Register our upgrade handlers
	app.RegisterUpgradeHandlers()

	// Load the chain
	if err := app.Load(loadLatest); err != nil {
		panic(err)
	}

	return app
}

func (app *AppV2) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)
	// register swagger API in app.go so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

func (app *AppV2) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler("v0.44", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// override versions for _new_ modules as to not skip InitGenesis
		fromVM[authz.ModuleName] = 0
		fromVM[feegrant.ModuleName] = 0

		// x/auth runs 1->2 migration
		fromVM[authtypes.ModuleName] = auth.AppModule{}.ConsensusVersion()

		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.0.4", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Upgrade contains
		// - Airdrop module fixes 2->3
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.0.5", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.IBCKeeper.ConnectionKeeper.SetParams(ctx, ibcconnectiontypes.DefaultParams())
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.1.0", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Upgrade contains
		// - Beam old to new queue data migration
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.2.0", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// set the ICS27 consensus version so InitGenesis is not run
		fromVM[icatypes.ModuleName] = app.ModuleManager.GetVersionMap()[icatypes.ModuleName]

		// create ICS27 Controller submodule params, controller module not enabled.
		controllerParams := icacontrollertypes.Params{}

		// create ICS27 Host submodule params
		hostParams := icahosttypes.Params{
			HostEnabled: true,
			AllowMessages: []string{
				sdk.MsgTypeURL(&banktypes.MsgSend{}),
				sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
				sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
				sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
				sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
				sdk.MsgTypeURL(&govtypesv1beta1.MsgVote{}),
				sdk.MsgTypeURL(&authz.MsgExec{}),
				sdk.MsgTypeURL(&authz.MsgGrant{}),
				sdk.MsgTypeURL(&authz.MsgRevoke{}),
				sdk.MsgTypeURL(&beamtypes.MsgOpenBeam{}),
				sdk.MsgTypeURL(&beamtypes.MsgUpdateBeam{}),
				sdk.MsgTypeURL(&beamtypes.MsgClaimBeam{}),
			},
		}

		// initialize ICS27 module
		icamodule, correctTypecast := app.ModuleManager.Modules[icatypes.ModuleName].(ica.AppModule)
		if !correctTypecast {
			panic("mm.Modules[icatypes.ModuleName] is not of type ica.AppModule")
		}
		icamodule.InitModule(ctx, controllerParams, hostParams)

		// Run the migrations
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.2.1", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Set the DFract module consensus version so InitGenesis is not run
		fromVM[dfracttypes.ModuleName] = app.ModuleManager.GetVersionMap()[dfracttypes.ModuleName]

		// Apply the initial parameters
		app.DFractKeeper.SetParams(ctx, dfracttypes.DefaultParams())

		app.Logger().Info("v1.2.1 upgrade applied. DFract module live")
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.2.2", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.3.0", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("v1.3.0 upgrade applied")
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.4.0", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Get the actual params
		icaHostParams := app.ICAHostKeeper.GetParams(ctx)

		// Patch the parameters - Enable the controller and patch the allowed messages for host
		icaControllerParams := icacontrollertypes.Params{
			ControllerEnabled: true,
		}
		icaHostParams.AllowMessages = append(icaHostParams.AllowMessages,
			// Change: added dfract and millions messages
			sdk.MsgTypeURL(&dfracttypes.MsgDeposit{}),
			sdk.MsgTypeURL(&millionstypes.MsgDeposit{}),
			sdk.MsgTypeURL(&millionstypes.MsgDepositRetry{}),
			sdk.MsgTypeURL(&millionstypes.MsgClaimPrize{}),
			sdk.MsgTypeURL(&millionstypes.MsgWithdrawDeposit{}),
			sdk.MsgTypeURL(&millionstypes.MsgWithdrawDepositRetry{}),
			sdk.MsgTypeURL(&millionstypes.MsgDrawRetry{}),
		)

		// Apply patched parameters
		app.ICAHostKeeper.SetParams(ctx, icaHostParams)
		app.ICAControllerKeeper.SetParams(ctx, icaControllerParams)

		// Set the ICA Callbacks, ICQueries and Millions modules versions so InitGenesis is not run
		fromVM[icacallbackstypes.ModuleName] = app.ModuleManager.GetVersionMap()[icacallbackstypes.ModuleName]
		fromVM[icqueriestypes.ModuleName] = app.ModuleManager.GetVersionMap()[icqueriestypes.ModuleName]
		fromVM[millionstypes.ModuleName] = app.ModuleManager.GetVersionMap()[millionstypes.ModuleName]

		// Apply initial millions state
		genState := millionstypes.DefaultGenesisState()
		app.MillionsKeeper.SetParams(ctx, genState.Params)
		app.MillionsKeeper.SetNextPoolID(ctx, genState.NextPoolId)
		app.MillionsKeeper.SetNextDepositID(ctx, genState.NextDepositId)
		app.MillionsKeeper.SetNextPrizeID(ctx, genState.NextPrizeId)
		app.MillionsKeeper.SetNextWithdrawalID(ctx, genState.NextWithdrawalId)

		app.Logger().Info("v1.4.0 upgrade applied: Millions module enabled and ICA configuration updated.")
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.4.1", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("v1.4.1 upgrade applied")
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.4.5", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Kill the first pool that shouldn't be used anymore after that upgrade
		_, err := app.MillionsKeeper.UnsafeKillPool(ctx, 1)
		if err != nil {
			return fromVM, err
		}

		// Continue normal upgrade processing
		app.Logger().Info("Pool killed. v1.4.5 upgrade applied")
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.5.0", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("Starting v1.5.0 upgrade")

		// Migrate the consensus module params
		app.Logger().Info("Migrate the consensus module params...")
		legacyParamSubspace := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		baseapp.MigrateParams(ctx, legacyParamSubspace, app.ConsensusParamsKeeper)

		// Migrate DFract params, set the first withdrawal address (can be patched later on through proposal)
		app.Logger().Info("Migrate the DFract params...")
		dfrParams := dfracttypes.DefaultParams()
		dfrParams.WithdrawalAddress = "lum1euhszjasgkeskujz6zr42r3lsxv58mfgsmlps0"
		app.DFractKeeper.SetParams(ctx, dfrParams)

		// Migrate ICA channel capabilities from IBC V5 to IBC V6
		app.Logger().Info("Migrate the ICS27 channel capabilities...")
		if err := icacontrollermigrations.MigrateICS27ChannelCapability(
			ctx,
			app.appCodec,
			app.GetKVStoreKey(capabilitytypes.StoreKey),
			app.CapabilityKeeper,
			millionstypes.ModuleName,
		); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to migrate ICA channel capabilities")
		}

		// Migrate clients, and add the localhost type
		app.Logger().Info("Migrate the IBC allowed clients...")
		params := app.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = append(params.AllowedClients, exported.Localhost)
		app.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		// Prune expired client states
		app.Logger().Info("Prune expired client states...")
		if _, err := ibctmmigrations.PruneExpiredConsensusStates(ctx, app.appCodec, app.IBCKeeper.ClientKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to prune expired consensus states")
		}

		// Run migrations to ensure we patch our IBC channels names
		vm, _ := app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)

		// Change the Millions Pool prize strategy
		// We check if we are able to find the pool with ID 2, but we don't error out in the other case, to allow running on testnet as well
		app.Logger().Info("Patch the Millions prize strategy...")
		pool, err := app.MillionsKeeper.GetPool(ctx, 2)
		if err == nil {
			prizeStrategy := millionstypes.PrizeStrategy{
				PrizeBatches: []millionstypes.PrizeBatch{
					{PoolPercent: 50, Quantity: 1, IsUnique: true, DrawProbability: sdk.NewDecWithPrec(20, 2)},
					{PoolPercent: 25, Quantity: 5, IsUnique: false, DrawProbability: sdk.NewDecWithPrec(20, 2)},
					{PoolPercent: 17, Quantity: 25, IsUnique: false, DrawProbability: sdk.NewDecWithPrec(20, 2)},
					{PoolPercent: 8, Quantity: 60, IsUnique: false, DrawProbability: sdk.NewDecWithPrec(90, 2)},
				},
			}
			err = app.MillionsKeeper.UpdatePool(ctx, pool.GetPoolId(), []string{}, nil, nil, nil, nil, &prizeStrategy, millionstypes.PoolState_Unspecified)
			if err != nil {
				return nil, err
			}
		}

		// Upgrade complete
		app.Logger().Info("v1.5.0 upgrade applied")
		return vm, err
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.5.1", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("Starting v1.5.1 upgrade")

		// Change the Millions Pool prize strategy
		// We check if we are able to find the pool with ID 2, but we don't error out in the other case, to allow running on testnet as well
		app.Logger().Info("Patch the Millions prize strategy...")
		pool, err := app.MillionsKeeper.GetPool(ctx, 2)
		if err == nil {
			prizeStrategy := millionstypes.PrizeStrategy{
				PrizeBatches: []millionstypes.PrizeBatch{
					{PoolPercent: 50, Quantity: 1, IsUnique: true, DrawProbability: sdk.NewDecWithPrec(20, 2)},
					{PoolPercent: 25, Quantity: 5, IsUnique: true, DrawProbability: sdk.NewDecWithPrec(20, 2)},
					{PoolPercent: 17, Quantity: 25, IsUnique: true, DrawProbability: sdk.NewDecWithPrec(20, 2)},
					{PoolPercent: 8, Quantity: 60, IsUnique: true, DrawProbability: sdk.NewDecWithPrec(90, 2)},
				},
			}
			err = app.MillionsKeeper.UpdatePool(ctx, pool.GetPoolId(), []string{}, nil, nil, nil, nil, &prizeStrategy, millionstypes.PoolState_Unspecified)
			if err != nil {
				return nil, err
			}
		}

		// Upgrade complete
		app.Logger().Info("v1.5.1 upgrade applied")
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.5.2", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("Starting v1.5.2 upgrade")

		app.Logger().Info("v1.5.2 upgrade applied")
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.6.0", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("Starting v1.6.0 upgrade")

		app.Logger().Info("v1.6.0 upgrade applied")
		return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	})

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if upgradeInfo.Name == "v0.44" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{authz.ModuleName, feegrant.ModuleName},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
	if upgradeInfo.Name == "v1.2.0" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{icacontrollertypes.StoreKey, icahosttypes.StoreKey},
		}

		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
	if upgradeInfo.Name == "v1.2.1" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{dfracttypes.StoreKey},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
	if upgradeInfo.Name == "v1.2.2" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if upgradeInfo.Name == "v1.3.0" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if upgradeInfo.Name == "v1.3.1" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Deleted: []string{ibcfeetypes.ModuleName},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if upgradeInfo.Name == "v1.4.0" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// We create 3 new modules: ICA Callbacks, ICQueries, Millions
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{icacallbackstypes.StoreKey, icqueriestypes.StoreKey, millionstypes.StoreKey},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if upgradeInfo.Name == "v1.4.1" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Deleted: []string{icacontrollertypes.StoreKey},
			Added:   []string{ICAControllerCustomStoreKey},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if upgradeInfo.Name == "v1.4.5" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if upgradeInfo.Name == "v1.5.0" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// We create 1 new module: Crisis
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{crisistypes.StoreKey, consensusparamtypes.StoreKey},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if upgradeInfo.Name == "v1.5.1" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if upgradeInfo.Name == "v1.5.2" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{epochstypes.StoreKey},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if upgradeInfo.Name == "v1.6.0" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

func (app *AppV2) kvStoreKeys() map[string]*storetypes.KVStoreKey {
	keys := make(map[string]*storetypes.KVStoreKey)
	for _, k := range app.GetStoreKeys() {
		if kv, ok := k.(*storetypes.KVStoreKey); ok {
			keys[kv.Name()] = kv
		}
	}

	return keys
}

func (app *AppV2) GetKVStoreKey(storeKey string) *storetypes.KVStoreKey {
	sk := app.UnsafeFindStoreKey(storeKey)
	kvStoreKey, ok := sk.(*storetypes.KVStoreKey)
	if !ok {
		return nil
	}
	return kvStoreKey
}

func (app *AppV2) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

func (app *AppV2) AppCodec() codec.Codec {
	return app.appCodec
}

func (app *AppV2) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

func (app *AppV2) TxConfig() client.TxConfig {
	return app.txConfig
}
