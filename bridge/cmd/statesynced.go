package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	hmtypes "github.com/maticnetwork/heimdall/types"
)

func tokenMapShowAll(eventProcessor *util.TokenMapProcessor) error {
	mmap, err := eventProcessor.GetTokenMap()
	if err != nil {
		return err
	}

	for k, v := range mmap {
		helper.Logger.Info(
			fmt.Sprintf("----------- %s chain [%d items] -----------", k, len(v)))

		for _, item := range v {
			msg, err := json.Marshal(item)
			if err != nil {
				return err
			}

			helper.Logger.Info(string(msg))
		}

		helper.Logger.Info("----------- ------------------- -----------")
	}

	return nil
}

func tokenMapShowOne(eventProcessor *util.TokenMapProcessor, chainType string) error {
	mlist, err := eventProcessor.GetTokenMapByRootType(chainType)
	if err != nil {
		return err
	}

	helper.Logger.Info(
		fmt.Sprintf("----------- %s chain [%d items] -----------",
			chainType, len(mlist)))

	for _, item := range mlist {
		msg, err := json.Marshal(item)
		if err != nil {
			return err
		}

		helper.Logger.Info(string(msg))
	}

	helper.Logger.Info("----------- ------------------- -----------")

	return nil
}

func tokenMapShowHash(eventProcessor *util.TokenMapProcessor) error {
	hashBytes, err := eventProcessor.GetHash()
	if err != nil {
		return err
	}

	if hashBytes == nil {
		helper.Logger.Info("Hash value not found")

		return nil
	}

	helper.Logger.Info(fmt.Sprintf("Hash Value: %s", hex.EncodeToString(hashBytes)))

	return nil
}

func queryTokenMapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokenmap-query [root-chain-type|all|hash|end-block|last-event-id]",
		Short: "query token items for one chain or all",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainType := args[0]

			bridgeDB := util.GetBridgeDBInstance(viper.GetString(util.BridgeDBFlag))
			if bridgeDB == nil {
				return fmt.Errorf("query parameter error, open db failed")
			}

			eventProcessor := util.NewTokenMapProcessor(context.NewCLIContext(), bridgeDB)

			switch chainType {
			case "all":
				return tokenMapShowAll(eventProcessor)

			case hmtypes.RootChainTypeEth, hmtypes.RootChainTypeTron, hmtypes.RootChainTypeBsc:
				return tokenMapShowOne(eventProcessor, chainType)

			case "hash":
				return tokenMapShowHash(eventProcessor)

			case "end-block":
				helper.Logger.Info(fmt.Sprintf("End Block: %d",
					eventProcessor.TokenMapCheckedEndBlock))

			case "last-event-id":
				helper.Logger.Info(fmt.Sprintf("Last Event ID: %d",
					eventProcessor.TokenMapLastEventID))

			default:
				return fmt.Errorf("query parameter error, should be [root-chain-type|all|hash|end-block|last-event-id]")
			}

			return nil
		},
	}

	return cmd
}

func setTokenMapEventIDCmd() *cobra.Command {
	logger := helper.Logger

	cmd := &cobra.Command{
		Use:   "tokenmap-set-event-id [event-id]",
		Short: "set event id for token map processor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventIDStr := args[0]

			eventID, err := strconv.ParseUint(eventIDStr, util.DigitBase, util.DigitBitSize)
			if err != nil {
				return err
			}

			bridgeDB := util.GetBridgeDBInstance(viper.GetString(util.BridgeDBFlag))
			if bridgeDB == nil {
				return fmt.Errorf("query parameter error, open db failed")
			}

			eventProcessor := util.NewTokenMapProcessor(context.NewCLIContext(), bridgeDB)
			eventIDNow := eventProcessor.TokenMapLastEventID

			if err := eventProcessor.UpdateTokenMapLastEventID(eventID); err != nil {
				return err
			}
			logger.Info(
				fmt.Sprintf("Modify last event ID from %d to %d", eventIDNow, eventID))

			return nil
		},
	}

	return cmd
}

func setTokenMapEndBlockCmd() *cobra.Command {
	logger := helper.Logger
	cmd := &cobra.Command{
		Use:   "tokenmap-set-end-block [block-id]",
		Short: "set end block for token map processor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			blockID, err := strconv.ParseInt(args[0], util.DigitBase, util.DigitBitSize)
			if err != nil {
				return err
			}

			bridgeDB := util.GetBridgeDBInstance(viper.GetString(util.BridgeDBFlag))
			if bridgeDB == nil {
				return fmt.Errorf("query parameter error, open db failed")
			}

			eventProcessor := util.NewTokenMapProcessor(context.NewCLIContext(), bridgeDB)
			blockIDNow := eventProcessor.TokenMapCheckedEndBlock

			if err := eventProcessor.UpdateTokenMapCheckedEndBlock(blockID); err != nil {
				return err
			}

			logger.Info(fmt.Sprintf("Modify end block from %d to %d", blockIDNow, blockID))

			return nil
		},
	}

	return cmd
}

func removeTokenMapCmd() *cobra.Command {
	logger := helper.Logger
	cmd := &cobra.Command{
		Use:   "tokenmap-remove",
		Short: "remove token map to regenerate it",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			bridgeDB := util.GetBridgeDBInstance(viper.GetString(util.BridgeDBFlag))
			if bridgeDB == nil {
				return fmt.Errorf("query parameter error, open db failed")
			}

			eventProcessor := util.NewTokenMapProcessor(context.NewCLIContext(), bridgeDB)
			eventIDNow := eventProcessor.TokenMapLastEventID
			endBlockNow := eventProcessor.TokenMapCheckedEndBlock
			hashNow := ""

			hashBytes, err := eventProcessor.GetHash()
			if err != nil {
				return err
			}
			if hashBytes != nil {
				hashNow = hex.EncodeToString(hashBytes)
			}

			if err := eventProcessor.ResetParameters(); err != nil {
				return err
			}

			logger.Info(
				fmt.Sprintf("\nRemove map, lastest status:\nLastEventID:\t%d\nEndBlock:\t%d\nHash:\t%s",
					eventIDNow, endBlockNow, hashNow))

			return nil
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(queryTokenMapCmd())
	rootCmd.AddCommand(setTokenMapEventIDCmd())
	rootCmd.AddCommand(setTokenMapEndBlockCmd())
	rootCmd.AddCommand(removeTokenMapCmd())
}
