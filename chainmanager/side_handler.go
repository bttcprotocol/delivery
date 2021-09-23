package chainmanager

import (
	"bytes"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethCommon "github.com/maticnetwork/bor/common"
	ethTypes "github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/heimdall/chainmanager/types"
	"github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/helper"
	hmTypes "github.com/maticnetwork/heimdall/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

// NewSideTxHandler returns a side handler for "chainmanager" type messages.
func NewSideTxHandler(k Keeper, contractCaller helper.IContractCaller) hmTypes.SideTxHandler {
	return func(ctx sdk.Context, msg sdk.Msg) abci.ResponseDeliverSideTx {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgNewChain:
			return SideHandleMsgNewChain(ctx, msg, k, contractCaller)

		default:
			return abci.ResponseDeliverSideTx{
				Code: uint32(sdk.CodeUnknownRequest),
			}
		}
	}
}

// SideHandleMsgNewChain side msg new chain
func SideHandleMsgNewChain(ctx sdk.Context, msg types.MsgNewChain, k Keeper, contractCaller helper.IContractCaller) (result abci.ResponseDeliverSideTx) {
	k.Logger(ctx).Debug("✅ Validating External call for new chain msg",
		"txHash", hmTypes.BytesToHeimdallHash(msg.TxHash.Bytes()),
		"logIndex", msg.LogIndex,
		"blockNumber", msg.BlockNumber,
	)
	// chainManager params
	params := k.GetParams(ctx)
	chainParams := params.ChainParams

	var (
		contractAddress ethCommon.Address
		receipt         *ethTypes.Receipt
		err             error
	)
	// get event log on tron
	receipt, err = contractCaller.GetTronTransactionReceipt(msg.TxHash.Hex())
	if err != nil || receipt == nil {
		return common.ErrorSideTx(k.Codespace(), common.CodeWaitFrConfirmation)
	}
	contractAddress = hmTypes.HexToTronAddress(chainParams.TronChainAddress)
	// decode validator join event
	eventLog, err := contractCaller.DecodeNewChainEvent(contractAddress, receipt, msg.LogIndex)
	if err != nil || eventLog == nil {
		return common.ErrorSideTx(k.Codespace(), common.CodeErrDecodeEvent)
	}
	if msg.BlockNumber != receipt.BlockNumber.Uint64() {
		k.Logger(ctx).Error("BlockNumber in message doesn't match with receipt",
			"MsgBlockNumber", msg.BlockNumber, "ReceiptBlockNumber", receipt.BlockNumber.Uint64())
		return common.ErrorSideTx(k.Codespace(), common.CodeInvalidMsg)
	}
	if uint64(hmTypes.GetRootChainID(msg.RootChainType)) != eventLog.RootChainId.Uint64() {
		k.Logger(ctx).Error("RootChainType in message doesn't match with receipt",
			"MsgRootChainType", msg.RootChainType, "ReceiptRootChainType", eventLog.RootChainId.Uint64())
		return common.ErrorSideTx(k.Codespace(), common.CodeInvalidMsg)
	}
	if msg.TxConfirmations != eventLog.TxConfirmations.Uint64() {
		k.Logger(ctx).Error("TxConfirmations in message doesn't match with receipt",
			"MsgTxConfirmations", msg.TxConfirmations, "ReceiptTxConfirmations", eventLog.TxConfirmations.Uint64())
		return common.ErrorSideTx(k.Codespace(), common.CodeInvalidMsg)
	}
	if msg.ActivationHeight != eventLog.ActivationHeight.Uint64() {
		k.Logger(ctx).Error("ActivationHeight in message doesn't match with receipt",
			"MsgActivationHeight", msg.ActivationHeight, "ReceiptActivationHeight", eventLog.ActivationHeight.Uint64())
		return common.ErrorSideTx(k.Codespace(), common.CodeInvalidMsg)
	}
	if !bytes.Equal(msg.RootChainAddress.Bytes(), eventLog.RootChainAddress.Bytes()) ||
		!bytes.Equal(msg.StateSenderAddress.Bytes(), eventLog.StateSenderAddress.Bytes()) ||
		!bytes.Equal(msg.StakingInfoAddress.Bytes(), eventLog.StakingInfoAddress.Bytes()) ||
		!bytes.Equal(msg.StakingManagerAddress.Bytes(), eventLog.StakingManagerAddress.Bytes()) {
		k.Logger(ctx).Error("ContractsAddresses in message doesn't match with receipt")
		return common.ErrorSideTx(k.Codespace(), common.CodeInvalidMsg)
	}

	k.Logger(ctx).Debug("✅ Succesfully validated External call for new chain msg")
	result.Result = abci.SideTxResultType_Yes
	return
}

// NewPostTxHandler returns a side handler for "chainmanager" type messages.
func NewPostTxHandler(k Keeper, contractCaller helper.IContractCaller) hmTypes.PostTxHandler {
	return func(ctx sdk.Context, msg sdk.Msg, sideTxResult abci.SideTxResultType) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgNewChain:
			return PostHandleMsgMsgNewChain(ctx, k, msg, sideTxResult)
		default:
			return sdk.ErrUnknownRequest("Unrecognized Staking Msg type").Result()
		}
	}
}

/*
	Post Handlers - update the state of the tx
**/

// PostHandleMsgMsgNewChain msg new chain
func PostHandleMsgMsgNewChain(ctx sdk.Context, k Keeper, msg types.MsgNewChain, sideTxResult abci.SideTxResultType) sdk.Result {

	// Skip handler if validator join is not approved
	if sideTxResult != abci.SideTxResultType_Yes {
		k.Logger(ctx).Debug("Skipping new chain msg since side-tx didn't get yes votes")
		return common.ErrSideTxValidation(k.Codespace()).Result()
	}

	k.Logger(ctx).Debug("Adding new chain to state", "sideTxResult", sideTxResult)

	// add validator to store=
	err := k.AddNewChainParams(ctx, types.ChainInfo{
		RootChainType:         msg.RootChainType,
		ActivationHeight:      msg.ActivationHeight,
		TxConfirmations:       msg.TxConfirmations,
		RootChainAddress:      msg.RootChainAddress,
		StateSenderAddress:    msg.StateSenderAddress,
		StakingManagerAddress: msg.StakingManagerAddress,
		StakingInfoAddress:    msg.StakingInfoAddress,
		TimeStamp:             uint64(ctx.BlockTime().Unix()),
	})
	if err != nil {
		k.Logger(ctx).Error("Unable to add new chain to state", "error", err, "root", msg.RootChainType)
		return common.ErrWrongRootChain(k.Codespace()).Result()
	}

	k.Logger(ctx).Debug("✅ New chain successfully joined", "root", msg.RootChainType)

	// TX bytes
	txBytes := ctx.TxBytes()
	hash := tmTypes.Tx(txBytes).Hash()

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeNewChain,
			sdk.NewAttribute(sdk.AttributeKeyAction, msg.Type()),                                  // action
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),                // module name
			sdk.NewAttribute(hmTypes.AttributeKeyTxHash, hmTypes.BytesToHeimdallHash(hash).Hex()), // tx hash
			sdk.NewAttribute(hmTypes.AttributeKeyTxLogIndex, strconv.FormatUint(msg.LogIndex, 10)),
			sdk.NewAttribute(hmTypes.AttributeKeySideTxResult, sideTxResult.String()), // result
			sdk.NewAttribute(types.AttributeKeyActivationHeight, strconv.FormatUint(msg.ActivationHeight, 10)),
			sdk.NewAttribute(types.AttributeKeyRootChain, msg.RootChainType),
		),
	})

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
