package types

// query endpoints supported by the auth Querier
const (
	QueryParams               = "params"
	QueryAckCount             = "ack-count"
	QueryEpoch                = "epoch"
	QueryCheckpoint           = "checkpoint"
	QueryCheckpointBuffer     = "checkpoint-buffer"
	QueryCheckpointSyncBuffer = "checkpoint-sync"
	QueryCheckpointActivation = "checkpoint-activation"
	QueryLastNoAck            = "last-no-ack"
	QueryCheckpointList       = "checkpoint-list"
	QueryNextCheckpoint       = "next-checkpoint"
	QueryProposer             = "is-proposer"
	QueryCurrentProposer      = "current-proposer"
	StakingQuerierRoute       = "staking"
)

// QueryCheckpointParams defines the params for querying accounts.
type QueryCheckpointParams struct {
	Number    uint64
	RootChain string
}

// NewQueryCheckpointParams creates a new instance of QueryCheckpointHeaderIndex.
func NewQueryCheckpointParams(number uint64, rootChain string) QueryCheckpointParams {
	return QueryCheckpointParams{
		Number:    number,
		RootChain: rootChain,
	}
}

// QueryBorChainID defines the params for querying with bor chain id
type QueryBorChainID struct {
	BorChainID string
}

// NewQueryBorChainID creates a new instance of QueryBorChainID with give chain id
func NewQueryBorChainID(chainID string) QueryBorChainID {
	return QueryBorChainID{BorChainID: chainID}
}
