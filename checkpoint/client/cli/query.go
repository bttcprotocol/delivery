package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/maticnetwork/heimdall/checkpoint/types"
	hmClient "github.com/maticnetwork/heimdall/client"
	"github.com/maticnetwork/heimdall/version"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	// Group supply queries under a subcommand
	supplyQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the checkpoint module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       hmClient.ValidateCmd,
	}

	// supply query command
	supplyQueryCmd.AddCommand(
		client.GetCommands(
			GetQueryParams(cdc),
			GetCheckpointBuffer(cdc),
			GetLastNoACK(cdc),
			GetHeaderFromIndex(cdc),
			GetCheckpointCount(cdc),
			GetQueryActivateHeight(cdc),
		)...,
	)

	return supplyQueryCmd
}

// GetQueryParams implements the params query command.
func GetQueryParams(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "show the current checkpoint parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as checkpoint parameters.

Example:
$ %s query checkpoint params
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)
			bz, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.Params
			if err := json.Unmarshal(bz, &params); err != nil {
				return nil
			}
			return cliCtx.PrintOutput(params)
		},
	}
}

// GetCheckpointBuffer get checkpoint present in buffer
func GetCheckpointBuffer(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkpoint-buffer",
		Short: "show checkpoint present in buffer",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			rootChain := viper.GetString(FlagRootChain)
			// get query params
			queryParams, err := cliCtx.Codec.MarshalJSON(types.NewQueryCheckpointParams(0, rootChain))
			if err != nil {
				return errors.New("rootChain Error :" + rootChain)
			}
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryCheckpointBuffer), queryParams)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return errors.New("No checkpoint buffer found")
			}

			fmt.Printf(string(res))
			return nil
		},
	}
	cmd.Flags().String(FlagRootChain, "", "--root-chain=<root-chain>")
	if err := cmd.MarkFlagRequired(FlagRootChain); err != nil {
		logger.Error("GetCheckpointBuffer | MarkFlagRequired | FlagHeaderNumber", "Error", err)
	}
	return cmd
}

// GetLastNoACK get last no ack time
func GetLastNoACK(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last-noack",
		Short: "get last no ack received time",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryLastNoAck), nil)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return errors.New("No last-no-ack count found")
			}

			var lastNoAck uint64
			if err := json.Unmarshal(res, &lastNoAck); err != nil {
				return err
			}

			fmt.Printf("LastNoACK received at %v", time.Unix(int64(lastNoAck), 0))
			return nil
		},
	}

	return cmd
}

// GetHeaderFromIndex get checkpoint given header index
func GetHeaderFromIndex(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "header",
		Short: "get checkpoint (header) from index",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			headerNumber := viper.GetUint64(FlagHeaderNumber)

			// get query params
			queryParams, err := cliCtx.Codec.MarshalJSON(types.NewQueryCheckpointParams(headerNumber, ""))
			if err != nil {
				return err
			}

			// fetch checkpoint
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryCheckpoint), queryParams)
			if err != nil {
				return err
			}

			fmt.Printf(string(res))
			return nil
		},
	}

	cmd.Flags().Uint64(FlagHeaderNumber, 0, "--header=<header-number>")
	if err := cmd.MarkFlagRequired(FlagHeaderNumber); err != nil {
		logger.Error("GetHeaderFromIndex | MarkFlagRequired | FlagHeaderNumber", "Error", err)
	}

	return cmd
}

// GetCheckpointCount get number of checkpoint received count
func GetCheckpointCount(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkpoint-count",
		Short: "get checkpoint counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryEpoch), nil)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return errors.New("No ack count found")
			}

			var ackCount uint64
			if err := cliCtx.Codec.UnmarshalJSON(res, &ackCount); err != nil {
				return err
			}

			fmt.Printf("Total number of checkpoint so far : %v", ackCount)
			return nil
		},
	}

	return cmd
}

func GetQueryActivateHeight(cdc *codec.Codec) *cobra.Command {
	//nolint: exhaustivestruct
	return &cobra.Command{
		Use:   "activate-height",
		Short: "show the target chain activate height",
		Long:  strings.TrimSpace("show the target chain activate height"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			if len(args) != 1 {
				return fmt.Errorf(`expect an arg to figure out chain name:
%s query checkpoint activate-height <target chain>`, version.ClientName)
			}

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryCheckpointActivation)

			queryParams, err := cliCtx.Codec.MarshalJSON(types.NewQueryCheckpointParams(0, args[0]))
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(route, queryParams)
			if err != nil {
				return err
			}

			var activationHeight uint64
			if err = json.Unmarshal(res, &activationHeight); err != nil {
				return err
			}
			// nolint
			fmt.Printf("Target chain activate height, root=%v, activateHeight=%v\n", args[0], activationHeight)

			return nil
		},
	}
}
