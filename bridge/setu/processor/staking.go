package processor

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	authTypes "github.com/maticnetwork/heimdall/auth/types"

	"github.com/RichardKnop/machinery/v1/tasks"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/bor/accounts/abi"
	"github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	chainmanagerTypes "github.com/maticnetwork/heimdall/chainmanager/types"
	"github.com/maticnetwork/heimdall/contracts/rootchain"
	"github.com/maticnetwork/heimdall/contracts/stakinginfo"
	"github.com/maticnetwork/heimdall/helper"
	stakingTypes "github.com/maticnetwork/heimdall/staking/types"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// StakingProcessor - process staking related events
type StakingProcessor struct {
	BaseProcessor

	// header listener subscription
	cancelStakingService context.CancelFunc

	stakingInfoAbi *abi.ABI
}

// StakingContext represents checkpoint context
type StakingContext struct {
	ChainmanagerParams *chainmanagerTypes.Params
	StakingParams      *stakingTypes.Params
}

// NewStakingProcessor - add  abi to staking processor
func NewStakingProcessor(stakingInfoAbi *abi.ABI) *StakingProcessor {
	stakingProcessor := &StakingProcessor{
		stakingInfoAbi: stakingInfoAbi,
	}
	return stakingProcessor
}

// Start starts new block subscription
func (sp *StakingProcessor) Start() error {
	sp.Logger.Info("Starting")
	// no-ack
	ackCtx, cancelStakingService := context.WithCancel(context.Background())
	sp.cancelStakingService = cancelStakingService
	sp.Logger.Info("Start polling for staking sync", "pollInterval", helper.GetConfig().StakingPollInterval)
	go sp.startPolling(ackCtx, helper.GetConfig().StakingPollInterval)
	return nil
}

// Stop stops all necessary go routines
func (sp *StakingProcessor) Stop() {
	sp.cancelStakingService()
}

// RegisterTasks - Registers staking tasks with machinery
func (sp *StakingProcessor) RegisterTasks() {
	sp.Logger.Info("Registering staking related tasks")
	if err := sp.queueConnector.Server.RegisterTask("sendValidatorJoinToHeimdall", sp.sendValidatorJoinToHeimdall); err != nil {
		sp.Logger.Error("RegisterTasks | sendValidatorJoinToHeimdall", "error", err)
	}
	if err := sp.queueConnector.Server.RegisterTask("sendUnstakeInitToHeimdall", sp.sendUnstakeInitToHeimdall); err != nil {
		sp.Logger.Error("RegisterTasks | sendUnstakeInitToHeimdall", "error", err)
	}
	if err := sp.queueConnector.Server.RegisterTask("sendStakeUpdateToHeimdall", sp.sendStakeUpdateToHeimdall); err != nil {
		sp.Logger.Error("RegisterTasks | sendStakeUpdateToHeimdall", "error", err)
	}
	if err := sp.queueConnector.Server.RegisterTask("sendSignerChangeToHeimdall", sp.sendSignerChangeToHeimdall); err != nil {
		sp.Logger.Error("RegisterTasks | sendSignerChangeToHeimdall", "error", err)
	}
	if err := sp.queueConnector.Server.RegisterTask("sendStakingSyncToHeimdall", sp.sendStakingSyncToHeimdall); err != nil {
		sp.Logger.Error("RegisterTasks | sendStakingSyncToHeimdall", "error", err)
	}
	if err := sp.queueConnector.Server.RegisterTask("sendStakingSyncToRootChain", sp.sendStakingSyncToRootchain); err != nil {
		sp.Logger.Error("RegisterTasks | sendStakingSyncToRootChain", "error", err)
	}
	if err := sp.queueConnector.Server.RegisterTask("sendStakingAckToHeimdall", sp.sendStakingAckToHeimdall); err != nil {
		sp.Logger.Error("RegisterTasks | sendStakingAckToHeimdall", "error", err)
	}
}

func (sp *StakingProcessor) sendValidatorJoinToHeimdall(eventName string, logBytes string) error {
	var vLog = types.Log{}
	if err := json.Unmarshal([]byte(logBytes), &vLog); err != nil {
		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	event := new(stakinginfo.StakinginfoStaked)
	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &vLog); err != nil {
		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		signerPubKey := event.SignerPubkey
		if len(signerPubKey) == 64 {
			signerPubKey = util.AppendPrefix(signerPubKey)
		}
		if isOld, _ := sp.isOldTx(sp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index)); isOld {
			sp.Logger.Info("Ignoring task to send validatorjoin to heimdall as already processed",
				"event", eventName,
				"validatorID", event.ValidatorId,
				"activationEpoch", event.ActivationEpoch,
				"nonce", event.Nonce,
				"amount", event.Amount,
				"totalAmount", event.Total,
				"SignerPubkey", hmTypes.NewPubKey(signerPubKey).String(),
				"txHash", hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
				"logIndex", uint64(vLog.Index),
				"blockNumber", vLog.BlockNumber,
			)
			return nil
		}

		// if account doesn't exists Retry with delay for topup to process first.
		if _, err := util.GetAccount(sp.cliCtx, hmTypes.HeimdallAddress(event.Signer)); err != nil {
			sp.Logger.Info(
				"Heimdall Account doesn't exist. Retrying validator-join after 10 seconds",
				"event", eventName,
				"signer", event.Signer,
			)
			return tasks.NewErrRetryTaskLater("account doesn't exist", util.RetryTaskDelay)
		}

		sp.Logger.Info(
			"✅ Received task to send validatorjoin to heimdall",
			"event", eventName,
			"validatorID", event.ValidatorId,
			"activationEpoch", event.ActivationEpoch,
			"nonce", event.Nonce,
			"amount", event.Amount,
			"totalAmount", event.Total,
			"SignerPubkey", hmTypes.NewPubKey(signerPubKey).String(),
			"txHash", hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
			"logIndex", uint64(vLog.Index),
			"blockNumber", vLog.BlockNumber,
		)

		// msg validator exit
		msg := stakingTypes.NewMsgValidatorJoin(
			hmTypes.BytesToHeimdallAddress(helper.GetAddress()),
			event.ValidatorId.Uint64(),
			event.ActivationEpoch.Uint64(),
			sdk.NewIntFromBigInt(event.Amount),
			hmTypes.NewPubKey(signerPubKey),
			hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
			uint64(vLog.Index),
			vLog.BlockNumber,
			event.Nonce.Uint64(),
		)

		// return broadcast to heimdall
		if err := sp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
			sp.Logger.Error("Error while broadcasting unstakeInit to heimdall", "validatorId", event.ValidatorId.Uint64(), "error", err)
			return err
		}
	}
	return nil
}

func (sp *StakingProcessor) sendUnstakeInitToHeimdall(eventName string, logBytes string) error {
	var vLog = types.Log{}
	if err := json.Unmarshal([]byte(logBytes), &vLog); err != nil {
		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	event := new(stakinginfo.StakinginfoUnstakeInit)
	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &vLog); err != nil {
		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		if isOld, _ := sp.isOldTx(sp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index)); isOld {
			sp.Logger.Info("Ignoring task to send unstakeinit to heimdall as already processed",
				"event", eventName,
				"validator", event.User,
				"validatorID", event.ValidatorId,
				"nonce", event.Nonce,
				"deactivatonEpoch", event.DeactivationEpoch,
				"amount", event.Amount,
				"txHash", hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
				"logIndex", uint64(vLog.Index),
				"blockNumber", vLog.BlockNumber,
			)
			return nil
		}

		sp.Logger.Info(
			"✅ Received task to send unstake-init to heimdall",
			"event", eventName,
			"validator", event.User,
			"validatorID", event.ValidatorId,
			"nonce", event.Nonce,
			"deactivatonEpoch", event.DeactivationEpoch,
			"amount", event.Amount,
			"txHash", hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
			"logIndex", uint64(vLog.Index),
			"blockNumber", vLog.BlockNumber,
		)

		// msg validator exit
		msg := stakingTypes.NewMsgValidatorExit(
			hmTypes.BytesToHeimdallAddress(helper.GetAddress()),
			event.ValidatorId.Uint64(),
			event.DeactivationEpoch.Uint64(),
			hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
			uint64(vLog.Index),
			vLog.BlockNumber,
			event.Nonce.Uint64(),
		)

		// return broadcast to heimdall
		if err := sp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
			sp.Logger.Error("Error while broadcasting unstakeInit to heimdall", "validatorId", event.ValidatorId.Uint64(), "error", err)
			return err
		}
	}
	return nil
}

func (sp *StakingProcessor) sendStakeUpdateToHeimdall(eventName string, logBytes string) error {
	var vLog = types.Log{}
	if err := json.Unmarshal([]byte(logBytes), &vLog); err != nil {
		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	event := new(stakinginfo.StakinginfoStakeUpdate)
	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &vLog); err != nil {
		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		if isOld, _ := sp.isOldTx(sp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index)); isOld {
			sp.Logger.Info("Ignoring task to send unstakeinit to heimdall as already processed",
				"event", eventName,
				"validatorID", event.ValidatorId,
				"nonce", event.Nonce,
				"newAmount", event.NewAmount,
				"txHash", hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
				"logIndex", uint64(vLog.Index),
				"blockNumber", vLog.BlockNumber,
			)
			return nil
		}
		sp.Logger.Info(
			"✅ Received task to send stake-update to heimdall",
			"event", eventName,
			"validatorID", event.ValidatorId,
			"nonce", event.Nonce,
			"newAmount", event.NewAmount,
			"txHash", hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
			"logIndex", uint64(vLog.Index),
			"blockNumber", vLog.BlockNumber,
		)

		// msg validator exit
		msg := stakingTypes.NewMsgStakeUpdate(
			hmTypes.BytesToHeimdallAddress(helper.GetAddress()),
			event.ValidatorId.Uint64(),
			sdk.NewIntFromBigInt(event.NewAmount),
			hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
			uint64(vLog.Index),
			vLog.BlockNumber,
			event.Nonce.Uint64(),
		)

		// return broadcast to heimdall
		if err := sp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
			sp.Logger.Error("Error while broadcasting stakeupdate to heimdall", "validatorId", event.ValidatorId.Uint64(), "error", err)
			return err
		}
	}
	return nil
}

func (sp *StakingProcessor) sendSignerChangeToHeimdall(eventName string, logBytes string) error {
	var vLog = types.Log{}
	if err := json.Unmarshal([]byte(logBytes), &vLog); err != nil {
		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	event := new(stakinginfo.StakinginfoSignerChange)
	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &vLog); err != nil {
		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		newSignerPubKey := event.SignerPubkey
		if len(newSignerPubKey) == 64 {
			newSignerPubKey = util.AppendPrefix(newSignerPubKey)
		}

		if isOld, _ := sp.isOldTx(sp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index)); isOld {
			sp.Logger.Info("Ignoring task to send unstakeinit to heimdall as already processed",
				"event", eventName,
				"validatorID", event.ValidatorId,
				"nonce", event.Nonce,
				"NewSignerPubkey", hmTypes.NewPubKey(newSignerPubKey).String(),
				"oldSigner", event.OldSigner.Hex(),
				"newSigner", event.NewSigner.Hex(),
				"txHash", hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
				"logIndex", uint64(vLog.Index),
				"blockNumber", vLog.BlockNumber,
			)
			return nil
		}
		sp.Logger.Info(
			"✅ Received task to send signer-change to heimdall",
			"event", eventName,
			"validatorID", event.ValidatorId,
			"nonce", event.Nonce,
			"NewSignerPubkey", hmTypes.NewPubKey(newSignerPubKey).String(),
			"oldSigner", event.OldSigner.Hex(),
			"newSigner", event.NewSigner.Hex(),
			"txHash", hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
			"logIndex", uint64(vLog.Index),
			"blockNumber", vLog.BlockNumber,
		)

		// signer change
		msg := stakingTypes.NewMsgSignerUpdate(
			hmTypes.BytesToHeimdallAddress(helper.GetAddress()),
			event.ValidatorId.Uint64(),
			hmTypes.NewPubKey(newSignerPubKey),
			hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
			uint64(vLog.Index),
			vLog.BlockNumber,
			event.Nonce.Uint64(),
		)

		// return broadcast to heimdall
		if err := sp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
			sp.Logger.Error("Error while broadcasting signerChainge to heimdall", "validatorId", event.ValidatorId.Uint64(), "error", err)
			return err
		}
	}
	return nil
}

// isOldTx  checks if tx is already processed or not
func (sp *StakingProcessor) isOldTx(cliCtx cliContext.CLIContext, txHash string, logIndex uint64) (bool, error) {
	queryParam := map[string]interface{}{
		"txhash":   txHash,
		"logindex": logIndex,
	}

	endpoint := helper.GetHeimdallServerEndpoint(util.StakingTxStatusURL)
	url, err := util.CreateURLWithQuery(endpoint, queryParam)
	if err != nil {
		sp.Logger.Error("Error in creating url", "endpoint", endpoint, "error", err)
		return false, err
	}

	res, err := helper.FetchFromAPI(sp.cliCtx, url)
	if err != nil {
		sp.Logger.Error("Error fetching tx status", "url", url, "error", err)
		return false, err
	}

	var status bool
	if err := json.Unmarshal(res.Result, &status); err != nil {
		sp.Logger.Error("Error unmarshalling tx status received from Heimdall Server", "error", err)
		return false, err
	}

	return status, nil
}

func (sp *StakingProcessor) startPolling(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	// stop ticker when everything done
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			go sp.checkAndSendStakingSync()
		case <-ctx.Done():
			sp.Logger.Info("No-ack Polling stopped")
			ticker.Stop()
			return
		}
	}
}

// checkAndSendStakingSync - Staking No-Ack handler
// 1. Fetch latest stake sync seq from rootchain
// 2. check if elapsed time is more than NoAck Wait time.
// 3. Send NoAck to heimdall if required.
func (sp *StakingProcessor) checkAndSendStakingSync() {
	// fetch fresh staking context
	stakingContext, err := sp.getStakingContext()
	if err != nil {
		return
	}
	isCurrentProposer, err := util.IsCurrentProposer(sp.cliCtx)
	if err != nil {
		sp.Logger.Error("Error checking isCurrentProposer in staking no ack handler", "error", err)
		return
	}
	if !isCurrentProposer {
		sp.Logger.Info("I am not the current proposer. Ignoring")
		return
	}

	for rootChain, _ := range hmTypes.GetRootChainIDMap() {
		if rootChain == hmTypes.RootChainTypeStake {
			continue
		}

		stakingRecord, shouldSend := sp.shouldSendStakingSync(stakingContext, rootChain)
		if shouldSend {
			_ = sp.createAndSendStakingSyncToHeimdall(rootChain, stakingRecord)
		}
	}
}

// sendStakingSyncToHeimdall - - handles stake confirmation event from heimdall.
// 1. check if i am the current proposer.
// 2. check if this stake has to be submitted to rootchain
// 3. if so, create and broadcast stake transaction to rootchain
func (sp *StakingProcessor) sendStakingSyncToHeimdall(eventBytes string, blockHeight int64) error {
	sp.Logger.Info("Received triggerValidatorJoin request", "eventBytes", eventBytes, "blockHeight", blockHeight)
	var event = sdk.StringEvent{}
	if err := json.Unmarshal([]byte(eventBytes), &event); err != nil {
		sp.Logger.Error("Error unmarshalling event from heimdall", "error", err)
		return err
	}

	sp.Logger.Info("processing stake confirmation event", "event_type", event.Type)
	isCurrentProposer, err := util.IsCurrentProposer(sp.cliCtx)
	if err != nil {
		sp.Logger.Error("Error checking isCurrentProposer in StakeConfirmation handler", "error", err)
		return err
	}
	if !isCurrentProposer {
		sp.Logger.Info("I am not the current proposer. Ignoring", "eventType", event.Type)
		return nil
	}
	stakingContext, err := sp.getStakingContext()
	if err != nil {
		return nil
	}
	for _, rootChain := range hmTypes.GetRootChainMap() {
		if rootChain == hmTypes.RootChainTypeStake {
			continue
		}

		stakingRecord, shouldSend := sp.shouldSendStakingSync(stakingContext, rootChain)
		if shouldSend {
			return sp.createAndSendStakingSyncToHeimdall(rootChain, stakingRecord)
		}
	}
	return nil
}

func (sp *StakingProcessor) getNonceFromRootChain(stakingContext *StakingContext, rootChain string, validatorID uint64) uint64 {
	var currentStakingID uint64
	switch rootChain {
	case hmTypes.RootChainTypeEth:
		//TODO get latest staking id from rootChain1
	case hmTypes.RootChainTypeTron:
		//TODO get latest staking id from rootChain2
	default:

	}
	return currentStakingID
}

func (sp *StakingProcessor) shouldSendStakingSync(stakingContext *StakingContext, rootChain string) (*stakingTypes.StakingRecord, bool) {
	//
	// Check staking buffer
	//
	bufferedStakingRecord, err := util.GetBufferedStakingRecord(sp.cliCtx, rootChain)
	if err == nil {
		timeStamp := uint64(time.Now().Unix())
		stakingBufferTime := uint64(stakingContext.StakingParams.StakingBufferTime.Seconds())

		currentNonce := sp.getNonceFromRootChain(stakingContext, rootChain, bufferedStakingRecord.ValidatorID.Uint64())
		bufferedRecordTime := bufferedStakingRecord.TimeStamp
		if bufferedStakingRecord != nil && bufferedStakingRecord.Nonce > currentNonce &&
			!(bufferedRecordTime == 0 || ((timeStamp > bufferedRecordTime) && timeStamp-bufferedRecordTime >= stakingBufferTime)) {
			sp.Logger.Info("Staking already exits in buffer", "Staking", bufferedStakingRecord.String())
			return nil, false
		}
	}

	res, err := util.GetNextStakingRecord(sp.cliCtx, rootChain)
	if err != nil {
		return nil, false
	}
	currentNonce := sp.getNonceFromRootChain(stakingContext, rootChain, res.ValidatorID.Uint64())
	if res.Nonce > currentNonce {
		return res, true
	}
	return nil, false
}

func (sp *StakingProcessor) createAndSendStakingSyncToHeimdall(rootChain string, stakingRecord *stakingTypes.StakingRecord) error {
	if stakingRecord == nil {
		sp.Logger.Error("Invalid stakingRecord", "root", rootChain)
		return nil
	}

	sp.Logger.Info("✅ Creating and broadcasting new staking sync",
		"root", rootChain,
		"ValidatorID", stakingRecord.ValidatorID,
		"nonce", stakingRecord.Nonce,
	)

	// create and send staking sync message
	msg := stakingTypes.NewMsgStakingSync(
		hmTypes.BytesToHeimdallAddress(helper.GetAddress()),
		rootChain,
		stakingRecord.ValidatorID,
		stakingRecord.Nonce,
		stakingRecord.TxHash,
	)

	// return broadcast to heimdall
	if err := sp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
		sp.Logger.Error("Error while broadcasting staking sync to heimdall", "error", err)
		return err
	}

	return nil
}

func (sp *StakingProcessor) createAndSendStakingToRootChain(stakingContext *StakingContext, rootChain string, height int64, stakingInfo *stakingTypes.StakingRecord) error {

	sp.Logger.Info("Preparing staking info to be pushed on chain",
		"root", rootChain, "txHash", stakingInfo.TxHash)
	// proof
	tx, err := helper.QueryTxWithProof(sp.cliCtx, stakingInfo.TxHash.Bytes())
	if err != nil {
		sp.Logger.Error("Error querying staking info tx proof", "txHash", stakingInfo.TxHash)
		return err
	}

	// fetch side txs sigs
	decoder := helper.GetTxDecoder(authTypes.ModuleCdc)
	stdTx, err := decoder(tx.Tx)
	if err != nil {
		sp.Logger.Error("Error while decoding staking info tx", "txHash", tx.Tx.Hash(), "error", err)
		return err
	}

	cmsg := stdTx.GetMsgs()[0]
	sideMsg, ok := cmsg.(hmTypes.SideTxMsg)
	if !ok {
		sp.Logger.Error("Invalid side-tx msg", "txHash", tx.Tx.Hash())
		return err
	}

	// side-tx data
	sideTxData := sideMsg.GetSideSignBytes()

	// get sigs
	sigs, err := helper.FetchSideTxSigs(sp.httpClient, height, tx.Tx.Hash(), sideTxData)
	if err != nil {
		sp.Logger.Error("Error fetching votes for staking info tx", "height", height)
		return err
	}

	//TODO add state sync abi
	{
		// chain manager params
		chainParams := stakingContext.ChainmanagerParams.ChainParams
		// root chain address
		rootChainAddress := chainParams.RootChainAddress.EthAddress()
		// root chain instance
		rootChainInstance, err := sp.contractConnector.GetRootChainInstance(rootChainAddress)
		if err != nil {
			sp.Logger.Info("Error while creating rootchain instance", "error", err)
			return err
		}

		if err := sp.contractConnector.SendCheckpoint(sideTxData, sigs, rootChainAddress, rootChainInstance); err != nil {
			sp.Logger.Info("Error submitting checkpoint to rootchain", "error", err)
			return err
		}
	}
	return nil
}

// sendStakingSyncToRootchain - handles staking confirmation event from heimdall.
// 1. check if i am the current proposer.
// 2. check if this staking sync has to be submitted to rootchain
// 3. if so, create and broadcast staking sync transaction to rootchain
func (sp *StakingProcessor) sendStakingSyncToRootchain(eventBytes string, blockHeight int64) error {
	sp.Logger.Info("Received sendStakingSyncToRootchain request", "eventBytes", eventBytes, "blockHeight", blockHeight)
	var event = sdk.StringEvent{}
	if err := json.Unmarshal([]byte(eventBytes), &event); err != nil {
		sp.Logger.Error("Error unmarshalling event from heimdall", "error", err)
		return err
	}

	// var tx = sdk.TxResponse{}
	// if err := json.Unmarshal([]byte(txBytes), &tx); err != nil {
	// 	sp.Logger.Error("Error unmarshalling txResponse", "error", err)
	// 	return err
	// }

	sp.Logger.Info("processing staking sync confirmation event", "eventType", event.Type)
	isCurrentProposer, err := util.IsCurrentProposer(sp.cliCtx)
	if err != nil {
		sp.Logger.Error("Error checking isCurrentProposer in StakingSyncConfirmation handler", "error", err)
		return err
	}

	var (
		validatorID uint64
		nonce       uint64
		stakingHash string
		rootChain   string
	)

	for _, attr := range event.Attributes {
		if attr.Key == stakingTypes.AttributeKeyValidatorID {
			validatorID, _ = strconv.ParseUint(attr.Value, 10, 64)
		}
		if attr.Key == stakingTypes.AttributeKeyValidatorNonce {
			nonce, _ = strconv.ParseUint(attr.Value, 10, 64)
		}
		if attr.Key == stakingTypes.AttributeKeyStakingHash {
			stakingHash = attr.Value
		}
		if attr.Key == stakingTypes.AttributeKeyRootChain {
			rootChain = attr.Value
		}
	}

	stakingContext, err := sp.getStakingContext()
	if err != nil {
		return err
	}

	currentNonce := sp.getNonceFromRootChain(stakingContext, rootChain, validatorID)
	if err != nil {
		return err
	}

	if nonce == currentNonce+1 && isCurrentProposer {
		if err := sp.createAndSendStakingToRootChain(stakingContext, rootChain, blockHeight, &stakingTypes.StakingRecord{
			ValidatorID: hmTypes.NewValidatorID(validatorID),
			TxHash:      hmTypes.HexToHeimdallHash(stakingHash),
			Nonce:       nonce,
			TimeStamp:   0,
		}); err != nil {
			sp.Logger.Error("Error sending staking sync to rootchain", "error", err)
			return err
		}
	} else {
		sp.Logger.Info("I am not the current proposer or staking sync already sent. Ignoring", "eventType", event.Type)
	}
	return nil
}

// sendStakingAckToHeimdall - handles checkpointAck event from rootchain
// 1. create and broadcast checkpointAck msg to heimdall.
func (sp *StakingProcessor) sendStakingAckToHeimdall(eventName string, StakingAckStr string, rootChain string) error {
	var log = types.Log{}
	if err := json.Unmarshal([]byte(StakingAckStr), &log); err != nil {
		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	//TODO abi
	event := new(rootchain.RootchainNewHeaderBlock)
	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &log); err != nil {
		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		sp.Logger.Info(
			"✅ Received task to send staking-ack to heimdall",
			"event", eventName,
			"start", event.Start,
			"end", event.End,
			"reward", event.Reward,
			"root", "0x"+hex.EncodeToString(event.Root[:]),
			"proposer", event.Proposer.Hex(),
			"txHash", hmTypes.BytesToHeimdallHash(log.TxHash.Bytes()),
			"logIndex", uint64(log.Index),
			"rootChain", rootChain,
		)

		// create msg staking ack message
		msg := stakingTypes.NewMsgStakingSyncAck(
			helper.GetFromAddress(sp.cliCtx),
			rootChain,
			1,
			1,
		)

		// return broadcast to heimdall
		if err := sp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
			sp.Logger.Error("Error while broadcasting staking-ack to heimdall", "error", err)
			return err
		}
	}
	return nil
}

//
// utils
//

func (sp *StakingProcessor) getStakingContext() (*StakingContext, error) {
	chainmanagerParams, err := util.GetChainmanagerParams(sp.cliCtx)
	if err != nil {
		sp.Logger.Error("Error while fetching chain manager params", "error", err)
		return nil, err
	}

	stakingParams, err := util.GetStakingParams(sp.cliCtx)
	if err != nil {
		sp.Logger.Error("Error while fetching checkpoint params", "error", err)
		return nil, err
	}

	return &StakingContext{
		ChainmanagerParams: chainmanagerParams,
		StakingParams:      stakingParams,
	}, nil
}
