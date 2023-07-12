package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/maticnetwork/heimdall/featuremanager/types"
	"github.com/maticnetwork/heimdall/version"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the transaction commands for this module.
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                types.ModuleName,
		Short:              "Querying commands for the featuremanager module",
		DisableFlagParsing: true,
		RunE:               client.ValidateCmd,
	}
	txCmd.AddCommand(
		client.GetCommands(
			GetTargetFeature(cdc),
			GetAllFeatures(cdc),
			GetAllSupportedFeatures(cdc),
		)...,
	)

	return txCmd
}

// GetTargetFeature implements the target-feature query command.
func GetTargetFeature(cdc *codec.Codec) *cobra.Command {
	//nolint: exhaustivestruct
	return &cobra.Command{
		Use:   "target-feature [target feature]",
		Short: "show the current featuremanager target feature information in proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Query values set as chain manager parameters.
Example:
$ %s query featuremanager target-features
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			if len(args) != 1 {
				return fmt.Errorf(`expect an arg to figure out feature name:
				%s query featuremanager target-feature <target feature>`, version.ClientName)
			}

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryTargetFeature)

			qp, _ := cliCtx.Codec.MarshalJSON(types.QueryTargetFeatureParam{
				TargetFeature: args[0],
			})

			data, _, err := cliCtx.QueryWithData(route, qp)
			if err != nil {
				return err
			}

			var params types.PlainFeatureData
			if err = json.Unmarshal(data, &params); err != nil {
				return err
			}

			return cliCtx.PrintOutput(params)
		},
	}
}

// GetAllFeatures implements the all-features query command.
func GetAllFeatures(cdc *codec.Codec) *cobra.Command {
	//nolint: exhaustivestruct
	return &cobra.Command{
		Use:   "all-features",
		Args:  cobra.NoArgs,
		Short: "show the current featuremanager parameters information in proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Query values set as feature manager parameters.
Example:
$ %s query featuremanager all-features
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryFeatureMap)

			data, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.FeatureParams
			if err = json.Unmarshal(data, &params); err != nil {
				return err
			}

			return cliCtx.PrintOutput(params)
		},
	}
}

// GetAllSupportedFeatures implements the supported-features query command.
func GetAllSupportedFeatures(cdc *codec.Codec) *cobra.Command {
	//nolint: exhaustivestruct
	return &cobra.Command{
		Use:   "supported-features",
		Args:  cobra.NoArgs,
		Short: "show the current featuremanager supported features",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Query all supported features in feature manager.
Example:
$ %s query featuremanager supported-features
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySupportedFeatures)

			data, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.FeatureSupport
			if err = json.Unmarshal(data, &params); err != nil {
				return err
			}

			return cliCtx.PrintOutput(params)
		},
	}
}
