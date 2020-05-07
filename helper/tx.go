package helper

import (
	"context"
	"encoding/hex"
	"math/big"

	ethereum "github.com/maticnetwork/bor"
	"github.com/maticnetwork/bor/accounts/abi/bind"
	"github.com/maticnetwork/bor/common"
	"github.com/maticnetwork/bor/crypto"
	"github.com/maticnetwork/bor/ethclient"
	"github.com/maticnetwork/heimdall/contracts/erc20"
	"github.com/maticnetwork/heimdall/contracts/rootchain"
	"github.com/maticnetwork/heimdall/contracts/stakemanager"
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
	// fetch nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return
	}

	// fetch gas limit
	callMsg.From = fromAddress
	gasLimit, err := client.EstimateGas(context.Background(), callMsg)

	// create auth
	auth = bind.NewKeyedTransactor(ecdsaPrivateKey)
	auth.GasPrice = gasprice
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(gasLimit) // uint64(gasLimit)

	return
}

// SendCheckpoint sends checkpoint to rootchain contract
// todo return err
func (c *ContractCaller) SendCheckpoint(signedData []byte, sigs []byte, rootChainAddress common.Address, rootChainInstance *rootchain.Rootchain) (er error) {
	data, err := c.RootChainABI.Pack("submitHeaderBlock", signedData, sigs)
	if err != nil {
		Logger.Error("Unable to pack tx for submitHeaderBlock", "error", err)
		return err
	}

	auth, err := GenerateAuthObj(GetMainClient(), rootChainAddress, data)
	if err != nil {
		Logger.Error("Unable to create auth object", "error", err)
		Logger.Info("Setting custom gaslimit", "gaslimit", GetConfig().MainchainGasLimit)
		auth.GasLimit = GetConfig().MainchainGasLimit
	}

	Logger.Debug("Sending new checkpoint",
		"sigs", hex.EncodeToString(sigs),
		"data", hex.EncodeToString(signedData),
	)

	tx, err := rootChainInstance.SubmitHeaderBlock(auth, signedData, sigs)
	if err != nil {
		Logger.Error("Error while submitting checkpoint", "error", err)
		return err
	}
	Logger.Info("Submitted new checkpoint to rootchain successfully", "txHash", tx.Hash().String())
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
