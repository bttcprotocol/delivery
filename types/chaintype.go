package types

const (
	RootChainTypeEth  = "eth"
	RootChainTypeTron = "tron"

	RootChainTypeStake   = RootChainTypeEth
	DefaultRootChainType = RootChainTypeEth
)

var chainIDMap = map[string]byte{RootChainTypeEth: 1, RootChainTypeTron: 2}

func GetRootChainID(rootChain string) byte {
	return chainIDMap[rootChain]
}

func GetRootChainIDMap() map[string]byte {
	return chainIDMap
}
