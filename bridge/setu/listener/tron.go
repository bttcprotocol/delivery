package listener

import (
	"context"
	"math/big"
	"strconv"
	"time"

	"github.com/maticnetwork/heimdall/types"

	"github.com/RichardKnop/machinery/v1/tasks"
	ethTypes "github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	chainmanagerTypes "github.com/maticnetwork/heimdall/chainmanager/types"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/maticnetwork/heimdall/tron"
)

const (
	tronLastBlockKey = "tron-last-block" // storage key
)

// TronListener - Listens to and process events from Tron
type TronListener struct {
	BaseListener
	tronClient *tron.Client
}

// NewTronListener - constructor func
func NewTronListener() *TronListener {
	TronListener := &TronListener{
		tronClient: helper.GetTronChainRPCClient(),
	}
	return TronListener
}

// Start starts new block subscription
func (tl *TronListener) Start() error {
	tl.Logger.Info("Starting")

	// create cancellable context
	headerCtx, cancelHeaderProcess := context.WithCancel(context.Background())
	tl.cancelHeaderProcess = cancelHeaderProcess

	// start header process
	go tl.StartHeaderProcess(headerCtx)

	// subscribe to new head
	pollInterval := helper.GetConfig().SyncerPollInterval

	tl.Logger.Info("Start polling for events", "pollInterval", pollInterval)
	// poll for new header using client object
	go tl.StartPolling(headerCtx, pollInterval)
	return nil
}

// startPolling starts polling
func (tl *TronListener) StartPolling(ctx context.Context, pollInterval time.Duration) {
	// How often to fire the passed in function in second
	interval := pollInterval

	// Setup the ticket and the channel to signal
	// the ending of the interval
	ticker := time.NewTicker(interval)

	// start listening
	for {
		select {
		case <-ticker.C:
			headerNum, err := tl.tronClient.GetNowBlock(ctx)
			if err == nil {
				// send data to channel
				tl.HeaderChannel <- &(ethTypes.Header{
					Number: big.NewInt(headerNum),
				})
			}
		case <-ctx.Done():
			tl.Logger.Info("Polling stopped")
			ticker.Stop()
			return
		}
	}
}

// ProcessHeader - process headerblock from rootchain
func (tl *TronListener) ProcessHeader(newHeader *ethTypes.Header) {
	tl.Logger.Debug("New block detected", "blockNumber", newHeader.Number)

	// fetch context
	chainManagerParams, err := tl.getChainManagerParams()
	if err != nil {
		return
	}
	latestNumber := newHeader.Number

	// confirmation
	confirmationBlocks := big.NewInt(int64(chainManagerParams.MainchainTxConfirmations))

	if latestNumber.Cmp(confirmationBlocks) <= 0 {
		tl.Logger.Error("Block number less than Confirmations required", "blockNumber", latestNumber.Uint64, "confirmationsRequired", confirmationBlocks.Uint64)
		return
	}
	latestNumber = latestNumber.Sub(latestNumber, confirmationBlocks)

	// default fromBlock
	fromBlock := latestNumber

	// get last block from storage
	hasLastBlock, _ := tl.storageClient.Has([]byte(tronLastBlockKey), nil)
	if hasLastBlock {
		lastBlockBytes, err := tl.storageClient.Get([]byte(tronLastBlockKey), nil)
		if err != nil {
			tl.Logger.Info("Error while fetching last block bytes from storage", "error", err)
			return
		}
		tl.Logger.Debug("Got last block from bridge storage", "lastBlock", string(lastBlockBytes))
		if result, err := strconv.ParseUint(string(lastBlockBytes), 10, 64); err == nil {
			if result >= newHeader.Number.Uint64() {
				return
			}
			fromBlock = big.NewInt(0).SetUint64(result + 1)
		}
	}

	// to block
	toBlock := latestNumber

	if toBlock.Cmp(fromBlock) == -1 {
		fromBlock = toBlock
	}

	// set last block to storage
	if err := tl.storageClient.Put([]byte(tronLastBlockKey), []byte(toBlock.String()), nil); err != nil {
		tl.Logger.Error("tl.storageClient.Put", "Error", err)
	}

	// query events
	tl.queryAndBroadcastEvents(chainManagerParams.ChainParams, fromBlock, toBlock)
}

func (tl *TronListener) queryAndBroadcastEvents(chainParams chainmanagerTypes.ChainParams, fromBlock *big.Int, toBlock *big.Int) {
	tl.Logger.Info("Query tron event logs", "fromBlock", fromBlock, "toBlock", toBlock)
	//TODO:
}

func (tl *TronListener) sendTaskWithDelay(taskName string, eventName string, logBytes []byte, delay time.Duration) {
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
				Value: types.RootChainTypeTron,
			},
		},
	}
	signature.RetryCount = 3

	// add delay for task so that multiple validators won't send same transaction at same time
	eta := time.Now().Add(delay)
	signature.ETA = &eta
	tl.Logger.Info("Sending task", "taskName", taskName, "currentTime", time.Now(), "delayTime", eta)
	_, err := tl.queueConnector.Server.SendTask(signature)
	if err != nil {
		tl.Logger.Error("Error sending task", "taskName", taskName, "error", err)
	}
}

//
// utils
//

func (tl *TronListener) getChainManagerParams() (*chainmanagerTypes.Params, error) {
	chainmanagerParams, err := util.GetChainmanagerParams(tl.cliCtx)
	if err != nil {
		tl.Logger.Error("Error while fetching chain manager params", "error", err)
		return nil, err
	}

	return chainmanagerParams, nil
}
