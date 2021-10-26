package processor

import (
	"math/big"
	"time"

	authTypes "github.com/maticnetwork/heimdall/auth/types"

	hmTypes "github.com/maticnetwork/heimdall/types"

	checkpointTypes "github.com/maticnetwork/heimdall/checkpoint/types"
	"github.com/maticnetwork/heimdall/helper"

	"github.com/maticnetwork/heimdall/bridge/setu/util"
)

func (cp *CheckpointProcessor) sendTronCheckpointToHeimdall(checkpointContext *CheckpointContext, latestConfirmedChildBlock uint64) {
	expectedCheckpointState, err := cp.nextExpectedTronCheckpoint(checkpointContext, latestConfirmedChildBlock)
	if err != nil {
		cp.Logger.Error("Error while calculate next expected checkpoint[tron]", "error", err)
		return
	}
	start := expectedCheckpointState.newStart
	end := expectedCheckpointState.newEnd

	//
	// Check checkpoint buffer
	//
	timeStamp := uint64(time.Now().Unix())
	checkpointBufferTime := uint64(checkpointContext.CheckpointParams.CheckpointBufferTime.Seconds())

	bufferedCheckpoint, err := util.GetBufferedCheckpoint(cp.cliCtx, hmTypes.RootChainTypeTron)
	if err != nil {
		cp.Logger.Debug("No buffered checkpoint[tron]", "bufferedCheckpoint", bufferedCheckpoint)
	}

	if bufferedCheckpoint != nil && !(bufferedCheckpoint.TimeStamp == 0 || ((timeStamp > bufferedCheckpoint.TimeStamp) && timeStamp-bufferedCheckpoint.TimeStamp >= checkpointBufferTime)) {
		cp.Logger.Info("Checkpoint[tron] already exists in buffer", "Checkpoint", bufferedCheckpoint.String())
		return
	}

	if err := cp.createAndSendTronCheckpointToHeimdall(checkpointContext, start, end); err != nil {
		cp.Logger.Error("Error sending checkpoint[tron] to heimdall", "error", err)
		return
	}
}

// nextExpectedTronCheckpoint - fetched contract checkpoint state and returns the next probable checkpoint that needs to be sent
func (cp *CheckpointProcessor) nextExpectedTronCheckpoint(checkpointContext *CheckpointContext, latestChildBlock uint64) (*ContractCheckpoint, error) {
	chainManagerParams := checkpointContext.ChainmanagerParams
	checkpointParams := checkpointContext.CheckpointParams

	// fetch current header block from tron contract
	_currentHeaderBlock, err := cp.contractConnector.TronChainRPC.CurrentHeaderBlock(chainManagerParams.ChainParams.TronChainAddress, checkpointParams.ChildBlockInterval)
	if err != nil {
		cp.Logger.Error("Error while fetching current header block number from tron", "error", err)
		return nil, err
	}

	// current header block
	currentHeaderBlockNumber := big.NewInt(0).SetUint64(_currentHeaderBlock)

	// get header info
	_, currentStart, currentEnd, lastCheckpointTime, _, err := cp.contractConnector.GetTronHeaderInfo(
		currentHeaderBlockNumber.Uint64(), chainManagerParams.ChainParams.TronChainAddress, checkpointParams.ChildBlockInterval)
	if err != nil {
		cp.Logger.Error("Error while fetching current header block object from tron", "error", err)
		return nil, err
	}

	// find next start/end
	var start, end uint64
	start = currentEnd

	// add 1 if start > 0
	if start > 0 {
		start = start + 1
	}

	// get diff
	diff := latestChildBlock - start + 1
	// process if diff > 0 (positive)
	if diff > 0 {
		expectedDiff := diff - diff%checkpointParams.AvgCheckpointLength
		if expectedDiff > 0 {
			expectedDiff = expectedDiff - 1
		}
		// cap with max checkpoint length
		if expectedDiff > checkpointParams.MaxCheckpointLength-1 {
			expectedDiff = checkpointParams.MaxCheckpointLength - 1
		}
		// get end result
		end = expectedDiff + start
		cp.Logger.Debug("Calculating checkpoint[tron] eligibility",
			"latest", latestChildBlock,
			"start", start,
			"end", end,
		)
	}
	// Handle when block producers go down
	if end == 0 || end == start || (0 < diff && diff < checkpointParams.AvgCheckpointLength) {
		cp.Logger.Debug("Fetching last header block to calculate time")

		currentTime := time.Now().UTC().Unix()
		defaultForcePushInterval := checkpointParams.MaxCheckpointLength * 2 // in seconds (1024 * 2 seconds)
		if currentTime-int64(lastCheckpointTime) > int64(defaultForcePushInterval) {
			end = latestChildBlock
			cp.Logger.Info("Force push checkpoint",
				"currentTime", currentTime,
				"lastCheckpointTime", lastCheckpointTime,
				"defaultForcePushInterval", defaultForcePushInterval,
				"start", start,
				"end", end,
			)
		}
	}
	// if end == 0 || start >= end {
	// 	c.Logger.Info("Waiting for 256 blocks or invalid start end formation", "start", start, "end", end)
	// 	return nil, errors.New("Invalid start end formation")
	// }
	return NewContractCheckpoint(start, end, &HeaderBlock{
		start:  currentStart,
		end:    currentEnd,
		number: currentHeaderBlockNumber,
	}), nil
}

// sendCheckpointToHeimdall - creates checkpoint msg and broadcasts to heimdall
func (cp *CheckpointProcessor) createAndSendTronCheckpointToHeimdall(checkpointContext *CheckpointContext, start uint64, end uint64) error {
	cp.Logger.Debug("Initiating checkpoint[tron] to Heimdall", "start", start, "end", end)

	if end == 0 || start >= end {
		cp.Logger.Info("Waiting[tron] for blocks or invalid start end formation", "start", start, "end", end)
		return nil
	}
	// fetch latest checkpoint
	latestCheckpoint, err := util.GetlastestCheckpoint(cp.cliCtx, hmTypes.RootChainTypeTron)
	// event checkpoint is older than or equal to latest checkpoint
	if err == nil && latestCheckpoint != nil && latestCheckpoint.EndBlock+1 < start {
		cp.Logger.Debug("Need to resubmit Checkpoint ack first", "start", start, "last_end", latestCheckpoint.EndBlock)
		err := cp.resubmitTronCheckpointAck(checkpointContext)
		if err != nil {
			cp.Logger.Info("Error while resubmit checkpoint ack", "root", hmTypes.RootChainTypeTron, "err", err)
			return err
		}
		return nil
	}
	// get checkpoint params
	checkpointParams := checkpointContext.CheckpointParams

	// Get root hash
	root, err := cp.contractConnector.GetRootHash(start, end, checkpointParams.MaxCheckpointLength)
	if err != nil {
		return err
	}
	cp.Logger.Info("[tron]Root hash calculated", "rootHash", hmTypes.BytesToHeimdallHash(root))
	var accountRootHash hmTypes.HeimdallHash
	//Get DividendAccountRoot from HeimdallServer
	if accountRootHash, err = cp.fetchDividendAccountRoot(); err != nil {
		cp.Logger.Info("Error while fetching initial account root hash from HeimdallServer[tron]", "err", err)
		return err
	}

	chainParams := checkpointContext.ChainmanagerParams.ChainParams

	cp.Logger.Info("✅ Creating and broadcasting new checkpoint",
		"proposer", hmTypes.BytesToHeimdallAddress(helper.GetAddress()),
		"start", start,
		"end", end,
		"root", hmTypes.BytesToHeimdallHash(root),
		"accountRoot", accountRootHash,
		"borChainId", chainParams.BorChainID,
		"epoch", cp.getCurrentEpoch(),
		"rootType", hmTypes.RootChainTypeTron,
	)

	// create and send checkpoint message
	msg := checkpointTypes.NewMsgCheckpointBlock(
		hmTypes.BytesToHeimdallAddress(helper.GetAddress()),
		start,
		end,
		hmTypes.BytesToHeimdallHash(root),
		accountRootHash,
		chainParams.BorChainID,
		cp.getCurrentEpoch(),
		hmTypes.RootChainTypeTron,
	)

	// return broadcast to heimdall
	if err := cp.txBroadcaster.BroadcastToHeimdall(&msg); err != nil {
		cp.Logger.Error("Error while broadcasting checkpoint[tron] to heimdall", "error", err)
		return err
	}

	return nil
}

// shouldSendTronCheckpoint checks if checkpoint with given start,end should be sent to rootchain or not.
func (cp *CheckpointProcessor) shouldSendTronCheckpoint(checkpointContext *CheckpointContext, start uint64, end uint64) (bool, error) {
	chainManagerParams := checkpointContext.ChainmanagerParams

	// current child block from contract
	currentChildBlock, err := cp.contractConnector.TronChainRPC.GetLastChildBlock(chainManagerParams.ChainParams.TronChainAddress)
	if err != nil {
		cp.Logger.Error("Error fetching tron current child block", "currentChildBlock", currentChildBlock, "error", err)
		return false, err
	}

	cp.Logger.Debug("Fetched tron current child block", "currentChildBlock", currentChildBlock)

	shouldSend := false
	// validate if checkpoint needs to be pushed to tron and submit
	cp.Logger.Info("Validating if checkpoint[tron] needs to be pushed", "commitedLastBlock", currentChildBlock, "startBlock", start)
	// check if we need to send checkpoint or not
	if ((currentChildBlock + 1) == start) || (currentChildBlock == 0 && start == 0) {
		cp.Logger.Info("Checkpoint[tron] Valid", "startBlock", start)
		shouldSend = true
	} else if currentChildBlock > start {
		cp.Logger.Info("Start block does not match, checkpoint[tron] already sent", "commitedLastBlock", currentChildBlock, "startBlock", start)
	} else if currentChildBlock > end {
		cp.Logger.Info("Checkpoint[tron] already sent", "commitedLastBlock", currentChildBlock, "startBlock", start)
	} else {
		cp.Logger.Info("No need to send checkpoint[tron]")
	}

	return shouldSend, nil
}

// createAndSendCheckpointToTron prepares the data required for rootchain checkpoint submission
// and sends a transaction to tron
func (cp *CheckpointProcessor) createAndSendCheckpointToTron(checkpointContext *CheckpointContext, start uint64, end uint64, height int64, txHash []byte) error {
	cp.Logger.Info("Preparing checkpoint[tron] to be pushed on chain", "height", height, "txHash", hmTypes.BytesToHeimdallHash(txHash), "start", start, "end", end)
	// proof
	tx, err := helper.QueryTxWithProof(cp.cliCtx, txHash)
	if err != nil {
		cp.Logger.Error("Error querying checkpoint[tron] tx proof", "txHash", txHash)
		return err
	}

	// fetch side txs sigs
	decoder := helper.GetTxDecoder(authTypes.ModuleCdc)
	stdTx, err := decoder(tx.Tx)
	if err != nil {
		cp.Logger.Error("Error while decoding checkpoint[tron] tx", "txHash", tx.Tx.Hash(), "error", err)
		return err
	}

	msg := stdTx.GetMsgs()[0]
	sideMsg, ok := msg.(hmTypes.SideTxMsg)
	if !ok {
		cp.Logger.Error("Invalid side-tx msg[tron]", "txHash", tx.Tx.Hash())
		return err
	}

	// side-tx data
	sideTxData := sideMsg.GetSideSignBytes()

	// get sigs
	sigs, err := helper.FetchSideTxSigs(cp.httpClient, height, tx.Tx.Hash(), sideTxData)
	if err != nil {
		cp.Logger.Error("Error fetching votes for checkpoint[tron] tx", "height", height)
		return err
	}

	shouldSend, err := cp.shouldSendTronCheckpoint(checkpointContext, start, end)
	if err != nil {
		return err
	}

	if shouldSend {
		// chain manager params
		chainParams := checkpointContext.ChainmanagerParams.ChainParams
		err := cp.contractConnector.SendTronCheckpoint(sideTxData, sigs, chainParams.TronChainAddress)
		if err != nil {
			cp.Logger.Error("Error submitting checkpoint[tron] to rootchain", "error", err)
			return err
		}
		cp.Logger.Info("Submitted new checkpoint[tron] to rootchain successfully")
	}

	return nil
}

// getLatestTronCheckpointTime - get latest checkpoint time from rootchain
func (cp *CheckpointProcessor) getLatestTronCheckpointTime(checkpointContext *CheckpointContext) (int64, error) {
	// get chain params
	chainParams := checkpointContext.ChainmanagerParams.ChainParams
	checkpointParams := checkpointContext.CheckpointParams

	// fetch last header number
	lastHeaderNumber, err := cp.contractConnector.TronChainRPC.CurrentHeaderBlock(chainParams.TronChainAddress, checkpointParams.ChildBlockInterval)
	if err != nil {
		cp.Logger.Error("Error while fetching current header block number", "error", err)
		return 0, err
	}

	// header block
	_, _, _, createdAt, _, err := cp.contractConnector.GetTronHeaderInfo(lastHeaderNumber, chainParams.TronChainAddress, checkpointParams.ChildBlockInterval)
	if err != nil {
		cp.Logger.Error("Error while fetching header block object", "error", err)
		return 0, err
	}
	return int64(createdAt), nil
}

// resubmitTronCheckpointAck - resubmit checkpoint ack of root
func (cp *CheckpointProcessor) resubmitTronCheckpointAck(checkpointContext *CheckpointContext) error {
	// get chain params
	chainParams := checkpointContext.ChainmanagerParams.ChainParams
	checkpointParams := checkpointContext.CheckpointParams

	// fetch last header number
	lastHeaderNumber, err := cp.contractConnector.TronChainRPC.CurrentHeaderBlock(chainParams.TronChainAddress, checkpointParams.ChildBlockInterval)
	if err != nil {
		cp.Logger.Error("Error while fetching current header block number", "error", err)
		return err
	}

	// header block
	root, start, end, _, proposer, err := cp.contractConnector.GetTronHeaderInfo(lastHeaderNumber, chainParams.TronChainAddress, checkpointParams.ChildBlockInterval)
	if err != nil {
		cp.Logger.Error("Error while fetching header block object", "error", err)
		return err
	}
	// create msg checkpoint ack message
	msg := checkpointTypes.NewMsgCheckpointAck(
		helper.GetFromAddress(cp.cliCtx),
		lastHeaderNumber,
		proposer,
		start,
		end,
		hmTypes.BytesToHeimdallHash(root.Bytes()),
		hmTypes.ZeroHeimdallHash,
		0,
		hmTypes.RootChainTypeTron,
	)

	// return broadcast to heimdall
	if err := cp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
		cp.Logger.Error("Error while broadcasting checkpoint to heimdall", "error", err)
		return err
	}

	return nil
}
