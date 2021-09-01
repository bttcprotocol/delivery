package cmd

import (
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	"github.com/maticnetwork/heimdall/helper"
	hmtypes "github.com/maticnetwork/heimdall/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math/big"
	"strconv"
)

const (
	ethLastRootBlockKey = "rootchain-last-block" // eth storage key
	tronLastBlockKey    = "tron-last-block"      // tron storage key
)

// resetCmd represents the start command
func CreateSetStartBLockCmd() *cobra.Command {
	var logger = helper.Logger.With("module", "bridge/cmd/")
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "set up bridge db data",
		Run: func(cmd *cobra.Command, args []string) {
			rootChainType := viper.GetString(rootChainTypeFlag)
			if rootChainType != hmtypes.RootChainTypeTron && rootChainType != hmtypes.RootChainTypeEth {
				logger.Error("-root-chain-type value should be either eth or tron")
				return
			}

			startListenBlock := viper.GetInt64(startListenBlockFlag)
			startBlock := big.NewInt(startListenBlock)
			bridgeDB := util.GetBridgeDBInstance(viper.GetString(util.BridgeDBFlag))
			if rootChainType == hmtypes.RootChainTypeEth {
				if err := bridgeDB.Put([]byte(ethLastRootBlockKey), []byte(startBlock.String()), nil); err != nil {
					logger.Error("bridgeDB.Put", "Error", err)
					return
				}
			} else {
				if err := bridgeDB.Put([]byte(tronLastBlockKey), []byte(startBlock.String()), nil); err != nil {
					logger.Error("bridgeDB.Put", "Error", err)
					return
				}
			}
			logger.Info("set bridge latest listen block","set rootChain ",rootChainType," start listen block",startBlock.String())
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
			if rootChainType != hmtypes.RootChainTypeTron && rootChainType != hmtypes.RootChainTypeEth {
				logger.Error("-root-chain-type should be either eth or tron")
				return
			}
			bridgeDB := util.GetBridgeDBInstance(viper.GetString(util.BridgeDBFlag))
			var DBkey string
			if rootChainType == hmtypes.RootChainTypeEth {
				DBkey = ethLastRootBlockKey
			} else {
				DBkey = tronLastBlockKey
			}
			lastBlockBytes, err := bridgeDB.Get([]byte(DBkey), nil)
			if err != nil {
				logger.Error("Error while fetching last block bytes from storage", "error", err)
				return
			}
			if result, err := strconv.ParseUint(string(lastBlockBytes), 10, 64); err == nil {
				logger.Info("list bridge latest listen block  ","list rootChain",rootChainType," startListBlock",result)
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
