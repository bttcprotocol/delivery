package helper

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethCrypto "github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/maticnetwork/heimdall/contracts/erc20"
	"github.com/maticnetwork/heimdall/contracts/rootchain"
	"github.com/maticnetwork/heimdall/contracts/slashmanager"
	"github.com/maticnetwork/heimdall/contracts/stakemanager"
	hmtypes "github.com/maticnetwork/heimdall/types"
	"google.golang.org/protobuf/proto"
)

func GenerateAuthObj(client *ethclient.Client, address common.Address, data []byte) (auth *bind.TransactOpts, err error) {
	// generate call msg
	callMsg := ethereum.CallMsg{
		To:   &address,
		Data: data,
	}

	// get priv key
	pkObject := GetPrivKey()

	// create ecdsa private key
	ecdsaPrivateKey, err := crypto.ToECDSA(pkObject[:])
	if err != nil {
		return
	}

	// from address
	fromAddress := common.BytesToAddress(pkObject.PubKey().Address().Bytes())
	// fetch gas price
	gasprice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	mainChainMaxGasPrice := GetConfig().MainchainMaxGasPrice
	// Check if configured or not, Use default in case of invalid value
	if mainChainMaxGasPrice <= 0 {
		mainChainMaxGasPrice = DefaultMainchainMaxGasPrice
	}
	if gasprice.Cmp(big.NewInt(mainChainMaxGasPrice)) == 1 {
		Logger.Error("Gas price is more than max gas price", "gasprice", gasprice)
		err = fmt.Errorf("gas price is more than max_gas_price, gasprice = %v, maxGasPrice = %d", gasprice, mainChainMaxGasPrice)
		return
	}

	// fetch nonce
	nonce, err := client.NonceAt(context.Background(), fromAddress, nil)
	if err != nil {
		return
	}

	// fetch gas limit
	callMsg.From = fromAddress
	gasLimit, err := client.EstimateGas(context.Background(), callMsg)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		Logger.Error("Unable to fetch ChainID", "error", err)

		return
	}

	// create auth
	auth, err = bind.NewKeyedTransactorWithChainID(ecdsaPrivateKey, chainID)
	if err != nil {
		Logger.Error("Unable to create auth object", "error", err)

		return
	}

	auth.GasPrice = gasprice
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(gasLimit) // uint64(gasLimit)

	return
}

// SendCheckpoint sends checkpoint to rootchain contract
// todo return err
func (c *ContractCaller) SendCheckpoint(signedData []byte, sigs [][3]*big.Int,
	rootChainAddress common.Address, rootChainInstance *rootchain.Rootchain, rootChain string) (er error) {
	data, err := c.RootChainABI.Pack("submitCheckpoint", signedData, sigs)
	if err != nil {
		Logger.Error("Unable to pack tx for submitCheckpoint", "error", err)
		return err
	}

	var client *ethclient.Client
	switch rootChain {
	case hmtypes.RootChainTypeEth:
		client = GetMainClient()
	case hmtypes.RootChainTypeBsc:
		client = GetBscClient()
	}
	auth, err := GenerateAuthObj(client, rootChainAddress, data)
	if err != nil {
		Logger.Error("Unable to create auth object", "error", err)
		return err
	}

	s := make([]string, 0)
	for i := 0; i < len(sigs); i++ {
		s = append(s, fmt.Sprintf("[%s,%s,%s]", sigs[i][0].String(), sigs[i][1].String(), sigs[i][2].String()))
	}

	Logger.Debug("Sending new checkpoint",
		"sigs", strings.Join(s, ","),
		"data", hex.EncodeToString(signedData),
	)

	tx, err := rootChainInstance.SubmitCheckpoint(auth, signedData, sigs)
	if err != nil {
		Logger.Error("Error while submitting checkpoint", "error", err)
		return err
	}
	Logger.Info("Submitted new checkpoint to rootchain successfully", "txHash", tx.Hash().String())
	return
}

// SendTick sends slash tick to rootchain contract
func (c *ContractCaller) SendTick(signedData []byte, sigs []byte, slashManagerAddress common.Address, slashManagerInstance *slashmanager.Slashmanager) (er error) {
	data, err := c.SlashManagerABI.Pack("updateSlashedAmounts", signedData, sigs)
	if err != nil {
		Logger.Error("Unable to pack tx for updateSlashedAmounts", "error", err)
		return err
	}

	auth, err := GenerateAuthObj(GetMainClient(), slashManagerAddress, data)
	if err != nil {
		Logger.Error("Unable to create auth object", "error", err)
		return err
	}

	Logger.Info("Sending new tick",
		"sigs", hex.EncodeToString(sigs),
		"data", hex.EncodeToString(signedData),
	)

	tx, err := slashManagerInstance.UpdateSlashedAmounts(auth, signedData, sigs)
	if err != nil {
		Logger.Error("Error while submitting tick", "error", err)
		return err
	}
	Logger.Info("Submitted new tick to slashmanager successfully", "txHash", tx.Hash().String())
	return
}

// StakeFor stakes for a validator
func (c *ContractCaller) StakeFor(val common.Address, stakeAmount *big.Int, feeAmount *big.Int, acceptDelegation bool, stakeManagerAddress common.Address, stakeManagerInstance *stakemanager.Stakemanager) error {
	signerPubkey := GetPubKey()
	signerPubkeyBytes := signerPubkey[1:] // remove 04 prefix

	// pack data based on method definition
	data, err := c.StakeManagerABI.Pack("stakeFor", val, stakeAmount, feeAmount, acceptDelegation, signerPubkeyBytes)
	if err != nil {
		Logger.Error("Unable to pack tx for stakeFor", "error", err)
		return err
	}

	auth, err := GenerateAuthObj(GetMainClient(), stakeManagerAddress, data)
	if err != nil {
		Logger.Error("Unable to create auth object", "error", err)
		return err
	}

	// stake for stake manager
	tx, err := stakeManagerInstance.StakeFor(
		auth,
		val,
		stakeAmount,
		feeAmount,
		acceptDelegation,
		signerPubkeyBytes,
	)

	if err != nil {
		Logger.Error("Error while submitting stake", "error", err)
		return err
	}

	Logger.Info("Submitted stake sucessfully", "txHash", tx.Hash().String())
	return nil
}

// ApproveTokens approves matic token for stake
func (c *ContractCaller) ApproveTokens(amount *big.Int, stakeManager common.Address, tokenAddress common.Address, maticTokenInstance *erc20.Erc20) error {
	data, err := c.MaticTokenABI.Pack("approve", stakeManager, amount)
	if err != nil {
		Logger.Error("Unable to pack tx for approve", "error", err)
		return err
	}

	auth, err := GenerateAuthObj(GetMainClient(), tokenAddress, data)
	if err != nil {
		Logger.Error("Unable to create auth object", "error", err)
		return err
	}

	tx, err := maticTokenInstance.Approve(auth, stakeManager, amount)
	if err != nil {
		Logger.Error("Error while approving approve", "error", err)
		return err
	}

	Logger.Info("Sent approve tx sucessfully", "txHash", tx.Hash().String())
	return nil
}

// SendMainStakingSync sends staking sync to rootchain contract
func (c *ContractCaller) SendTronCheckpoint(signedData []byte, sigs [][3]*big.Int, rootChainAddress string) error {
	data, err := c.RootChainABI.Pack("submitCheckpoint", signedData, sigs)
	if err != nil {
		return err
	}
	privateKey := GetPrivKey()
	// trigger
	trx, err := c.TronChainRPC.TriggerContract(privateKey.PubKey().Address().String(), rootChainAddress, data)
	if err != nil {
		return err
	}
	trx.RawData.FeeLimit = int64(GetConfig().TronchainFeeLimit)
	rawData, _ := proto.Marshal(trx.GetRawData())
	hash, err := Hash(rawData)
	if err != nil {
		return err
	}

	signature, err := ethCrypto.Sign(hash, privateKey[:])
	if err != nil {
		return err
	}

	trx.Signature = append(trx.GetSignature(), signature)

	err = c.TronChainRPC.BroadcastTransaction(context.Background(), trx)
	if err != nil {
		return err
	}
	return nil
}

// SendCheckpointSyncToTron sends staking sync to tron stake manager contract
func (c *ContractCaller) SendCheckpointSyncToTron(signedData []byte, sigs [][3]*big.Int, stakeManagerAddress string) error {
	data, err := c.StakeManagerABI.Pack("submitCheckpointSync", signedData, sigs)
	if err != nil {
		return err
	}
	privateKey := GetPrivKey()
	// trigger
	trx, err := c.TronChainRPC.TriggerContract(privateKey.PubKey().Address().String(), stakeManagerAddress, data)
	if err != nil {
		return err
	}
	trx.RawData.FeeLimit = int64(GetConfig().TronchainFeeLimit)
	rawData, _ := proto.Marshal(trx.GetRawData())
	hash, err := Hash(rawData)
	if err != nil {
		return err
	}

	signature, err := ethCrypto.Sign(hash, privateKey[:])
	if err != nil {
		return err
	}

	trx.Signature = append(trx.GetSignature(), signature)

	err = c.TronChainRPC.BroadcastTransaction(context.Background(), trx)
	if err != nil {
		return err
	}
	return nil
}

// SendMainStakingSync sends staking sync to rootchain contract
func (c *ContractCaller) SendMainStakingSync(syncMethod string, signedData []byte, sigs [][3]*big.Int, stakingManager common.Address, stakingManagerInstance *stakemanager.Stakemanager, rootChain string) (er error) {
	data, err := c.StakeManagerABI.Pack(syncMethod, signedData, sigs)
	if err != nil {
		Logger.Error("Unable to pack tx for submitStakingSync", "error", err, "syncMethod", syncMethod)
		return err
	}
	var client *ethclient.Client
	switch rootChain {
	case hmtypes.RootChainTypeEth:
		client = GetMainClient()
	case hmtypes.RootChainTypeBsc:
		client = GetBscClient()
	}
	auth, err := GenerateAuthObj(client, stakingManager, data)
	if err != nil {
		Logger.Error("Unable to create auth object", "error", err)
		return err
	}

	s := make([]string, 0)
	for i := 0; i < len(sigs); i++ {
		s = append(s, fmt.Sprintf("[%s,%s,%s]", sigs[i][0].String(), sigs[i][1].String(), sigs[i][2].String()))
	}

	Logger.Debug("Sending new staking sync",
		"sigs", strings.Join(s, ","),
		"data", hex.EncodeToString(signedData),
		"address", stakingManager.Hex(),
	)

	var tx *types.Transaction
	switch syncMethod {
	case "validatorJoin":
		tx, err = stakingManagerInstance.ValidatorJoin(auth, signedData, sigs)
	case "validatorExit":
		tx, err = stakingManagerInstance.ValidatorExit(auth, signedData, sigs)
	case "signerUpdate":
		tx, err = stakingManagerInstance.SignerUpdate(auth, signedData, sigs)
	default:
		Logger.Error("Submitted new staking sync to stake chain failed", "method", syncMethod)
		return
	}

	if err != nil {
		Logger.Error("Error while submitting staking sync", "error", err)
		return err
	}
	Logger.Info("Submitted new staking sync to stake chain successfully", "txHash", tx.Hash().String())
	return
}

// SendTronStakingSync sends staking sync to tron contract
func (c *ContractCaller) SendTronStakingSync(syncMethod string, signedData []byte, sigs [][3]*big.Int, stakingManagerAddress string) (er error) {
	data, err := c.StakeManagerABI.Pack(syncMethod, signedData, sigs)
	if err != nil {
		return err
	}
	privateKey := GetPrivKey()

	// trigger
	trx, err := c.TronChainRPC.TriggerContract(privateKey.PubKey().Address().String(), stakingManagerAddress, data)
	if err != nil {
		return err
	}
	trx.RawData.FeeLimit = int64(GetConfig().TronchainFeeLimit)
	rawData, _ := proto.Marshal(trx.GetRawData())
	hash, err := Hash(rawData)
	if err != nil {
		return err
	}

	signature, err := ethCrypto.Sign(hash, privateKey[:])
	if err != nil {
		return err
	}

	trx.Signature = append(trx.GetSignature(), signature)

	s := make([]string, 0)
	for i := 0; i < len(sigs); i++ {
		s = append(s, fmt.Sprintf("[%s,%s,%s]", sigs[i][0].String(), sigs[i][1].String(), sigs[i][2].String()))
	}
	Logger.Debug("Sending new staking sync",
		"sigs", strings.Join(s, ","),
		"data", hex.EncodeToString(signedData),
	)

	err = c.TronChainRPC.BroadcastTransaction(context.Background(), trx)
	if err != nil {
		return err
	}
	Logger.Info("Submitted new staking to tron successfully")
	return nil
}
