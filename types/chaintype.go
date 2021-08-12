package types

const (
	RootChainTypeEth  = "eth"
	RootChainTypeTron = "tron"

	RootChainTypeStake   = RootChainTypeEth
	DefaultRootChainType = RootChainTypeEth
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

func ConvertChainTypeToInt(chainType string) int64 {
	switch chainType {
	case RootChainTypeEth:
		return 1
	case RootChainTypeTron:
		return 2
	default:
		return 0
	}
}
