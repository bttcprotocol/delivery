package featuremanager

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/featuremanager/types"
	"github.com/maticnetwork/heimdall/gov"
	"github.com/maticnetwork/heimdall/params/subspace"
	"github.com/tendermint/tendermint/libs/log"
)

// Keeper stores all related data.
type Keeper struct {
	cdc *codec.Codec
	// codespace
	codespace sdk.CodespaceType

	// gov keeper to handle subprocess of related proposal
	govKeeper gov.Keeper

	featureTable map[string]bool

	// param space
	paramSpace subspace.Subspace
}

// NewKeeper create new keeper.
func NewKeeper(
	cdc *codec.Codec,
	paramSpace subspace.Subspace,
	govKeeper gov.Keeper,
	codespace sdk.CodespaceType,
) Keeper {
	// create keeper
	keeper := Keeper{
		cdc:          cdc,
		paramSpace:   paramSpace.WithKeyTable(types.ParamKeyTable()),
		govKeeper:    govKeeper,
		featureTable: make(map[string]bool),
		codespace:    codespace,
	}
	keeper.RegistreFeature()

	return keeper
}

// Codespace returns the codespace.
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

func (k Keeper) addFeature(feature string) {
	k.featureTable[feature] = true
}

// RegistreFeature keeps all adden features to vote for change.
func (k Keeper) RegistreFeature() {
	// all new type of features should be registered here.
	k.addFeature("feature-x")
}

func (k Keeper) HasFeature(feature string) bool {
	_, ok := k.featureTable[feature]

	return ok
}

// -----------------------------------------------------------------------------
// Params

// SetParams sets the featuremanager module's parameters.
func (k Keeper) SeFeatureParams(ctx sdk.Context, params types.FeatureParams) error {
	k.paramSpace.SetParamSet(ctx, &params)

	return nil
}

// GetParams gets the featuremanager module's parameters.
func (k Keeper) GetFeatureParams(ctx sdk.Context) (params types.FeatureParams) {
	defer func() {
		if err := recover(); err != nil {
			params = types.DefaultFeatureParams()
		}
	}()
	k.paramSpace.GetParamSet(ctx, &params)

	return
}
