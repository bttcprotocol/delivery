package util

import (
	"sync"

	"github.com/maticnetwork/heimdall/params/subspace"

	sdk "github.com/cosmos/cosmos-sdk/types"
	featuremanagerTypes "github.com/maticnetwork/heimdall/featuremanager/types"
)

var (
	featureConfigManagerOnce        sync.Once
	featureConfigureManagerInstance *FeatureConfig
)

type FeatureConfig struct {
	featureSpace subspace.Subspace
}

func GetFeatureConfig() *FeatureConfig {
	featureConfigManagerOnce.Do(func() {
		featureConfigureManagerInstance = new(FeatureConfig)
	})

	return featureConfigureManagerInstance
}

func InitFeatureConfig(paramSpace subspace.Subspace) {
	featureManager := GetFeatureConfig()
	featureManager.featureSpace = paramSpace
}

// GetFeature is used to get target feature config.
func (m *FeatureConfig) GetFeature(ctx sdk.Context, feature string) (config featuremanagerTypes.PlainFeatureData) {
	defer func() {
		// if featureSpace cannot get FeatureParams, it will recover from here.
		if err := recover(); err != nil {
			config = featuremanagerTypes.PlainFeatureData{
				IsOpen:     false,
				IntConf:    make(map[string]int),
				StringConf: make(map[string]string),
			}
		}
	}()

	params := featuremanagerTypes.FeatureParams{}
	m.featureSpace.GetParamSet(ctx, &params)

	rawData, ok := params.FeatureParamMap[feature]
	if !ok {
		config = featuremanagerTypes.PlainFeatureData{}
	} else {
		config = rawData.Plainify()
	}

	return
}

// GetSupportedFeature is used to get whether the target feature is consensus supported feature.
func (m *FeatureConfig) GetSupportedFeature(ctx sdk.Context, feature string) (isSupported bool) {
	defer func() {
		// if featureSpace cannot get FeatureSupport, it will recover from here.
		if err := recover(); err != nil {
			isSupported = false
		}
	}()

	params := featuremanagerTypes.FeatureSupport{}
	m.featureSpace.GetParamSet(ctx, &params)

	res, ok := params.FeatureSupportMap[feature]
	if !ok {
		isSupported = false
	} else {
		isSupported = res
	}

	return
}
