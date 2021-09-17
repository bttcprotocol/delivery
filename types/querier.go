package types

// QueryPaginationParams defines the params for querying accounts.
type QueryPaginationParams struct {
	Page      uint64
	Limit     uint64
	RootChain string
}

// NewQueryPaginationParams creates a new instance of QueryPaginationParams.
func NewQueryPaginationParams(page uint64, limit uint64, rootChain string) QueryPaginationParams {
	return QueryPaginationParams{Page: page, Limit: limit, RootChain: rootChain}
}
