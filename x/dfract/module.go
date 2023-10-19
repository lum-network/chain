package dfract

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lum-network/chain/x/dfract/migrations"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/lum-network/chain/x/dfract/client/cli"
	"github.com/lum-network/chain/x/dfract/client/rest"
	"github.com/lum-network/chain/x/dfract/keeper"
	"github.com/lum-network/chain/x/dfract/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

type AppModuleBasic struct {
	cdc codec.Codec
}

func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// Name returns the capability module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

func (a AppModuleBasic) DefaultGenesis(jsonCodec codec.JSONCodec) json.RawMessage {
	return jsonCodec.MustMarshalJSON(types.DefaultGenesisState())
}

func (a AppModuleBasic) ValidateGenesis(jsonCodec codec.JSONCodec, config client.TxEncodingConfig, message json.RawMessage) error {
	var data types.GenesisState
	if err := jsonCodec.UnmarshalJSON(message, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return nil
}

func (a AppModuleBasic) RegisterRESTRoutes(context client.Context, router *mux.Router) {
	rest.RegisterRoutes(context, router)
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	_ = types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(types.StoreKey)
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

func (a AppModule) Name() string {
	return a.AppModuleBasic.Name()
}

func (a AppModule) InitGenesis(context sdk.Context, jsonCodec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	jsonCodec.MustUnmarshalJSON(message, &genState)

	InitGenesis(context, a.keeper, genState)
	return []abci.ValidatorUpdate{}
}

func (a AppModule) ExportGenesis(context sdk.Context, jsonCodec codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(context, a.keeper)
	return jsonCodec.MustMarshalJSON(genState)
}

func (a AppModule) RegisterInvariants(registry sdk.InvariantRegistry) {}

func (a AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

func (a AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(a.keeper))
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(a.keeper))

	// Register the migrations
	migrator := migrations.NewMigrator(a.keeper)
	if err := cfg.RegisterMigration(types.ModuleName, 1, migrator.Migrate1To2); err != nil {
		panic(err)
	}
}

// BeginBlock executes all ABCI BeginBlock logic respective to the capability module.
func (a AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock executes all ABCI EndBlock logic respective to the capability module. It
// returns no validator updates.
func (a AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	EndBlocker(ctx, a.keeper)
	return []abci.ValidatorUpdate{}
}

func (a AppModule) ConsensusVersion() uint64 {
	return types.ModuleVersion
}
