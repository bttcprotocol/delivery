package listener

import (
	"context"
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	hmtypes "github.com/maticnetwork/heimdall/types"

	"github.com/RichardKnop/machinery/v1/tasks"
	ethereum "github.com/maticnetwork/bor"
	"github.com/maticnetwork/bor/accounts/abi"
	ethCommon "github.com/maticnetwork/bor/common"
	ethTypes "github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	chainmanagerTypes "github.com/maticnetwork/heimdall/chainmanager/types"
	"github.com/maticnetwork/heimdall/helper"
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
	pollInterval   time.Duration

	busyLimit      int
	maxQueryBlocks int64

	stateSyncedCountWithDecay uint64
}

const (
	lastEthBlockKey = "eth-last-block" // storage key
	lastBscBlockKey = "bsc-last-block"

	decayPerSecond = 30
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
	rootChainListener := &RootChainListener{
		abis:           abis,
		stakingInfoAbi: &contractCaller.StakingInfoABI,
		rootChainType:  rootChain,
	}
	switch rootChain {
	case hmtypes.RootChainTypeEth:
		rootChainListener.blockKey = lastEthBlockKey
		rootChainListener.pollInterval = helper.GetConfig().EthSyncerPollInterval
		rootChainListener.busyLimit = helper.GetConfig().EthUnconfirmedTxsBusyLimit
		rootChainListener.maxQueryBlocks = helper.GetConfig().EthMaxQueryBlocks
	case hmtypes.RootChainTypeBsc:
		rootChainListener.blockKey = lastBscBlockKey
		rootChainListener.pollInterval = helper.GetConfig().BscSyncerPollInterval
		rootChainListener.busyLimit = helper.GetConfig().BscUnconfirmedTxsBusyLimit
		rootChainListener.maxQueryBlocks = helper.GetConfig().BscMaxQueryBlocks
	default:
		panic("wrong chain type for root chain")
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
			"root", rl.rootChainType, "pollInterval", rl.pollInterval)
		go rl.StartPolling(ctx, rl.pollInterval)
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

	// check if heimdall is busy
	if rl.busyLimit != 0 {
		// event decay
		decay := decayPerSecond * uint64(rl.pollInterval.Seconds())
		if rl.stateSyncedCountWithDecay > decay {
			rl.stateSyncedCountWithDecay -= decay
		} else {
			rl.stateSyncedCountWithDecay = 0
		}
		if rl.stateSyncedCountWithDecay > uint64(rl.busyLimit) {
			rl.Logger.Debug("heimdall is busy now", "busyLimit", rl.busyLimit, "stateSyncedCountWithDecay", rl.stateSyncedCountWithDecay)
			return
		}

		numUnconfirmedTxs, err := helper.GetNumUnconfirmedTxs(rl.cliCtx)
		if err != nil {
			rl.Logger.Debug("heimdall is busy now", "error", err)
			return
		}
		if numUnconfirmedTxs.Total > rl.busyLimit {
			rl.Logger.Debug("heimdall is busy now", "busyLimit", rl.busyLimit, "UnconfirmedTxs", numUnconfirmedTxs.Total)
			return
		}
	}
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
	if rl.maxQueryBlocks != 0 && big.NewInt(0).Sub(toBlock, fromBlock).Cmp(big.NewInt(rl.maxQueryBlocks)) > 0 {
		toBlock = toBlock.Add(fromBlock, big.NewInt(rl.maxQueryBlocks))
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
						rl.stateSyncedCountWithDecay++
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
