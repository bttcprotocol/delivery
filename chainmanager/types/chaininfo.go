package types

import (
	"fmt"

	"github.com/maticnetwork/heimdall/types"
)

// ChainInfo represents new eth fork chain info
type ChainInfo struct {
	RootChainType         string                `json:"root_chain_type" yaml:"root_chain_type"`
	ActivationHeight      uint64                `json:"activation_height" yaml:"activation_height"`
	TxConfirmations       uint64                `json:"tx_confirmations" yaml:"tx_confirmations"`
	RootChainAddress      types.HeimdallAddress `json:"root_chain_address" yaml:"root_chain_address"`
	StateSenderAddress    types.HeimdallAddress `json:"state_sender_address" yaml:"state_sender_address"`
	StakingManagerAddress types.HeimdallAddress `json:"staking_manager_address" yaml:"staking_manager_address"`
	StakingInfoAddress    types.HeimdallAddress `json:"staking_info_address" yaml:"staking_info_address"`
	TimeStamp             uint64                `json:"timestamp"`
}

// String returns the string representation of chain info
func (s *ChainInfo) String() string {
	return fmt.Sprintf(
		"RootChainType: %v, ActivationHeight %v,TxConfirmations: %v, RootChainAddress: %v, StateSenderAddress: %v, StakingManagerAddress: %v, StakingInfoAddress: %v",
		s.RootChainType, s.ActivationHeight, s.TxConfirmations, s.RootChainAddress, s.StateSenderAddress, s.StakingManagerAddress, s.StakingInfoAddress,
	)
}
