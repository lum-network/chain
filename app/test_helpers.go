package app

import (
	"encoding/json"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"time"

	"github.com/stretchr/testify/suite"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/cosmos/ibc-go/v7/testing/simapp"
)

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(o string) interface{} {
	return nil
}

var (
	LumChainID     = "LUM-NETWORK"
	HostChainID    = "GAIA-DEVNET"
	TestIcaVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ibctesting.FirstConnectionID,
		HostConnectionId:       ibctesting.FirstConnectionID,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
)

var DefaultSupplyWeight int64 = 100000001000000

func SetupForTesting(isCheckTx bool) *App {
	// Create our mocked validator
	privVal := mock.NewPV()
	pubKey, _ := privVal.GetPubKey()

	// Create validator set with 1 validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// Add a genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(CoinBondDenom, sdk.NewInt(100000000000000))),
	}

	// Initialize the application
	db := dbm.NewMemDB()
	app := New("simapp", log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, MakeEncodingConfig(), EmptyAppOptions{})
	if !isCheckTx {
		genesisState := NewDefaultGenesisState()
		genesisState = GenesisStateWithValSet(app, genesisState, valSet, []authtypes.GenesisAccount{acc}, balance)

		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simtestutil.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

// SetupForIBCTesting Configure the testing application and cast it as an IBC testing application
func SetupForIBCTesting() (ibctesting.TestingApp, map[string]json.RawMessage) {
	app := SetupForTesting(false)
	return app, NewDefaultGenesisState()
}

func GenesisStateWithValSet(app *App, genesisState GenesisState, valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) GenesisState {
	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	// Initialize each validator in the validator set
	for _, val := range valSet.Validators {
		pk, _ := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		pkAny, _ := codectypes.NewAnyWithValue(pk)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdk.OneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), sdk.OneDec()))
	}

	// Initialize our delegations
	stakingparams := stakingtypes.DefaultParams()
	stakingparams.BondDenom = CoinBondDenom
	stakingGenesis := stakingtypes.NewGenesisState(stakingparams, validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		totalSupply = totalSupply.Add(sdk.NewCoin(CoinBondDenom, bondAmt))
	}

	// Add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(CoinBondDenom, bondAmt)},
	})

	// Update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	return genesisState
}

type TestPackage struct {
	suite.Suite
	App     *App
	HostApp *simapp.SimApp

	IbcEnabled   bool
	Coordinator  *ibctesting.Coordinator
	LumChain     *ibctesting.TestChain
	HostChain    *ibctesting.TestChain
	TransferPath *ibctesting.Path

	QueryHelper  *baseapp.QueryServiceTestHelper
	TestAccs     []sdk.AccAddress
	ICAAddresses map[string]string
	Ctx          sdk.Context
}

func (p *TestPackage) Setup() {
	p.App = SetupForTesting(false)
	p.Ctx = p.App.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: LumChainID}).WithBlockTime(time.Now().UTC())
	p.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: p.App.GRPCQueryRouter(),
		Ctx:             p.Ctx,
	}
	p.IbcEnabled = false
	p.ICAAddresses = make(map[string]string)
}

// SetupIBC Initialize the IBC configuration for test environment
func (p *TestPackage) SetupIBC() {
	p.Coordinator = ibctesting.NewCoordinator(p.T(), 0)

	// Initialize a stride testing app by casting a LumApp -> TestingApp
	ibctesting.DefaultTestingAppInit = SetupForIBCTesting
	p.LumChain = ibctesting.NewTestChain(p.T(), p.Coordinator, LumChainID)

	// Initialize a host testing app using SimApp -> TestingApp
	ibctesting.DefaultTestingAppInit = ibctesting.SetupTestingApp
	p.HostChain = ibctesting.NewTestChain(p.T(), p.Coordinator, HostChainID)

	// Update coordinator
	p.Coordinator.Chains = map[string]*ibctesting.TestChain{
		LumChainID:  p.LumChain,
		HostChainID: p.HostChain,
	}
	p.IbcEnabled = true
}

// NewTransferPath Creates a transfer channel between two chains
func NewTransferPath(chainA *ibctesting.TestChain, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = ibctransfertypes.Version
	path.EndpointB.ChannelConfig.Version = ibctransfertypes.Version
	return path
}

// NewIcaPath Creates an ICA channel between two chains
func NewIcaPath(chainA *ibctesting.TestChain, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointB.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = TestIcaVersion
	path.EndpointB.ChannelConfig.Version = TestIcaVersion
	return path
}

// CopyConnectionAndClientToPath
// In ibctesting, there's no easy way to create a new channel on an existing connection
// To get around this, this helper function will copy the client/connection info from an existing channel
// We use this when creating ICA channels, because we want to reuse the same connections/clients from the transfer channel
func CopyConnectionAndClientToPath(path *ibctesting.Path, pathToCopy *ibctesting.Path) *ibctesting.Path {
	path.EndpointA.ClientID = pathToCopy.EndpointA.ClientID
	path.EndpointB.ClientID = pathToCopy.EndpointB.ClientID
	path.EndpointA.ConnectionID = pathToCopy.EndpointA.ConnectionID
	path.EndpointB.ConnectionID = pathToCopy.EndpointB.ConnectionID
	path.EndpointA.ClientConfig = pathToCopy.EndpointA.ClientConfig
	path.EndpointB.ClientConfig = pathToCopy.EndpointB.ClientConfig
	path.EndpointA.ConnectionConfig = pathToCopy.EndpointA.ConnectionConfig
	path.EndpointB.ConnectionConfig = pathToCopy.EndpointB.ConnectionConfig
	return path
}

// CreateTransferChannel Creates clients, connections, and a transfer channel between Lum and a host chain
func (p *TestPackage) CreateTransferChannel(hostChainID string) {
	// If we have yet to create the host chain, do that here
	if !p.IbcEnabled {
		p.SetupIBC()
	}

	// Make sure the chain IDs are the same
	p.Require().Equal(p.HostChain.ChainID, hostChainID, "The testing app has already been initialized with a different chainID (%s)", p.HostChain.ChainID)

	// Create clients, connections, and a transfer channel
	p.TransferPath = NewTransferPath(p.LumChain, p.HostChain)
	p.Coordinator.Setup(p.TransferPath)

	// Replace stride and host apps with those from TestingApp
	p.App = p.LumChain.App.(*App)
	p.HostApp = p.HostChain.GetSimApp()
	p.Ctx = p.LumChain.GetContext()

	// Finally confirm the channel was configured properly
	p.Require().Equal(ibctesting.FirstClientID, p.TransferPath.EndpointA.ClientID, "lum clientID")
	p.Require().Equal(ibctesting.FirstConnectionID, p.TransferPath.EndpointA.ConnectionID, "lum connectionID")
	p.Require().Equal(ibctesting.FirstChannelID, p.TransferPath.EndpointA.ChannelID, "lum transfer channelID")

	p.Require().Equal(ibctesting.FirstClientID, p.TransferPath.EndpointB.ClientID, "host clientID")
	p.Require().Equal(ibctesting.FirstConnectionID, p.TransferPath.EndpointB.ConnectionID, "host connectionID")
	p.Require().Equal(ibctesting.FirstChannelID, p.TransferPath.EndpointB.ChannelID, "host transfer channelID")
}

// CreateICAChannel Creates an ICA channel through ibctesting, also creates a transfer channel if it hasn't been done yet
func (p *TestPackage) CreateICAChannel(owner string) string {
	// If we have yet to create a client/connection (through creating a transfer channel), do that here
	_, transferChannelExists := p.App.IBCKeeper.ChannelKeeper.GetChannel(p.Ctx, ibctesting.TransferPort, ibctesting.FirstChannelID)
	if !transferChannelExists {
		p.CreateTransferChannel(HostChainID)
	}

	// Create ICA Path and then copy over the client and connection from the transfer path
	icaPath := NewIcaPath(p.LumChain, p.HostChain)
	icaPath = CopyConnectionAndClientToPath(icaPath, p.TransferPath)

	// Register the ICA and complete the handshake
	p.RegisterInterchainAccount(icaPath.EndpointA, owner)

	// Try to open the channel
	err := icaPath.EndpointB.ChanOpenTry()
	p.Require().NoError(err, "ChanOpenTry error")

	// Try to acknowledge channel opening
	err = icaPath.EndpointA.ChanOpenAck()
	p.Require().NoError(err, "ChanOpenAck error")

	// Try to confirm opening
	err = icaPath.EndpointB.ChanOpenConfirm()
	p.Require().NoError(err, "ChanOpenConfirm error")

	// Acquire the context from controller chain
	p.Ctx = p.LumChain.GetContext()

	// Confirm the ICA channel was created properly
	portID := icaPath.EndpointA.ChannelConfig.PortID
	channelID := icaPath.EndpointA.ChannelID
	_, found := p.App.IBCKeeper.ChannelKeeper.GetChannel(p.Ctx, portID, channelID)
	p.Require().True(found, "Channel not found after creation, PortID: %s, ChannelID: %s", portID, channelID)

	// Store the account address
	icaAddress, found := p.App.ICAControllerKeeper.GetInterchainAccountAddress(p.Ctx, ibctesting.FirstConnectionID, portID)
	p.Require().True(found, "can't get ICA address")
	p.ICAAddresses[owner] = icaAddress

	// Finally set the active channel
	p.App.ICAControllerKeeper.SetActiveChannelID(p.Ctx, ibctesting.FirstConnectionID, portID, channelID)
	return channelID
}

// RegisterInterchainAccount Register's a new ICA account on the next channel available, this function assumes a connection already exists
func (p *TestPackage) RegisterInterchainAccount(endpoint *ibctesting.Endpoint, owner string) {
	// Get the port ID from the owner name (i.e. "icacontroller-{owner}")
	portID, err := icatypes.NewControllerPortID(owner)
	p.Require().NoError(err, "owner to portID error")

	// Get the next channel available and register the ICA
	channelSequence := p.App.IBCKeeper.ChannelKeeper.GetNextChannelSequence(p.Ctx)

	err = p.App.ICAControllerKeeper.RegisterInterchainAccount(p.Ctx, endpoint.ConnectionID, owner, TestIcaVersion)
	p.Require().NoError(err, "register interchain account error")

	// Commit the state
	endpoint.Chain.NextBlock()

	// Update the endpoint object to the newly created port + channel
	endpoint.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	endpoint.ChannelConfig.PortID = portID
}

func (p *TestPackage) FundModuleAccount(moduleName string, amount sdk.Coin) {
	coins := sdk.NewCoins(amount)
	err := p.App.BankKeeper.MintCoins(p.Ctx, minttypes.ModuleName, coins)
	p.Require().NoError(err)
	err = p.App.BankKeeper.SendCoinsFromModuleToModule(p.Ctx, minttypes.ModuleName, moduleName, coins)
	p.Require().NoError(err)
}

func (p *TestPackage) FundAccount(acc sdk.AccAddress, amount sdk.Coin) {
	coins := sdk.NewCoins(amount)
	err := p.App.BankKeeper.MintCoins(p.Ctx, minttypes.ModuleName, coins)
	p.Require().NoError(err)
	err = p.App.BankKeeper.SendCoinsFromModuleToAccount(p.Ctx, minttypes.ModuleName, acc, coins)
	p.Require().NoError(err)
}
