package listener

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	"github.com/maticnetwork/heimdall/helper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	checkpointTypes "github.com/maticnetwork/heimdall/checkpoint/types"
	clerkTypes "github.com/maticnetwork/heimdall/clerk/types"
	featureManagerTypes "github.com/maticnetwork/heimdall/featuremanager/types"
	slashingTypes "github.com/maticnetwork/heimdall/slashing/types"
	stakingTypes "github.com/maticnetwork/heimdall/staking/types"
	htype "github.com/maticnetwork/heimdall/types"
)

const (
	heimdallLastBlockKey = "heimdall-last-block" // storage key
)

// HeimdallListener - Listens to and process events from heimdall
type HeimdallListener struct {
	BaseListener

	stateSyncedInitializationRun uint32 // atomic
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

	go hl.StartPolling(headerCtx, pollInterval, false)

	go hl.StartPollingEventRecord(headerCtx, pollInterval, false)

	return nil
}

// ProcessHeader -
func (hl *HeimdallListener) ProcessHeader(*types.Header) {
}

// StartPolling - starts polling for heimdall events
// needAlign is used to decide whether the ticker is align to 1970 UTC.
// if true, the ticker will always tick as it begins at 1970 UTC.
func (hl *HeimdallListener) StartPolling(ctx context.Context, pollInterval time.Duration, needAlign bool) {
	// How often to fire the passed in function in second
	interval := pollInterval
	firstInterval := interval
	if needAlign {
		now := time.Now()
		baseTime := time.Unix(0, 0)
		firstInterval = interval - (now.UTC().Sub(baseTime) % interval)
	}

	// Setup the ticket and the channel to signal
	// the ending of the interval
	ticker := time.NewTicker(firstInterval)

	var tickerOnce sync.Once
	// var eventTypes []string
	// eventTypes = append(eventTypes, "message.action='checkpoint'")
	// eventTypes = append(eventTypes, "message.action='event-record'")
	// eventTypes = append(eventTypes, "message.action='tick'")
	// ADD EVENT TYPE for SLASH-LIMIT

	// start listening
	for {
		select {
		case <-ticker.C:
			tickerOnce.Do(func() {
				ticker.Reset(interval)
			})
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

// StartPollingEventRecord - starts polling for heimdall event records
// needAlign is used to decide whether the ticker is align to 1970 UTC.
// if true, the ticker will always tick as it begins at 1970 UTC.
func (hl *HeimdallListener) StartPollingEventRecord(ctx context.Context, pollInterval time.Duration, needAlign bool) {
	// How often to fire the passed in function in second
	interval := pollInterval
	firstInterval := interval

	if needAlign {
		now := time.Now()
		baseTime := time.Unix(0, 0)
		firstInterval = interval - (now.UTC().Sub(baseTime) % interval)
	}

	// Setup the ticket and the channel to signal
	// the ending of the interval
	ticker := time.NewTicker(firstInterval)

	var tickerOnce sync.Once
	// start listening
	for {
		select {
		case <-ticker.C:
			tickerOnce.Do(func() {
				ticker.Reset(interval)
			})

			targetFeature, err := util.GetTargetFeatureConfig(hl.cliCtx, featureManagerTypes.DynamicCheckpoint)
			if err == nil && targetFeature.IsOpen {
				hl.loadEventRecords(ctx, pollInterval)
			}

		case <-ctx.Done():
			hl.Logger.Info("Polling stopped")
			ticker.Stop()

			return
		}
	}
}

func (hl *HeimdallListener) loadEventRecords(ctx context.Context, pollInterval time.Duration) {
	if atomic.LoadUint32(&hl.stateSyncedInitializationRun) == 1 {
		hl.Logger.Info("IsInitializationDone not finished... goroutine exists")

		return
	}

	nodeStatus, err := helper.GetNodeStatus(hl.cliCtx)
	if err != nil {
		hl.Logger.Error("Error while fetching heimdall node status", "error", err)

		return
	}

	toBlock := nodeStatus.SyncInfo.LatestBlockHeight
	toBlockTime := nodeStatus.SyncInfo.LatestBlockTime.Unix()

	eventProcessor := util.NewTokenMapProcessor(hl.cliCtx, hl.storageClient)
	if eventProcessor != nil {
		done, err := eventProcessor.IsInitializationDoneWithBlock(toBlock)
		if err != nil {
			hl.Logger.Error("Error check IsInitializationDone...skipping events query", "error", err)

			return
		}

		if !done {
			go hl.loadHistoricalEventRecords(ctx, eventProcessor, toBlock, toBlockTime, pollInterval)
		} else {
			hl.Logger.Info("Fetching new events between",
				"lastEventID", eventProcessor.TokenMapLastEventID,
				"endBlock", eventProcessor.TokenMapCheckedEndBlock,
				"toBlock", toBlock)
			hl.processEventRecords(ctx, eventProcessor, toBlock, toBlockTime, pollInterval)
		}
	} else {
		hl.Logger.Error("Error construct eventProcessor")
	}
}

func (hl *HeimdallListener) loadHistoricalEventRecords(ctx context.Context,
	eventProcessor *util.TokenMapProcessor,
	blockHeight, toBlockTime int64, pollInterval time.Duration,
) {
	if atomic.CompareAndSwapUint32(&hl.stateSyncedInitializationRun, 0, 1) {
		hl.Logger.Info("IsInitializationDone not finished... start goroutine")

		defer atomic.StoreUint32(&hl.stateSyncedInitializationRun, 0)

		hl.Logger.Info("Fetching all events between",
			"lastEventID", eventProcessor.TokenMapLastEventID,
			"endBlock", eventProcessor.TokenMapCheckedEndBlock,
			"toBlock", blockHeight)
		hl.processEventRecords(ctx, eventProcessor, blockHeight, toBlockTime, pollInterval)
	} else {
		hl.Logger.Info("IsInitializationDone not finished... goroutine exists")
	}
}

func (hl *HeimdallListener) processEventRecords(
	ctx context.Context,
	eventProcessor *util.TokenMapProcessor,
	blockHeight, toBlockTime int64, timeout time.Duration,
) {
	epochLen := 500
	circleLen := 50
	doneC := make(chan bool)
	events := make([]*clerkTypes.EventRecord, 0)

	go hl.getNEventRecords(eventProcessor, toBlockTime, circleLen, epochLen, &events, doneC)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

LOOPS:
	for {
		select {
		case <-doneC:
			if events != nil {
				finalEpoch := hl.processEvents(events, epochLen, blockHeight)
				if !finalEpoch {
					time.Sleep(time.Duration(1) * time.Second)
					go hl.getNEventRecords(eventProcessor, toBlockTime, circleLen, epochLen, &events, doneC)
					timer.Reset(timeout)
				} else {
					break LOOPS
				}
			} else {
				hl.Logger.Info("no events")

				break LOOPS
			}

		case <-timer.C:
			hl.Logger.Error("job timeout")

			return
		case <-ctx.Done():
			hl.Logger.Info("job stopped")

			return
		}
	}
}

func (hl *HeimdallListener) processEvents(events []*clerkTypes.EventRecord,
	epochLen int, blockHeight int64,
) bool {
	finalEpoch := false

	eventsLen := len(events)
	if eventsLen == 0 {
		defaultEvent := clerkTypes.NewEventRecord(
			htype.ZeroHeimdallHash, 0, 0,
			htype.ZeroHeimdallAddress,
			nil,
			"",
			time.Now(),
			"",
		)
		hl.processEvent(&defaultEvent, 1, blockHeight)

		return true
	}

	if eventsLen < epochLen {
		finalEpoch = true
	}

	for index, event := range events {
		if finalEpoch && eventsLen == index+1 {
			hl.processEvent(event, 1, blockHeight)

			return true
		}

		hl.processEvent(event, 0, blockHeight)
	}

	return false
}

func (hl *HeimdallListener) getNEventRecords(
	eventProcessor *util.TokenMapProcessor, toBlockTime int64,
	circleLen, epochLen int, events *[]*clerkTypes.EventRecord,
	doneC chan bool,
) {
	defer func() {
		doneC <- true
	}()

	*events = make([]*clerkTypes.EventRecord, 0)

	eventsTmp, err := eventProcessor.FetchStateSyncEventsWithTotal(toBlockTime, circleLen, epochLen)
	if err != nil {
		hl.Logger.Error("Fetch event failed", "fetcher", hl.name, "error", err)

		return
	}

	hl.Logger.Info("process event",
		"fromEventIndex", eventProcessor.TokenMapLastEventID,
		"fetcher", hl.name, "getEventSize", len(eventsTmp))

	*events = append(*events, eventsTmp...)
}

func (hl *HeimdallListener) processEvent(event *clerkTypes.EventRecord, final int32, blockHeight int64) {
	hl.Logger.Info("Received event record from Heimdall",
		"RootChainType", event.RootChainType,
		"time", event.RecordTime.UTC().String(),
		"EventID", event.ID,
		"TxHash", event.TxHash)

	eventBytes, err := json.Marshal(event)
	if err != nil {
		hl.Logger.Error("Error while parsing block event", "error", err, "TxHash", event.TxHash)

		return
	}

	hl.sendEventTask("processEventRecordFromHeimdall", eventBytes, hl.name, final, blockHeight)
}

func (hl *HeimdallListener) sendEventTask(taskName string,
	eventBytes []byte, chainType string, final int32,
	blockHeight int64,
) {
	// create machinery task
	signature := &tasks.Signature{
		Name: taskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: chainType,
			},
			{
				Type:  "string",
				Value: string(eventBytes),
			},
			{
				Type:  "int32",
				Value: final,
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
		"taskName", taskName, "eventBytes", eventBytes, "currentTime", time.Now())

	// send task
	_, err := hl.queueConnector.Server.SendTask(signature)
	if err != nil {
		hl.Logger.Error("Error sending block level task", "taskName", taskName, "error", err)
	}
}
