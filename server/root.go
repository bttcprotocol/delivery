package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmLog "github.com/tendermint/tendermint/libs/log"
	rpcserver "github.com/tendermint/tendermint/rpc/lib/server"

	"github.com/maticnetwork/heimdall/app"
	tx "github.com/maticnetwork/heimdall/client/tx"
	"github.com/maticnetwork/heimdall/helper"

	// unnamed import of statik for swagger UI support
	_ "github.com/maticnetwork/heimdall/server/statik"
)

const (
	healthCheckInterval = 10 * time.Millisecond
)

func StartRestServer(mainCtx context.Context, cdc *codec.Codec,
	registerRoutesFn func(*lcd.RestServer), restCh chan struct{},
) error {
	restServer := lcd.NewRestServer(cdc)
	registerRoutesFn(restServer)

	go restServerHealthCheck(restCh)

	logger := tmLog.NewTMLogger(tmLog.NewSyncWriter(os.Stdout)).With("module", "rest-server")

	cfg := rpcserver.DefaultConfig()
	cfg.MaxOpenConnections = viper.GetInt(client.FlagMaxOpenConnections)
	cfg.ReadTimeout = 0 * time.Second
	cfg.WriteTimeout = 0 * time.Second

	listener, err := rpcserver.Listen(viper.GetString(client.FlagListenAddr), cfg)
	if err != nil {
		logger.Error("Cannot start REST server", "Error", err)
		return err
	}

	logger.Info(
		fmt.Sprintf(
			"Starting application REST service (chain-id: %q)...",
			viper.GetString(flags.FlagChainID),
		),
	)

	// create HTTP server so we can perform graceful shutdown
	httpSrv := &http.Server{
		Handler:      restServer.Mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Graceful shutdown: when context is cancelled (Ctrl+C / SIGTERM),
	// give in-flight requests a window to finish before forcing close.
	go func() {
		<-mainCtx.Done()
		logger.Info("Shutting down REST server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpSrv.Shutdown(shutdownCtx); err != nil && err != http.ErrServerClosed {
			logger.Error("Error during REST server shutdown", "err", err)
		}
	}()

	// Serve will return http.ErrServerClosed when Shutdown is called.
	if err := httpSrv.Serve(listener); err != nil && err != http.ErrServerClosed {
		logger.Error("REST server stopped with error", "err", err)
		return err
	}

	logger.Info("REST server stopped")
	return nil
}

// ServeCommands will generate a long-running rest server
// (aka Light Client Daemon) that exposes functionality similar
// to the cli, but over rest
func ServeCommands(cdc *codec.Codec, registerRoutesFn func(*lcd.RestServer)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start LCD (light-client daemon), a local REST server",
		RunE: func(cmd *cobra.Command, args []string) error {
			helper.InitDeliveryConfig("")
			restCh := make(chan struct{}, 1)

			// use cmd.Context() so that Ctrl+C / SIGTERM from the root command
			// is propagated down to the REST server for graceful shutdown.
			return StartRestServer(cmd.Context(), cdc, registerRoutesFn, restCh)
		},
	}

	DecorateWithRestFlags(cmd)

	return cmd
}

// function is called whenever is the reste server flags has to be added to command.
func DecorateWithRestFlags(cmd *cobra.Command) {
	cmd.Flags().String(client.FlagListenAddr, "tcp://0.0.0.0:1317", "The address for the server to listen on")
	cmd.Flags().Bool(client.FlagTrustNode, true, "Trust connected full node (don't verify proofs for responses)")
	cmd.Flags().Int(client.FlagMaxOpenConnections, 1000, "The number of maximum open connections")

	// heimdall specific flags for rest server start
	cmd.Flags().String(client.FlagChainID, "", "The chain ID to connect to")
	cmd.Flags().String(client.FlagNode, helper.DefaultTendermintNode, "Address of the node to connect to")
}

// RegisterRoutes register routes of all modules
func RegisterRoutes(rs *lcd.RestServer) {
	registerSwaggerUI(rs)

	rpc.RegisterRPCRoutes(rs.CliCtx, rs.Mux)
	tx.RegisterRoutes(rs.CliCtx, rs.Mux)

	// auth.RegisterRoutes(rs.CliCtx, rs.Mux)
	// bank.RegisterRoutes(rs.CliCtx, rs.Mux)

	// checkpoint.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	// staking.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	// bor.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	// clerk.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)

	// register rest routes
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
}

func registerSwaggerUI(rs *lcd.RestServer) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}
	staticServer := http.FileServer(statikFS)
	rs.Mux.PathPrefix("/swagger-ui/").Handler(http.StripPrefix("/swagger-ui/", staticServer))
}

// Check locally if rest server port has been opened.
func restServerHealthCheck(restCh chan struct{}) {
	address := viper.GetString(client.FlagListenAddr)

	for {
		conn, err := net.Dial("tcp", address[6:])
		if err != nil {
			time.Sleep(healthCheckInterval)

			continue
		}

		if conn != nil {
			defer conn.Close()
		}

		close(restCh)

		break
	}
}
