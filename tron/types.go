package tron

import (
	"github.com/maticnetwork/bor/core/types"
	"math/rand"
	"strconv"
)

const (
	JsonRpcVersion       = "2.0"
	GetLogsMethod        = "eth_getLogs"
	GetTransactionByHash = "eth_getTransactionReceipt"
	GetBlockByNumber     = "eth_getBlockByNumber"
	MAXQueryAddress      = 3
)

type NewFilter struct {
	Address   []string `json:"address"`
	FromBlock string   `json:"fromBlock"`
	ToBlock   string   `json:"toBlock"`
}
type FilterEventParams struct {
	BaseQueryParam
	Method  string      `json:"method"`
	Params  []NewFilter `json:"params"`
}
type FilterOtherParams struct {
	BaseQueryParam
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}

type BaseQueryParam struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      string      `json:"id"`
}
type FilterEventResponse struct {
	BaseQueryParam
	Result  []types.Log `json:result`
}

type FilterTxResponse struct {
	BaseQueryParam
	Result  types.Receipt `json:result`
}

func  GetDefaultBaseParm() BaseQueryParam {
	param := BaseQueryParam{
		Jsonrpc: JsonRpcVersion,
		Id:      strconv.FormatInt(int64(rand.Int()%100), 10),
	}
	return param
}

type Block struct {
	BlockHeader *BlockHeader  `json:"block_header"`
}
type BlockHeader struct {
	RawData *RawData `json:"raw_data"`
}
type RawData struct {
	Number *int64  `json:"number"`
}
