package types

const (
	RootChainTypeEth  = "eth"
	RootChainTypeTron = "tron"

	RootChainTypeStake   = RootChainTypeTron
	DefaultRootChainType = RootChainTypeEth
)

var chainIDMap = map[string]byte{RootChainTypeEth: 1, RootChainTypeTron: 2}

func GetRootChainID(rootChain string) byte {
	return chainIDMap[rootChain]
}

func GetRootChainName(rootChainID uint64) string {
	for chainName, chainID := range chainIDMap {
		if uint64(chainID) == rootChainID {
			return chainName
		}
	}
	return "no-chain"
}

func GetRootChainIDMap() map[string]byte {
	return chainIDMap
}
