package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/maticnetwork/heimdall/chainmanager/types"
	"github.com/maticnetwork/heimdall/version"
)

// GetQueryCmd returns the transaction commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the chainmanager module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		client.GetCommands(
			GetQueryParams(cdc),
			GetQueryProposalChainParam(cdc),
		)...,
	)
	return txCmd
}

// GetQueryParams implements the params query command.
func GetQueryParams(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "show the current chainmanager parameters information of target chain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as chain manager parameters.
Available args are: tron / bsc / eth.
Example:
$ %s query chainmanager params bsc
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			if len(args) != 1 {
				return fmt.Errorf(`expect an arg to figure out chain name:
%s query chainmanager params <target chain>`, version.ClientName)
			}

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryNewChainParam)

			qp, _ := cliCtx.Codec.MarshalJSON(types.QueryChainParams{
				RootChain: args[0],
			})

			res, _, err := cliCtx.QueryWithData(route, qp)
			if err != nil {
				return err
			}

			var params types.Params
			if err = json.Unmarshal(res, &params); err != nil {
				return err
			}

			return cliCtx.PrintOutput(params)
		},
	}
}

func GetQueryProposalChainParam(cdc *codec.Codec) *cobra.Command {
	//nolint: exhaustivestruct
	return &cobra.Command{
		Use:   "proposal-params",
		Args:  cobra.NoArgs,
		Short: "show the current chainmanager parameters information in proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Query values set as chain manager parameters.
Example:
$ %s query chainmanager params
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryProposalChainParamMap)

			bz, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.ParamsWithMultiChains
			if err = json.Unmarshal(bz, &params); err != nil {
				return err
			}

			return cliCtx.PrintOutput(params)
		},
	}
}
