package helper

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maticnetwork/heimdall/tron"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/maticnetwork/heimdall/file"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	logger "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/privval"

	tmTypes "github.com/tendermint/tendermint/types"
)

const (
	NodeFlag               = "node"
	WithDeliveryConfigFlag = "with-delivery-config"
	HomeFlag               = "home"
	FlagClientHome         = "home-client"
	ChainFlag              = "chain"
	LogLevel               = "log_level"

	// ---
	// TODO Move these to common client flags
	// BroadcastBlock defines a tx broadcasting mode where the client waits for
	// the tx to be committed in a block.
	BroadcastBlock = "block"

	// BroadcastSync defines a tx broadcasting mode where the client waits for
	// a CheckTx execution response only.
	BroadcastSync = "sync"

	// BroadcastAsync defines a tx broadcasting mode where the client returns
	// immediately.
	BroadcastAsync = "async"

	DefaultMainRPCUrl = "http://localhost:9545"
	DefaultBttcRPCUrl = "http://localhost:8545"
	DefaultBscRPCUrl  = "http://localhost:7545"

	// tron
	DefaultTronRPCUrl  = "http://localhost:50051"
	DefaultTronGridUrl = "http://localhost:30080" // get log host

	DefaultTronGridApiKey = "xxxxxxxx"
	// Services

	// DefaultAmqpURL represents default AMQP url
	DefaultAmqpURL           = "amqp://guest:guest@localhost:5672/"
	DefaultDeliveryServerURL = "http://0.0.0.0:1317"
	DefaultTendermintNodeURL = "http://0.0.0.0:26657"

	DefaultTendermintNode = "tcp://localhost:26657"

	NoACKWaitTime = 1800 * time.Second // Time ack service waits to clear buffer and elect new proposer (1800 seconds ~ 30 mins)

	DefaultCheckpointerPollInterval = 30 * time.Minute
	DefaultSyncerPollInterval       = 1 * time.Minute
	DefaultBscSyncerPollInterval    = 10 * time.Second
	DefaultTronSyncerPollInterval   = 3 * time.Second
	DefaultNoACKPollInterval        = 1010 * time.Second
	DefaultClerkPollInterval        = 10 * time.Second
	DefaultSpanPollInterval         = 1 * time.Minute
	DefaultStakingPollInterval      = 1 * time.Minute
	DefaultStartListenBlock         = 0

	DefaultMainchainMaxGasPrice = 400000000000 // 400 Gwei
	DefaultTronFeeLimit         = uint64(200000000)

	DefaultEthBusyLimitTxs  = 1000
	DefaultBscBusyLimitTxs  = 1000
	DefaultTronBusyLimitTxs = 20000

	DefaultEthMaxQueryBlocks  = 100
	DefaultBscMaxQueryBlocks  = 5
	DefaultTronMaxQueryBlocks = 5

	DefaultBttcChainID string = "15001"

	secretFilePerm = 0600
)

var (
	DefaultCLIHome  = os.ExpandEnv("$HOME/.deliverycli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.deliveryd")
	MinBalance      = big.NewInt(100000000000000000) // aka 0.1 Ether
)

var cdc = amino.NewCodec()

func init() {
	cdc.RegisterConcrete(secp256k1.PubKeySecp256k1{}, secp256k1.PubKeyAminoName, nil)
	cdc.RegisterConcrete(secp256k1.PrivKeySecp256k1{}, secp256k1.PrivKeyAminoName, nil)
	Logger = logger.NewTMLogger(logger.NewSyncWriter(os.Stdout))
}

// Configuration represents heimdall config
type Configuration struct {
	EthRPCUrl        string `mapstructure:"eth_rpc_url"`        // RPC endpoint for main chain
	TronRPCUrl       string `mapstructure:"tron_rpc_url"`       // RPC endpoint for tron chain
	BscRPCUrl        string `mapstructure:"bsc_rpc_url"`        // RPC endpoint for bsc chain
	BttcRPCUrl       string `mapstructure:"bttc_rpc_url"`       // RPC endpoint for bttc chain
	TendermintRPCUrl string `mapstructure:"tendermint_rpc_url"` // tendemint node url

	TronGridUrl       string `mapstructure:"tron_grid_url"`        // tron grid url
	AmqpURL           string `mapstructure:"amqp_url"`             // amqp url
	DeliveryServerURL string `mapstructure:"delivery_rest_server"` // delivery server url

	// tron
	TronGridApiKey string `mapstructure:"tron_grid_api_key"` // tron api key

	TronchainFeeLimit uint64 `mapstructure:"tron_chain_fee_limit"` // gas limit to tron transaction. eg....submit checkpoint.

	MainchainMaxGasPrice int64 `mapstructure:"main_chain_max_gas_price"` // max gas price to mainchain transaction. eg....submit checkpoint.

	// config related to bridge
	CheckpointerPollInterval time.Duration `mapstructure:"checkpoint_poll_interval"`  // Poll interval for checkpointer service to send new checkpoints or missing ACK
	EthSyncerPollInterval    time.Duration `mapstructure:"eth_syncer_poll_interval"`  // Poll interval for syncher service to sync for changes on eth chain
	BscSyncerPollInterval    time.Duration `mapstructure:"bsc_syncer_poll_interval"`  // Poll interval for syncher service to sync for changes on bsc chain
	TronSyncerPollInterval   time.Duration `mapstructure:"tron_syncer_poll_interval"` // Poll interval for syncher service to sync for changes on tron chain
	NoACKPollInterval        time.Duration `mapstructure:"noack_poll_interval"`       // Poll interval for ack service to send no-ack in case of no checkpoints
	StakingPollInterval      time.Duration `mapstructure:"staking_poll_interval"`     // Poll interval for staking service
	ClerkPollInterval        time.Duration `mapstructure:"clerk_poll_interval"`
	SpanPollInterval         time.Duration `mapstructure:"span_poll_interval"`

	// wait time related options
	NoACKWaitTime time.Duration `mapstructure:"no_ack_wait_time"` // Time ack service waits to clear buffer and elect new proposer

	TronStartListenBlock uint64 `mapstructure:"tron_start_listen_block"` // tron chain start listen block on bridge
	EthStartListenBlock  uint64 `mapstructure:"eth_start_listen_block"`  // eth chain start listen block on bridge

	EthUnconfirmedTxsBusyLimit  int `mapstructure:"eth_unconfirmed_txs_busy_limit"`  // the busy limit of unconfirmed txs on heimdall for eth
	BscUnconfirmedTxsBusyLimit  int `mapstructure:"bsc_unconfirmed_txs_busy_limit"`  // the busy limit of unconfirmed txs on heimdall for bsc
	TronUnconfirmedTxsBusyLimit int `mapstructure:"tron_unconfirmed_txs_busy_limit"` // the busy limit of unconfirmed txs on heimdall for tron

	EthMaxQueryBlocks  int64 `mapstructure:"eth_max_query_blocks"`  // eth max number of blocks in one query logs
	BscMaxQueryBlocks  int64 `mapstructure:"bsc_max_query_blocks"`  // bsc max number of blocks in one query logs
	TronMaxQueryBlocks int64 `mapstructure:"tron_max_query_blocks"` // tron max number of blocks in one query logs
}

var conf Configuration

// MainChainClient stores eth client for Main chain Network
var mainChainClient *ethclient.Client
var mainRPCClient *rpc.Client

// BscChainClient stores eth client for Bsc chain Network
var bscChainClient *ethclient.Client
var bscRPCClient *rpc.Client

var tronRPCClient *tron.Client

// MaticClient stores eth/rpc client for Matic Network
var maticClient *ethclient.Client
var maticRPCClient *rpc.Client

var maticEthClient *eth.EthAPIBackend

// private key object
var privObject secp256k1.PrivKeySecp256k1

var pubObject secp256k1.PubKeySecp256k1

// Logger stores global logger object
var Logger logger.Logger

// GenesisDoc contains the genesis file
var GenesisDoc tmTypes.GenesisDoc

// Contracts
// var RootChain types.Contract
// var DepositManager types.Contract

// InitDeliveryConfig initializes with viper config (from heimdall configuration)
func InitDeliveryConfig(homeDir string) {
	if strings.Compare(homeDir, "") == 0 {
		// get home dir from viper
		homeDir = viper.GetString(HomeFlag)
	}

	// get heimdall config filepath from viper/cobra flag
	deliveryConfigFilePath := viper.GetString(WithDeliveryConfigFlag)

	// init heimdall with changed config files
	InitDeliveryConfigWith(homeDir, deliveryConfigFilePath)
}

// InitDeliveryConfigWith initializes passed delivery/tendermint config files
func InitDeliveryConfigWith(homeDir string, deliveryConfigFilePath string) {
	if strings.Compare(homeDir, "") == 0 {
		return
	}

	if strings.Compare(conf.BttcRPCUrl, "") != 0 {
		return
	}

	configDir := filepath.Join(homeDir, "config")

	heimdallViper := viper.New()
	if deliveryConfigFilePath == "" {
		heimdallViper.SetConfigName("delivery-config") // name of config file (without extension)
		heimdallViper.AddConfigPath(configDir)         // call multiple times to add many search paths
	} else {
		heimdallViper.SetConfigFile(deliveryConfigFilePath) // set config file explicitly
	}

	err := heimdallViper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Fatal(err)
	}

	if err = heimdallViper.UnmarshalExact(&conf); err != nil {
		log.Fatalln("Unable to unmarshall config", "Error", err)
	}

	if mainRPCClient, err = rpc.Dial(conf.EthRPCUrl); err != nil {
		log.Fatalln("Unable to dial via ethClient", "URL=", conf.EthRPCUrl, "chain=eth", "Error", err)
	}

	mainChainClient = ethclient.NewClient(mainRPCClient)
	if maticRPCClient, err = rpc.Dial(conf.BttcRPCUrl); err != nil {
		log.Fatal(err)
	}

	if bscRPCClient, err = rpc.Dial(conf.BscRPCUrl); err != nil {
		log.Fatalln("Unable to dial via ethClient", "URL=", conf.BscRPCUrl, "chain=bsc", "Error", err)
	}
	bscChainClient = ethclient.NewClient(bscRPCClient)

	tronRPCClient = tron.NewClient(conf.TronRPCUrl)

	maticClient = ethclient.NewClient(maticRPCClient)
	// Loading genesis doc
	genDoc, err := tmTypes.GenesisDocFromFile(filepath.Join(configDir, "genesis.json"))
	if err != nil {
		log.Fatal(err)
	}
	GenesisDoc = *genDoc

	// load pv file, unmarshall and set to privObject
	err = file.PermCheck(file.Rootify("priv_validator_key.json", configDir), secretFilePerm)
	if err != nil {
		Logger.Error(err.Error())
	}
	privVal := privval.LoadFilePV(filepath.Join(configDir, "priv_validator_key.json"), filepath.Join(configDir, "priv_validator_key.json"))
	cdc.MustUnmarshalBinaryBare(privVal.Key.PrivKey.Bytes(), &privObject)
	cdc.MustUnmarshalBinaryBare(privObject.PubKey().Bytes(), &pubObject)
}

// GetDefaultHeimdallConfig returns configration with default params
func GetDefaultHeimdallConfig() Configuration {
	return Configuration{
		EthRPCUrl:        DefaultMainRPCUrl,
		TronRPCUrl:       DefaultTronRPCUrl,
		BttcRPCUrl:       DefaultBttcRPCUrl,
		BscRPCUrl:        DefaultBscRPCUrl,
		TendermintRPCUrl: DefaultTendermintNodeURL,

		TronGridUrl:       DefaultTronGridUrl,
		AmqpURL:           DefaultAmqpURL,
		DeliveryServerURL: DefaultDeliveryServerURL,

		TronchainFeeLimit: DefaultTronFeeLimit,

		MainchainMaxGasPrice: DefaultMainchainMaxGasPrice,

		CheckpointerPollInterval: DefaultCheckpointerPollInterval,
		EthSyncerPollInterval:    DefaultSyncerPollInterval,
		BscSyncerPollInterval:    DefaultBscSyncerPollInterval,
		TronSyncerPollInterval:   DefaultTronSyncerPollInterval,
		NoACKPollInterval:        DefaultNoACKPollInterval,
		ClerkPollInterval:        DefaultClerkPollInterval,
		SpanPollInterval:         DefaultSpanPollInterval,
		StakingPollInterval:      DefaultStakingPollInterval,

		NoACKWaitTime: NoACKWaitTime,

		TronGridApiKey:       DefaultTronGridApiKey,
		TronStartListenBlock: DefaultStartListenBlock,
		EthStartListenBlock:  DefaultStartListenBlock,

		EthUnconfirmedTxsBusyLimit:  DefaultEthBusyLimitTxs,
		BscUnconfirmedTxsBusyLimit:  DefaultBscBusyLimitTxs,
		TronUnconfirmedTxsBusyLimit: DefaultTronBusyLimitTxs,

		EthMaxQueryBlocks:  DefaultEthMaxQueryBlocks,
		BscMaxQueryBlocks:  DefaultBscMaxQueryBlocks,
		TronMaxQueryBlocks: DefaultTronMaxQueryBlocks,
	}
}

// GetConfig returns cached configuration object
func GetConfig() Configuration {
	return conf
}

func GetGenesisDoc() tmTypes.GenesisDoc {
	return GenesisDoc
}

// TEST PURPOSE ONLY
// SetTestConfig sets test configuration
func SetTestConfig(_conf Configuration) {
	conf = _conf
}

//
// Get main/matic clients
//

// GetMainChainRPCClient returns main chain RPC client
func GetMainChainRPCClient() *rpc.Client {
	return mainRPCClient
}

// GetMainClient returns main chain's eth client
func GetMainClient() *ethclient.Client {
	return mainChainClient
}

// GetBscChainRPCClient returns bsc chain RPC client
func GetBscChainRPCClient() *rpc.Client {
	return bscRPCClient
}

// GetBscClient returns bsc chain's eth client
func GetBscClient() *ethclient.Client {
	return bscChainClient
}

// GetTronChainRPCClient returns main chain RPC client
func GetTronChainRPCClient() *tron.Client {
	return tronRPCClient
}

// GetMaticClient returns matic's eth client
func GetMaticClient() *ethclient.Client {
	return maticClient
}

// GetMaticRPCClient returns matic's RPC client
func GetMaticRPCClient() *rpc.Client {
	return maticRPCClient
}

// GetMaticEthClient returns matic's Eth client
func GetMaticEthClient() *eth.EthAPIBackend {
	return maticEthClient
}

// GetPrivKey returns priv key object
func GetPrivKey() secp256k1.PrivKeySecp256k1 {
	return privObject
}

// GetECDSAPrivKey return ecdsa private key
func GetECDSAPrivKey() *ecdsa.PrivateKey {
	// get priv key
	pkObject := GetPrivKey()

	// create ecdsa private key
	ecdsaPrivateKey, _ := ethCrypto.ToECDSA(pkObject[:])
	return ecdsaPrivateKey
}

// GetPubKey returns pub key object
func GetPubKey() secp256k1.PubKeySecp256k1 {
	return pubObject
}

// GetAddress returns address object
func GetAddress() []byte {
	return GetPubKey().Address().Bytes()
}

// GetValidChains returns all the valid chains.
func GetValidChains() []string {
	return []string{"mainnet", "donau", "local"}
}
