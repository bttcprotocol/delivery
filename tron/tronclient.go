package tron

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/maticnetwork/heimdall/contracts/rootchain"
	"github.com/maticnetwork/heimdall/tron/pb"
	"google.golang.org/grpc"
)

// Client defines typed wrappers for the Tron RPC API.
type Client struct {
	client       pb.WalletClient
	rootchainABI abi.ABI
}

// NewClient creates a client that uses the given RPC client.
func NewClient(url string) *Client {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		os.Exit(0)
	}
	rootchainABI, err := getABI(rootchain.RootchainABI)
	if err != nil {
		os.Exit(0)
	}
	return &Client{
		client:       pb.NewWalletClient(conn),
		rootchainABI: rootchainABI,
	}
}

// private abi methods
func getABI(data string) (abi.ABI, error) {
	return abi.JSON(strings.NewReader(data))
}

func (tc *Client) TriggerContract(ownerAddress, contractAddress string, data []byte) (*pb.Transaction, error) {
	response, err := tc.client.TriggerContract(context.Background(),
		&pb.TriggerSmartContract{
			OwnerAddress:    common.FromHex("41" + ownerAddress),
			ContractAddress: common.FromHex(contractAddress),
			CallValue:       0,
			Data:            data,
			CallTokenValue:  0,
			TokenId:         0,
		})
	if err != nil {
		return nil, err
	}
	if response.Result.Code != pb.Return_SUCCESS {
		return nil, fmt.Errorf("code:%v message:%v", response.Result.Code, string(response.Result.Message))
	}
	return response.Transaction, nil
}

func (tc *Client) TriggerConstantContract(contractAddress string, data []byte) ([]byte, error) {
	response, err := tc.client.TriggerConstantContract(context.Background(),
		&pb.TriggerSmartContract{
			OwnerAddress:    nil,
			ContractAddress: common.FromHex(contractAddress),
			CallValue:       0,
			Data:            data,
			CallTokenValue:  0,
			TokenId:         0,
		})
	if err != nil {
		return nil, err
	}
	if response.Result.Code != pb.Return_SUCCESS || response.Transaction.GetRet()[0].Ret == pb.Transaction_Result_FAILED {
		return nil, fmt.Errorf("code:%v message:%v", response.Result.Code, string(response.Result.Message))
	}
	return response.ConstantResult[0], nil
}

func (tc *Client) TriggerConstantContractWithRetry(contractAddress string, data []byte) ([]byte, error) {
	const maxRetries = 5

	var response []byte
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		response, err = tc.TriggerConstantContract(contractAddress, data)

		if err == nil && response != nil {
			log.Info("Successfully trigger tron constant contract", "attempt", attempt)
			break
		}
		log.Error("Failed to trigger tron constant contract",
			"err", err, "attempt", attempt, "maxRetries", maxRetries)
		if attempt < maxRetries-1 {
			delay := attempt + 1
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}
	return response, err
}
func (tc *Client) GetNowBlock(ctx context.Context) (int64, error) {
	block, err := tc.client.GetNowBlock2(ctx, &pb.EmptyMessage{})
	if err != nil {
		return 0, err
	}
	return block.BlockHeader.RawData.Number, nil
}

// CurrentHeaderBlock is a free data retrieval call binding the contract method 0xec7e4855.
//
// Solidity: function currentHeaderBlock() view returns(uint256)
func (tc *Client) CurrentHeaderBlock(contractAddress string, childBlockInterval uint64) (uint64, error) {
	// Pack the input
	btsPack, err := tc.rootchainABI.Pack("currentHeaderBlock")
	if err != nil {
		return 0, err
	}

	// Call
	data, err := tc.TriggerConstantContractWithRetry(contractAddress, btsPack)
	if err != nil {
		return 0, err
	}

	// Unpack the results
	var (
		ret0 = new(*big.Int)
	)

	if err := tc.rootchainABI.UnpackIntoInterface(ret0, "currentHeaderBlock", data); err != nil {
		return 0, nil
	}
	return (*ret0).Uint64() / childBlockInterval, nil
}

// GetLastChildBlock is a free data retrieval call binding the contract method 0xb87e1b66.
//
// Solidity: function getLastChildBlock() view returns(uint256)
func (tc *Client) GetLastChildBlock(contractAddress string) (uint64, error) {
	// Pack the input
	btsPack, err := tc.rootchainABI.Pack("getLastChildBlock")
	if err != nil {
		return 0, err
	}
	data, err := tc.TriggerConstantContractWithRetry(contractAddress, btsPack)
	if err != nil {
		return 0, err
	}
	// Unpack the results
	var (
		ret0 = new(*big.Int)
	)

	if err := tc.rootchainABI.UnpackIntoInterface(ret0, "getLastChildBlock", data); err != nil {
		return 0, nil
	}
	return (*ret0).Uint64(), nil
}

func (tc *Client) BroadcastTransaction(ctx context.Context, trx *pb.Transaction) error {
	result, err := tc.client.BroadcastTransaction(ctx, trx)
	if err != nil {
		return err
	}
	if result.Code != pb.Return_SUCCESS {
		return fmt.Errorf("code:%v message:%v", result.Code, string(result.Message))
	}
	return nil
}
