package types

const (
	RootChainTypeEth  = "eth"
	RootChainTypeTron = "tron"
	RootChainTypeBsc  = "bsc"

	RootChainTypeStake = RootChainTypeTron
)

var chainIDMap = map[string]byte{RootChainTypeTron: 1, RootChainTypeEth: 2, RootChainTypeBsc: 3}

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
