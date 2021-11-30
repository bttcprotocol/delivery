package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/maticnetwork/heimdall/tron"

	lru "github.com/hashicorp/golang-lru"
	"github.com/maticnetwork/bor/accounts/abi"
	"github.com/maticnetwork/bor/common"
	ethTypes "github.com/maticnetwork/bor/core/types"
	"github.com/maticnetwork/bor/ethclient"
	"github.com/maticnetwork/bor/rpc"
	"github.com/maticnetwork/heimdall/contracts/erc20"
	"github.com/maticnetwork/heimdall/contracts/rootchain"
	"github.com/maticnetwork/heimdall/contracts/slashmanager"
	"github.com/maticnetwork/heimdall/contracts/stakemanager"
	"github.com/maticnetwork/heimdall/contracts/stakinginfo"
	"github.com/maticnetwork/heimdall/contracts/statereceiver"
	"github.com/maticnetwork/heimdall/contracts/statesender"
	"github.com/maticnetwork/heimdall/contracts/validatorset"

	"github.com/maticnetwork/heimdall/types"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// IContractCaller represents contract caller
type IContractCaller interface {
	GetHeaderInfo(headerID uint64, rootChainInstance *rootchain.Rootchain, childBlockInterval uint64) (root common.Hash, start, end, createdAt uint64, proposer types.HeimdallAddress, err error)
	GetRootHash(start uint64, end uint64, checkpointLength uint64) ([]byte, error)
	GetValidatorInfo(valID types.ValidatorID, stakingInfoInstance *stakinginfo.Stakinginfo) (validator types.Validator, err error)
	GetLastChildBlock(rootChainInstance *rootchain.Rootchain) (uint64, error)
	CurrentHeaderBlock(rootChainInstance *rootchain.Rootchain, childBlockInterval uint64) (uint64, error)
	GetBalance(address common.Address) (*big.Int, error)
	SendCheckpoint(sigedData []byte, sigs [][3]*big.Int, rootchainAddress common.Address, rootChainInstance *rootchain.Rootchain, rootChain string) (err error)
	SendTronCheckpoint(signedData []byte, sigs [][3]*big.Int, rootChainAddress string) error
	SendTick(sigedData []byte, sigs []byte, slashManagerAddress common.Address, slashManagerInstance *slashmanager.Slashmanager) (err error)
	GetCheckpointSign(txHash common.Hash) ([]byte, []byte, []byte, error)
	GetMainChainBlock(*big.Int, string) (*ethTypes.Header, error)
	GetMaticChainBlock(*big.Int) (*ethTypes.Header, error)
	GetConfirmedTxReceipt(common.Hash, uint64, string) (*ethTypes.Receipt, error)
	GetBlockNumberFromTxHash(common.Hash) (*big.Int, error)

	// decode header event
	DecodeNewHeaderBlockEvent(common.Address, *ethTypes.Receipt, uint64) (*rootchain.RootchainNewHeaderBlock, error)
	// decode validator events
	DecodeValidatorTopupFeesEvent(common.Address, *ethTypes.Receipt, uint64) (*stakinginfo.StakinginfoTopUpFee, error)
	DecodeValidatorJoinEvent(common.Address, *ethTypes.Receipt, uint64) (*stakinginfo.StakinginfoStaked, error)
	DecodeValidatorStakeUpdateEvent(common.Address, *ethTypes.Receipt, uint64) (*stakinginfo.StakinginfoStakeUpdate, error)
	DecodeValidatorExitEvent(common.Address, *ethTypes.Receipt, uint64) (*stakinginfo.StakinginfoUnstakeInit, error)
	DecodeSignerUpdateEvent(common.Address, *ethTypes.Receipt, uint64) (*stakinginfo.StakinginfoSignerChange, error)
	// decode state events
	DecodeStateSyncedEvent(common.Address, *ethTypes.Receipt, uint64) (*statesender.StatesenderStateSynced, error)

	// decode slashing events
	DecodeSlashedEvent(common.Address, *ethTypes.Receipt, uint64) (*stakinginfo.StakinginfoSlashed, error)
	DecodeUnJailedEvent(common.Address, *ethTypes.Receipt, uint64) (*stakinginfo.StakinginfoUnJailed, error)

	GetMainTxReceipt(common.Hash, string) (*ethTypes.Receipt, error)
	GetMaticTxReceipt(common.Hash) (*ethTypes.Receipt, error)
	ApproveTokens(*big.Int, common.Address, common.Address, *erc20.Erc20) error
	StakeFor(common.Address, *big.Int, *big.Int, bool, common.Address, *stakemanager.Stakemanager) error
	CurrentAccountStateRoot(stakingInfoInstance *stakinginfo.Stakinginfo) ([32]byte, error)

	// bor related contracts
	CurrentSpanNumber(validatorset *validatorset.Validatorset) (Number *big.Int)
	GetSpanDetails(id *big.Int, validatorset *validatorset.Validatorset) (*big.Int, *big.Int, *big.Int, error)
	CurrentStateCounter(stateSenderInstance *statesender.Statesender) (Number *big.Int)
	CheckIfBlocksExist(end uint64) bool

	// staking sync
	GetMainStakingSyncNonce(validatorID uint64, stakingManagerInstance *stakemanager.Stakemanager) (nonce uint64)
	GetTronStakingSyncNonce(validatorID uint64, stakingManagerAddress string) (nonce uint64)
	SendMainStakingSync(stakingType string, sigedData []byte, sigs [][3]*big.Int, stakingManagerAddress common.Address, stakingManagerInstance *stakemanager.Stakemanager, rootChain string) (err error)
	SendTronStakingSync(stakingType string, sigedData []byte, sigs [][3]*big.Int, stakingManagerAddress string) (err error)

	GetRootChainInstance(rootchainAddress common.Address, rootChain string) (*rootchain.Rootchain, error)
	GetStakingInfoInstance(stakingInfoAddress common.Address, rootChain string) (*stakinginfo.Stakinginfo, error)
	GetValidatorSetInstance(validatorSetAddress common.Address) (*validatorset.Validatorset, error)
	GetStakeManagerInstance(stakingManagerAddress common.Address, rootChain string) (*stakemanager.Stakemanager, error)
	GetSlashManagerInstance(slashManagerAddress common.Address) (*slashmanager.Slashmanager, error)
	GetStateSenderInstance(stateSenderAddress common.Address) (*statesender.Statesender, error)
	GetStateReceiverInstance(stateReceiverAddress common.Address) (*statereceiver.Statereceiver, error)
	GetMaticTokenInstance(maticTokenAddress common.Address) (*erc20.Erc20, error)

	GetTronHeaderInfo(headerID uint64, rootChainAddress string, childBlockInterval uint64) (root common.Hash, start, end, createdAt uint64, proposer types.HeimdallAddress, err error)
	GetTronEventsByContractAddress(address []string, from, to int64) ([]ethTypes.Log, error)
	GetTronTransactionReceipt(txID string) (*ethTypes.Receipt, error)
	GetTronLatestBlockNumber() (int64, error)

	// checkpoint sync
	GetSyncedCheckpointId(rootChain string, contractAddress string) (currentHeader uint64, err error)
	GetStartListenBlock(rootChainType string) uint64

	// new chain
	DecodeNewChainEvent(common.Address, *ethTypes.Receipt, uint64) (*rootchain.RootchainNewChain, error)
}

// ContractCaller contract caller
type ContractCaller struct {
	MainChainClient  *ethclient.Client
	MainChainRPC     *rpc.Client
	TronChainRPC     *tron.Client
	MaticChainClient *ethclient.Client
	MaticChainRPC    *rpc.Client
	BscChainClient   *ethclient.Client
	BscChainRPC      *rpc.Client

	RootChainABI     abi.ABI
	StakingInfoABI   abi.ABI
	ValidatorSetABI  abi.ABI
	StateReceiverABI abi.ABI
	StateSenderABI   abi.ABI
	StakeManagerABI  abi.ABI
	SlashManagerABI  abi.ABI
	MaticTokenABI    abi.ABI

	ReceiptCache     *lru.Cache
	LatestBlockCache map[string]uint64

	ContractInstanceCache map[string]interface{}
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

type rpcTransaction struct {
	txExtraInfo
}

// NewContractCaller contract caller
func NewContractCaller() (contractCallerObj ContractCaller, err error) {
	contractCallerObj.MainChainClient = GetMainClient()
	contractCallerObj.MaticChainClient = GetMaticClient()
	contractCallerObj.BscChainClient = GetBscClient()
	contractCallerObj.TronChainRPC = GetTronChainRPCClient()
	contractCallerObj.MainChainRPC = GetMainChainRPCClient()
	contractCallerObj.BscChainRPC = GetBscChainRPCClient()
	contractCallerObj.MaticChainRPC = GetMaticRPCClient()
	contractCallerObj.ReceiptCache, _ = NewLru(5000)

	//
	// ABIs
	//

	if contractCallerObj.RootChainABI, err = getABI(string(rootchain.RootchainABI)); err != nil {
		return
	}

	if contractCallerObj.StakingInfoABI, err = getABI(string(stakinginfo.StakinginfoABI)); err != nil {
		return
	}

	if contractCallerObj.ValidatorSetABI, err = getABI(string(validatorset.ValidatorsetABI)); err != nil {
		return
	}

	if contractCallerObj.StateReceiverABI, err = getABI(string(statereceiver.StatereceiverABI)); err != nil {
		return
	}

	if contractCallerObj.StateSenderABI, err = getABI(string(statesender.StatesenderABI)); err != nil {
		return
	}

	if contractCallerObj.StakeManagerABI, err = getABI(string(stakemanager.StakemanagerABI)); err != nil {
		return
	}

	if contractCallerObj.SlashManagerABI, err = getABI(string(slashmanager.SlashmanagerABI)); err != nil {
		return
	}

	if contractCallerObj.MaticTokenABI, err = getABI(string(erc20.Erc20ABI)); err != nil {
		return
	}

	contractCallerObj.ContractInstanceCache = make(map[string]interface{})

	return
}

// GetRootChainInstance returns RootChain contract instance for selected base chain
func (c *ContractCaller) GetRootChainInstance(rootchainAddress common.Address, rootChain string) (*rootchain.Rootchain, error) {
	cacheKey := rootchainAddress.String() + rootChain
	contractInstance, ok := c.ContractInstanceCache[cacheKey]
	if !ok {
		var client *ethclient.Client
		switch rootChain {
		case hmTypes.RootChainTypeEth:
			client = mainChainClient
		case hmTypes.RootChainTypeBsc:
			client = bscChainClient
		}
		ci, err := rootchain.NewRootchain(rootchainAddress, client)
		c.ContractInstanceCache[cacheKey] = ci
		return ci, err
	}
	return contractInstance.(*rootchain.Rootchain), nil
}

// GetStakingInfoInstance returns stakinginfo contract instance for selected base chain
func (c *ContractCaller) GetStakingInfoInstance(stakingInfoAddress common.Address, rootChain string) (*stakinginfo.Stakinginfo, error) {
	cacheKey := stakingInfoAddress.String() + rootChain
	contractInstance, ok := c.ContractInstanceCache[cacheKey]
	if !ok {
		var client *ethclient.Client
		switch rootChain {
		case hmTypes.RootChainTypeEth:
			client = mainChainClient
		case hmTypes.RootChainTypeBsc:
			client = bscChainClient
		}
		ci, err := stakinginfo.NewStakinginfo(stakingInfoAddress, client)
		c.ContractInstanceCache[cacheKey] = ci
		return ci, err
	}
	return contractInstance.(*stakinginfo.Stakinginfo), nil
}

// GetValidatorSetInstance returns stakinginfo contract instance for selected base chain
func (c *ContractCaller) GetValidatorSetInstance(validatorSetAddress common.Address) (*validatorset.Validatorset, error) {
	contractInstance, ok := c.ContractInstanceCache[validatorSetAddress.String()]
	if !ok {
		ci, err := validatorset.NewValidatorset(validatorSetAddress, mainChainClient)
		c.ContractInstanceCache[validatorSetAddress.String()] = ci
		return ci, err

	}
	return contractInstance.(*validatorset.Validatorset), nil
}

// GetStakeManagerInstance returns stakinginfo contract instance for selected base chain
func (c *ContractCaller) GetStakeManagerInstance(stakingManagerAddress common.Address, rootChain string) (*stakemanager.Stakemanager, error) {
	cacheKey := stakingManagerAddress.String() + rootChain
	contractInstance, ok := c.ContractInstanceCache[cacheKey]
	if !ok {
		var client *ethclient.Client
		switch rootChain {
		case hmTypes.RootChainTypeEth:
			client = mainChainClient
		case hmTypes.RootChainTypeBsc:
			client = bscChainClient
		}
		ci, err := stakemanager.NewStakemanager(stakingManagerAddress, client)
		c.ContractInstanceCache[cacheKey] = ci
		return ci, err
	}
	return contractInstance.(*stakemanager.Stakemanager), nil
}

// GetSlashManagerInstance returns slashManager contract instance for selected base chain
func (c *ContractCaller) GetSlashManagerInstance(slashManagerAddress common.Address) (*slashmanager.Slashmanager, error) {
	contractInstance, ok := c.ContractInstanceCache[slashManagerAddress.String()]
	if !ok {
		ci, err := slashmanager.NewSlashmanager(slashManagerAddress, mainChainClient)
		c.ContractInstanceCache[slashManagerAddress.String()] = ci
		return ci, err
	}
	return contractInstance.(*slashmanager.Slashmanager), nil
}

// GetStateSenderInstance returns stakinginfo contract instance for selected base chain
func (c *ContractCaller) GetStateSenderInstance(stateSenderAddress common.Address) (*statesender.Statesender, error) {
	contractInstance, ok := c.ContractInstanceCache[stateSenderAddress.String()]
	if !ok {
		ci, err := statesender.NewStatesender(stateSenderAddress, mainChainClient)
		c.ContractInstanceCache[stateSenderAddress.String()] = ci
		return ci, err
	}
	return contractInstance.(*statesender.Statesender), nil
}

// GetStateReceiverInstance returns stakinginfo contract instance for selected base chain
func (c *ContractCaller) GetStateReceiverInstance(stateReceiverAddress common.Address) (*statereceiver.Statereceiver, error) {
	contractInstance, ok := c.ContractInstanceCache[stateReceiverAddress.String()]
	if !ok {
		ci, err := statereceiver.NewStatereceiver(stateReceiverAddress, mainChainClient)
		c.ContractInstanceCache[stateReceiverAddress.String()] = ci
		return ci, err
	}
	return contractInstance.(*statereceiver.Statereceiver), nil
}

// GetMaticTokenInstance returns stakinginfo contract instance for selected base chain
func (c *ContractCaller) GetMaticTokenInstance(maticTokenAddress common.Address) (*erc20.Erc20, error) {
	contractInstance, ok := c.ContractInstanceCache[maticTokenAddress.String()]
	if !ok {
		ci, err := erc20.NewErc20(maticTokenAddress, mainChainClient)
		c.ContractInstanceCache[maticTokenAddress.String()] = ci
		return ci, err
	}
	return contractInstance.(*erc20.Erc20), nil
}

// NewLru create instance of lru
func NewLru(size int) (*lru.Cache, error) {
	lruObj, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return lruObj, nil
}

// GetHeaderInfo get header info from checkpoint number
func (c *ContractCaller) GetHeaderInfo(number uint64, rootChainInstance *rootchain.Rootchain, childBlockInterval uint64) (
	root common.Hash,
	start uint64,
	end uint64,
	createdAt uint64,
	proposer types.HeimdallAddress,
	err error,
) {
	// get header from rootchain
	checkpointBigInt := big.NewInt(0).Mul(big.NewInt(0).SetUint64(number), big.NewInt(0).SetUint64(childBlockInterval))
	headerBlock, err := rootChainInstance.HeaderBlocks(nil, checkpointBigInt)
	if err != nil {
		Logger.Error("Unable to fetch checkpoint block", "error", err)
		return root, start, end, createdAt, proposer, errors.New("Unable to fetch checkpoint block")
	}

	return headerBlock.Root,
		headerBlock.Start.Uint64(),
		headerBlock.End.Uint64(),
		headerBlock.CreatedAt.Uint64(),
		types.BytesToHeimdallAddress(headerBlock.Proposer.Bytes()),
		nil
}

// GetRootHash get root hash from bor chain
func (c *ContractCaller) GetRootHash(start uint64, end uint64, checkpointLength uint64) ([]byte, error) {
	noOfBlock := end - start + 1

	if start > end {
		return nil, errors.New("start is greater than end")
	}

	if noOfBlock > checkpointLength {
		return nil, errors.New("number of headers requested exceeds")
	}

	rootHash, err := c.MaticChainClient.GetRootHash(context.Background(), start, end)
	if err != nil {
		return nil, errors.New("Could not fetch roothash from matic chain")
	}

	return common.FromHex(rootHash), nil
}

// GetLastChildBlock fetch current child block
func (c *ContractCaller) GetLastChildBlock(rootChainInstance *rootchain.Rootchain) (uint64, error) {
	GetLastChildBlock, err := rootChainInstance.GetLastChildBlock(nil)
	if err != nil {
		Logger.Error("Could not fetch current child block from rootchain contract", "Error", err)
		return 0, err
	}
	return GetLastChildBlock.Uint64(), nil
}

// CurrentHeaderBlock fetches current header block
func (c *ContractCaller) CurrentHeaderBlock(rootChainInstance *rootchain.Rootchain, childBlockInterval uint64) (uint64, error) {
	currentHeaderBlock, err := rootChainInstance.CurrentHeaderBlock(nil)
	if err != nil {
		Logger.Error("Could not fetch current header block from rootchain contract", "Error", err)
		return 0, err
	}
	return currentHeaderBlock.Uint64() / childBlockInterval, nil
}

// GetBalance get balance of account (returns big.Int balance wont fit in uint64)
func (c *ContractCaller) GetBalance(address common.Address) (*big.Int, error) {
	balance, err := c.MainChainClient.BalanceAt(context.Background(), address, nil)
	if err != nil {
		Logger.Error("Unable to fetch balance of account from root chain", "Error", err, "Address", address.String())
		return big.NewInt(0), err
	}

	return balance, nil
}

// GetValidatorInfo get validator info
func (c *ContractCaller) GetValidatorInfo(valID types.ValidatorID, stakingInfoInstance *stakinginfo.Stakinginfo) (validator types.Validator, err error) {
	// amount, startEpoch, endEpoch, signer, status, err := c.StakingInfoInstance.GetStakerDetails(nil, big.NewInt(int64(valID)))
	stakerDetails, err := stakingInfoInstance.GetStakerDetails(nil, big.NewInt(int64(valID)))
	if err != nil {
		Logger.Error("Error fetching validator information from stake manager", "error", err, "validatorId", valID, "status", stakerDetails.Status)
		return
	}

	newAmount, err := GetPowerFromAmount(stakerDetails.Amount)
	if err != nil {
		return
	}

	// newAmount
	validator = types.Validator{
		ID:          valID,
		VotingPower: newAmount.Int64(),
		StartEpoch:  stakerDetails.ActivationEpoch.Uint64(),
		EndEpoch:    stakerDetails.DeactivationEpoch.Uint64(),
		Signer:      types.BytesToHeimdallAddress(stakerDetails.Signer.Bytes()),
	}

	return validator, nil
}

// GetMainChainBlock returns main chain block header
func (c *ContractCaller) GetMainChainBlock(blockNum *big.Int, rootChain string) (header *ethTypes.Header, err error) {
	var latestBlock *ethTypes.Header
	switch rootChain {
	case hmTypes.RootChainTypeEth:
		latestBlock, err = c.MainChainClient.HeaderByNumber(context.Background(), blockNum)
	case hmTypes.RootChainTypeBsc:
		latestBlock, err = c.BscChainClient.HeaderByNumber(context.Background(), blockNum)
	default:
		return nil, errors.New("wrong chain type")
	}
	if err != nil {
		Logger.Error("Unable to connect to main chain", "Error", err)
		return
	}
	return latestBlock, nil
}

// GetMaticChainBlock returns child chain block header
func (c *ContractCaller) GetMaticChainBlock(blockNum *big.Int) (header *ethTypes.Header, err error) {
	latestBlock, err := c.MaticChainClient.HeaderByNumber(context.Background(), blockNum)
	if err != nil {
		Logger.Error("Unable to connect to matic chain", "Error", err)
		return
	}
	return latestBlock, nil
}

// GetBlockNumberFromTxHash gets block number of transaction
func (c *ContractCaller) GetBlockNumberFromTxHash(tx common.Hash) (*big.Int, error) {
	var rpcTx rpcTransaction
	if err := c.MainChainRPC.CallContext(context.Background(), &rpcTx, "eth_getTransactionByHash", tx); err != nil {
		return nil, err
	}

	if rpcTx.BlockNumber == nil {
		return nil, errors.New("No tx found")
	}

	blkNum := big.NewInt(0)
	blkNum, ok := blkNum.SetString(*rpcTx.BlockNumber, 0)
	if !ok {
		return nil, errors.New("unable to set string")
	}
	return blkNum, nil
}

// GetConfirmedTxReceipt returns confirmed tx receipt
func (c *ContractCaller) GetConfirmedTxReceipt(tx common.Hash, requiredConfirmations uint64, rootChain string) (*ethTypes.Receipt, error) {

	var receipt *ethTypes.Receipt = nil
	receiptCache, ok := c.ReceiptCache.Get(tx.String())

	if !ok {
		var err error

		// get main tx receipt
		receipt, err = c.GetMainTxReceipt(tx, rootChain)
		if err != nil {
			Logger.Error("Error while fetching mainchain receipt", "error", err, "txHash", tx.Hex())
			return nil, err
		}

		c.ReceiptCache.Add(tx.String(), receipt)
	} else {
		receipt, _ = receiptCache.(*ethTypes.Receipt)
	}

	Logger.Debug("Tx included in block", "root", rootChain, "block", receipt.BlockNumber.Uint64(), "tx", tx)

	latestBlkNumber := c.LatestBlockCache[rootChain]
	if latestBlkNumber-receipt.BlockNumber.Uint64() >= requiredConfirmations {
		Logger.Debug("receipt block is confirmed by cache",
			"root", rootChain, "latestBlockCached", latestBlkNumber, "receiptBlock", receipt.BlockNumber.Uint64())
		return receipt, nil
	}
	// get main chain block
	latestBlk, err := c.GetMainChainBlock(nil, rootChain)
	if err != nil {
		Logger.Error("error getting latest block from main chain", "Error", err)
		return nil, err
	}
	Logger.Debug("Latest block on main chain obtained", "root", rootChain, "Block", latestBlk.Number.Uint64())
	c.LatestBlockCache[rootChain] = latestBlk.Number.Uint64()
	diff := latestBlk.Number.Uint64() - receipt.BlockNumber.Uint64()
	if diff < requiredConfirmations {
		return nil, errors.New("not enough confirmations")
	}

	return receipt, nil
}

//
// Validator decode events
//

// DecodeNewHeaderBlockEvent represents new header block event
func (c *ContractCaller) DecodeNewHeaderBlockEvent(contractAddress common.Address, receipt *ethTypes.Receipt, logIndex uint64) (*rootchain.RootchainNewHeaderBlock, error) {
	event := new(rootchain.RootchainNewHeaderBlock)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex && bytes.Equal(vLog.Address.Bytes(), contractAddress.Bytes()) {
			found = true
			if err := UnpackLog(&c.RootChainABI, event, "NewHeaderBlock", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// DecodeValidatorTopupFeesEvent represents topup for fees tokens
func (c *ContractCaller) DecodeValidatorTopupFeesEvent(contractAddress common.Address, receipt *ethTypes.Receipt, logIndex uint64) (*stakinginfo.StakinginfoTopUpFee, error) {
	event := new(stakinginfo.StakinginfoTopUpFee)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex && bytes.Equal(vLog.Address.Bytes(), contractAddress.Bytes()) {
			found = true
			if err := UnpackLog(&c.StakingInfoABI, event, "TopUpFee", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// DecodeValidatorJoinEvent represents validator staked event
func (c *ContractCaller) DecodeValidatorJoinEvent(contractAddress common.Address, receipt *ethTypes.Receipt, logIndex uint64) (*stakinginfo.StakinginfoStaked, error) {
	event := new(stakinginfo.StakinginfoStaked)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex && bytes.Equal(vLog.Address.Bytes(), contractAddress.Bytes()) {
			found = true
			if err := UnpackLog(&c.StakingInfoABI, event, "Staked", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// DecodeValidatorStakeUpdateEvent represents validator stake update event
func (c *ContractCaller) DecodeValidatorStakeUpdateEvent(contractAddress common.Address, receipt *ethTypes.Receipt, logIndex uint64) (*stakinginfo.StakinginfoStakeUpdate, error) {
	event := new(stakinginfo.StakinginfoStakeUpdate)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex && bytes.Equal(vLog.Address.Bytes(), contractAddress.Bytes()) {
			found = true
			if err := UnpackLog(&c.StakingInfoABI, event, "StakeUpdate", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// DecodeValidatorExitEvent represents validator stake unstake event
func (c *ContractCaller) DecodeValidatorExitEvent(contractAddress common.Address, receipt *ethTypes.Receipt, logIndex uint64) (*stakinginfo.StakinginfoUnstakeInit, error) {
	event := new(stakinginfo.StakinginfoUnstakeInit)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex && bytes.Equal(vLog.Address.Bytes(), contractAddress.Bytes()) {
			found = true
			if err := UnpackLog(&c.StakingInfoABI, event, "UnstakeInit", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// DecodeSignerUpdateEvent represents sig update event
func (c *ContractCaller) DecodeSignerUpdateEvent(contractAddress common.Address, receipt *ethTypes.Receipt, logIndex uint64) (*stakinginfo.StakinginfoSignerChange, error) {
	event := new(stakinginfo.StakinginfoSignerChange)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex && bytes.Equal(vLog.Address.Bytes(), contractAddress.Bytes()) {
			found = true
			if err := UnpackLog(&c.StakingInfoABI, event, "SignerChange", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// DecodeStateSyncedEvent decode state sync data
func (c *ContractCaller) DecodeStateSyncedEvent(contractAddress common.Address, receipt *ethTypes.Receipt, logIndex uint64) (*statesender.StatesenderStateSynced, error) {
	event := new(statesender.StatesenderStateSynced)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex && bytes.Equal(vLog.Address.Bytes(), contractAddress.Bytes()) {
			found = true
			if err := UnpackLog(&c.StateSenderABI, event, "StateSynced", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// decode slashing events

// DecodeSlashedEvent represents tick ack on contract
func (c *ContractCaller) DecodeSlashedEvent(contractAddress common.Address, receipt *ethTypes.Receipt, logIndex uint64) (*stakinginfo.StakinginfoSlashed, error) {
	event := new(stakinginfo.StakinginfoSlashed)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex && bytes.Equal(vLog.Address.Bytes(), contractAddress.Bytes()) {
			found = true
			if err := UnpackLog(&c.StakingInfoABI, event, "Slashed", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// DecodeUnJailedEvent represents unjail on contract
func (c *ContractCaller) DecodeUnJailedEvent(contractAddress common.Address, receipt *ethTypes.Receipt, logIndex uint64) (*stakinginfo.StakinginfoUnJailed, error) {
	event := new(stakinginfo.StakinginfoUnJailed)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex && bytes.Equal(vLog.Address.Bytes(), contractAddress.Bytes()) {
			found = true
			if err := UnpackLog(&c.StakingInfoABI, event, "UnJailed", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

//
// Account root related functions
//

// CurrentAccountStateRoot get current account root from on chain
func (c *ContractCaller) CurrentAccountStateRoot(stakingInfoInstance *stakinginfo.Stakinginfo) ([32]byte, error) {
	accountStateRoot, err := stakingInfoInstance.GetAccountStateRoot(nil)

	if err != nil {
		Logger.Error("Unable to get current account state roor", "Error", err)
		var emptyArr [32]byte
		return emptyArr, err
	}

	return accountStateRoot, nil
}

//
// Span related functions
//

// CurrentSpanNumber get current span
func (c *ContractCaller) CurrentSpanNumber(validatorSetInstance *validatorset.Validatorset) (Number *big.Int) {
	result, err := validatorSetInstance.CurrentSpanNumber(nil)
	if err != nil {
		Logger.Error("Unable to get current span number", "Error", err)
		return nil
	}

	return result
}

// GetSpanDetails get span details
func (c *ContractCaller) GetSpanDetails(id *big.Int, validatorSetInstance *validatorset.Validatorset) (
	*big.Int,
	*big.Int,
	*big.Int,
	error,
) {
	d, err := validatorSetInstance.GetSpan(nil, id)
	return d.Number, d.StartBlock, d.EndBlock, err
}

// CurrentStateCounter get state counter
func (c *ContractCaller) CurrentStateCounter(stateSenderInstance *statesender.Statesender) (Number *big.Int) {
	result, err := stateSenderInstance.Counter(nil)
	if err != nil {
		Logger.Error("Unable to get current counter number", "Error", err)
		return nil
	}

	return result
}

// CheckIfBlocksExist - check if latest block number is greater than end block
func (c *ContractCaller) CheckIfBlocksExist(end uint64) bool {
	// Get Latest block number.
	var latestBlock *ethTypes.Header

	err := c.MaticChainRPC.Call(&latestBlock, "eth_getBlockByNumber", "latest", false)
	if err != nil {
		return false
	}

	if end > latestBlock.Number.Uint64() {
		return false
	}

	return true
}

//
// Receipt functions
//

// GetMainTxReceipt returns main tx receipt
func (c *ContractCaller) GetMainTxReceipt(txHash common.Hash, rootChain string) (*ethTypes.Receipt, error) {
	switch rootChain {
	case hmTypes.RootChainTypeEth:
		return c.getTxReceipt(c.MainChainClient, txHash)
	case hmTypes.RootChainTypeBsc:
		return c.getTxReceipt(c.BscChainClient, txHash)
	}
	return nil, errors.New("wrong chain type")
}

// GetMaticTxReceipt returns matic tx receipt
func (c *ContractCaller) GetMaticTxReceipt(txHash common.Hash) (*ethTypes.Receipt, error) {
	return c.getTxReceipt(c.MaticChainClient, txHash)
}

func (c *ContractCaller) getTxReceipt(client *ethclient.Client, txHash common.Hash) (*ethTypes.Receipt, error) {
	return client.TransactionReceipt(context.Background(), txHash)
}
func (c *ContractCaller) GetTronTransactionReceipt(txID string) (*ethTypes.Receipt, error) {
	// create filter
	var txIDs = []string{txID}
	queryFilter := tron.FilterOtherParams{
		BaseQueryParam: tron.GetDefaultBaseParm(),
		Method:         tron.GetTransactionByHash,
		Params:         txIDs,
	}
	queryByte, err := json.Marshal(queryFilter)
	req, err := http.NewRequest("POST", GetTronGridEndpoint("/jsonrpc"), bytes.NewBuffer(queryByte))
	if err != nil {
		return nil, err
	}
	result, err := MakeRequest(req)
	if err != nil {
		return nil, err
	}
	var transactionReceipt tron.FilterTxResponse
	if err := json.Unmarshal(result, &transactionReceipt); err != nil {
		return nil, err
	}
	return &transactionReceipt.Result, nil
}

//
// private abi methods
//

func getABI(data string) (abi.ABI, error) {
	return abi.JSON(strings.NewReader(data))
}

// GetCheckpointSign returns sigs input of committed checkpoint tranasction
func (c *ContractCaller) GetCheckpointSign(txHash common.Hash) ([]byte, []byte, []byte, error) {
	mainChainClient := GetMainClient()
	transaction, isPending, err := mainChainClient.TransactionByHash(context.Background(), txHash)
	if err != nil {
		Logger.Error("Error while Fetching Transaction By hash from MainChain", "error", err)
		return []byte{}, []byte{}, []byte{}, err
	} else if isPending {
		return []byte{}, []byte{}, []byte{}, errors.New("Transaction is still pending")
	}

	payload := transaction.Data()
	abi := c.RootChainABI
	return UnpackSigAndVotes(payload, abi)
}

//
// Staking sync methods
//

// GetMainStakingSyncNonce return validator nonce
func (c *ContractCaller) GetMainStakingSyncNonce(validatorID uint64, stakingManagerInstance *stakemanager.Stakemanager) (nonce uint64) {
	validatorNonce, err := stakingManagerInstance.ValidatorNonce(nil, big.NewInt(int64(validatorID)))
	if err != nil {
		Logger.Error("Error fetching validator nonce from stake manager",
			"error", err, "validatorId", validatorID)
		return 0
	}
	return validatorNonce.Uint64()
}

// GetMainStakingSyncNonce return validator nonce
func (c *ContractCaller) GetTronStakingSyncNonce(validatorID uint64, stakingManagerAddress string) (nonce uint64) {
	// Pack the input
	data, err := c.StakeManagerABI.Pack("validatorNonce", big.NewInt(0).SetUint64(validatorID))
	if err != nil {
		Logger.Error("Error pack validator nonce", "error", err, "validatorId", validatorID)

		return 0
	}
	result, err := c.TronChainRPC.TriggerConstantContract(stakingManagerAddress, data)
	if err != nil {
		Logger.Error("Error fetching validator nonce from stake manager",
			"error", err, "validatorId", validatorID)
		return 0
	}
	// Unpack the results
	var (
		ret0 = new(*big.Int)
	)
	if err = c.StakeManagerABI.Unpack(ret0, "validatorNonce", result); err != nil {
		Logger.Error("Error unpack validator nonce", "error", err, "validatorId", validatorID)
		return 0
	}
	return (*ret0).Uint64()
}

func (c *ContractCaller) GetTronEventsByContractAddress(address []string, from, to int64) ([]ethTypes.Log, error) {
	var decodedAddress []string
	for _, adr := range address {
		decodedAddress = append(decodedAddress, adr[2:])
	}
	//create filter
	filter := tron.NewFilter{
		Address:   decodedAddress,
		FromBlock: "0x" + strconv.FormatInt(from, 16),
		ToBlock:   "0x" + strconv.FormatInt(to, 16),
	}
	filtersArray := []tron.NewFilter{filter}
	queryFilter := tron.FilterEventParams{
		BaseQueryParam: tron.GetDefaultBaseParm(),
		Method:         tron.GetLogsMethod,
		Params:         filtersArray,
	}

	queryByte, err := json.Marshal(queryFilter)
	req, err := http.NewRequest("POST", GetTronGridEndpoint("/jsonrpc"), bytes.NewBuffer(queryByte))
	if err != nil {
		return nil, err
	}
	result, err := MakeRequest(req)
	if err != nil {
		return nil, err
	}
	var filterChangeResult tron.FilterEventResponse
	if err := json.Unmarshal(result, &filterChangeResult); err != nil {
		return nil, err
	}
	return filterChangeResult.Result, nil
}

func (c *ContractCaller) GetTronLatestBlockNumber() (int64, error) {
	var empty []string
	queryFilter := tron.FilterOtherParams{
		BaseQueryParam: tron.GetDefaultBaseParm(),
		Method:         tron.GetBlockByNumber,
		Params:         empty,
	}

	queryByte, err := json.Marshal(queryFilter)
	req, err := http.NewRequest("POST", GetTronGridEndpoint("/jsonrpc"), bytes.NewBuffer(queryByte))
	if err != nil {
		return 0, err
	}
	result, err := MakeRequest(req)
	if err != nil {
		return 0, err
	}
	var Block tron.FilterTxNumberResponse
	if err := json.Unmarshal(result, &Block); err != nil {
		return 0, err
	}
	if len(Block.Result) > 2 && Block.Result[:2] == "0x" {
		Block.Result = Block.Result[2:]
	}
	blockNumber, err := strconv.ParseUint(Block.Result, 16, 64)
	if err != nil {
		return 0, err
	}
	return int64(blockNumber), nil
}

func (c *ContractCaller) GetStartListenBlock(rootChainType string) uint64 {
	if rootChainType == hmTypes.RootChainTypeTron {
		return GetConfig().TronStartListenBlock
	} else if rootChainType == hmTypes.RootChainTypeEth {
		return GetConfig().EthStartListenBlock
	} else {
		return 0
	}
}

func (c *ContractCaller) GetTronHeaderInfo(headerID uint64, contractAddress string, childBlockInterval uint64) (
	root common.Hash, start, end, createdAt uint64, proposer types.HeimdallAddress, err error) {
	// Pack the input
	btsPack, err := c.RootChainABI.Pack("headerBlocks",
		big.NewInt(0).Mul(big.NewInt(0).SetUint64(headerID), big.NewInt(0).SetUint64(childBlockInterval)))
	if err != nil {
		return root, 0, 0, 0, types.HeimdallAddress{}, err
	}

	// Call
	data, err := c.TronChainRPC.TriggerConstantContract(contractAddress, btsPack)
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
	if err = c.RootChainABI.Unpack(ret, "headerBlocks", data); err != nil {
		return root, 0, 0, 0, types.HeimdallAddress{}, err
	}

	return ret.Root, ret.Start.Uint64(), ret.End.Uint64(),
		ret.CreatedAt.Uint64(), types.HeimdallAddress(ret.Proposer), nil
}

func (c *ContractCaller) GetSyncedCheckpointId(contractAddress string, rootChain string) (currentHeader uint64, err error) {
	// Pack the input
	chainID := types.GetRootChainID(rootChain)
	btsPack, err := c.StakeManagerABI.Pack("getCurrentSyncedCheckpoint", big.NewInt(int64(chainID)))
	if err != nil {
		return 0, err
	}

	// Call
	data, err := c.TronChainRPC.TriggerConstantContract(contractAddress, btsPack)
	if err != nil {
		return 0, err
	}

	// Unpack the results
	ret := new(*big.Int)

	if err = c.StakeManagerABI.Unpack(ret, "getCurrentSyncedCheckpoint", data); err != nil {
		return 0, err
	}

	return (*ret).Uint64(), nil
}

// DecodeNewChainEvent represents validator stake update event
func (c *ContractCaller) DecodeNewChainEvent(contractAddress common.Address, receipt *ethTypes.Receipt, logIndex uint64) (*rootchain.RootchainNewChain, error) {
	event := new(rootchain.RootchainNewChain)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex && bytes.Equal(vLog.Address.Bytes(), contractAddress.Bytes()) {
			found = true
			if err := UnpackLog(&c.RootChainABI, event, "NewChain", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}
