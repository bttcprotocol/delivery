package featuremanager

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"

	"github.com/maticnetwork/heimdall/featuremanager/client/cli"
	"github.com/maticnetwork/heimdall/featuremanager/client/rest"
	"github.com/maticnetwork/heimdall/featuremanager/types"
	hmTypes "github.com/maticnetwork/heimdall/types"
	hmModule "github.com/maticnetwork/heimdall/types/module"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule             = AppModule{}
	_ module.AppModuleBasic        = AppModuleBasic{}
	_ hmModule.HeimdallModuleBasic = AppModule{}
)

// AppModuleBasic defines the basic application module.
type AppModuleBasic struct{}

// Name returns the auth module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterCodec registers the auth module's types for the given codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis default genesis state.
func (AppModuleBasic) DefaultGenesis() json.RawMessage { return nil }

// ValidateGenesis module validate genesis.
func (AppModuleBasic) ValidateGenesis(_ json.RawMessage) error { return nil }

// VerifyGenesis module.
func (AppModuleBasic) VerifyGenesis(bz map[string]json.RawMessage) error { return nil }

// RegisterRESTRoutes register rest routes.
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr)
}

// GetTxCmd get the root tx command of this module.
func (a AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc, nil)
}

// GetQueryCmd get the root query command of this module.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(cdc)
}

// ____________________________________________________________________________
// AppModule implements an application module for the checkpoint module.
type AppModule struct {
	AppModuleBasic

	keeper Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// InitGenesis performs genesis initialization.
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// Name returns the auth module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants performs a no-op.
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the auth module.
func (AppModule) Route() string {
	return types.RouterKey
}

// QuerierRoute returns the auth module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// NewHandler returns an sdk.Handler for the module.
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// NewQuerierHandler returns the auth module sdk.Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// ExportGenesis returns the exported genesis state as raw bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return nil
}

// BeginBlock returns the begin blocker for the auth module.
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock returns the end blocker for the auth module. It returns no validator
// updates.
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

//
// Side module
//

// NewSideTxHandler side tx handler.
func (am AppModule) NewSideTxHandler() hmTypes.SideTxHandler {
	return NewSideTxHandler(am.keeper)
}

// NewPostTxHandler side tx handler.
func (am AppModule) NewPostTxHandler() hmTypes.PostTxHandler {
	return NewPostTxHandler(am.keeper)
}
