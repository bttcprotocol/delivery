package types

import (
	"bytes"
	"math/big"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	hmCommon "github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/maticnetwork/heimdall/types"
)

//
// Checkpoint Msg
//

var _ sdk.Msg = &MsgCheckpoint{}

// MsgCheckpoint represents checkpoint
type MsgCheckpoint struct {
	Proposer        types.HeimdallAddress `json:"proposer"`
	StartBlock      uint64                `json:"start_block"`
	EndBlock        uint64                `json:"end_block"`
	RootHash        types.HeimdallHash    `json:"root_hash"`
	AccountRootHash types.HeimdallHash    `json:"account_root_hash"`
	BorChainID      string                `json:"bor_chain_id"`
	Epoch           uint64                `json:"epoch"`
	RootChainType   string                `json:"root_chain_type"`
}

// NewMsgCheckpointBlock creates new checkpoint message using mentioned arguments
func NewMsgCheckpointBlock(
	proposer types.HeimdallAddress,
	startBlock uint64,
	endBlock uint64,
	roothash types.HeimdallHash,
	accountRootHash types.HeimdallHash,
	borChainID string,
	epoch uint64,
	rootChain string,
) MsgCheckpoint {
	return MsgCheckpoint{
		Proposer:        proposer,
		StartBlock:      startBlock,
		EndBlock:        endBlock,
		RootHash:        roothash,
		AccountRootHash: accountRootHash,
		BorChainID:      borChainID,
		Epoch:           epoch,
		RootChainType:   rootChain,
	}
}

// Type returns message type
func (msg MsgCheckpoint) Type() string {
	return "checkpoint"
}

func (msg MsgCheckpoint) Route() string {
	return RouterKey
}

// GetSigners returns address of the signer
func (msg MsgCheckpoint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.HeimdallAddressToAccAddress(msg.Proposer)}
}

func (msg MsgCheckpoint) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgCheckpoint) ValidateBasic() sdk.Error {
	if bytes.Equal(msg.RootHash.Bytes(), helper.ZeroHash.Bytes()) {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid rootHash %v", msg.RootHash.String())
	}

	if msg.Proposer.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid proposer %v", msg.Proposer.String())
	}

	if msg.StartBlock >= msg.EndBlock || msg.EndBlock == 0 {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid startBlock %v or/and endBlock %v", msg.StartBlock, msg.EndBlock)
	}

	return nil
}

// GetSideSignBytes returns side sign bytes
func (msg MsgCheckpoint) GetSideSignBytes() []byte {
	// keccak256(abi.encoded(proposer, startBlock, endBlock, rootHash, accountRootHash, bor chain id))
	borChainID, _ := strconv.ParseUint(msg.BorChainID, 10, 64)
	return appendBytes32(
		msg.Proposer.Bytes(),
		new(big.Int).SetUint64(msg.StartBlock).Bytes(),
		new(big.Int).SetUint64(msg.EndBlock).Bytes(),
		msg.RootHash.Bytes(),
		msg.AccountRootHash.Bytes(),
		new(big.Int).SetUint64(borChainID).Bytes(),
		new(big.Int).SetUint64(msg.Epoch).Bytes(),
	)
}

//
// Msg Checkpoint Ack
//

var _ sdk.Msg = &MsgCheckpointAck{}

// MsgCheckpointAck Add mainchain commit transaction hash to MsgCheckpointAck
type MsgCheckpointAck struct {
	From          types.HeimdallAddress `json:"from"`
	Number        uint64                `json:"number"`
	Proposer      types.HeimdallAddress `json:"proposer"`
	StartBlock    uint64                `json:"start_block"`
	EndBlock      uint64                `json:"end_block"`
	RootHash      types.HeimdallHash    `json:"root_hash"`
	TxHash        types.HeimdallHash    `json:"tx_hash"`
	LogIndex      uint64                `json:"log_index"`
	RootChainType string                `json:"root_chain_type"`
}

func NewMsgCheckpointAck(
	from types.HeimdallAddress,
	number uint64,
	proposer types.HeimdallAddress,
	startBlock uint64,
	endBlock uint64,
	rootHash types.HeimdallHash,
	txHash types.HeimdallHash,
	logIndex uint64,
	rootChain string,
) MsgCheckpointAck {

	return MsgCheckpointAck{
		From:          from,
		Number:        number,
		Proposer:      proposer,
		StartBlock:    startBlock,
		EndBlock:      endBlock,
		RootHash:      rootHash,
		TxHash:        txHash,
		LogIndex:      logIndex,
		RootChainType: rootChain,
	}
}

func (msg MsgCheckpointAck) Type() string {
	return "checkpoint-ack"
}

func (msg MsgCheckpointAck) Route() string {
	return RouterKey
}

// GetSigners returns signers
func (msg MsgCheckpointAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.HeimdallAddressToAccAddress(msg.From)}
}

// GetSignBytes returns sign bytes
func (msg MsgCheckpointAck) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// ValidateBasic validate basic
func (msg MsgCheckpointAck) ValidateBasic() sdk.Error {
	if msg.From.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid from %v", msg.From.String())
	}

	if msg.Proposer.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid empty proposer")
	}

	if msg.RootHash.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid empty root hash")
	}

	return nil
}

// GetTxHash Returns tx hash
func (msg MsgCheckpointAck) GetTxHash() types.HeimdallHash {
	return msg.TxHash
}

// GetLogIndex Returns log index
func (msg MsgCheckpointAck) GetLogIndex() uint64 {
	return msg.LogIndex
}

// GetSideSignBytes returns side sign bytes
func (msg MsgCheckpointAck) GetSideSignBytes() []byte {
	return nil
}

//
// Msg Checkpoint No Ack
//

var _ sdk.Msg = &MsgCheckpointNoAck{}

type MsgCheckpointNoAck struct {
	From types.HeimdallAddress `json:"from"`
}

func NewMsgCheckpointNoAck(from types.HeimdallAddress) MsgCheckpointNoAck {
	return MsgCheckpointNoAck{
		From: from,
	}
}

func (msg MsgCheckpointNoAck) Type() string {
	return "checkpoint-no-ack"
}

func (msg MsgCheckpointNoAck) Route() string {
	return RouterKey
}

func (msg MsgCheckpointNoAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.HeimdallAddressToAccAddress(msg.From)}
}

func (msg MsgCheckpointNoAck) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgCheckpointNoAck) ValidateBasic() sdk.Error {
	if msg.From.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid from %v", msg.From.String())
	}

	return nil
}

//
// Msg Checkpoint Sync
//

var _ sdk.Msg = &MsgCheckpointSync{}

type MsgCheckpointSync struct {
	From          types.HeimdallAddress `json:"from"`
	Number        uint64                `json:"Number"`
	Proposer      types.HeimdallAddress `json:"proposer"`
	StartBlock    uint64                `json:"start_block"`
	EndBlock      uint64                `json:"end_block"`
	RootChainType string                `json:"root_chain_type"`
}

func NewMsgCheckpointSync(from, proposer types.HeimdallAddress, number, start, end uint64, rootChain string) MsgCheckpointSync {
	return MsgCheckpointSync{
		From:          from,
		Number:        number,
		Proposer:      proposer,
		StartBlock:    start,
		EndBlock:      end,
		RootChainType: rootChain,
	}
}

func (msg MsgCheckpointSync) Type() string {
	return "checkpoint-sync"
}

func (msg MsgCheckpointSync) Route() string {
	return RouterKey
}

func (msg MsgCheckpointSync) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.HeimdallAddressToAccAddress(msg.From)}
}

func (msg MsgCheckpointSync) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgCheckpointSync) ValidateBasic() sdk.Error {
	if msg.Proposer.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid from %v", msg.Proposer.String())
	}
	return nil
}

// GetSideSignBytes returns side sign bytes
func (msg MsgCheckpointSync) GetSideSignBytes() []byte {
	// data: (address proposer, uint256 start, uint256 end, uint256 headerBlockId, uint256 chainID)
	return appendBytes32(
		msg.Proposer.Bytes(),
		new(big.Int).SetUint64(msg.StartBlock).Bytes(),
		new(big.Int).SetUint64(msg.EndBlock).Bytes(),
		new(big.Int).SetUint64(msg.Number).Bytes(),
		new(big.Int).SetUint64(uint64(types.GetRootChainID(msg.RootChainType))).Bytes(),
	)
}

//
// Msg Checkpoint Sync
//

var _ sdk.Msg = &MsgCheckpointSyncAck{}

type MsgCheckpointSyncAck MsgCheckpointSync

func NewMsgCheckpointSyncAck(proposer types.HeimdallAddress, number, start, end uint64, rootChain string) MsgCheckpointSyncAck {
	return MsgCheckpointSyncAck{
		Number:        number,
		Proposer:      proposer,
		StartBlock:    start,
		EndBlock:      end,
		RootChainType: rootChain,
	}
}

func (msg MsgCheckpointSyncAck) Type() string {
	return "checkpoint-sync-ack"
}

func (msg MsgCheckpointSyncAck) Route() string {
	return RouterKey
}

func (msg MsgCheckpointSyncAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.HeimdallAddressToAccAddress(msg.Proposer)}
}

func (msg MsgCheckpointSyncAck) GetSignBytes() []byte {
	b, err := ModuleCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgCheckpointSyncAck) ValidateBasic() sdk.Error {
	if msg.Proposer.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid from %v", msg.Proposer.String())
	}
	return nil
}

// GetSideSignBytes returns side sign bytes
func (msg MsgCheckpointSyncAck) GetSideSignBytes() []byte {
	return nil
}
