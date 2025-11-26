package types

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/maticnetwork/heimdall/params/subspace"
)

// Default parameter values
const (
	DefaultCheckpointBufferTime   time.Duration = 1000 * time.Second // Time checkpoint is allowed to stay in buffer (1000 seconds ~ 17 mins)
	DefaultAvgCheckpointLength    uint64        = 256
	DefaultMaxCheckpointLength    uint64        = 1024
	DefaultChildBlockInterval     uint64        = 10000
	DefaultCheckpointPollInterval time.Duration = 30 * time.Minute // Poll interval for checkpoint service to send new checkpoints or missing ACK
)

// Parameter keys
var (
	KeyCheckpointBufferTime   = []byte("CheckpointBufferTime")
	KeyAvgCheckpointLength    = []byte("AvgCheckpointLength")
	KeyMaxCheckpointLength    = []byte("MaxCheckpointLength")
	KeyChildBlockInterval     = []byte("ChildBlockInterval")
	KeyCheckpointPollInterval = []byte("CheckpointPollInterval")
)

var _ subspace.ParamSet = &Params{}

// Params defines the parameters for the auth module.
type Params struct {
	CheckpointBufferTime   time.Duration `json:"checkpoint_buffer_time" yaml:"checkpoint_buffer_time"`
	AvgCheckpointLength    uint64        `json:"avg_checkpoint_length" yaml:"avg_checkpoint_length"`
	MaxCheckpointLength    uint64        `json:"max_checkpoint_length" yaml:"max_checkpoint_length"`
	ChildBlockInterval     uint64        `json:"child_chain_block_interval" yaml:"child_chain_block_interval"`
	CheckpointPollInterval time.Duration `json:"checkpoint_poll_interval" yaml:"checkpoint_poll_interval"`
}

// NewParams creates a new Params object
func NewParams(
	checkpointBufferTime time.Duration,
	checkpointLength uint64,
	maxCheckpointLength uint64,
	childBlockInterval uint64,
	checkpointPollInterval time.Duration,
) Params {
	return Params{
		CheckpointBufferTime:   checkpointBufferTime,
		AvgCheckpointLength:    checkpointLength,
		MaxCheckpointLength:    maxCheckpointLength,
		ChildBlockInterval:     childBlockInterval,
		CheckpointPollInterval: checkpointPollInterval,
	}
}

// ParamKeyTable for auth module
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{KeyCheckpointBufferTime, &p.CheckpointBufferTime},
		{KeyAvgCheckpointLength, &p.AvgCheckpointLength},
		{KeyMaxCheckpointLength, &p.MaxCheckpointLength},
		{KeyChildBlockInterval, &p.ChildBlockInterval},
		{KeyCheckpointPollInterval, &p.CheckpointPollInterval},
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		CheckpointBufferTime:   DefaultCheckpointBufferTime,
		AvgCheckpointLength:    DefaultAvgCheckpointLength,
		MaxCheckpointLength:    DefaultMaxCheckpointLength,
		ChildBlockInterval:     DefaultChildBlockInterval,
		CheckpointPollInterval: DefaultCheckpointPollInterval,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("CheckpointBufferTime: %s\n", p.CheckpointBufferTime))
	sb.WriteString(fmt.Sprintf("AvgCheckpointLength: %d\n", p.AvgCheckpointLength))
	sb.WriteString(fmt.Sprintf("MaxCheckpointLength: %d\n", p.MaxCheckpointLength))
	sb.WriteString(fmt.Sprintf("ChildBlockInterval: %d\n", p.ChildBlockInterval))
	sb.WriteString(fmt.Sprintf("CheckpointPollInterval: %s\n", p.CheckpointPollInterval))
	return sb.String()
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if p.MaxCheckpointLength == 0 || p.AvgCheckpointLength == 0 {
		return fmt.Errorf("MaxCheckpointLength, AvgCheckpointLength should be non-zero")
	}

	if p.MaxCheckpointLength < p.AvgCheckpointLength {
		return fmt.Errorf("AvgCheckpointLength should not be greater than MaxCheckpointLength")
	}

	if p.ChildBlockInterval == 0 {
		return fmt.Errorf("ChildBlockInterval should be greater than zero")
	}

	if p.CheckpointPollInterval == 0 {
		return fmt.Errorf("CheckpointPollInterval should be greater than zero")
	}

	return nil
}
