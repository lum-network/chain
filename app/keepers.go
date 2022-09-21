package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
	icahost "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v3/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	airdropkeeper "github.com/lum-network/chain/x/airdrop/keeper"
	airdroptypes "github.com/lum-network/chain/x/airdrop/types"
	beamkeeper "github.com/lum-network/chain/x/beam/keeper"
	beamtypes "github.com/lum-network/chain/x/beam/types"
	"github.com/lum-network/chain/x/dfract"
	dfractkeeper "github.com/lum-network/chain/x/dfract/keeper"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
)

type AppKeepers struct {
	// Special Keepers
	ParamsKeeper     *paramskeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper

	// Normal Keepers
	AccountKeeper  *authkeeper.AccountKeeper
	BankKeeper     *bankkeeper.BaseKeeper
	AuthzKeeper    *authzkeeper.Keeper
	StakingKeeper  *stakingkeeper.Keeper
	DistrKeeper    *distrkeeper.Keeper
	SlashingKeeper *slashingkeeper.Keeper
	IBCKeeper      *ibckeeper.Keeper
	ICAHostKeeper  *icahostkeeper.Keeper
	TransferKeeper *ibctransferkeeper.Keeper
	EvidenceKeeper *evidencekeeper.Keeper
	MintKeeper     *mintkeeper.Keeper
	GovKeeper      *govkeeper.Keeper
	FeeGrantKeeper *feegrantkeeper.Keeper

	// Custom Keepers
	AirdropKeeper *airdropkeeper.Keeper
	BeamKeeper    *beamkeeper.Keeper
	DFractKeeper  *dfractkeeper.Keeper
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

	// Init params keeper
	paramsKeeper := app.InitParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
	app.ParamsKeeper = &paramsKeeper

	// Set the BaseApp's parameter store
	baseApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// Add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	app.ScopedICAHostKeeper = app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.ScopedTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	app.CapabilityKeeper.Seal()

	// Init the crisis keeper
	crisisKeeper := crisiskeeper.NewKeeper(
		app.GetSubspace(crisistypes.ModuleName), invCheckPeriod, app.BankKeeper, authtypes.FeeCollectorName,
	)
	app.CrisisKeeper = &crisisKeeper

	// Initialize the upgrade keeper
	upgradeKeeper := upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath, baseApp)
	app.UpgradeKeeper = &upgradeKeeper

	// Register the upgrade handlers
	app.registerUpgradeHandlers()
}

// InitNormalKeepers Initialize the normal keepers
func (app *App) InitNormalKeepers() {
	appCodec := app.appCodec
	baseApp := app.BaseApp
	keys := app.keys

	// Initialize the account keeper
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec, keys[authtypes.StoreKey], app.GetSubspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)
	app.AccountKeeper = &accountKeeper

	// Initialize the bank keeper
	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.GetSubspace(banktypes.ModuleName),
		app.BlockedAddrs(),
	)
	app.BankKeeper = &bankKeeper

	// Initialize the authz keeper
	authzKeeper := authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey],
		appCodec,
		baseApp.MsgServiceRouter(),
	)
	app.AuthzKeeper = &authzKeeper

	// Initialize the staking keeper
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(stakingtypes.ModuleName),
	)
	app.StakingKeeper = &stakingKeeper

	// Initialize the distribution keeper
	distrKeeper := distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey],
		app.GetSubspace(distrtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, authtypes.FeeCollectorName, app.BlockedAddrs(),
	)
	app.DistrKeeper = &distrKeeper

	// Initialize the slashing keeper
	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec, keys[slashingtypes.StoreKey], app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName),
	)
	app.SlashingKeeper = &slashingKeeper

	// Initialize the IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, keys[ibchost.StoreKey], app.GetSubspace(ibchost.ModuleName), app.StakingKeeper, app.UpgradeKeeper, app.ScopedIBCKeeper,
	)

	// Initialize the IBC transfer keeper
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec, keys[ibctransfertypes.StoreKey], app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		app.AccountKeeper, app.BankKeeper, app.ScopedTransferKeeper,
	)
	app.TransferKeeper = &transferKeeper
	app.transferModule = transfer.NewAppModule(*app.TransferKeeper)
	transferIBCModule := transfer.NewIBCModule(*app.TransferKeeper)

	// Initialize the ICA host keeper
	icaHostKeeper := icahostkeeper.NewKeeper(
		appCodec, keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.ScopedICAHostKeeper,
		app.MsgServiceRouter(),
	)
	app.ICAHostKeeper = &icaHostKeeper
	icaHostIBCModule := icahost.NewIBCModule(*app.ICAHostKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).AddRoute(ibctransfertypes.ModuleName, transferIBCModule)
	app.IBCKeeper.SetRouter(ibcRouter)

	// Initialize the evidence keeper
	app.EvidenceKeeper = evidencekeeper.NewKeeper(appCodec, keys[evidencetypes.StoreKey], app.StakingKeeper, app.SlashingKeeper)

	// Initialize the mint keeper
	mintKeeper := mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey], app.GetSubspace(minttypes.ModuleName), app.StakingKeeper,
		app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName,
	)
	app.MintKeeper = &mintKeeper

	// Initialize the fee grant keeper
	feeGrantKeeper := feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)
	app.FeeGrantKeeper = &feeGrantKeeper

	// Initialize our custom beam keeper
	beamKeeper := *beamkeeper.NewKeeper(
		appCodec, keys[beamtypes.StoreKey], keys[beamtypes.MemStoreKey],
		*app.AccountKeeper, app.BankKeeper, *app.StakingKeeper,
	)
	app.BeamKeeper = &beamKeeper

	// Initialize our custom airdrop keeper
	airdropKeeper := airdropkeeper.NewKeeper(appCodec, keys[airdroptypes.StoreKey], keys[airdroptypes.MemStoreKey], *app.AccountKeeper, app.BankKeeper, *app.StakingKeeper, *app.DistrKeeper)
	app.AirdropKeeper = airdropKeeper

	// Initialize our custom dfract keeper
	dfractKeeper := dfractkeeper.NewKeeper(appCodec, keys[dfracttypes.StoreKey], keys[dfracttypes.StoreKey], *app.AccountKeeper, app.BankKeeper, *app.StakingKeeper)
	app.DFractKeeper = dfractKeeper

	// Initialize the governance router
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*app.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(*app.DistrKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(*app.UpgradeKeeper)).
		AddRoute(ibchost.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(dfracttypes.RouterKey, dfract.NewDFractProposalHandler(*app.DFractKeeper))

	// Initialize the governance keeper
	govKeeper := govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], app.GetSubspace(govtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, govRouter,
	)
	app.GovKeeper = &govKeeper
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
func (app *App) InitParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypes.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(feegrant.ModuleName)
	paramsKeeper.Subspace(authz.ModuleName)

	paramsKeeper.Subspace(beamtypes.ModuleName)
	paramsKeeper.Subspace(dfracttypes.ModuleName)

	return paramsKeeper
}
