package main

import (
	"os"

	"github.com/maticnetwork/heimdall/bridge/cmd"
	"github.com/maticnetwork/heimdall/helper"
)

func main() {
	rootCmd := cmd.BridgeCommands()
	logger := helper.Logger.With("module", "bridge/cmd/")

	if err := rootCmd.Execute(); err != nil {
		logger.Error("bridge cmd failed", "Error", err)
		os.Exit(1)
	}
}
