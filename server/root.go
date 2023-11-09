package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/go-kit/kit/log"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmLog "github.com/tendermint/tendermint/libs/log"

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

	logger := tmLog.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "rest-server")

	err := restServer.Start(
		viper.GetString(client.FlagListenAddr),
		viper.GetInt(client.FlagMaxOpenConnections),
		0,
		0,
	)
	if err != nil {
		logger.Error("Cannot start REST server", "Error", err)
	}

	return err
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
			err := StartRestServer(context.Background(), cdc, registerRoutesFn, restCh)

			return err
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
