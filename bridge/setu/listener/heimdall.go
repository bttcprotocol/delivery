package listener

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	stakingTypes "github.com/maticnetwork/heimdall/staking/types"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/heimdall/helper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	checkpointTypes "github.com/maticnetwork/heimdall/checkpoint/types"
	slashingTypes "github.com/maticnetwork/heimdall/slashing/types"
)

const (
	heimdallLastBlockKey = "heimdall-last-block" // storage key
)

// HeimdallListener - Listens to and process events from heimdall
type HeimdallListener struct {
	BaseListener
}

// NewHeimdallListener - constructor func
func NewHeimdallListener() *HeimdallListener {
	return &HeimdallListener{}
}

// Start starts new block subscription
func (hl *HeimdallListener) Start() error {
	hl.Logger.Info("Starting")

	// create cancellable context
	headerCtx, cancelHeaderProcess := context.WithCancel(context.Background())
	hl.cancelHeaderProcess = cancelHeaderProcess

	// Heimdall pollIntervall = (minimal pollInterval of rootchain and matichain)
	pollInterval := helper.GetConfig().EthSyncerPollInterval
	if helper.GetConfig().CheckpointerPollInterval < helper.GetConfig().EthSyncerPollInterval {
		pollInterval = helper.GetConfig().CheckpointerPollInterval
	}

	hl.Logger.Info("Start polling for events", "pollInterval", pollInterval)
	hl.StartPolling(headerCtx, pollInterval, false)
	return nil
}

// ProcessHeader -
func (hl *HeimdallListener) ProcessHeader(*types.Header) {

}

// StartPolling - starts polling for heimdall events
func (hl *HeimdallListener) StartPolling(ctx context.Context, pollInterval time.Duration, needAlign bool) {
	// How often to fire the passed in function in second
	interval := pollInterval

	// needAlign is used to decide whether the ticker is align to 1970 UTC.
	// if true, the ticker will always tick as it begins at 1970 UTC.
	// otherwise, ticker begins to tick depended on the time when program run.
	firstInterval := interval
	if needAlign {
		now := time.Now()
		baseTime := time.Unix(0, 0)
		firstInterval = interval - (now.UTC().Sub(baseTime) % interval)
	}

	// Setup the ticket and the channel to signal
	// the ending of the interval
	ticker := time.NewTicker(firstInterval)

	// var eventTypes []string
	// eventTypes = append(eventTypes, "message.action='checkpoint'")
	// eventTypes = append(eventTypes, "message.action='event-record'")
	// eventTypes = append(eventTypes, "message.action='tick'")
	// ADD EVENT TYPE for SLASH-LIMIT

	// start listening
	for {
		select {
		case <-ticker.C:
			fromBlock, toBlock, err := hl.fetchFromAndToBlock()
			if err != nil {
				hl.Logger.Error("Error fetching fromBlock and toBlock...skipping events query", "error", err)
			} else if fromBlock < toBlock {

				hl.Logger.Info("Fetching new events between", "fromBlock", fromBlock, "toBlock", toBlock)

				// Querying and processing Begin events
				for i := fromBlock; i <= toBlock; i++ {
					events, err := helper.GetBeginBlockEvents(hl.httpClient, int64(i))
					if err != nil {
						hl.Logger.Error("Error fetching begin block events", "error", err)
					}
					for _, event := range events {
						hl.ProcessBlockEvent(sdk.StringifyEvent(event), int64(i))
					}
				}

				// Querying and processing tx Events. Below for loop is kept for future purpose to process events from tx
				/* 		for _, eventType := range eventTypes {
					var query []string
					query = append(query, eventType)
					query = append(query, fmt.Sprintf("tx.height>=%v", fromBlock))
					query = append(query, fmt.Sprintf("tx.height<=%v", toBlock))

					limit := 50
					for page := 1; page > 0; {
						searchResult, err := helper.QueryTxsByEvents(hl.cliCtx, query, page, limit)
						hl.Logger.Debug("Fetching new events using search query", "query", query, "page", page, "limit", limit)

						if err != nil {
							hl.Logger.Error("Error while searching events", "eventType", eventType, "error", err)
							break
						}

						for _, tx := range searchResult.Txs {
							for _, log := range tx.Logs {
								event := helper.FilterEvents(log.Events, func(et sdk.StringEvent) bool {
									return et.Type == checkpointTypes.EventTypeCheckpoint || et.Type == clerkTypes.EventTypeRecord
								})
								if event != nil {
									hl.ProcessEvent(*event, tx)
								}
							}
						}

						if len(searchResult.Txs) == limit {
							page = page + 1
						} else {
							page = 0
						}
					}
				} */
				// set last block to storage
				if err := hl.storageClient.Put([]byte(heimdallLastBlockKey), []byte(strconv.FormatUint(toBlock, 10)), nil); err != nil {
					hl.Logger.Error("hl.storageClient.Put", "Error", err)
				}
			}
			ticker.Reset(interval)

		case <-ctx.Done():
			hl.Logger.Info("Polling stopped")
			ticker.Stop()
			return
		}
	}
}

func (hl *HeimdallListener) fetchFromAndToBlock() (uint64, uint64, error) {
	// toBlock - get latest blockheight from heimdall node
	fromBlock := uint64(0)
	toBlock := uint64(0)

	nodeStatus, err := helper.GetNodeStatus(hl.cliCtx)
	if err != nil {
		hl.Logger.Error("Error while fetching heimdall node status", "error", err)
		return fromBlock, toBlock, err
	}
	toBlock = uint64(nodeStatus.SyncInfo.LatestBlockHeight)

	// fromBlock - get last block from storage
	hasLastBlock, _ := hl.storageClient.Has([]byte(heimdallLastBlockKey), nil)
	if hasLastBlock {
		lastBlockBytes, err := hl.storageClient.Get([]byte(heimdallLastBlockKey), nil)
		if err != nil {
			hl.Logger.Info("Error while fetching last block bytes from storage", "error", err)
			return fromBlock, toBlock, err
		}

		if result, err := strconv.ParseUint(string(lastBlockBytes), 10, 64); err == nil {
			hl.Logger.Debug("Got last block from bridge storage", "lastBlock", result)
			fromBlock = uint64(result) + 1
		} else {
			hl.Logger.Info("Error parsing last block bytes from storage", "error", err)
			toBlock = 0
			return fromBlock, toBlock, err
		}
	}
	return fromBlock, toBlock, err
}

// ProcessBlockEvent - process Blockevents (BeginBlock, EndBlock events) from heimdall.
func (hl *HeimdallListener) ProcessBlockEvent(event sdk.StringEvent, blockHeight int64) {
	hl.Logger.Info("Received block event from Heimdall", "eventType", event.Type, "height", blockHeight)
	eventBytes, err := json.Marshal(event)
	if err != nil {
		hl.Logger.Error("Error while parsing block event", "error", err, "eventType", event.Type)
		return
	}

	switch event.Type {
	case checkpointTypes.EventTypeCheckpoint:
		hl.sendBlockTask("sendCheckpointToRootchain", eventBytes, blockHeight)
	case checkpointTypes.EventTypeCheckpointSync:
		hl.sendBlockTask("sendCheckpointSyncToStakeChain", eventBytes, blockHeight)
	case slashingTypes.EventTypeSlashLimit:
		hl.sendBlockTask("sendTickToHeimdall", eventBytes, blockHeight)
	case slashingTypes.EventTypeTickConfirm:
		hl.sendBlockTask("sendTickToRootchain", eventBytes, blockHeight)

	case stakingTypes.EventTypeValidatorJoin,
		stakingTypes.EventTypeSignerUpdate,
		stakingTypes.EventTypeValidatorExit,
		stakingTypes.EventTypeStakingSyncAck:
		hl.sendBlockTask("sendStakingSyncToHeimdall", eventBytes, blockHeight)
	case stakingTypes.EventTypeStakingSync:
		hl.sendBlockTask("sendStakingSyncToRootChain", eventBytes, blockHeight)
	default:
		hl.Logger.Debug("BlockEvent Type mismatch", "eventType", event.Type)
	}
}

func (hl *HeimdallListener) sendBlockTask(taskName string, eventBytes []byte, blockHeight int64) {
	// create machinery task
	signature := &tasks.Signature{
		Name: taskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(eventBytes),
			},
			{
				Type:  "int64",
				Value: blockHeight,
			},
		},
	}
	signature.RetryCount = 3
	signature.RetryTimeout = 3
	hl.Logger.Info("Sending block level task",
		"taskName", taskName, "eventBytes", eventBytes, "currentTime", time.Now(), "blockHeight", blockHeight)
	// send task
	_, err := hl.queueConnector.Server.SendTask(signature)
	if err != nil {
		hl.Logger.Error("Error sending block level task", "taskName", taskName, "blockHeight", blockHeight, "error", err)
	}
}
