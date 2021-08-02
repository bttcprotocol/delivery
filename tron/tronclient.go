package tron

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"

	ethCrypto "github.com/maticnetwork/bor/crypto/secp256k1"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/maticnetwork/bor/accounts/abi"
	"github.com/maticnetwork/bor/common"
	"github.com/maticnetwork/heimdall/contracts/rootchain"
	"github.com/maticnetwork/heimdall/tron/pb"
	"github.com/maticnetwork/heimdall/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
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

//
// private abi methods
//
func getABI(data string) (abi.ABI, error) {
	return abi.JSON(strings.NewReader(data))
}

//Package goLang sha256 hash algorithm.
func hash(s []byte) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write(s)
	if err != nil {
		return nil, err
	}
	bs := h.Sum(nil)
	return bs, nil
}

func (tc *Client) triggerContract(ownerAddress, contractAddress string, data []byte) (*pb.Transaction, error) {
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

func (tc *Client) triggerConstantContract(contractAddress string, data []byte) ([]byte, error) {
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
	if response.Result.Code != pb.Return_SUCCESS {
		return nil, fmt.Errorf("code:%v message:%v", response.Result.Code, string(response.Result.Message))
	}
	return response.ConstantResult[0], nil
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
	data, err := tc.triggerConstantContract(contractAddress, btsPack)
	if err != nil {
		return 0, err
	}

	// Unpack the results
	var (
		ret0 = new(*big.Int)
	)
	if err = tc.rootchainABI.Unpack(ret0, "currentHeaderBlock", data); err != nil {
		return 0, nil
	}
	return (*ret0).Uint64() / childBlockInterval, nil
}

// HeaderBlocks is a free data retrieval call binding the contract method 0x41539d4a.
//
// Solidity: function headerBlocks(uint256 ) view returns(bytes32 root, uint256 start, uint256 end, uint256 createdAt, address proposer)
func (tc *Client) GetHeaderInfo(number uint64, contractAddress string, childBlockInterval uint64) (
	root common.Hash,
	start uint64,
	end uint64,
	createdAt uint64,
	proposer types.HeimdallAddress,
	err error,
) {
	// Pack the input
	btsPack, err := tc.rootchainABI.Pack("headerBlocks",
		big.NewInt(0).Mul(big.NewInt(0).SetUint64(number), big.NewInt(0).SetUint64(childBlockInterval)))
	if err != nil {
		return root, 0, 0, 0, types.HeimdallAddress{}, err
	}

	// Call
	data, err := tc.triggerConstantContract(contractAddress, btsPack)
	if err != nil {
		return root, 0, 0, 0, types.HeimdallAddress{}, err
	}

	// Unpack the results
	ret := new(struct {
		Root      [32]byte
		Start     *big.Int
		End       *big.Int
		CreatedAt *big.Int
		Proposer  common.Address
	})
	if err = tc.rootchainABI.Unpack(ret, "headerBlocks", data); err != nil {
		return root, 0, 0, 0, types.HeimdallAddress{}, err
	}

	return ret.Root, ret.Start.Uint64(), ret.End.Uint64(),
		ret.CreatedAt.Uint64(), types.HeimdallAddress(ret.Proposer), nil
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
	data, err := tc.triggerConstantContract(contractAddress, btsPack)
	if err != nil {
		return 0, err
	}
	// Unpack the results
	var (
		ret0 = new(*big.Int)
	)
	if err = tc.rootchainABI.Unpack(ret0, "getLastChildBlock", data); err != nil {
		return 0, nil
	}
	return (*ret0).Uint64(), nil
}

func (tc *Client) SendCheckpoint(privateKey secp256k1.PrivKeySecp256k1, signedData []byte, sigs [][3]*big.Int, rootChainAddress string) (string, error) {
	data, err := tc.rootchainABI.Pack("submitCheckpoint", signedData, sigs)
	if err != nil {
		return "", err
	}
	privateKey.PubKey()
	// trigger
	trx, err := tc.triggerContract(privateKey.PubKey().Address().String(), rootChainAddress, data)
	if err != nil {
		return "", err
	}
	rawData, _ := proto.Marshal(trx.GetRawData())
	hash, err := hash(rawData)
	if err != nil {
		return "", err
	}

	signature, err := ethCrypto.Sign(hash, privateKey[:])
	if err != nil {
		return "", err
	}

	trx.Signature = append(trx.GetSignature(), signature)

	result, err := tc.client.BroadcastTransaction(context.Background(), trx)
	if err != nil {
		return "", err
	}
	if result.Code != pb.Return_SUCCESS {
		return "", fmt.Errorf("code:%v message:%v", result.Code, string(result.Message))
	}
	return hex.EncodeToString(result.Message), nil
}
