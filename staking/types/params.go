package types

import (
	"time"

	"github.com/maticnetwork/heimdall/params/subspace"
)

const (

	// DefaultProposerBonusPercent - Proposer Signer Reward Ratio
	DefaultProposerBonusPercent = int64(10)
)

// Parameter keys
var (
	KeyStakingBufferTime = []byte("StakingBufferTime")
)

// Params defines the parameters for the auth module.
type Params struct {
	StakingBufferTime time.Duration `json:"staking_buffer_time" yaml:"staking_buffer_time"`
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{KeyStakingBufferTime, &p.StakingBufferTime},
	}
}

// ParamStoreKeyProposerBonusPercent - Store's Key for Reward amount
var ParamStoreKeyProposerBonusPercent = []byte("proposerbonuspercent")

// ParamKeyTable type declaration for parameters
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable(
		ParamStoreKeyProposerBonusPercent, DefaultProposerBonusPercent,
	)
}
