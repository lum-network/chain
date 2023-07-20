package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"

	tendermintlightclient "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontrollermigrations "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/migrations/v6"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctmmigrations "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/migrations"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	ibctestingtypes "github.com/cosmos/ibc-go/v7/testing/types"

	appparams "github.com/lum-network/chain/app/params"
	"github.com/lum-network/chain/x/airdrop"
	airdroptypes "github.com/lum-network/chain/x/airdrop/types"
	"github.com/lum-network/chain/x/beam"
	beamtypes "github.com/lum-network/chain/x/beam/types"
	"github.com/lum-network/chain/x/dfract"
	dfractclient "github.com/lum-network/chain/x/dfract/client"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
	"github.com/lum-network/chain/x/icacallbacks"
	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	"github.com/lum-network/chain/x/icqueries"
	icqueriestypes "github.com/lum-network/chain/x/icqueries/types"
	"github.com/lum-network/chain/x/millions"
	millionsclient "github.com/lum-network/chain/x/millions/client"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			[]govclient.ProposalHandler{
				paramsclient.ProposalHandler,
				upgradeclient.LegacyProposalHandler,
				upgradeclient.LegacyCancelProposalHandler,
				ibcclientclient.UpdateClientProposalHandler,
				ibcclientclient.UpgradeProposalHandler,
				dfractclient.UpdateParamsProposalHandler,
				millionsclient.RegisterPoolProposalHandler,
				millionsclient.UpdatePoolProposalHandler,
				millionsclient.UpdateParamsProposalHandler,
			},
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ica.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		consensus.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		vesting.AppModuleBasic{},
		tendermintlightclient.AppModuleBasic{},
		icacallbacks.AppModuleBasic{},
		icqueries.AppModuleBasic{},
		beam.AppModuleBasic{},
		airdrop.AppModuleBasic{},
		dfract.AppModuleBasic{},
		millions.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		icatypes.ModuleName:            nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		beamtypes.ModuleName:           nil,
		icqueriestypes.ModuleName:      nil,
		airdroptypes.ModuleName:        {authtypes.Minter, authtypes.Burner},
		dfracttypes.ModuleName:         {authtypes.Minter, authtypes.Burner},
		millionstypes.ModuleName:       {authtypes.Minter, authtypes.Burner, authtypes.Staking},
	}
)

var (
	_ LumApp                = (*App)(nil)
	_ ibctesting.TestingApp = (*App)(nil)
)

func init() {
	// Initialize the bech32 prefixes and coin type
	SetConfig()

	// Initialize the default denom regex
	sdk.SetCoinDenomRegex(sdk.DefaultCoinDenomRegex)

	// Acquire the homedir
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".lumd")
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	AppKeepers

	appName string

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// IBC module
	transferModule transfer.AppModule

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}

// New returns a reference to an initialized App.
func New(
	appName string,
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, ICAControllerCustomStoreKey, icahosttypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey, authzkeeper.StoreKey, crisistypes.StoreKey, consensusparamtypes.StoreKey,
		icacallbackstypes.StoreKey, icqueriestypes.StoreKey,
		beamtypes.StoreKey, airdroptypes.StoreKey, dfracttypes.StoreKey, millionstypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &App{
		BaseApp:           bApp,
		appName:           appName,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	// Initialize the keepers
	app.InitSpecialKeepers(skipUpgradeHeights, homePath, invCheckPeriod)
	app.registerUpgradeHandlers()
	app.InitNormalKeepers()
	app.SetupHooks()

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, *app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(*app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, *app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, *app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		distr.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(*app.EvidenceKeeper),
		consensus.NewAppModule(appCodec, *app.ConsensusParamsKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(*app.ParamsKeeper),
		app.transferModule,
		ica.NewAppModule(app.ICAControllerKeeper, app.ICAHostKeeper),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		beam.NewAppModule(appCodec, *app.BeamKeeper),
		airdrop.NewAppModule(appCodec, *app.AirdropKeeper),
		dfract.NewAppModule(appCodec, *app.DFractKeeper),
		millions.NewAppModule(appCodec, *app.MillionsKeeper),
		icacallbacks.NewAppModule(appCodec, *app.ICACallbacksKeeper, app.AccountKeeper, app.BankKeeper),
		icqueries.NewAppModule(appCodec, *app.ICQueriesKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibchost.ModuleName,
		icatypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		icqueriestypes.ModuleName,
		icacallbackstypes.ModuleName,
		beamtypes.ModuleName,
		airdroptypes.ModuleName,
		dfracttypes.ModuleName,
		millionstypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		ibchost.ModuleName,
		icatypes.ModuleName,
		ibctransfertypes.ModuleName,
		icqueriestypes.ModuleName,
		icacallbackstypes.ModuleName,
		beamtypes.ModuleName,
		airdroptypes.ModuleName,
		dfracttypes.ModuleName,
		millionstypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibchost.ModuleName,
		icatypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		ibctransfertypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		authz.ModuleName,
		icqueriestypes.ModuleName,
		icacallbackstypes.ModuleName,
		beamtypes.ModuleName,
		airdroptypes.ModuleName,
		dfracttypes.ModuleName,
		millionstypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// Register the simulation manager
	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, *app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, *app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, *app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		distr.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		evidence.NewAppModule(*app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		ica.NewAppModule(nil, app.ICAHostKeeper),
		params.NewAppModule(*app.ParamsKeeper),
		app.transferModule,
		ica.NewAppModule(app.ICAControllerKeeper, app.ICAHostKeeper),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
	)
	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   app.AccountKeeper,
		BankKeeper:      app.BankKeeper,
		FeegrantKeeper:  app.FeeGrantKeeper,
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
	})
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}
	app.SetAnteHandler(anteHandler)
	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(fmt.Sprintf("failed to load latest version: %s", err))
		}
	}
	return app
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application
func (app *App) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// GetStakingKeeper implements the TestingApp interface.
func (app *App) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return *app.StakingKeeper
}

// GetTransferKeeper implements the TestingApp interface.
func (app *App) GetTransferKeeper() *ibctransferkeeper.Keeper {
	return app.TransferKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *App) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *App) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetTxConfig implements the TestingApp interface.
func (app *App) GetTxConfig() client.TxConfig {
	cfg := MakeEncodingConfig()
	return cfg.TxConfig
}

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedModuleAccountAddrs returns all the app's blocked module account addresses
func (app *App) BlockedModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns Gaia's app codec.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Gaia's InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// GetSubspace returns a param subspace for a given module name.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// RegisterNodeService registers the node gRPC Query service.
func (app *App) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(ctx client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticServer))
	rtr.PathPrefix("/swagger/").Handler(staticServer)
}

func (app *App) registerUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler("v0.44", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// override versions for _new_ modules as to not skip InitGenesis
		fromVM[authz.ModuleName] = 0
		fromVM[feegrant.ModuleName] = 0

		// x/auth runs 1->2 migration
		fromVM[authtypes.ModuleName] = auth.AppModule{}.ConsensusVersion()

		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.0.4", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Upgrade contains
		// - Airdrop module fixes 2->3
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.0.5", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.IBCKeeper.ConnectionKeeper.SetParams(ctx, ibcconnectiontypes.DefaultParams())
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.1.0", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Upgrade contains
		// - Beam old to new queue data migration
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.2.0", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// set the ICS27 consensus version so InitGenesis is not run
		fromVM[icatypes.ModuleName] = app.mm.GetVersionMap()[icatypes.ModuleName]

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
		icamodule, correctTypecast := app.mm.Modules[icatypes.ModuleName].(ica.AppModule)
		if !correctTypecast {
			panic("mm.Modules[icatypes.ModuleName] is not of type ica.AppModule")
		}
		icamodule.InitModule(ctx, controllerParams, hostParams)

		// Run the migrations
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.2.1", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Set the DFract module consensus version so InitGenesis is not run
		fromVM[dfracttypes.ModuleName] = app.mm.GetVersionMap()[dfracttypes.ModuleName]

		// Apply the initial parameters
		app.DFractKeeper.SetParams(ctx, dfracttypes.DefaultParams())

		app.Logger().Info("v1.2.1 upgrade applied. DFract module live")
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.2.2", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.3.0", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("v1.3.0 upgrade applied")
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
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
		fromVM[icacallbackstypes.ModuleName] = app.mm.GetVersionMap()[icacallbackstypes.ModuleName]
		fromVM[icqueriestypes.ModuleName] = app.mm.GetVersionMap()[icqueriestypes.ModuleName]
		fromVM[millionstypes.ModuleName] = app.mm.GetVersionMap()[millionstypes.ModuleName]

		// Apply initial millions state
		genState := millionstypes.DefaultGenesisState()
		app.MillionsKeeper.SetParams(ctx, genState.Params)
		app.MillionsKeeper.SetNextPoolID(ctx, genState.NextPoolId)
		app.MillionsKeeper.SetNextDepositID(ctx, genState.NextDepositId)
		app.MillionsKeeper.SetNextPrizeID(ctx, genState.NextPrizeId)
		app.MillionsKeeper.SetNextWithdrawalID(ctx, genState.NextWithdrawalId)

		app.Logger().Info("v1.4.0 upgrade applied: Millions module enabled and ICA configuration updated.")
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.4.1", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("v1.4.1 upgrade applied")
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})

	app.UpgradeKeeper.SetUpgradeHandler("v1.4.5", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Kill the first pool that shouldn't be used anymore after that upgrade
		_, err := app.MillionsKeeper.UnsafeKillPool(ctx, 1)
		if err != nil {
			return fromVM, err
		}

		// Continue normal upgrade processing
		app.Logger().Info("Pool killed. v1.4.5 upgrade applied")
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
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
			app.keys[capabilitytypes.StoreKey],
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
		vm, _ := app.mm.RunMigrations(ctx, app.configurator, fromVM)

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
			err = app.MillionsKeeper.UpdatePool(ctx, pool.GetPoolId(), []string{}, nil, nil, &prizeStrategy, millionstypes.PoolState_Unspecified)
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
			err = app.MillionsKeeper.UpdatePool(ctx, pool.GetPoolId(), []string{}, nil, nil, &prizeStrategy, millionstypes.PoolState_Unspecified)
			if err != nil {
				return nil, err
			}
		}

		// Upgrade complete
		app.Logger().Info("v1.5.1 upgrade applied")
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
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
}
