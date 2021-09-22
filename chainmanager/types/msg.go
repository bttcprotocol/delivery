package types

import (
	hmCommon "github.com/maticnetwork/heimdall/common"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/maticnetwork/heimdall/types"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

var cdc = codec.New()

//
// Validator Join
//

var _ sdk.Msg = &MsgNewChain{}

type MsgNewChain struct {
	From                  hmTypes.HeimdallAddress `json:"from"`
	RootChainType         string                  `json:"root_chain_type" yaml:"root_chain_type"`
	ActivationHeight      uint64                  `json:"activation_height" yaml:"activation_height"`
	TxConfirmations       uint64                  `json:"tx_confirmations" yaml:"tx_confirmations"`
	RootChainAddress      types.HeimdallAddress   `json:"root_chain_address" yaml:"root_chain_address"`
	StateSenderAddress    types.HeimdallAddress   `json:"state_sender_address" yaml:"state_sender_address"`
	StakingManagerAddress types.HeimdallAddress   `json:"staking_manager_address" yaml:"staking_manager_address"`
	StakingInfoAddress    types.HeimdallAddress   `json:"staking_info_address" yaml:"staking_info_address"`
	TxHash                hmTypes.HeimdallHash    `json:"tx_hash"`
	LogIndex              uint64                  `json:"log_index"`
	BlockNumber           uint64                  `json:"block_number"`
}

// NewMsgNewChain creates new validator-join
func NewMsgNewChain(
	from hmTypes.HeimdallAddress,
	rootChain string,
	activationHeight uint64,
	txConfirmations uint64,
	RootChainAddress hmTypes.HeimdallAddress,
	StateSenderAddress hmTypes.HeimdallAddress,
	StakingManagerAddress hmTypes.HeimdallAddress,
	StakingInfoAddress hmTypes.HeimdallAddress,
	txhash hmTypes.HeimdallHash,
	logIndex uint64,
	blockNumber uint64,
) MsgNewChain {

	return MsgNewChain{
		From:                  from,
		RootChainType:         rootChain,
		ActivationHeight:      activationHeight,
		TxConfirmations:       txConfirmations,
		RootChainAddress:      RootChainAddress,
		StateSenderAddress:    StateSenderAddress,
		StakingManagerAddress: StakingManagerAddress,
		StakingInfoAddress:    StakingInfoAddress,
		TxHash:                txhash,
		LogIndex:              logIndex,
		BlockNumber:           blockNumber,
	}
}

func (msg MsgNewChain) Type() string {
	return "new-chain"
}

func (msg MsgNewChain) Route() string {
	return RouterKey
}

func (msg MsgNewChain) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{hmTypes.HeimdallAddressToAccAddress(msg.From)}
}

func (msg MsgNewChain) GetSignBytes() []byte {
	b, err := cdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgNewChain) ValidateBasic() sdk.Error {
	if msg.RootChainType == "" {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid validator ID %v", msg.RootChainType)
	}

	if msg.RootChainAddress.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid root address %v", msg.RootChainAddress.String())
	}

	if msg.StateSenderAddress.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid state send address %v", msg.StateSenderAddress.String())
	}

	if msg.StakingManagerAddress.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid staking manager address %v", msg.StakingManagerAddress.String())
	}

	if msg.StakingInfoAddress.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid staking info address %v", msg.StakingInfoAddress.String())
	}

	if msg.From.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid proposer %v", msg.From.String())
	}

	return nil
}

// GetSideSignBytes returns side sign bytes
func (msg MsgNewChain) GetSideSignBytes() []byte {
	return nil
}
