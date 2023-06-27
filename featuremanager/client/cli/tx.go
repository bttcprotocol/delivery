package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/featuremanager/types"
	"github.com/maticnetwork/heimdall/helper"
	hmTypes "github.com/maticnetwork/heimdall/types"
	"github.com/maticnetwork/heimdall/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logger = helper.Logger.With("module", "featuremanager/client/cli")

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd(cdc *codec.Codec, pcmds []*cobra.Command) *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:                types.ModuleName,
		Short:              "Featuremanager transactions subcommands",
		DisableFlagParsing: true,
		RunE:               client.ValidateCmd,
	}

	cmdSubmitProp := GetCmdSubmitProposal(cdc)
	for _, pcmd := range pcmds {
		cmdSubmitProp.AddCommand(client.PostCommands(pcmd)[0])
	}

	govTxCmd.AddCommand(client.PostCommands(
		cmdSubmitProp,
	)...)

	return govTxCmd
}

// GetCmdSubmitProposal implements a command handler for submitting a feature change
// proposal transaction.
func GetCmdSubmitProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feature-change [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a feature change proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a feature change proposal along with an initial deposit.
The proposal details must be supplied via a JSON file. For values that contains
objects, only non-empty fields will be updated.

IMPORTANT: Currently feature change changes are evaluated but not validated, so it is
very important that any "value" change is valid (ie. correct type and within bounds)
for its respective parameter, eg. "MaxValidators" should be an integer and not a decimal.

Proper vetting of a feature change proposal should prevent this from happening
(no deposits should occur during the governance process), but it should be noted
regardless.

Example:
$ %s tx featuremanager feature-change <path/to/proposal.json> --validator-id <validator_id> --chain-id <chain_id>

Where proposal.json contains:

{
  "title": "Feature-X Change",
  "description": "Add feature-x",
  "changes": [
    {
      "value": {
		"feature_param_map": {
			"feature_x" : {
				"is_open": false,
				int_conf: {},
				string_conf: {}
			}
		}
	  }
    }
  ],
  "deposit": [
    {
      "denom": "btt",
      "amount": "1000000000000000000" 
    }
  ]
}
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			proposal, err := ParseFeatureChangeProposalJSON(cdc, args[0])
			if err != nil {
				return err
			}

			validatorID := viper.GetUint64(FlagValidatorID)
			if validatorID == 0 {
				return fmt.Errorf("valid validator ID required")
			}

			from := helper.GetFromAddress(cliCtx)
			content := types.NewFeatureChangeProposal(proposal.Title, proposal.Description, proposal.Changes.ToFeatureChanges())

			// create submit proposal
			msg := types.NewMsgSubmitProposal(content, proposal.Deposit, from, hmTypes.NewValidatorID(validatorID))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return helper.BroadcastMsgsWithCLI(cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Int(FlagValidatorID, 0, "--validator-id=<validator ID here>")

	if err := cmd.MarkFlagRequired(FlagValidatorID); err != nil {
		logger.Error("GetCmdSubmitProposal | MarkFlagRequired | FlagValidatorID", "Error", err)
	}

	return cmd
}
