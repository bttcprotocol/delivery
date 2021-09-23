package listener

import (
	"context"
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	"github.com/RichardKnop/machinery/v1/tasks"
	ethereum "github.com/maticnetwork/bor"
	"github.com/maticnetwork/bor/accounts/abi"
	ethCommon "github.com/maticnetwork/bor/common"
	ethTypes "github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	chainmanagerTypes "github.com/maticnetwork/heimdall/chainmanager/types"
	"github.com/maticnetwork/heimdall/helper"
	hmtypes "github.com/maticnetwork/heimdall/types"
)

// RootChainListenerContext root chain listener context
type RootChainListenerContext struct {
	ChainmanagerParams *chainmanagerTypes.Params
}

// RootChainListener - Listens to and process events from rootchain
type RootChainListener struct {
	BaseListener
	// ABIs
	abis []*abi.ABI

	stakingInfoAbi *abi.ABI
	rootChainType  string
	blockKey       string
}

const (
	lastEthBlockKey = "eth-last-block" // storage key
	lastBscBlockKey = "bsc-last-block"
)

// NewRootChainListener - constructor func
func NewRootChainListener(rootChain string) *RootChainListener {
	contractCaller, err := helper.NewContractCaller()
	if err != nil {
		panic(err)
	}
	abis := []*abi.ABI{
		&contractCaller.RootChainABI,
		&contractCaller.StateSenderABI,
		&contractCaller.StakingInfoABI,
	}
	blockKey := ""
	switch rootChain {
	case hmtypes.RootChainTypeEth:
		blockKey = lastEthBlockKey
	case hmtypes.RootChainTypeBsc:
		blockKey = lastBscBlockKey
	default:
		panic("wrong chain type for root chain")
	}

	rootChainListener := &RootChainListener{
		abis:           abis,
		stakingInfoAbi: &contractCaller.StakingInfoABI,
		rootChainType:  rootChain,
		blockKey:       blockKey,
	}
	return rootChainListener
}

// Start starts new block subscription
func (rl *RootChainListener) Start() error {
	rl.Logger.Info("Starting", "root", rl.rootChainType)

	// create cancellable context
	ctx, cancelSubscription := context.WithCancel(context.Background())
	rl.cancelSubscription = cancelSubscription

	// create cancellable context
	headerCtx, cancelHeaderProcess := context.WithCancel(context.Background())
	rl.cancelHeaderProcess = cancelHeaderProcess

	// set start listen block
	startListenBlock := rl.contractConnector.GetStartListenBlock(rl.rootChainType)
	if startListenBlock != 0 {
		_ = rl.setStartListenBLock(startListenBlock, rl.blockKey)
	}

	// start header process
	go rl.StartHeaderProcess(headerCtx)

	// subscribe to new head
	subscription, err := rl.chainClient.SubscribeNewHead(ctx, rl.HeaderChannel)
	if err != nil {
		// start go routine to poll for new header using client object
		rl.Logger.Info("Start polling for root chain header blocks",
			"root", rl.rootChainType, "pollInterval", helper.GetConfig().SyncerPollInterval)
		go rl.StartPolling(ctx, helper.GetConfig().SyncerPollInterval)
	} else {
		// start go routine to listen new header using subscription
		go rl.StartSubscription(ctx, subscription)
	}

	// subscribed to new head
	rl.Logger.Info("Subscribed to new head", "root", rl.rootChainType)

	return nil
}

// ProcessHeader - process headerblock from rootchain
func (rl *RootChainListener) ProcessHeader(newHeader *ethTypes.Header) {
	rl.Logger.Debug("New block detected", "root", rl.rootChainType, "blockNumber", newHeader.Number)

	// fetch context
	rootchainContext, err := rl.getRootChainContext()
	if err != nil {
		return
	}
	requiredConfirmations := rootchainContext.ChainmanagerParams.MainchainTxConfirmations
	latestNumber := newHeader.Number

	// confirmation
	confirmationBlocks := big.NewInt(0).SetUint64(requiredConfirmations)

	if latestNumber.Cmp(confirmationBlocks) <= 0 {
		rl.Logger.Error("Block number less than Confirmations required",
			"root", rl.rootChainType, "blockNumber", latestNumber.Uint64, "confirmationsRequired", confirmationBlocks.Uint64)
		return
	}
	latestNumber = latestNumber.Sub(latestNumber, confirmationBlocks)

	// default fromBlock
	fromBlock := latestNumber

	// get last block from storage
	hasLastBlock, _ := rl.storageClient.Has([]byte(rl.blockKey), nil)
	if hasLastBlock {
		lastBlockBytes, err := rl.storageClient.Get([]byte(rl.blockKey), nil)
		if err != nil {
			rl.Logger.Info("Error while fetching last block bytes from storage", "root", rl.rootChainType, "error", err)
			return
		}
		rl.Logger.Debug("Got last block from bridge storage", "root", rl.rootChainType, "lastBlock", string(lastBlockBytes))
		if result, err := strconv.ParseUint(string(lastBlockBytes), 10, 64); err == nil {
			if result >= newHeader.Number.Uint64() {
				return
			}
			if result+1 < fromBlock.Uint64() { // only start from solidity block
				fromBlock = big.NewInt(0).SetUint64(result + 1)
			}
		}

	}

	// to block
	toBlock := latestNumber

	if toBlock.Cmp(fromBlock) == -1 {
		fromBlock = toBlock
	}

	// query events
	rl.queryAndBroadcastEvents(rootchainContext, fromBlock, toBlock)
}

func (rl *RootChainListener) queryAndBroadcastEvents(rootchainContext *RootChainListenerContext, fromBlock *big.Int, toBlock *big.Int) {
	rl.Logger.Info("Query rootchain event logs", "root", rl.rootChainType, "fromBlock", fromBlock, "toBlock", toBlock)

	// get chain params
	chainParams := rootchainContext.ChainmanagerParams.ChainParams

	// draft a query

	queryAddresses := []ethCommon.Address{
		chainParams.RootChainAddress.EthAddress(),
		chainParams.StakingInfoAddress.EthAddress(),
		chainParams.StateSenderAddress.EthAddress(),
	}

	query := ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: queryAddresses}
	// get logs from root chain by filter
	logs, err := rl.chainClient.FilterLogs(context.Background(), query)
	if err != nil {
		rl.Logger.Error("Error while filtering logs", "error", err)
		return
	} else if len(logs) > 0 {
		rl.Logger.Debug("New logs found", "numberOfLogs", len(logs))
	}

	// set last block to storage
	if err := rl.storageClient.Put([]byte(rl.blockKey), []byte(toBlock.String()), nil); err != nil {
		rl.Logger.Error("rl.storageClient.Put", "Error", err)
	}

	// process filtered log
	for _, vLog := range logs {
		topic := vLog.Topics[0].Bytes()
		for _, abiObject := range rl.abis {
			selectedEvent := helper.EventByID(abiObject, topic)
			logBytes, _ := json.Marshal(vLog)
			if selectedEvent != nil {
				rl.Logger.Debug("ReceivedEvent", "eventname", selectedEvent.Name, "root", rl.rootChainType)
				switch selectedEvent.Name {
				case "NewHeaderBlock":
					if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx); isCurrentValidator {
						rl.sendTaskWithDelay("sendCheckpointAckToHeimdall", selectedEvent.Name, logBytes, delay)
					}

				case "StateSynced":
					if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx); isCurrentValidator {
						rl.sendTaskWithDelay("sendStateSyncedToHeimdall", selectedEvent.Name, logBytes, delay)
					}

				case "StakeAck":
					if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx); isCurrentValidator {
						rl.sendTaskWithDelay("sendStakingAckToHeimdall", selectedEvent.Name, logBytes, delay)
					}
				}
			}
		}
	}
}

func (rl *RootChainListener) sendTaskWithDelay(taskName string, eventName string, logBytes []byte, delay time.Duration) {
	signature := &tasks.Signature{
		Name: taskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: eventName,
			},
			{
				Type:  "string",
				Value: string(logBytes),
			},
			{
				Type:  "string",
				Value: rl.rootChainType,
			},
		},
	}
	signature.RetryCount = 3
	signature.RetryTimeout = 3
	// add delay for task so that multiple validators won't send same transaction at same time
	eta := time.Now().Add(delay)
	signature.ETA = &eta
	rl.Logger.Info("Sending task", "root", rl.rootChainType, "taskName", taskName, "currentTime", time.Now(), "delayTime", eta)
	_, err := rl.queueConnector.Server.SendTask(signature)
	if err != nil {
		rl.Logger.Error("Error sending task", "taskName", taskName, "error", err)
	}
}

//
// utils
//

func (rl *RootChainListener) getRootChainContext() (*RootChainListenerContext, error) {
	chainmanagerParams, err := util.GetNewChainParams(rl.cliCtx, rl.rootChainType)
	if err != nil {
		rl.Logger.Error("Error while fetching chain manager params", "error", err)
		return nil, err
	}

	return &RootChainListenerContext{
		ChainmanagerParams: chainmanagerParams,
	}, nil
}
