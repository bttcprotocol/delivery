package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypes "github.com/maticnetwork/heimdall/gov/types"
)

const (
	// ProposalTypeChange defines the type for a FeatureChangeProposal.
	ProposalTypeChange = "FeatureChange"
)

// Assert FeatureChangeProposal implements govtypes.Content at compile-time.
var _ govTypes.Content = FeatureChangeProposal{}

func init() {
	govTypes.RegisterProposalType(ProposalTypeChange)
	govTypes.RegisterProposalTypeCodec(FeatureChangeProposal{}, "heimdall/FeatureChangeProposal")
}

// FeatureChangeProposal defines a proposal which contains multiple parameter changes.
type FeatureChangeProposal struct {
	Title       string          `json:"title" yaml:"title"`
	Description string          `json:"description" yaml:"description"`
	Changes     []FeatureChange `json:"changes" yaml:"changes"`
}

func NewFeatureChangeProposal(title, description string, changes []FeatureChange) FeatureChangeProposal {
	return FeatureChangeProposal{title, description, changes}
}

// GetTitle returns the title of a parameter change proposal.
func (pcp FeatureChangeProposal) GetTitle() string { return pcp.Title }

// GetDescription returns the description of a parameter change proposal.
func (pcp FeatureChangeProposal) GetDescription() string { return pcp.Description }

// GetDescription returns the routing key of a parameter change proposal.
func (pcp FeatureChangeProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a parameter change proposal.
func (pcp FeatureChangeProposal) ProposalType() string { return ProposalTypeChange }

// ValidateBasic validates the parameter change proposal.
func (pcp FeatureChangeProposal) ValidateBasic() sdk.Error {
	err := govTypes.ValidateAbstract(DefaultCodespace, pcp)
	if err != nil {
		return err
	}

	return ValidateChanges(pcp.Changes)
}

// String implements the Stringer interface.
func (pcp FeatureChangeProposal) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`Parameter Change Proposal:
  Title:       %s
  Description: %s
  Changes:
`, pcp.Title, pcp.Description))

	for _, pc := range pcp.Changes {
		builder.WriteString(fmt.Sprintf(`    Param Change:
      Value:    %X
`, pc.Value))
	}

	return builder.String()
}

// ParamChange defines a parameter change.
type FeatureChange struct {
	Value string `json:"value" yaml:"value"`
}

func NewFeatureChange(value string) FeatureChange {
	return FeatureChange{value}
}

// String implements the Stringer interface.
func (pc FeatureChange) String() string {
	return fmt.Sprintf(`Param Change:
  Value:    %X
`, pc.Value)
}

// ValidateChange performs basic validation checks over a set of ParamChange. It
// returns an error if any ParamChange is invalid.
func ValidateChanges(changes []FeatureChange) sdk.Error {
	if len(changes) == 0 {
		return ErrEmptyChanges(DefaultCodespace)
	}

	for _, pc := range changes {
		if len(pc.Value) == 0 {
			return ErrEmptyValue(DefaultCodespace)
		}
	}

	return nil
}
