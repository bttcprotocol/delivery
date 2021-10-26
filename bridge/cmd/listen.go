package cmd

import (
	"math/big"
	"strconv"

	"github.com/maticnetwork/heimdall/bridge/setu/util"
	"github.com/maticnetwork/heimdall/helper"
	hmtypes "github.com/maticnetwork/heimdall/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ethLastRootBlockKey = "eth-last-block"  // eth storage key
	tronLastBlockKey    = "tron-last-block" // tron storage key
	bscLastBlockKey     = "bsc-last-block"  // bsc storage key
)

// resetCmd represents the start command
func CreateSetStartBLockCmd() *cobra.Command {
	var logger = helper.Logger.With("module", "bridge/cmd/")
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "set up bridge db data",
		Run: func(cmd *cobra.Command, args []string) {
			rootChainType := viper.GetString(rootChainTypeFlag)
			lastBlockKey := ""
			switch rootChainType {
			case hmtypes.RootChainTypeEth:
				lastBlockKey = ethLastRootBlockKey
			case hmtypes.RootChainTypeBsc:
				lastBlockKey = bscLastBlockKey
			case hmtypes.RootChainTypeTron:
				lastBlockKey = tronLastBlockKey
			default:
				logger.Error("-root-chain-type value should be in [eth,bsc,tron]")
				return
			}
			startListenBlock := viper.GetInt64(startListenBlockFlag)
			startBlock := big.NewInt(startListenBlock)
			bridgeDB := util.GetBridgeDBInstance(viper.GetString(util.BridgeDBFlag))
			if err := bridgeDB.Put([]byte(lastBlockKey), []byte(startBlock.String()), nil); err != nil {
				logger.Error("bridgeDB.Put", "Error", err)
				return
			}
			logger.Info("set bridge latest listen block", "set rootChain ", rootChainType, " start listen block", startBlock.String())
		},
	}
	cmd.Flags().String(rootChainTypeFlag, "", "--rootChainType=<root-chain-type>")

	cmd.Flags().Int64(startListenBlockFlag, 0, "--startListenBlock=<start-listen-block>")

	if err := cmd.MarkFlagRequired(rootChainTypeFlag); err != nil {
		logger.Error("setUpBridgeStartListenBlock |rootChainTypeFlag", "Error", err)
	}

	if err := cmd.MarkFlagRequired(startListenBlockFlag); err != nil {
		logger.Error("setUpBridgeStartListenBlock |startListenBlockFlag", "Error", err)
	}
	return cmd

}

func CreateListStartListenBlockCmd() *cobra.Command {
	var logger = helper.Logger.With("module", "bridge/cmd/")
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list bridge start block",
		Run: func(cmd *cobra.Command, args []string) {
			rootChainType := viper.GetString(rootChainTypeFlag)
			bridgeDB := util.GetBridgeDBInstance(viper.GetString(util.BridgeDBFlag))
			lastBlockKey := ""
			switch rootChainType {
			case hmtypes.RootChainTypeEth:
				lastBlockKey = ethLastRootBlockKey
			case hmtypes.RootChainTypeBsc:
				lastBlockKey = bscLastBlockKey
			case hmtypes.RootChainTypeTron:
				lastBlockKey = tronLastBlockKey
			default:
				logger.Error("-root-chain-type value should be in [eth,bsc,tron]")
				return
			}
			lastBlockBytes, err := bridgeDB.Get([]byte(lastBlockKey), nil)
			if err != nil {
				logger.Error("Error while fetching last block bytes from storage", "error", err)
				return
			}
			if result, err := strconv.ParseUint(string(lastBlockBytes), 10, 64); err == nil {
				logger.Info("list bridge latest listen block  ", "list rootChain", rootChainType, " startListBlock", result)
			}
		},
	}
	cmd.Flags().String(rootChainTypeFlag, "", "--rootChainType=<root-chain-type>")

	if err := cmd.MarkFlagRequired(rootChainTypeFlag); err != nil {
		logger.Error("setUpBridgeStartListenBlock |rootChainTypeFlag", "Error", err)
	}
	return cmd

}

func init() {
	rootCmd.AddCommand(CreateSetStartBLockCmd())
	rootCmd.AddCommand(CreateListStartListenBlockCmd())
}
