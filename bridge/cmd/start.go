package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/common"
	httpClient "github.com/tendermint/tendermint/rpc/client"

	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/maticnetwork/heimdall/app"
	"github.com/maticnetwork/heimdall/bridge/setu/broadcaster"
	"github.com/maticnetwork/heimdall/bridge/setu/listener"
	"github.com/maticnetwork/heimdall/bridge/setu/processor"
	"github.com/maticnetwork/heimdall/bridge/setu/queue"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	"github.com/maticnetwork/heimdall/helper"
)

const (
	waitDuration = 1 * time.Minute
)

func StartBridge(shutdownCtx context.Context, isStandAlone bool) error {
	// create codec
	cdc := app.MakeCodec()
	// queue connector & http client
	_queueConnector := queue.NewQueueConnector(helper.GetConfig().AmqpURL)
	_queueConnector.StartWorker()

	_txBroadcaster := broadcaster.NewTxBroadcaster(cdc)
	_httpClient := httpClient.NewHTTP(helper.GetConfig().TendermintRPCUrl, "/websocket")

	// selected services to start
	services := []common.Service{}
	services = append(services,
		listener.NewListenerService(cdc, _queueConnector, _httpClient),
		processor.NewProcessorService(cdc, _queueConnector, _httpClient, _txBroadcaster),
	)

	var waitGroup errgroup.Group

	if isStandAlone {
		// go routine to catch signal
		catchSignal := make(chan os.Signal, 1)
		signal.Notify(catchSignal, os.Interrupt, syscall.SIGTERM)

		cancelCtx, cancel := context.WithCancel(shutdownCtx)
		shutdownCtx = cancelCtx

		waitGroup.Go(func() error {
			// sig is a ^C, handle it
			for range catchSignal {
				// stop processes
				logger.Info("Received stop signal - Stopping all services")
				for _, service := range services {
					if err := service.Stop(); err != nil {
						logger.Error("GetStartCmd | service.Stop", "Error", err)
					}
				}

				// stop http client
				if err := _httpClient.Stop(); err != nil {
					logger.Error("GetStartCmd | _httpClient.Stop", "Error", err)
				}

				// stop db instance
				util.CloseBridgeDBInstance()

				cancel()
			}

			return nil
		})
	} else {
		// if not standalone, wait for shutdown context to stop services
		waitGroup.Go(func() error {
			<-shutdownCtx.Done()

			logger.Info("Received stop signal - Stopping all delivery bridge services")

			for _, service := range services {
				srv := service
				if srv.IsRunning() {
					if err := srv.Stop(); err != nil {
						logger.Error("GetStartCmd | service.Stop", "Error", err)

						return err
					}
				}
			}

			// stop http client
			if err := _httpClient.Stop(); err != nil {
				logger.Error("GetStartCmd | _httpClient.Stop", "Error", err)

				return err
			}

			// stop db instance
			util.CloseBridgeDBInstance()

			return nil
		})
	}

	// Start http client
	err := _httpClient.Start()
	if err != nil {
		logger.Error("Error connecting to server: %v", err)

		return err
	}

	// cli context
	cliCtx := cliContext.NewCLIContext().WithCodec(cdc)
	cliCtx.BroadcastMode = client.BroadcastAsync
	cliCtx.TrustNode = true

	// start bridge service only when node fully synced
	loop := true
	for loop {
		select {
		case <-shutdownCtx.Done():
			return nil

		case <-time.After(waitDuration):
			if !util.IsCatchingUp(cliCtx) {
				logger.Info("Node up to date, starting bridge services")

				loop = false
			} else {
				logger.Info("Waiting for heimdall to be synced")
			}
		}
	}

	// start all services
	for _, service := range services {
		srv := service

		waitGroup.Go(func() error {
			if err := srv.Start(); err != nil {
				logger.Error("GetStartCmd | serv.Start", "Error", err)

				return err
			}

			<-srv.Quit()

			return nil
		})
	}

	if err := waitGroup.Wait(); err != nil {
		logger.Error("Bridge stopped", "err", err)

		return err
	}

	if isStandAlone {
		os.Exit(1)
	}

	return nil
}

// GetStartCmd returns the start command to start bridge.
func GetStartCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start bridge server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = StartBridge(context.Background(), true)
		}}

	// log level
	startCmd.Flags().String(helper.LogLevel, "info", "Log level for bridge")

	if err := viper.BindPFlag(helper.LogLevel, startCmd.Flags().Lookup(helper.LogLevel)); err != nil {
		logger.Error("GetStartCmd | BindPFlag | logLevel", "Error", err)
	}

	startCmd.Flags().Bool("all", false, "start all bridge services")

	if err := viper.BindPFlag("all", startCmd.Flags().Lookup("all")); err != nil {
		logger.Error("GetStartCmd | BindPFlag | all", "Error", err)
	}

	startCmd.Flags().StringSlice("only", []string{}, "comma separated bridge services to start")

	if err := viper.BindPFlag("only", startCmd.Flags().Lookup("only")); err != nil {
		logger.Error("GetStartCmd | BindPFlag | only", "Error", err)
	}

	return startCmd
}

func init() {
	rootCmd.AddCommand(GetStartCmd())
}
