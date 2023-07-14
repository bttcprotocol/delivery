package cli

import (
	"encoding/json"
	"io/ioutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/featuremanager/types"
	hmTypes "github.com/maticnetwork/heimdall/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

type (
	FeatureChangesJSON []FeatureChangeJSON

	// FeatureChangeJSON defines a parameter change used in JSON input. This
	// allows values to be specified in raw JSON instead of being string encoded.
	FeatureChangeJSON struct {
		Value json.RawMessage `json:"value" yaml:"value"`
	}

	// FeatureChangeProposalJSON defines a FeatureChangeProposal with a deposit used
	// to parse parameter change proposals from a JSON file.
	FeatureChangeProposalJSON struct {
		Title       string              `json:"title" yaml:"title"`
		Description string              `json:"description" yaml:"description"`
		Changes     FeatureChangesJSON  `json:"changes" yaml:"changes"`
		Deposit     sdk.Coins           `json:"deposit" yaml:"deposit"`
		Validator   hmTypes.ValidatorID `json:"validator" yaml:"validator"`
	}
)

func NewFeatureChangeJSON(value json.RawMessage) FeatureChangeJSON {
	return FeatureChangeJSON{value}
}

// ToParamChange converts a FeatureChangeJSON object to ParamChange.
func (fcj FeatureChangeJSON) ToParamChange() types.FeatureChange {
	return types.NewFeatureChange(string(fcj.Value))
}

// ToFeatureChanges converts a slice of FeatureChangeJSON objects to a slice of
// FeatureChange.
func (fcj FeatureChangesJSON) ToFeatureChanges() []types.FeatureChange {
	res := make([]types.FeatureChange, len(fcj))
	for i, pc := range fcj {
		res[i] = pc.ToParamChange()
	}

	return res
}

// ParseParamChangeProposalJSON reads and parses a ParamChangeProposalJSON from
// file.
func ParseFeatureChangeProposalJSON(cdc *codec.Codec, proposalFile string) (FeatureChangeProposalJSON, error) {
	proposal := FeatureChangeProposalJSON{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
