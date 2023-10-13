package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/maticnetwork/heimdall/helper"
	tendermintLogger "github.com/tendermint/tendermint/libs/log"
)

const (
	bridgeDBFlag         = "bridge-db"
	bttcChainIDFlag      = "bttc-chain-id"
	rootChainTypeFlag    = "root-chain-type"
	startListenBlockFlag = "start-listen-block"
)

var logger = helper.Logger.With("module", "bridge/cmd/")

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bridge",
	Short: "Delivery bridge deamon",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// initialize tendermint viper config
		InitTendermintViperConfig(cmd)
	},
}

func BridgeCommands(viperInstance *viper.Viper, loggerInstance tendermintLogger.Logger, caller string) *cobra.Command {
	DecorateWithBridgeRootFlags(rootCmd, viperInstance, loggerInstance, caller)
	return rootCmd
}

func DecorateWithBridgeRootFlags(
	cmd *cobra.Command, viperInstance *viper.Viper, loggerInstance tendermintLogger.Logger, caller string,
) {
	cmd.PersistentFlags().StringP(helper.TendermintNodeFlag, "n", helper.DefaultTendermintNode, "Node to connect to")

	if err := viperInstance.BindPFlag(helper.TendermintNodeFlag,
		cmd.PersistentFlags().Lookup(helper.TendermintNodeFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, helper.TendermintNodeFlag), "Error", err)
	}

	cmd.PersistentFlags().String(helper.HomeFlag, helper.DefaultNodeHome, "directory for config and data")

	if err := viperInstance.BindPFlag(helper.HomeFlag, cmd.PersistentFlags().Lookup(helper.HomeFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, helper.HomeFlag), "Error", err)
	}

	// bridge storage db
	cmd.PersistentFlags().String(
		bridgeDBFlag,
		"",
		"Bridge db path (default <home>/bridge/storage)",
	)

	if err := viperInstance.BindPFlag(bridgeDBFlag, cmd.PersistentFlags().Lookup(bridgeDBFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, bridgeDBFlag), "Error", err)
	}

	// bridge chain id
	cmd.PersistentFlags().String(
		bttcChainIDFlag,
		helper.DefaultBttcChainID,
		"Bttc chain id",
	)

	if err := viperInstance.BindPFlag(bttcChainIDFlag, cmd.PersistentFlags().Lookup(bttcChainIDFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, bttcChainIDFlag), "Error", err)
	}
}

// InitTendermintViperConfig sets global viper configuration needed to heimdall
func InitTendermintViperConfig(cmd *cobra.Command) {
	// set appropriate bridge db.
	AdjustBridgeDBValue(cmd, viper.GetViper())

	// start heimdall config.
	helper.InitDeliveryConfig("")
}

// function is called to set appropriate bridge db path.
func AdjustBridgeDBValue(cmd *cobra.Command, v *viper.Viper) {
	tendermintNode, _ := cmd.Flags().GetString(helper.TendermintNodeFlag)
	homeValue, _ := cmd.Flags().GetString(helper.HomeFlag)
	withDeliveryConfigFlag, _ := cmd.Flags().GetString(helper.WithDeliveryConfigFlag)
	bridgeDBValue, _ := cmd.Flags().GetString(bridgeDBFlag)
	bttcChainIDValue, _ := cmd.Flags().GetString(bttcChainIDFlag)
	rootChainTypeValue, _ := cmd.Flags().GetString(rootChainTypeFlag)
	startListenBlockValue, _ := cmd.Flags().GetInt64(startListenBlockFlag)

	// bridge-db directory (default storage)
	if bridgeDBValue == "" {
		bridgeDBValue = filepath.Join(homeValue, "bridge", "storage")
	}

	// set to viper
	viper.Set(helper.TendermintNodeFlag, tendermintNode)
	viper.Set(helper.HomeFlag, homeValue)
	viper.Set(helper.WithDeliveryConfigFlag, withDeliveryConfigFlag)
	viper.Set(bridgeDBFlag, bridgeDBValue)
	viper.Set(bttcChainIDFlag, bttcChainIDValue)
	viper.Set(rootChainTypeFlag, rootChainTypeValue)
	viper.Set(startListenBlockFlag, startListenBlockValue)
}
