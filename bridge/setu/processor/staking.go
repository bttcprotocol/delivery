package processor

import (
	"context"
	"encoding/json"
	"time"

	checkpointTypes "github.com/maticnetwork/heimdall/checkpoint/types"

	authTypes "github.com/maticnetwork/heimdall/auth/types"

	"github.com/RichardKnop/machinery/v1/tasks"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/bor/accounts/abi"
	"github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	chainmanagerTypes "github.com/maticnetwork/heimdall/chainmanager/types"
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
	//if err := sp.queueConnector.Server.RegisterTask("sendStakeUpdateToHeimdall", sp.sendStakeUpdateToHeimdall); err != nil {
	//	sp.Logger.Error("RegisterTasks | sendStakeUpdateToHeimdall", "error", err)
	//}
	if err := sp.queueConnector.Server.RegisterTask("sendSignerChangeToHeimdall", sp.sendSignerChangeToHeimdall); err != nil {
		sp.Logger.Error("RegisterTasks | sendSignerChangeToHeimdall", "error", err)
	}
	if err := sp.queueConnector.Server.RegisterTask("sendStakingSyncToHeimdall", sp.sendStakingSyncToHeimdall); err != nil {
		sp.Logger.Error("RegisterTasks | sendStakingSyncToRootChain", "error", err)
	}
	if err := sp.queueConnector.Server.RegisterTask("sendStakingSyncToRootChain", sp.sendStakingSyncToRootchain); err != nil {
		sp.Logger.Error("RegisterTasks | sendStakingSyncToRootChain", "error", err)
	}
	if err := sp.queueConnector.Server.RegisterTask("sendStakingAckToHeimdall", sp.sendStakingAckToHeimdall); err != nil {
		sp.Logger.Error("RegisterTasks | sendStakingAckToHeimdall", "error", err)
	}
}

func (sp *StakingProcessor) sendValidatorJoinToHeimdall(eventName string, logBytes string, rootChain string) error {
	if rootChain != hmTypes.RootChainTypeStake {
		sp.Logger.Error("There should be no messages from un-stake.", "root", rootChain)
		return nil
	}

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

func (sp *StakingProcessor) sendUnstakeInitToHeimdall(eventName string, logBytes string, rootChain string) error {
	if rootChain != hmTypes.RootChainTypeStake {
		sp.Logger.Error("There should be no messages from un-stake.", "root", rootChain)
		return nil
	}

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

//func (sp *StakingProcessor) sendStakeUpdateToHeimdall(eventName string, logBytes string, rootChain string) error {
//	if rootChain != hmTypes.RootChainTypeStake {
//		sp.Logger.Error("There should be no messages from un-stake.", "root", rootChain)
//		return nil
//	}
//
//	var vLog = types.Log{}
//	if err := json.Unmarshal([]byte(logBytes), &vLog); err != nil {
//		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
//		return err
//	}
//
//	event := new(stakinginfo.StakinginfoStakeUpdate)
//	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &vLog); err != nil {
//		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
//	} else {
//		if isOld, _ := sp.isOldTx(sp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index)); isOld {
//			sp.Logger.Info("Ignoring task to send unstakeinit to heimdall as already processed",
//				"event", eventName,
//				"validatorID", event.ValidatorId,
//				"nonce", event.Nonce,
//				"newAmount", event.NewAmount,
//				"txHash", hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
//				"logIndex", uint64(vLog.Index),
//				"blockNumber", vLog.BlockNumber,
//			)
//			return nil
//		}
//		sp.Logger.Info(
//			"✅ Received task to send stake-update to heimdall",
//			"event", eventName,
//			"validatorID", event.ValidatorId,
//			"nonce", event.Nonce,
//			"newAmount", event.NewAmount,
//			"txHash", hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
//			"logIndex", uint64(vLog.Index),
//			"blockNumber", vLog.BlockNumber,
//		)
//
//		// msg validator exit
//		msg := stakingTypes.NewMsgStakeUpdate(
//			hmTypes.BytesToHeimdallAddress(helper.GetAddress()),
//			event.ValidatorId.Uint64(),
//			sdk.NewIntFromBigInt(event.NewAmount),
//			hmTypes.BytesToHeimdallHash(vLog.TxHash.Bytes()),
//			uint64(vLog.Index),
//			vLog.BlockNumber,
//			event.Nonce.Uint64(),
//		)
//
//		// return broadcast to heimdall
//		if err := sp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
//			sp.Logger.Error("Error while broadcasting stakeupdate to heimdall", "validatorId", event.ValidatorId.Uint64(), "error", err)
//			return err
//		}
//	}
//	return nil
//}

func (sp *StakingProcessor) sendSignerChangeToHeimdall(eventName string, logBytes string, rootChain string) error {
	if rootChain != hmTypes.RootChainTypeStake {
		sp.Logger.Error("There should be no messages from un-stake.", "root", rootChain)
		return nil
	}

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
			go sp.checkStakingSyncAck()
		case <-ctx.Done():
			sp.Logger.Info("No-ack Polling stopped")
			ticker.Stop()
			return
		}
	}
}

// checkStakingSyncAck - Staking Ack handler
// 1. Fetch latest validator nonce from rootchain
// 2. check if nonce == queue_nonce.
// 3. Send Ack to heimdall if required.
// 4. start next staking sync task.
func (sp *StakingProcessor) checkStakingSyncAck() {
	isCurrentProposer, err := util.IsCurrentProposer(sp.cliCtx)
	if err != nil {
		sp.Logger.Error("Error checking isCurrentProposer in staking no ack handler", "error", err)
		return
	}
	if !isCurrentProposer {
		sp.Logger.Info("I am not the current proposer. Ignoring")
		return
	}

	for rootChain := range hmTypes.GetRootChainIDMap() {
		if rootChain == hmTypes.RootChainTypeStake {
			continue
		}
		// fetch fresh staking context
		stakingContext, err := sp.getStakingContext(rootChain)
		if err != nil {
			return
		}
		// fetch next staking record from queue
		res, err := util.GetNextStakingRecord(sp.cliCtx, rootChain)
		if err != nil || res.Nonce == 0 {
			continue
		}
		currentNonce := sp.getNonceFromRootChain(stakingContext, rootChain, res.ValidatorID.Uint64())
		if res.Nonce == currentNonce {
			// create msg staking ack message
			msg := stakingTypes.NewMsgStakingSyncAck(
				helper.GetFromAddress(sp.cliCtx),
				rootChain,
				res.ValidatorID,
				res.Nonce,
			)

			// return broadcast to heimdall
			if err := sp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
				sp.Logger.Error("Error while broadcasting staking-ack to heimdall", "error", err)
			}
		}
		sp.checkAndSendStakingSync(rootChain, false)
	}

}

// checkAndSendStakingSync
// 1. Fetch latest validator nonce from rootchain
// 2. check if should send.
// 3. Send staking sync to root if required.
func (sp *StakingProcessor) checkAndSendStakingSync(rootChain string, toRootChain bool) {
	// fetch fresh staking context
	stakingContext, err := sp.getStakingContext(rootChain)
	if err != nil {
		return
	}

	stakingRecord, shouldSend := sp.shouldSendStakingSync(stakingContext, rootChain)
	if shouldSend {
		if toRootChain {
			sp.createAndSendStakingSyncToRootChain(stakingContext, rootChain, stakingRecord)
		} else {
			// create msg staking ack message
			msg := stakingTypes.NewMsgStakingSync(
				helper.GetFromAddress(sp.cliCtx),
				rootChain,
				stakingRecord.ValidatorID,
				stakingRecord.Nonce,
			)

			// return broadcast to heimdall
			if err := sp.txBroadcaster.BroadcastToHeimdall(msg); err != nil {
				sp.Logger.Error("Error while broadcasting staking-sync to heimdall", "error", err)
			}
		}
	}
}

func (sp *StakingProcessor) getNonceFromRootChain(stakingContext *StakingContext, rootChain string, validatorID uint64) uint64 {
	switch rootChain {
	case hmTypes.RootChainTypeEth, hmTypes.RootChainTypeBsc:
		stakingManagerAddress := stakingContext.ChainmanagerParams.ChainParams.StakingManagerAddress.EthAddress()
		stakingManagerInstance, _ := sp.contractConnector.GetStakeManagerInstance(stakingManagerAddress, rootChain)
		return sp.contractConnector.GetMainStakingSyncNonce(validatorID, stakingManagerInstance)
	case hmTypes.RootChainTypeTron:
		stakingManagerAddress := stakingContext.ChainmanagerParams.ChainParams.TronStakingManagerAddress
		return sp.contractConnector.GetTronStakingSyncNonce(validatorID, stakingManagerAddress)
	}
	return 0
}

func (sp *StakingProcessor) shouldSendStakingSync(stakingContext *StakingContext, rootChain string) (*stakingTypes.StakingRecord, bool) {
	// fetch next staking record from queue
	res, err := util.GetNextStakingRecord(sp.cliCtx, rootChain)
	if err != nil || res.Nonce == 0 {
		return nil, false
	}
	currentNonce := sp.getNonceFromRootChain(stakingContext, rootChain, res.ValidatorID.Uint64())
	if res.Nonce > currentNonce {
		return res, true
	}
	return nil, false
}

func (sp *StakingProcessor) createAndSendStakingSyncToRootChain(stakingContext *StakingContext, rootChain string, stakingInfo *stakingTypes.StakingRecord) {

	sp.Logger.Info("Preparing staking sync to be pushed on chain",
		"root", rootChain, "type", stakingInfo.Type,
		"id", stakingInfo.ValidatorID, "nonce", stakingInfo.Nonce)
	// proof
	tx, err := helper.QueryTxWithProof(sp.cliCtx, stakingInfo.TxHash.Bytes())
	if err != nil {
		sp.Logger.Error("Error querying staking info tx proof", "txHash", stakingInfo.TxHash, "error", err)
		return
	}

	// fetch side txs sigs
	decoder := helper.GetTxDecoder(authTypes.ModuleCdc)
	stdTx, err := decoder(tx.Tx)
	if err != nil {
		sp.Logger.Error("Error while decoding staking info tx", "txHash", tx.Tx.Hash(), "error", err)
		return
	}

	cmsg := stdTx.GetMsgs()[0]
	sideMsg, ok := cmsg.(hmTypes.SideTxMsg)
	if !ok {
		sp.Logger.Error("Invalid side-tx msg", "txHash", tx.Tx.Hash())
		return
	}

	// side-tx data
	sideTxData := sideMsg.GetSideSignBytes()

	// get sigs
	sigs, err := helper.FetchSideTxSigs(sp.httpClient, stakingInfo.Height, tx.Tx.Hash(), sideTxData)
	if err != nil {
		sp.Logger.Error("Error fetching votes for staking record tx", "height", stakingInfo.Height, "error", err)
		return
	}
	// chain manager params
	chainParams := stakingContext.ChainmanagerParams.ChainParams
	stakingManagerAddress := chainParams.StakingManagerAddress.EthAddress()
	// staking manager instance
	stakingManagerInstance, err := sp.contractConnector.GetStakeManagerInstance(stakingManagerAddress, rootChain)
	if err != nil {
		sp.Logger.Error("Error while creating staking instance", "error", err)
		return
	}
	if err := sp.contractConnector.SendMainStakingSync(stakingInfo.Type, sideTxData, sigs, stakingManagerAddress, stakingManagerInstance, rootChain); err != nil {
		sp.Logger.Error("Error submitting staking sync to rootchain", "error", err)
		return
	}
}

// sendStakingSyncToRootchain - handles staking confirmation event from heimdall.
// 1. check if i am the current proposer.
// 2. check if this staking sync has to be submitted to rootchain
// 3. if so, create and broadcast staking sync transaction to heimdall
func (sp *StakingProcessor) sendStakingSyncToHeimdall(eventBytes string, blockHeight int64) error {
	sp.Logger.Info("Received sendStakingSyncToHeimdall request", "eventBytes", eventBytes, "blockHeight", blockHeight)

	isCurrentProposer, err := util.IsCurrentProposer(sp.cliCtx)
	if err != nil {
		sp.Logger.Error("Error checking isCurrentProposer in staking send heimdall handler", "error", err)
		return nil
	}
	if !isCurrentProposer {
		sp.Logger.Info("I am not the current proposer. Ignoring")
		return nil
	}

	for rootChain := range hmTypes.GetRootChainIDMap() {
		if rootChain == hmTypes.RootChainTypeStake {
			continue
		}
		sp.checkAndSendStakingSync(rootChain, false)
	}
	return nil
}

// sendStakingSyncToRootchain - handles staking confirmation event from heimdall.
// 1. check if i am the current proposer.
// 2. check if this staking sync has to be submitted to rootchain
// 3. if so, create and broadcast staking sync transaction to rootchain
func (sp *StakingProcessor) sendStakingSyncToRootchain(eventBytes string, blockHeight int64) error {
	sp.Logger.Info("Received sendStakingSyncToRootchain request", "eventBytes", eventBytes, "blockHeight", blockHeight)
	isCurrentProposer, err := util.IsCurrentProposer(sp.cliCtx)
	if err != nil {
		sp.Logger.Error("Error checking isCurrentProposer in staking send rootchain handler", "error", err)
		return nil
	}
	if !isCurrentProposer {
		sp.Logger.Info("I am not the current proposer. Ignoring")
		return nil
	}

	var event = sdk.StringEvent{}
	if err := json.Unmarshal([]byte(eventBytes), &event); err != nil {
		sp.Logger.Error("Error unmarshalling event from heimdall", "error", err)
		return err
	}
	sp.Logger.Info("processing staking sync confirmation event", "eventtype", event.Type)
	var rootChain string
	for _, attr := range event.Attributes {
		if attr.Key == checkpointTypes.AttributeKeyRootChain {
			rootChain = attr.Value
		}
	}

	sp.Logger.Info("Received sendStakingSyncToRootchain request", "rootChain", rootChain)
	sp.checkAndSendStakingSync(rootChain, true)
	return nil
}

// sendStakingAckToHeimdall - handles checkpointAck event from rootchain
// 1. create and broadcast checkpointAck msg to heimdall.
func (sp *StakingProcessor) sendStakingAckToHeimdall(eventName string, StakingAckStr string, rootChain string) error {
	if rootChain == hmTypes.RootChainTypeStake {
		sp.Logger.Error("There should be no messages from stake chain.", "root", rootChain)
		return nil
	}

	var log = types.Log{}
	if err := json.Unmarshal([]byte(StakingAckStr), &log); err != nil {
		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	event := new(stakinginfo.StakinginfoStakeAck)
	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &log); err != nil {
		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		sp.Logger.Info(
			"✅ Received task to send staking-ack to heimdall",
			"event", eventName,
			"validatorID", event.ValidatorId.Uint64(),
			"nonce", event.Nonce.Uint64(),
		)

		// create msg staking ack message
		msg := stakingTypes.NewMsgStakingSyncAck(
			helper.GetFromAddress(sp.cliCtx),
			rootChain,
			hmTypes.NewValidatorID(event.ValidatorId.Uint64()),
			event.Nonce.Uint64(),
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

func (sp *StakingProcessor) getStakingContext(rootChain string) (*StakingContext, error) {
	chainmanagerParams, err := util.GetNewChainParams(sp.cliCtx, rootChain)
	if err != nil {
		sp.Logger.Error("Error while fetching chain manager params", "error", err)
		return nil, err
	}

	return &StakingContext{
		ChainmanagerParams: chainmanagerParams,
	}, nil
}
