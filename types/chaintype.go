package types

const (
	RootChainTypeEth  = "eth"
	RootChainTypeTron = "tron"

	RootChainTypeStake = RootChainTypeEth
)

var chainMap = []string{RootChainTypeEth, RootChainTypeTron}
var chainIDMap = map[string]byte{RootChainTypeEth: 1, RootChainTypeTron: 2}

func GetRootChainID(rootChain string) byte {
	return chainIDMap[rootChain]
}

func GetRootChainMap() []string {
	return chainMap
}

func GetRootChainIDMap() map[string]byte {
	return chainIDMap
}
