package clerk

import (
	"bytes"
	"math/big"
	"strconv"

	ethCommon "github.com/maticnetwork/bor/common"

	ethTypes "github.com/maticnetwork/bor/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/maticnetwork/heimdall/clerk/types"
	"github.com/maticnetwork/heimdall/common"
	hmCommon "github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/helper"
	hmTypes "github.com/maticnetwork/heimdall/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

// NewSideTxHandler returns a side handler for "topup" type messages.
func NewSideTxHandler(k Keeper, contractCaller helper.IContractCaller) hmTypes.SideTxHandler {
	return func(ctx sdk.Context, msg sdk.Msg) abci.ResponseDeliverSideTx {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgEventRecord:
			return SideHandleMsgEventRecord(ctx, k, msg, contractCaller)
		default:
			return abci.ResponseDeliverSideTx{
				Code: uint32(sdk.CodeUnknownRequest),
			}
		}
	}
}

// NewPostTxHandler returns a side handler for "bank" type messages.
func NewPostTxHandler(k Keeper, contractCaller helper.IContractCaller) hmTypes.PostTxHandler {
	return func(ctx sdk.Context, msg sdk.Msg, sideTxResult abci.SideTxResultType) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgEventRecord:
			return PostHandleMsgEventRecord(ctx, k, msg, sideTxResult)
		default:
			return sdk.ErrUnknownRequest("Unknown msg type").Result()
		}
	}
}

func SideHandleMsgEventRecord(ctx sdk.Context, k Keeper, msg types.MsgEventRecord, contractCaller helper.IContractCaller) (result abci.ResponseDeliverSideTx) {

	k.Logger(ctx).Debug("✅ Validating External call for clerk msg",
		"txHash", hmTypes.BytesToHeimdallHash(msg.TxHash.Bytes()),
		"logIndex", uint64(msg.LogIndex),
		"blockNumber", msg.BlockNumber,
		"rootChainType", msg.RootChainType,
	)

	// chainManager params
	params := k.chainKeeper.GetParams(ctx)
	chainParams := params.ChainParams

	// get confirmed tx receipt
	var receipt *ethTypes.Receipt
	var err error
	// get main tx receipt
	var contractAddress ethCommon.Address
	switch msg.RootChainType {
	case hmTypes.RootChainTypeEth:
		receipt, err = contractCaller.GetConfirmedTxReceipt(msg.TxHash.EthHash(), params.MainchainTxConfirmations,
			hmTypes.RootChainTypeEth)
		if err != nil || receipt == nil {
			return hmCommon.ErrorSideTx(k.Codespace(), common.CodeWaitFrConfirmation)
		}
		contractAddress = chainParams.StateSenderAddress.EthAddress()
	case hmTypes.RootChainTypeBsc:
		bscChain, err := k.chainKeeper.GetChainParams(ctx, hmTypes.RootChainTypeBsc)
		if err != nil {
			k.Logger(ctx).Error("RootChain type: ", msg.RootChainType, " does not  match bsc")
			return hmCommon.ErrorSideTx(k.Codespace(), common.CodeWrongRootChainType)
		}
		receipt, err = contractCaller.GetConfirmedTxReceipt(msg.TxHash.EthHash(), bscChain.TxConfirmations,
			hmTypes.RootChainTypeBsc)
		if err != nil || receipt == nil {
			return hmCommon.ErrorSideTx(k.Codespace(), common.CodeWaitFrConfirmation)
		}
		contractAddress = bscChain.StateSenderAddress.EthAddress()
	case hmTypes.RootChainTypeTron:
		receipt, err = contractCaller.GetTronTransactionReceipt(msg.TxHash.Hex())
		if err != nil || receipt == nil {
			return hmCommon.ErrorSideTx(k.Codespace(), common.CodeWaitFrConfirmation)
		}
		contractAddress = hmTypes.HexToTronAddress(chainParams.TronStateSenderAddress)
	default:
		k.Logger(ctx).Error("RootChain type: ", msg.RootChainType, " does not  match eth or tron")
		return hmCommon.ErrorSideTx(k.Codespace(), common.CodeWrongRootChainType)
	}

	eventLog, err := contractCaller.DecodeStateSyncedEvent(contractAddress, receipt, msg.LogIndex)
	if err != nil || eventLog == nil {
		k.Logger(ctx).Error("Error fetching log from txhash")
		return hmCommon.ErrorSideTx(k.Codespace(), common.CodeErrDecodeEvent)
	}

	if receipt.BlockNumber.Uint64() != msg.BlockNumber {
		k.Logger(ctx).Error("BlockNumber in message doesn't match block umber in receipt", "MsgBlockNumber", msg.BlockNumber, "ReceiptBlockNumber", receipt.BlockNumber.Uint64())
		return hmCommon.ErrorSideTx(k.Codespace(), common.CodeInvalidMsg)
	}

	// check if message and event log matches
	if eventLog.Id.Uint64() != msg.ID {
		k.Logger(ctx).Error("ID in message doesn't match with id in log", "msgId", msg.ID, "stateIdFromTx", eventLog.Id)
		return hmCommon.ErrorSideTx(k.Codespace(), common.CodeInvalidMsg)
	}

	if !bytes.Equal(eventLog.ContractAddress.Bytes(), msg.ContractAddress.Bytes()) {
		k.Logger(ctx).Error(
			"ContractAddress from event does not match with Msg ContractAddress",
			"EventContractAddress", eventLog.ContractAddress.String(),
			"MsgContractAddress", msg.ContractAddress.String(),
		)
		return hmCommon.ErrorSideTx(k.Codespace(), common.CodeInvalidMsg)
	}

	if !bytes.Equal(eventLog.Data, msg.Data) {
		k.Logger(ctx).Error(
			"Data from event does not match with Msg Data",
			"EventData", hmTypes.BytesToHexBytes(eventLog.Data),
			"MsgData", hmTypes.BytesToHexBytes(msg.Data),
		)
		return hmCommon.ErrorSideTx(k.Codespace(), common.CodeInvalidMsg)
	}

	result.Result = abci.SideTxResultType_Yes
	return
}

func PostHandleMsgEventRecord(ctx sdk.Context, k Keeper, msg types.MsgEventRecord, sideTxResult abci.SideTxResultType) sdk.Result {

	// Skip handler if clerk is not approved
	if sideTxResult != abci.SideTxResultType_Yes {
		k.Logger(ctx).Debug("Skipping new clerk since side-tx didn't get yes votes")
		return common.ErrSideTxValidation(k.Codespace()).Result()
	}
	// check for replay
	if k.HasRootChainEventRecord(ctx, msg.RootChainType, msg.ID) {
		k.Logger(ctx).Debug("Skipping new clerk record as it's already processed")
		return hmCommon.ErrOldTx(k.Codespace()).Result()
	}
	// heimdall latestID
	latestMsgID := k.GetLatestID(ctx)
	//if latestMsgID < msg.ID {
	//	k.Logger(ctx).Debug("Heimdall msg latest ID ", latestMsgID, " less than rootChain", msg.RootChainType, " ID: ", msg.ID)
	//	return hmCommon.ErrOldTx(k.Codespace()).Result()
	//}
	k.Logger(ctx).Debug("Persisting clerk state", "sideTxResult", sideTxResult)

	// sequence id
	blockNumber := new(big.Int).SetUint64(msg.BlockNumber)
	sequence := helper.CalculateSequence(blockNumber, msg.LogIndex, msg.RootChainType)
	// create event record
	record := types.NewEventRecord(
		msg.TxHash,
		msg.LogIndex,
		msg.ID,
		msg.ContractAddress,
		msg.Data,
		msg.ChainID,
		ctx.BlockTime(),
		msg.RootChainType,
	)

	// save event into state
	if err := k.SetEventRecord(ctx, record); err != nil {
		k.Logger(ctx).Error("Unable to update event record", "error", err, "id", msg.ID)
		return types.ErrEventUpdate(k.Codespace()).Result()
	}
	processedLatestID := k.GetLatestID(ctx)
	if latestMsgID+1 != processedLatestID {
		k.Logger(ctx).Debug("Difference between Heimdall msg previous latest ID ", latestMsgID, " and current processed ID  ", processedLatestID, " is bigger than 1")
		return hmCommon.ErrValidatorSave(k.Codespace()).Result()
	}
	// save record sequence
	k.SetRecordSequence(ctx, sequence.String())

	// TX bytes
	txBytes := ctx.TxBytes()
	hash := tmTypes.Tx(txBytes).Hash()
	// add events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRecord,
			sdk.NewAttribute(sdk.AttributeKeyAction, msg.Type()),                                  // action
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),                // module name
			sdk.NewAttribute(hmTypes.AttributeKeyTxHash, hmTypes.BytesToHeimdallHash(hash).Hex()), // tx hash
			sdk.NewAttribute(types.AttributeKeyRecordTxLogIndex, strconv.FormatUint(msg.LogIndex, 10)),
			sdk.NewAttribute(hmTypes.AttributeKeySideTxResult, sideTxResult.String()), // result
			sdk.NewAttribute(types.AttributeKeyRecordID, strconv.FormatUint(processedLatestID, 10)),
			sdk.NewAttribute(types.AttributeKeyRecordContract, msg.ContractAddress.String()),
		),
	})

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
