package types

// query endpoints supported by the chain-manager Querier
const (
	QueryParams        = "params"
	QueryNewChainParam = "chain-params"
)

// QueryChainParams defines the params for querying accounts.
type QueryChainParams struct {
	RootChain string
}

// NewQueryChainParams creates a new instance of NewQueryChainParams.
func NewQueryChainParams(rootChain string) QueryChainParams {
	return QueryChainParams{
		RootChain: rootChain,
	}
}
