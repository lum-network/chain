package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibcfeekeeper "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	airdropkeeper "github.com/lum-network/chain/x/airdrop/keeper"
	airdroptypes "github.com/lum-network/chain/x/airdrop/types"
	beamkeeper "github.com/lum-network/chain/x/beam/keeper"
	beamtypes "github.com/lum-network/chain/x/beam/types"
	"github.com/lum-network/chain/x/dfract"
	dfractkeeper "github.com/lum-network/chain/x/dfract/keeper"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
	icacallbackskeeper "github.com/lum-network/chain/x/icacallbacks/keeper"
	icacallbackstypes "github.com/lum-network/chain/x/icacallbacks/types"
	icquerieskeeper "github.com/lum-network/chain/x/icqueries/keeper"
	icqueriestypes "github.com/lum-network/chain/x/icqueries/types"
	"github.com/lum-network/chain/x/millions"
	millionskeeper "github.com/lum-network/chain/x/millions/keeper"
	millionstypes "github.com/lum-network/chain/x/millions/types"
)

type AppKeepers struct {
	// Special Keepers
	ParamsKeeper          *paramskeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ConsensusParamsKeeper *consensusparamkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedICACallbacksKeeper  capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedMillionsKeeper      capabilitykeeper.ScopedKeeper

	// Normal Keepers
	AccountKeeper       *authkeeper.AccountKeeper
	BankKeeper          *bankkeeper.BaseKeeper
	AuthzKeeper         *authzkeeper.Keeper
	StakingKeeper       *stakingkeeper.Keeper
	DistrKeeper         *distrkeeper.Keeper
	SlashingKeeper      *slashingkeeper.Keeper
	IBCKeeper           *ibckeeper.Keeper
	IBCFeeKeeper        *ibcfeekeeper.Keeper
	ICAHostKeeper       *icahostkeeper.Keeper
	ICAControllerKeeper *icacontrollerkeeper.Keeper
	TransferKeeper      *ibctransferkeeper.Keeper
	EvidenceKeeper      *evidencekeeper.Keeper
	MintKeeper          *mintkeeper.Keeper
	GovKeeper           *govkeeper.Keeper
	FeeGrantKeeper      *feegrantkeeper.Keeper

	// Custom Keepers
	ICACallbacksKeeper *icacallbackskeeper.Keeper
	ICQueriesKeeper    *icquerieskeeper.Keeper
	AirdropKeeper      *airdropkeeper.Keeper
	BeamKeeper         *beamkeeper.Keeper
	DFractKeeper       *dfractkeeper.Keeper
	MillionsKeeper     *millionskeeper.Keeper
}

// InitSpecialKeepers Init the "special" keepers in the order of definition
func (app *App) InitSpecialKeepers(
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
) {
	// Acquire variables from app structure
	appCodec := app.appCodec
	baseApp := app.BaseApp
	cdc := app.cdc
	keys := app.keys
	tkeys := app.tkeys
	memKeys := app.memKeys
	govModuleAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// Init params keeper
	paramsKeeper := app.InitParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
	app.ParamsKeeper = &paramsKeeper

	// Set the BaseApp's parameter store
	consensusParamsKeeper := consensusparamkeeper.NewKeeper(appCodec, keys[upgradetypes.StoreKey], govModuleAddr)
	app.ConsensusParamsKeeper = &consensusParamsKeeper
	baseApp.SetParamStore(app.ConsensusParamsKeeper)

	// Add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	app.ScopedICAControllerKeeper = app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	app.ScopedICAHostKeeper = app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.ScopedICACallbacksKeeper = app.CapabilityKeeper.ScopeToModule(icacallbackstypes.ModuleName)
	app.ScopedTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	app.ScopedMillionsKeeper = app.CapabilityKeeper.ScopeToModule(millionstypes.ModuleName)
	app.CapabilityKeeper.Seal()

	// Init the crisis keeper
	crisisKeeper := crisiskeeper.NewKeeper(appCodec, keys[crisistypes.StoreKey],
		invCheckPeriod, app.BankKeeper, authtypes.FeeCollectorName, govModuleAddr,
	)
	app.CrisisKeeper = crisisKeeper

	// Initialize the upgrade keeper
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		govModuleAddr,
	)

	// Register the upgrade handlers
	app.registerUpgradeHandlers()
}

// InitNormalKeepers Initialize the normal keepers
func (app *App) InitNormalKeepers() {
	appCodec := app.appCodec
	baseApp := app.BaseApp
	keys := app.keys
	govModuleAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// Initialize the account keeper
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		AccountAddressPrefix,
		govModuleAddr,
	)
	app.AccountKeeper = &accountKeeper

	// Initialize the bank keeper
	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.BlockedModuleAccountAddrs(),
		govModuleAddr,
	)
	app.BankKeeper = &bankKeeper

	// Initialize the authz keeper
	authzKeeper := authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey],
		appCodec,
		baseApp.MsgServiceRouter(),
		accountKeeper,
	)
	app.AuthzKeeper = &authzKeeper

	// Initialize the staking keeper
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		govModuleAddr,
	)

	// Initialize the distribution keeper
	distrKeeper := distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		govModuleAddr,
	)
	app.DistrKeeper = &distrKeeper

	// Initialize the slashing keeper
	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec,
		app.LegacyAmino(),
		keys[slashingtypes.StoreKey],
		app.StakingKeeper,
		govModuleAddr,
	)
	app.SlashingKeeper = &slashingKeeper

	// Initialize the IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		app.GetSubspace(ibchost.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		app.ScopedIBCKeeper,
	)

	// Initialize the IBC Fee keeper
	ibcFeeKeeper := ibcfeekeeper.NewKeeper(
		appCodec,
		app.keys[ibcfeetypes.StoreKey],
		app.IBCKeeper.ChannelKeeper, // may be replaced with IBC middleware
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
	)
	app.IBCFeeKeeper = &ibcFeeKeeper

	// Initialize the IBC transfer keeper
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.ScopedTransferKeeper,
	)
	app.TransferKeeper = &transferKeeper
	app.transferModule = transfer.NewAppModule(*app.TransferKeeper)

	// Initialize the ICA controller keeper
	icaControllerKeeper := icacontrollerkeeper.NewKeeper(
		appCodec, keys[ICAControllerCustomStoreKey],
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCFeeKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.ScopedICAControllerKeeper,
		app.MsgServiceRouter(),
	)
	app.ICAControllerKeeper = &icaControllerKeeper

	// Initialize the ICA host keeper
	icaHostKeeper := icahostkeeper.NewKeeper(
		appCodec, keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCFeeKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.ScopedICAHostKeeper,
		app.MsgServiceRouter(),
	)
	app.ICAHostKeeper = &icaHostKeeper

	// Initialize the evidence keeper
	app.EvidenceKeeper = evidencekeeper.NewKeeper(
		appCodec,
		keys[evidencetypes.StoreKey],
		app.StakingKeeper,
		app.SlashingKeeper,
	)

	// Initialize the mint keeper
	mintKeeper := mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		govModuleAddr,
	)
	app.MintKeeper = &mintKeeper

	// Initialize the fee grant keeper
	feeGrantKeeper := feegrantkeeper.NewKeeper(
		appCodec,
		keys[feegrant.StoreKey],
		app.AccountKeeper,
	)
	app.FeeGrantKeeper = &feeGrantKeeper

	// Initialize our ICA callbacks keeper
	app.ICACallbacksKeeper = icacallbackskeeper.NewKeeper(
		appCodec,
		keys[icacallbackstypes.StoreKey],
		keys[icacallbackstypes.MemStoreKey],
		app.GetSubspace(icacallbackstypes.ModuleName),
		app.ScopedICACallbacksKeeper,
		*app.IBCKeeper,
		*app.ICAControllerKeeper,
	)

	// Initialize our ICQueries keeper
	app.ICQueriesKeeper = icquerieskeeper.NewKeeper(appCodec, keys[icqueriestypes.StoreKey], app.IBCKeeper)

	// Initialize our custom beam keeper
	app.BeamKeeper = beamkeeper.NewKeeper(
		appCodec,
		keys[beamtypes.StoreKey],
		keys[beamtypes.MemStoreKey],
		*app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
	)

	// Initialize our custom airdrop keeper
	app.AirdropKeeper = airdropkeeper.NewKeeper(
		appCodec,
		keys[airdroptypes.StoreKey],
		keys[airdroptypes.MemStoreKey],
		*app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
	)

	// Initialize our custom dfract keeper
	app.DFractKeeper = dfractkeeper.NewKeeper(
		appCodec,
		keys[dfracttypes.StoreKey],
		keys[dfracttypes.StoreKey],
		app.GetSubspace(dfracttypes.ModuleName),
		*app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
	)

	// Initialize our custom millions keeper
	app.MillionsKeeper = millionskeeper.NewKeeper(
		appCodec,
		keys[millionstypes.StoreKey],
		app.GetSubspace(millionstypes.ModuleName),
		app.ScopedMillionsKeeper,
		*app.AccountKeeper,
		*app.IBCKeeper,
		*app.TransferKeeper,
		*app.ICAControllerKeeper,
		*app.ICACallbacksKeeper,
		*app.ICQueriesKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		app.StakingKeeper,
	)

	// First stack contains
	// - Transfer IBC Module
	// - Millions IBC middleware
	// - base app
	transferIBCModule := transfer.NewIBCModule(*app.TransferKeeper)
	transferStack := millions.NewIBCMiddleware(*app.MillionsKeeper, transferIBCModule)

	// Second stack contains
	// - ICAHost IBC Module
	// - base app
	icaHostIBCModule := icahost.NewIBCModule(*app.ICAHostKeeper)

	// Third stack contains
	// - Millions IBC Module
	// - ICAController IBC Middleware
	// - base app
	millionsIBCModule := millions.NewIBCModule(*app.MillionsKeeper)
	icaControllerIBCModule := icacontroller.NewIBCMiddleware(millionsIBCModule, *app.ICAControllerKeeper)

	// Register our ICACallbacks handlers
	err := app.ICACallbacksKeeper.SetICACallbackHandler(millionstypes.ModuleName, app.MillionsKeeper.ICACallbackHandler())
	if err != nil {
		panic(err)
	}

	// Register our ICQueries handlers
	err = app.ICQueriesKeeper.SetCallbackHandler(millionstypes.ModuleName, app.MillionsKeeper.ICQCallbackHandler())
	if err != nil {
		panic(err)
	}

	// Create static IBC router, then seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerIBCModule).
		AddRoute(millionstypes.ModuleName, icaControllerIBCModule).
		AddRoute(ibctransfertypes.ModuleName, transferStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	// Initialize the governance router
	govRouter := govtypesv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*app.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(ibchost.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(dfracttypes.RouterKey, dfract.NewDFractProposalHandler(*app.DFractKeeper)).
		AddRoute(millionstypes.RouterKey, millions.NewMillionsProposalHandler(*app.MillionsKeeper))

	// Initialize the governance keeper
	app.GovKeeper = govkeeper.NewKeeper(
		appCodec,
		keys[govtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		baseApp.MsgServiceRouter(),
		govtypes.DefaultConfig(),
		govModuleAddr,
	)
	app.GovKeeper.SetLegacyRouter(govRouter)
}

func (app *App) SetupHooks() {
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks(), app.AirdropKeeper.Hooks()),
	)

	app.GovKeeper.SetHooks(govtypes.NewMultiGovHooks(
		govtypes.NewMultiGovHooks(app.AirdropKeeper.Hooks()),
	))
}

// InitParamsKeeper init params keeper and its subspaces
func (app *App) InitParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	// Base modules
	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypesv1.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(feegrant.ModuleName)
	paramsKeeper.Subspace(authz.ModuleName)

	// Custom modules
	paramsKeeper.Subspace(icacallbackstypes.ModuleName)
	paramsKeeper.Subspace(icqueriestypes.ModuleName)
	paramsKeeper.Subspace(beamtypes.ModuleName)
	paramsKeeper.Subspace(dfracttypes.ModuleName)
	paramsKeeper.Subspace(millionstypes.ModuleName)

	return paramsKeeper
}
