package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/maticnetwork/heimdall/helper"
)

const (
	bridgeDBFlag         = "bridge-db"
	bttcChainIDFlag      = "bttc-chain-id"
	rootChainTypeFlag    = "root-chain-type"
	startListenBlockFlag = "start-listen-block"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bridge",
	Short: "Delivery bridge deamon",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// initialize tendermint viper config
		InitTendermintViperConfig(cmd)
	},
}

// BridgeCommands returns command for bridge service
func BridgeCommands() *cobra.Command {
	return rootCmd
}

// InitTendermintViperConfig sets global viper configuration needed to heimdall
func InitTendermintViperConfig(cmd *cobra.Command) {
	tendermintNode, _ := cmd.Flags().GetString(helper.NodeFlag)
	homeValue, _ := cmd.Flags().GetString(helper.HomeFlag)
	withDeliveryConfigValue, _ := cmd.Flags().GetString(helper.WithDeliveryConfigFlag)
	bridgeDBValue, _ := cmd.Flags().GetString(bridgeDBFlag)
	borChainIDValue, _ := cmd.Flags().GetString(bttcChainIDFlag)
	rootChainTypeValue, _ := cmd.Flags().GetString(rootChainTypeFlag)
	startListenBlockValue, _ := cmd.Flags().GetInt64(startListenBlockFlag)
	// bridge-db directory (default storage)
	if bridgeDBValue == "" {
		bridgeDBValue = filepath.Join(homeValue, "bridge", "storage")
	}

	// set to viper
	viper.Set(helper.NodeFlag, tendermintNode)
	viper.Set(helper.HomeFlag, homeValue)
	viper.Set(helper.WithDeliveryConfigFlag, withDeliveryConfigValue)
	viper.Set(bridgeDBFlag, bridgeDBValue)
	viper.Set(bttcChainIDFlag, borChainIDValue)
	viper.Set(rootChainTypeFlag, rootChainTypeValue)
	viper.Set(startListenBlockFlag, startListenBlockValue)
	// start heimdall config
	helper.InitDeliveryConfig("")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	var logger = helper.Logger.With("module", "bridge/cmd/")
	rootCmd.PersistentFlags().StringP(helper.NodeFlag, "n", "tcp://localhost:26657", "Node to connect to")
	rootCmd.PersistentFlags().String(helper.HomeFlag, os.ExpandEnv("$HOME/.deliveryd"), "directory for config and data")
	rootCmd.PersistentFlags().String(
		helper.WithDeliveryConfigFlag,
		"",
		"Delivery config file path (default <home>/config/delivery-config.toml)",
	)
	// bridge storage db
	rootCmd.PersistentFlags().String(
		bridgeDBFlag,
		"",
		"Bridge db path (default <home>/bridge/storage)",
	)
	// bridge chain id
	rootCmd.PersistentFlags().String(
		bttcChainIDFlag,
		helper.DefaultBttcChainID,
		"Bor chain id",
	)

	// bind all flags with viper
	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		logger.Error("init | BindPFlag | rootCmd.Flags", "Error", err)
	}
}
