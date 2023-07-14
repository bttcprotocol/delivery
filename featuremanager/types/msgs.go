package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypes "github.com/maticnetwork/heimdall/gov/types"
	"github.com/maticnetwork/heimdall/helper"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

var _ sdk.Msg = MsgSubmitProposal{}

const (
	TypeMsgSubmitProposal = "submit_feature_proposal"
)

// MsgSubmitProposal represents submit proposal message.
type MsgSubmitProposal struct {
	govTypes.MsgSubmitProposal //  Validator id
}

func (msg MsgSubmitProposal) Route() string { return RouterKey }
func (msg MsgSubmitProposal) Type() string  { return TypeMsgSubmitProposal }

// NewMsgSubmitProposal creates new submit proposal.
func NewMsgSubmitProposal(
	content Content,
	initialDeposit sdk.Coins,
	proposer hmTypes.HeimdallAddress,
	validator hmTypes.ValidatorID,
) MsgSubmitProposal {
	return MsgSubmitProposal{
		MsgSubmitProposal: govTypes.MsgSubmitProposal{
			Content:        content,
			InitialDeposit: initialDeposit,
			Proposer:       proposer,
			Validator:      validator,
		},
	}
}

// GetSideSignBytes returns side sign bytes.
func (msg MsgSubmitProposal) GetSideSignBytes() []byte {
	// stakeUpdate:(uint256 validatorId, uint256 nonce, uint256 newAmount)
	return helper.AppendBytes32()
}
