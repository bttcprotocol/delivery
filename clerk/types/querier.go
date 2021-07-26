package types

import (
	"time"
)

// query endpoints supported by the auth Querier
const (
	QueryRecord             = "record"
	QueryRecordList         = "record-list"
	QueryRecordListWithTime = "record-list-time"
	QueryRecordSequence     = "record-sequence"
)

// QueryRecordParams defines the params for querying accounts.
type QueryRecordParams struct {
	RecordID uint64
}

// QueryRootChainRecordParams defines the params for querying accounts.
type QueryRootChainRecordParams struct {
	RecordID      uint64
	RootChainType string
}

// QueryRecordSequenceParams defines the params for querying an account Sequence.
type QueryRecordSequenceParams struct {
	TxHash        string
	LogIndex      uint64
	RootChainType string
}

// QueryRecordTimePaginationParams defines the params for querying records with time.
type QueryRecordTimePaginationParams struct {
	FromTime time.Time
	ToTime   time.Time
	Page     uint64
	Limit    uint64
}

// NewQueryRecordParams creates a new instance of QueryRecordParams.
func NewQueryRecordParams(recordID uint64) QueryRecordParams {
	return QueryRecordParams{RecordID: recordID}
}

// NewQueryRecordParams creates a new instance of QueryRecordParams.
func NewQueryRecordRootChainParams(recordID uint64, rootChainType string) QueryRootChainRecordParams {
	return QueryRootChainRecordParams{RecordID: recordID, RootChainType: rootChainType}
}

// NewQueryRecordSequenceParams creates a new instance of QuerySequenceParams.
func NewQueryRecordSequenceParams(txHash string, logIndex uint64, rootChainTyoe string) QueryRecordSequenceParams {
	return QueryRecordSequenceParams{TxHash: txHash, LogIndex: logIndex, RootChainType: rootChainTyoe}
}

// NewQueryTimeRangePaginationParams creates a new instance of NewQueryTimeRangePaginationParams.
func NewQueryTimeRangePaginationParams(fromTime, toTime time.Time, page, limit uint64) QueryRecordTimePaginationParams {
	return QueryRecordTimePaginationParams{FromTime: fromTime, ToTime: toTime, Page: page, Limit: limit}
}
