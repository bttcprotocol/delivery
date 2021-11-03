package processor

import (
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/bor/common"
	"github.com/maticnetwork/bor/core/types"
	authTypes "github.com/maticnetwork/heimdall/auth/types"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	checkpointTypes "github.com/maticnetwork/heimdall/checkpoint/types"
	"github.com/maticnetwork/heimdall/helper"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// handleCheckpointSync - Checkpoint Sync handler
// 1. Fetch CheckpointSync from buffer
// 2. check if elapsed.
// 3. Send CheckpointSync to heimdall if required.
func (cp *CheckpointProcessor) handleCheckpointSync() {
	currentTime := time.Now().UTC()

	for rootChain := range hmTypes.GetRootChainIDMap() {
		if rootChain == hmTypes.RootChainTypeStake {
			continue
		}
		// fetch fresh checkpoint context
		checkpointContext, err := cp.getCheckpointContext(rootChain)
		if err != nil {
			return
		}
		checkpointParams := checkpointContext.CheckpointParams
		bufferedCheckpoint, err := util.GetBufferedCheckpointSync(cp.cliCtx, rootChain)
		if err == nil {
			bufferedTime := time.Unix(int64(bufferedCheckpoint.TimeStamp), 0)
			if currentTime.Sub(bufferedTime).Seconds() < checkpointParams.CheckpointBufferTime.Seconds()/5 {
				cp.Logger.Debug("Cannot send multiple checkpoint sync in short time", "error", err)
				continue
			}
		}

		// fetch latest syncHeaderBlock from stakeChain
		lastSyncedCheckpointNumber, err := cp.getLastSyncedCheckpointNumber(checkpointContext, rootChain)
		if err != nil {
			cp.Logger.Error("Error fetching syncedHeaderNumber from stakeChain", "root", rootChain, "error", err)
			continue
		}
		nextCheckpointNumber := lastSyncedCheckpointNumber + 1
		// fetch nextHeaderBlock from rootChain
		start, end, proposer, err := cp.getHeaderBlock(checkpointContext, nextCheckpointNumber, rootChain)
		if err != nil {
			cp.Logger.Error("Error fetching currentHeaderBlock from rootChain", "root", rootChain, "error", err)
			continue
		}

		if end != 0 {
			// send checkpoint sync
			msg := checkpointTypes.NewMsgCheckpointSync(hmTypes.BytesToHeimdallAddress(helper.GetAddress()),
				proposer, nextCheckpointNumber, start, end, rootChain)
			// return broadcast to heimdall
			if err := cp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
				cp.Logger.Error("Error while broadcasting checkpoint-sync to heimdall",
					"root", rootChain, "number", nextCheckpointNumber, "start", start, "end", end, "error", err)
				continue
			}
			cp.Logger.Info("checkpoint sync transaction sent successfully",
				"root", rootChain, "number", nextCheckpointNumber, "start", start, "end", end)
		}
	}
}

// sendCheckpointSyncToStakeChain - handles checkpoint sync confirmation event from heimdall.
// 1. check if i am the current proposer.
// 2. check if this checkpoint has to be submitted to stake chain
// 3. if so, create and broadcast checkpoint transaction to stake chain
func (cp *CheckpointProcessor) sendCheckpointSyncToStakeChain(eventBytes string, blockHeight int64) error {
	cp.Logger.Info("Received sendCheckpointSyncToStakeChain request", "eventBytes", eventBytes, "blockHeight", blockHeight)
	var event = sdk.StringEvent{}
	if err := json.Unmarshal([]byte(eventBytes), &event); err != nil {
		cp.Logger.Error("Error unmarshalling event from heimdall", "error", err)
		return err
	}

	cp.Logger.Info("processing checkpoint sync confirmation event", "event", event.Type)
	isCurrentProposer, err := util.IsCurrentProposer(cp.cliCtx)
	if err != nil {
		cp.Logger.Error("Error checking isCurrentProposer in CheckpointConfirmation handler", "error", err)
		return err
	}
	if !isCurrentProposer {
		cp.Logger.Info("I am not the current proposer. Ignoring", "eventType", event.Type)
		return nil
	}

	var (
		startBlock, endBlock, number uint64
		txHash                       string
		rootChain                    string
	)

	for _, attr := range event.Attributes {
		if attr.Key == checkpointTypes.AttributeKeyStartBlock {
			startBlock, _ = strconv.ParseUint(attr.Value, 10, 64)
		}
		if attr.Key == checkpointTypes.AttributeKeyEndBlock {
			endBlock, _ = strconv.ParseUint(attr.Value, 10, 64)
		}
		if attr.Key == hmTypes.AttributeKeyTxHash {
			txHash = attr.Value
		}
		if attr.Key == checkpointTypes.AttributeKeyRootChain {
			rootChain = attr.Value
		}
		if attr.Key == checkpointTypes.AttributeKeyHeaderIndex {
			number, _ = strconv.ParseUint(attr.Value, 10, 64)
		}
	}
	cp.Logger.Info("Received sendCheckpointSyncToStakeChain request", "number", number, "root", rootChain)

	checkpointContext, err := cp.getCheckpointContext(rootChain)
	if err != nil {
		return err
	}

	// fetch latest syncHeaderBlock from stakeChain
	lastSyncedHeaderNumber, err := cp.getLastSyncedCheckpointNumber(checkpointContext, rootChain)
	if err != nil {
		cp.Logger.Error("Error fetching syncedHeaderNumber from stakeChain", "root", rootChain, "error", err)
		return err
	}
	if number != lastSyncedHeaderNumber+1 {
		cp.Logger.Error("Checkpoint sync number mismatch", "eventType", event.Type,
			"number", number, "lastSynced", lastSyncedHeaderNumber)
	} else {
		txHash := common.FromHex(txHash)
		if err := cp.createAndSendCheckpointSyncToTron(checkpointContext, number, startBlock, endBlock, rootChain, blockHeight, txHash); err != nil {
			cp.Logger.Error("Error sending checkpoint to rootchain", "error", err)
			return err
		}
	}
	return nil
}

// sendCheckpointSyncAckToHeimdall - handles checkpointSyncAck event from stake chain
// 1. create and broadcast checkpointSyncAck msg to heimdall.
func (cp *CheckpointProcessor) sendCheckpointSyncAckToHeimdall(eventName string, checkpointSyncAckStr string, rootChain string) error {
	// fetch fresh checkpoint context
	checkpointContext, err := cp.getCheckpointContext(rootChain)
	if err != nil {
		return err
	}
	checkpointParams := checkpointContext.CheckpointParams

	var log = types.Log{}
	if err := json.Unmarshal([]byte(checkpointSyncAckStr), &log); err != nil {
		cp.Logger.Error("Error while unmarshalling event from stake chain", "error", err)
		return err
	}

	event := new(struct {
		Proposer     common.Address
		ChainId      *big.Int
		CheckpointId *big.Int
		Start        *big.Int
		End          *big.Int
	})
	if err := helper.UnpackLog(cp.stakingInfoAbi, event, eventName, &log); err != nil {
		cp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		checkpointChain := hmTypes.GetRootChainName(event.ChainId.Uint64())

		cp.Logger.Info(
			"✅ Received task to send checkpoint-sync-ack to heimdall",
			"event", eventName,
			"start", event.Start,
			"end", event.End,
			"root", checkpointChain,
			"proposer", event.Proposer.Hex(),
			"checkpointNumber", event.CheckpointId,
		)
		// fetch checkpoint sync buffer
		bufferedCheckpoint, err := util.GetBufferedCheckpointSync(cp.cliCtx, checkpointChain)
		if err == nil {
			bufferedTime := time.Unix(int64(bufferedCheckpoint.TimeStamp), 0)
			currentTime := time.Now().UTC()
			if currentTime.Sub(bufferedTime).Seconds() > checkpointParams.CheckpointBufferTime.Seconds()/5 {
				cp.Logger.Debug("checkpoint sync buffer has expired, ignore this ack")
				return nil
			}
		}

		// create msg checkpoint ack message
		msg := checkpointTypes.NewMsgCheckpointSyncAck(
			helper.GetFromAddress(cp.cliCtx),
			event.CheckpointId.Uint64(),
			event.Start.Uint64(),
			event.End.Uint64(),
			checkpointChain,
		)

		// return broadcast to heimdall
		if err := cp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
			cp.Logger.Error("Error while broadcasting checkpoint-ack to heimdall", "error", err)
			return err
		}
	}
	return nil
}

// getHeaderBlock - get header block info from rootchain
func (cp *CheckpointProcessor) getHeaderBlock(checkpointContext *CheckpointContext, headerNumber uint64, rootChain string) (start, end uint64, proposer hmTypes.HeimdallAddress, err error) {
	// get chain params
	chainParams := checkpointContext.ChainmanagerParams.ChainParams
	checkpointParams := checkpointContext.CheckpointParams

	switch rootChain {
	case hmTypes.RootChainTypeEth, hmTypes.RootChainTypeBsc:
		rootChainInstance, err := cp.contractConnector.GetRootChainInstance(chainParams.RootChainAddress.EthAddress(), rootChain)
		if err != nil {
			return 0, 0, hmTypes.ZeroHeimdallAddress, err
		}
		_, start, end, _, proposer, err := cp.contractConnector.GetHeaderInfo(headerNumber, rootChainInstance, checkpointParams.ChildBlockInterval)
		if err != nil {
			cp.Logger.Error("Error while fetching header block", "root", rootChain, "error", err)
			return 0, 0, hmTypes.ZeroHeimdallAddress, err
		}
		return start, end, proposer, nil
	case hmTypes.RootChainTypeTron:
		_, start, end, _, proposer, err := cp.contractConnector.GetTronHeaderInfo(headerNumber, chainParams.TronChainAddress, checkpointParams.ChildBlockInterval)

		if err != nil {
			cp.Logger.Error("Error while fetching header block", "root", rootChain, "error", err)
			return 0, 0, hmTypes.ZeroHeimdallAddress, err
		}
		return start, end, proposer, nil
	}
	return 0, 0, hmTypes.ZeroHeimdallAddress, nil
}

// getLastSyncedCheckpointNumber - get last checkpoint header number from stake chain
func (cp *CheckpointProcessor) getLastSyncedCheckpointNumber(checkpointContext *CheckpointContext, rootChain string) (uint64, error) {
	chainParams := checkpointContext.ChainmanagerParams.ChainParams
	syncedHeaderNumber, err := cp.contractConnector.GetSyncedCheckpointId(chainParams.TronStakingManagerAddress, rootChain)
	if err != nil {
		cp.Logger.Error("Error while fetching current synced header block number from stake chain", "error", err)
		return 0, err
	}
	return syncedHeaderNumber, nil
}

// createAndSendCheckpointSyncToTron prepares the data required for checkpoint sync submission
// and sends a transaction to tron
func (cp *CheckpointProcessor) createAndSendCheckpointSyncToTron(
	checkpointContext *CheckpointContext, number, start, end uint64, rootChain string, height int64, txHash []byte) error {
	cp.Logger.Info("Preparing checkpoint sync to be pushed on tron",
		"height", height, "txHash", hmTypes.BytesToHeimdallHash(txHash), "number", number, "root", rootChain, "start", start, "end", end)
	// proof
	tx, err := helper.QueryTxWithProof(cp.cliCtx, txHash)
	if err != nil {
		cp.Logger.Error("Error querying checkpoint sync tx proof", "txHash", txHash)
		return err
	}

	// fetch side txs sigs
	decoder := helper.GetTxDecoder(authTypes.ModuleCdc)
	stdTx, err := decoder(tx.Tx)
	if err != nil {
		cp.Logger.Error("Error while decoding checkpoint sync tx", "txHash", tx.Tx.Hash(), "error", err)
		return err
	}

	msg := stdTx.GetMsgs()[0]
	sideMsg, ok := msg.(hmTypes.SideTxMsg)
	if !ok {
		cp.Logger.Error("Invalid side-tx msg checkpoint sync", "txHash", tx.Tx.Hash())
		return err
	}

	// side-tx data
	sideTxData := sideMsg.GetSideSignBytes()

	// get sigs
	sigs, err := helper.FetchSideTxSigs(cp.httpClient, height, tx.Tx.Hash(), sideTxData)
	if err != nil {
		cp.Logger.Error("Error fetching votes for checkpoint sync tx", "height", height)
		return err
	}

	// chain manager params
	chainParams := checkpointContext.ChainmanagerParams.ChainParams
	err = cp.contractConnector.SendCheckpointSyncToTron(sideTxData, sigs, chainParams.TronStakingManagerAddress)
	if err != nil {
		cp.Logger.Error("Error submitting checkpoint sync to tron", "error", err)
		return err
	}
	cp.Logger.Info("Submitted new checkpoint sync to tron successfully")

	return nil
}
