package types

import (
	"fmt"

	"github.com/maticnetwork/heimdall/params/subspace"
)

// Parameter keys.
var (
	KeyFeatureParams  = []byte("FeatureParams")
	KeySupportFeature = []byte("SupportFeature")
)

// nolint
type FeatureData struct {
	IsOpen *bool `json:"is_open" yaml:"is_open"`

	// extra data
	IntConf    map[string]int    `json:"int_conf" yaml:"int_conf"`
	StringConf map[string]string `json:"string_conf" yaml:"string_conf"`
}

// nolint
type PlainFeatureData struct {
	IsOpen bool `json:"is_open" yaml:"is_open"`

	// extra data
	IntConf    map[string]int    `json:"int_conf" yaml:"int_conf"`
	StringConf map[string]string `json:"string_conf" yaml:"string_conf"`
}

func (fd FeatureData) Plainify() (pd PlainFeatureData) {
	if fd.IsOpen != nil {
		pd.IsOpen = *fd.IsOpen
	}

	pd.IntConf = fd.IntConf
	pd.StringConf = fd.StringConf

	return
}

func (pfd PlainFeatureData) String() string {
	var ret string

	ret += fmt.Sprintf(`
	IsOpen: 					%v,
	IntConf: 					%v,
	StringConf:					%v
	`,
		pfd.IsOpen, pfd.IntConf, pfd.StringConf)

	return ret
}

// FeatureParams indicates feature config.
// nolint
type FeatureParams struct {
	FeatureParamMap map[string]FeatureData `json:"feature_param_map" yaml:"feature_param_map"`
}

// DefaultParams returns a default set of parameters.
func DefaultFeatureParams() FeatureParams {
	return FeatureParams{
		FeatureParamMap: make(map[string]FeatureData),
	}
}

func (fp *FeatureParams) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{Key: KeyFeatureParams, Value: fp},
	}
}

func (fp FeatureParams) String() string {
	var ret string

	for key, val := range fp.FeatureParamMap {
		pVal := val.Plainify()

		ret += fmt.Sprintf(`
		[feature]: 				%s
		IsOpen: 				%v,
		IntConf: 				%v,
		StringConf:				%v,				
		`,
			key, pVal.IsOpen, pVal.IntConf, pVal.StringConf)
	}

	return ret
}

// FeatureSupport indicates current chain supported features via sidechannel.
// nolint
type FeatureSupport struct {
	FeatureSupportMap map[string]bool `json:"feature_support_map" yaml:"feature_support_map"`
}

// DefaultFeatureSupport returns a default set of feature supported.
func DefaultFeatureSupport() FeatureSupport {
	return FeatureSupport{
		FeatureSupportMap: make(map[string]bool),
	}
}

func (fs *FeatureSupport) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{Key: KeySupportFeature, Value: fs},
	}
}

func (fs FeatureSupport) String() string {
	var ret string

	for key, val := range fs.FeatureSupportMap {
		ret += fmt.Sprintf(`
		[feature]: 				%s
		IsSupported: 				%v,		
		`,
			key, val)
	}

	return ret
}

// ParamKeyTable for auth module.
func ParamKeyTable() subspace.KeyTable {
	//nolint: exhaustivestruct
	return subspace.NewKeyTable().
		RegisterParamSet(&FeatureParams{}).
		RegisterParamSet(&FeatureSupport{})
}
