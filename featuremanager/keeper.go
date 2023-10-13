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
	keeper.RegisterFeature()

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

// RegisterFeature keeps all adden features to vote for change.
func (k Keeper) RegisterFeature() {
	// all new type of features should be registered here.
	k.addFeature("feature-x")
	k.addFeature(types.DynamicCheckpoint)
	k.addFeature(types.SupportMapMarshaling)
}

func (k Keeper) HasFeature(feature string) bool {
	_, ok := k.featureTable[feature]

	return ok
}

// -----------------------------------------------------------------------------
// Params

// SetParams sets the featuremanager module's parameters.
func (k Keeper) SetFeatureParams(ctx sdk.Context, params types.FeatureParams) error {
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

// -----------------------------------------------------------------------------
// Supported Features.

// AddSupportedFeature add the featuremanager module's supported features.
func (k Keeper) AddSupportedFeature(ctx sdk.Context, feature string) error {
	supportedMap := types.FeatureSupport{
		FeatureSupportMap: map[string]bool{
			feature: true,
		},
	}

	supportedMapRawData, err := k.cdc.MarshalJSON(supportedMap)
	if err != nil {
		return err
	}

	err = k.paramSpace.Update(ctx, types.KeySupportFeature, supportedMapRawData)

	return err
}

// GetSupportedFeature gets the featuremanager module's all supported features.
func (k Keeper) GetSupportedFeature(ctx sdk.Context) (params types.FeatureSupport) {
	defer func() {
		if err := recover(); err != nil {
			params = types.DefaultFeatureSupport()
		}
	}()
	k.paramSpace.GetParamSet(ctx, &params)

	return
}
