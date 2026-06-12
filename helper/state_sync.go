package helper

import (
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type StateSyncEventType int

const (
	StateSyncEventUnknown StateSyncEventType = iota
	StateSyncEventDeposit
	StateSyncEventMapToken
)

const MintableERC20TokenType = "0x5ffef61af1560b9aefc0e42aaa0f9464854ab113ab7b8bfab271be94cdb1d053"

var (
	StateSyncDepositTypeHash  = crypto.Keccak256Hash([]byte("DEPOSIT"))
	StateSyncMapTokenTypeHash = crypto.Keccak256Hash([]byte("MAP_TOKEN"))
	MintableERC20TokenHash    = common.HexToHash(MintableERC20TokenType)

	stateSyncBytes32Type, _ = abi.NewType("bytes32", "", nil)
	stateSyncBytesType, _   = abi.NewType("bytes", "", nil)
	stateSyncAddressType, _ = abi.NewType("address", "", nil)
	stateSyncUint256Type, _ = abi.NewType("uint256", "", nil)

	rootChainManagerProxyABI, _ = abi.JSON(strings.NewReader(`[{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"tokenToType","outputs":[{"name":"","type":"bytes32"}],"payable":false,"stateMutability":"view","type":"function"}]`))
)

type StateSyncData struct {
	EventType  StateSyncEventType
	RootToken  common.Address
	ChildToken common.Address
}

// ParseStateSyncData decodes StateSender data encoded as abi.encode(eventType, syncData).
func ParseStateSyncData(data []byte) (*StateSyncData, error) {
	stateData, err := parseStateSyncPayload(data)
	if err == nil && stateData.EventType != StateSyncEventUnknown {
		return stateData, nil
	}

	return stateData, nil
}

func parseStateSyncPayload(data []byte) (*StateSyncData, error) {
	arguments := abi.Arguments{
		{Type: stateSyncBytes32Type},
		{Type: stateSyncBytesType},
	}

	stateData, err := arguments.Unpack(data)
	if err != nil {
		return nil, err
	}
	if len(stateData) != len(arguments) {
		return nil, errors.New("invalid state sync data")
	}

	eventTypeBytes, ok := stateData[0].([32]byte)
	if !ok {
		return nil, errors.New("invalid state sync event type")
	}

	syncData, ok := stateData[1].([]byte)
	if !ok {
		return nil, errors.New("invalid state sync payload")
	}

	eventType := common.BytesToHash(eventTypeBytes[:])
	switch eventType {
	case StateSyncDepositTypeHash:
		return parseDepositStateSyncData(syncData)
	case StateSyncMapTokenTypeHash:
		return parseMapTokenStateSyncData(syncData)
	default:
		return &StateSyncData{EventType: StateSyncEventUnknown}, nil
	}
}

func unwrapStateSyncBytes(data []byte) ([]byte, error) {
	arguments := abi.Arguments{
		{Type: stateSyncBytesType},
	}

	stateData, err := arguments.Unpack(data)
	if err != nil {
		return nil, err
	}
	if len(stateData) != len(arguments) {
		return nil, errors.New("invalid wrapped state sync data")
	}

	wrappedData, ok := stateData[0].([]byte)
	if !ok {
		return nil, errors.New("invalid wrapped state sync payload")
	}

	return wrappedData, nil
}

func parseDepositStateSyncData(data []byte) (*StateSyncData, error) {
	arguments := abi.Arguments{
		{Type: stateSyncAddressType},
		{Type: stateSyncAddressType},
		{Type: stateSyncUint256Type},
		{Type: stateSyncBytesType},
	}

	ret, err := arguments.Unpack(data)
	if err != nil {
		return nil, err
	}
	if len(ret) != len(arguments) {
		return nil, errors.New("invalid deposit state sync data")
	}

	rootToken, ok := ret[1].(common.Address)
	if !ok {
		return nil, errors.New("invalid deposit root token")
	}
	if _, ok := ret[2].(*big.Int); !ok {
		return nil, errors.New("invalid deposit chain id")
	}

	return &StateSyncData{
		EventType: StateSyncEventDeposit,
		RootToken: rootToken,
	}, nil
}

func parseMapTokenStateSyncData(data []byte) (*StateSyncData, error) {
	arguments := abi.Arguments{
		{Type: stateSyncAddressType},
		{Type: stateSyncAddressType},
		{Type: stateSyncUint256Type},
		{Type: stateSyncBytes32Type},
	}

	ret, err := arguments.Unpack(data)
	if err != nil {
		return nil, err
	}
	if len(ret) != len(arguments) {
		return nil, errors.New("invalid map token state sync data")
	}

	rootToken, ok := ret[0].(common.Address)
	if !ok {
		return nil, errors.New("invalid map token root token")
	}
	childToken, ok := ret[1].(common.Address)
	if !ok {
		return nil, errors.New("invalid map token child token")
	}
	if _, ok := ret[2].(*big.Int); !ok {
		return nil, errors.New("invalid map token chain id")
	}

	return &StateSyncData{
		EventType:  StateSyncEventMapToken,
		RootToken:  rootToken,
		ChildToken: childToken,
	}, nil
}
