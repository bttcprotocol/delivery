package types

import (
	"math/big"
)

const (
	// DefaultLogIndexUnit default tx hash + log index unit
	DefaultLogIndexUnit = 100000

	DefaultChainIdUnit = 100
)

var (
	// CoinDecimals is the amount of staking tokens required for 1 unit
	CoinDecimals = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
)
