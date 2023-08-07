package types

// query endpoints supported by the chain-manager Querier.
const (
	QueryTargetFeature     = "target-feature"
	QueryFeatureMap        = "feature-map"
	QuerySupportedFeatures = "supported-features"
)

// QueryChainParams defines the params for querying accounts.
type QueryTargetFeatureParam struct {
	TargetFeature string
}

// NewQueryChainParams creates a new instance of NewQueryChainParams.
func NewQueryTargeFeatureParams(targetFeature string) QueryTargetFeatureParam {
	return QueryTargetFeatureParam{
		TargetFeature: targetFeature,
	}
}
