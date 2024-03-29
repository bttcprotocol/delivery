package processor

import (
	"encoding/json"

	"github.com/maticnetwork/heimdall/bridge/setu/util"

	clerkTypes "github.com/maticnetwork/heimdall/clerk/types"
)

// EventRecordProcessor - processor for record events.
type EventRecordProcessor struct {
	BaseProcessor
}

// Start - start processor.
func (rp *EventRecordProcessor) Start() error {
	rp.Logger.Info("Starting")

	return nil
}

// RegisterTasks - register tasks to server.
func (rp *EventRecordProcessor) RegisterTasks() {
	rp.Logger.Info("Registering record event tasks")

	err := rp.queueConnector.Server.RegisterTask(
		"processEventRecordFromHeimdall",
		rp.processEventRecordFromHeimdall)
	if err != nil {
		rp.Logger.Error("RegisterTasks | processEventRecordFromHeimdall", "error", err)
	}
}

func (rp *EventRecordProcessor) processEventRecordFromHeimdall(
	chainType string, txBytes string, final int32, blockHeight int64,
) error {
	rp.Logger.Info("Received processEventRecordFromHeimdall request", "txBytes", txBytes)

	event := clerkTypes.EventRecord{}
	if err := json.Unmarshal([]byte(txBytes), &event); err != nil {
		rp.Logger.Error("Error unmarshalling event from heimdall", "error", err)

		return err
	}

	eventProcessor := util.NewTokenMapProcessor(rp.cliCtx, rp.storageClient)
	if eventProcessor == nil {
		rp.Logger.Error("Error init eventProcessor",
			"nowEventID", event.ID)

		return nil
	}

	eventProcessor.LockForUpdate()
	defer eventProcessor.UnLockForUpdate()

	if event.ID > 0 {
		if event.ID != eventProcessor.TokenMapLastEventID+1 {
			rp.Logger.Error("Error event ID",
				"lastEventID", eventProcessor.TokenMapLastEventID,
				"nowEventID", event.ID)

			return nil
		}

		rp.Logger.Info("Process event ID",
			"lastEventID", eventProcessor.TokenMapLastEventID,
			"nowEventID", event.ID,
			"final", final)

		if err := rp.processStateSyncedEvent(eventProcessor, &event, chainType); err != nil {
			return err
		}

		if final == 1 {
			if err := eventProcessor.FlushCacheTokenMapItem(); err != nil {
				rp.Logger.Error("Error update token map hash to db", "error", err)

				return err
			}

			if err := eventProcessor.UpdateHash(); err != nil {
				rp.Logger.Error("Error update token map hash to db", "error", err)

				return err
			}

			if err := eventProcessor.UpdateTokenMapLastEventID(event.ID); err != nil {
				rp.Logger.Error("Error update last event id to db", "error", err)

				return err
			}

			if err := eventProcessor.LoadLastEventIDFromDB(); err != nil {
				rp.Logger.Error("Error update last event id from db", "error", err)

				return err
			}

			// only for test
			mmap, _ := eventProcessor.GetTokenMap()
			for k, v := range mmap {
				for _, item := range v {
					rp.Logger.Info("map", "root", k, "item", item)
				}
			}
		}
	}

	if final == 1 {
		if err := eventProcessor.UpdateTokenMapCheckedEndBlock(blockHeight); err != nil {
			rp.Logger.Error("Error update token map checked end block to db", "error", err)

			return err
		}
	}

	return nil
}

func (rp *EventRecordProcessor) processStateSyncedEvent(
	eventProcessor *util.TokenMapProcessor, event *clerkTypes.EventRecord,
	chainType string,
) error {
	eventSynced := util.NewStateSyncedEvent(event, chainType)
	if err := eventSynced.ParshEvent(); err != nil {
		rp.Logger.Error("Error parse event from heimdall", "error", err)

		return err
	}

	if eventSynced.IsMapTokenEvent() {
		item, err := eventSynced.ConvertToTokenMap()
		if err != nil {
			rp.Logger.Error("Error parse event from heimdall", "error", err)

			return err
		}

		rp.Logger.Info("Process event, chached", "eventID", event.ID)

		if err := eventProcessor.CacheTokenMapItem(item); err != nil {
			rp.Logger.Error("Error cache token map item to db", "error", err)

			return err
		}
	}

	eventProcessor.TokenMapLastEventID = event.ID

	return nil
}
