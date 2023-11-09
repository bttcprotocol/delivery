package util

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	stakingTypes "github.com/maticnetwork/heimdall/staking/types"

	mLog "github.com/RichardKnop/machinery/v1/log"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	httpClient "github.com/tendermint/tendermint/rpc/client"
	tmTypes "github.com/tendermint/tendermint/types"

	authTypes "github.com/maticnetwork/heimdall/auth/types"
	chainManagerTypes "github.com/maticnetwork/heimdall/chainmanager/types"
	checkpointTypes "github.com/maticnetwork/heimdall/checkpoint/types"
	featureManagerTypes "github.com/maticnetwork/heimdall/featuremanager/types"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/maticnetwork/heimdall/types"
	hmtypes "github.com/maticnetwork/heimdall/types"
)

const (
	AccountDetailsURL       = "/auth/accounts/%v"
	LastNoAckURL            = "/checkpoints/last-no-ack"
	CurrentEpochURL         = "/checkpoints/epoch"
	CheckpointParamsURL     = "/checkpoints/params"
	CheckpointActivationURL = "/checkpoints/activation-height/%v"
	ChainManagerParamsURL   = "/chainmanager/params"
	// for new eth forkChain such as bsc, replace address of params.
	ChainNewParamsURL         = "/chainmanager/newparams/%v"
	TaragetFeatureConfigURL   = "/featuremanager/target-feature/%v"
	AllFeatureConfigURL       = "/featuremanager/feature-map"
	ProposersURL              = "/staking/proposer/%v"
	BufferedCheckpointURL     = "/checkpoints/buffer/%v"
	BufferedCheckpointSyncURL = "/checkpoints/sync/%v"
	LatestCheckpointURL       = "/checkpoints/latest/%v"
	CurrentProposerURL        = "/staking/current-proposer"
	LatestSpanURL             = "/bor/latest-span"
	NextSpanInfoURL           = "/bor/prepare-next-span"
	NextSpanSeedURL           = "/bor/next-span-seed"
	DividendAccountRootURL    = "/topup/dividend-account-root"
	ValidatorURL              = "/staking/validator/%v"
	CurrentValidatorSetURL    = "staking/validator-set"
	StakingTxStatusURL        = "/staking/isoldtx"
	NextStakingRecordURL      = "/staking/next/%v"
	TopupTxStatusURL          = "/topup/isoldtx"
	ClerkTxStatusURL          = "/clerk/isoldtx"
	LatestSlashInfoBytesURL   = "/slashing/latest_slash_info_bytes"
	TickSlashInfoListURL      = "/slashing/tick_slash_infos"
	SlashingTxStatusURL       = "/slashing/isoldtx"
	SlashingTickCountURL      = "/slashing/tick-count"

	TransactionTimeout      = 1 * time.Minute
	CommitTimeout           = 2 * time.Minute
	TaskDelayBetweenEachVal = 24 * time.Second
	RetryTaskDelay          = 12 * time.Second

	BridgeDBFlag          = "bridge-db"
	ProposersURLSizeLimit = 100

	WithdrawToEventSig = "0x67b714876402c93362735688659e2283b4a37fb21bab24bc759ca759ae851fd8"
)

var withDrawToTopics = [][]string{{WithdrawToEventSig}}

var (
	logger     log.Logger
	loggerOnce sync.Once
)

// Logger returns logger singleton instance
func Logger() log.Logger {
	loggerOnce.Do(func() {
		defaultLevel := "info"
		logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
		option, err := log.AllowLevel(viper.GetString("log_level"))
		if err != nil {
			// cosmos sdk is using different style of log format
			// and levels don't map well, config.toml
			// see: https://github.com/cosmos/cosmos-sdk/pull/8072
			logger.Error("Unable to parse logging level", "Error", err)
			logger.Info("Using default log level")

			option, err = log.AllowLevel(defaultLevel)
			if err != nil {
				logger.Error("failed to allow default log level", "Level", defaultLevel, "Error", err)
			}
		}
		logger = log.NewFilter(logger, option)

		// set no-op logger if log level is not debug for machinery
		if viper.GetString("log_level") != "debug" {
			mLog.SetDebug(NoopLogger{})
		}
	})

	return logger
}

// IsProposer  checks if we are proposer.
// index starts from 0, and it will return whether it is the current proposer
// if index = 0.
func IsProposerByIndex(cliCtx cliContext.CLIContext, index uint64) (bool, error) {
	var proposers []hmtypes.Validator
	count := index + 1
	result, err := helper.FetchFromAPI(cliCtx,
		helper.GetHeimdallServerEndpoint(fmt.Sprintf(ProposersURL, strconv.FormatUint(count, 10))),
	)
	if err != nil {
		logger.Error("Error fetching proposers", "url", ProposersURL, "error", err)
		return false, err
	}

	err = json.Unmarshal(result.Result, &proposers)
	if err != nil {
		logger.Error("error unmarshalling proposer slice", "error", err)
		return false, err
	}

	proposerSize := len(proposers)
	if proposerSize == 0 {
		return false, nil
	}
	proposerIndex := index % uint64(proposerSize)

	if bytes.Equal(proposers[proposerIndex].Signer.Bytes(), helper.GetAddress()) {
		return true, nil
	}

	return false, nil
}

func IsValidator(cliCtx cliContext.CLIContext) (bool, error) {
	var validatorSet hmtypes.ValidatorSet

	result, err := helper.FetchFromAPI(cliCtx,
		helper.GetHeimdallServerEndpoint(CurrentValidatorSetURL),
	)
	if err != nil {
		logger.Error("Error fetching proposers", "url", CurrentValidatorSetURL, "error", err)

		return false, fmt.Errorf("failed to query heimdall server: %w", err)
	}

	err = json.Unmarshal(result.Result, &validatorSet)
	if err != nil {
		logger.Error("error unmarshalling proposer slice", "error", err)

		return false, fmt.Errorf("failed to unmarshal validatorSet: %w", err)
	}

	for _, validator := range validatorSet.Validators {
		if bytes.Equal(validator.Signer.Bytes(), helper.GetAddress()) {
			return true, nil
		}
	}

	return false, nil
}

// IsInProposerList checks if we are in current proposer
func IsInProposerList(cliCtx cliContext.CLIContext, count uint64) (bool, error) {
	logger.Debug("Skipping proposers", "count", strconv.FormatUint(count, 10))
	response, err := helper.FetchFromAPI(
		cliCtx,
		helper.GetHeimdallServerEndpoint(fmt.Sprintf(ProposersURL, strconv.FormatUint(count, 10))),
	)
	if err != nil {
		logger.Error("Unable to send request for next proposers", "url", ProposersURL, "error", err)
		return false, err
	}

	// unmarshall data from buffer
	var proposers []hmtypes.Validator

	if err := json.Unmarshal(response.Result, &proposers); err != nil {
		logger.Error("Error unmarshalling validator data ", "error", err)
		return false, err
	}

	logger.Debug("Fetched proposers list", "numberOfProposers", count)
	for _, proposer := range proposers {
		if bytes.Equal(proposer.Signer.Bytes(), helper.GetAddress()) {
			return true, nil
		}
	}
	return false, nil
}

// default offset 0.
func CalculateTaskDelay(cliCtx cliContext.CLIContext) (bool, time.Duration) {
	return CalculateTaskDelayWithOffset(cliCtx, 0)
}

// CalculateTaskDelay calculates delay required for current validator to propose the tx
// It solves for multiple validators sending same transaction.
// with offset
func CalculateTaskDelayWithOffset(cliCtx cliContext.CLIContext, offset int) (bool, time.Duration) {
	// calculate validator position
	valPosition := 0
	isCurrentValidator := false

	proposersURL := fmt.Sprintf(ProposersURL, ProposersURLSizeLimit)
	proposersResponse, err := helper.FetchFromAPI(cliCtx, helper.GetHeimdallServerEndpoint(proposersURL))
	if err != nil {
		logger.Error("Unable to send request for proposers ", "url", proposersURL, "error", err)
		return isCurrentValidator, 0
	}

	var proposers []hmtypes.Validator
	err = json.Unmarshal(proposersResponse.Result, &proposers)
	if err != nil {
		logger.Error("Error unmarshalling proposers data ", "error", err)
		return isCurrentValidator, 0
	}

	logger.Info("Fetched proposers ", "currentValidatorsCount", len(proposers))
	for i, validator := range proposers {
		if bytes.Equal(validator.Signer.Bytes(), helper.GetAddress()) {
			valPosition = i + offset
			isCurrentValidator = true
			break
		}
	}

	// calculate delay
	taskDelay := time.Duration(valPosition) * TaskDelayBetweenEachVal
	return isCurrentValidator, taskDelay
}

// IsCurrentProposer checks if we are current proposer
func IsCurrentProposer(cliCtx cliContext.CLIContext) (bool, error) {
	var proposer hmtypes.Validator
	result, err := helper.FetchFromAPI(cliCtx, helper.GetHeimdallServerEndpoint(CurrentProposerURL))
	if err != nil {
		logger.Error("Error fetching proposers", "error", err)
		return false, err
	}

	err = json.Unmarshal(result.Result, &proposer)
	if err != nil {
		logger.Error("error unmarshalling validator", "error", err)
		return false, err
	}
	logger.Debug("Current proposer fetched", "validator", proposer.String())

	if bytes.Equal(proposer.Signer.Bytes(), helper.GetAddress()) {
		return true, nil
	}

	return false, nil
}

// IsEventSender check if we are the EventSender
func IsEventSender(cliCtx cliContext.CLIContext, validatorID uint64) bool {
	var validator hmtypes.Validator

	result, err := helper.FetchFromAPI(cliCtx,
		helper.GetHeimdallServerEndpoint(fmt.Sprintf(ValidatorURL, strconv.FormatUint(validatorID, 10))),
	)
	if err != nil {
		logger.Error("Error fetching proposers", "error", err)
		return false
	}

	err = json.Unmarshal(result.Result, &validator)
	if err != nil {
		logger.Error("error unmarshalling proposer slice", "error", err)
		return false
	}
	logger.Debug("Current event sender received", "validator", validator.String())

	return bytes.Equal(validator.Signer.Bytes(), helper.GetAddress())
}

// CreateURLWithQuery receives the uri and parameters in key value form
// it will return the new url with the given query from the parameter.
func CreateURLWithQuery(uri string, param map[string]interface{}) (string, error) {
	urlObj, err := url.Parse(uri)
	if err != nil {
		return uri, err
	}

	query := urlObj.Query()
	for k, v := range param {
		query.Set(k, fmt.Sprintf("%v", v))
	}

	urlObj.RawQuery = query.Encode()
	return urlObj.String(), nil
}

// WaitForOneEvent subscribes to a websocket event for the given
// event time and returns upon receiving it one time, or
// when the timeout duration has expired.
//
// This handles subscribing and unsubscribing under the hood
func WaitForOneEvent(tx tmTypes.Tx, client *httpClient.HTTP) (tmTypes.TMEventData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), CommitTimeout)
	defer cancel()

	// subscriber
	subscriber := hex.EncodeToString(tx.Hash())

	// query
	query := tmTypes.EventQueryTxFor(tx).String()

	// register for the next event of this type
	eventCh, err := client.Subscribe(ctx, subscriber, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to subscribe")
	}

	// make sure to unregister after the test is over
	defer func() {
		if err := client.UnsubscribeAll(ctx, subscriber); err != nil {
			logger.Error("WaitForOneEvent | UnsubscribeAll", "Error", err)
		}
	}()

	select {
	case event := <-eventCh:
		return event.Data.(tmTypes.TMEventData), nil
	case <-ctx.Done():
		return nil, errors.New("timed out waiting for event")
	}
}

// IsCatchingUp checks if the heimdall node you are connected to is fully synced or not
// returns true when synced
func IsCatchingUp(cliCtx cliContext.CLIContext) bool {
	resp, err := helper.GetNodeStatus(cliCtx)
	if err != nil {
		return true
	}
	return resp.SyncInfo.CatchingUp
}

// GetAccount returns heimdall auth account
func GetAccount(cliCtx cliContext.CLIContext, address types.HeimdallAddress) (account authTypes.Account, err error) {
	url := helper.GetHeimdallServerEndpoint(fmt.Sprintf(AccountDetailsURL, address))
	// call account rest api
	response, err := helper.FetchFromAPI(cliCtx, url)
	if err != nil {
		return
	}
	if err = cliCtx.Codec.UnmarshalJSON(response.Result, &account); err != nil {
		logger.Error("Error unmarshalling account details", "url", url)
		return
	}
	return
}

// GetChainmanagerParams return chain manager params
func GetChainmanagerParams(cliCtx cliContext.CLIContext) (*chainManagerTypes.Params, error) {
	response, err := helper.FetchFromAPI(
		cliCtx,
		helper.GetHeimdallServerEndpoint(ChainManagerParamsURL),
	)
	if err != nil {
		logger.Error("Error fetching chainmanager params", "err", err)
		return nil, err
	}

	var params chainManagerTypes.Params
	if err := json.Unmarshal(response.Result, &params); err != nil {
		logger.Error("Error unmarshalling chainmanager params", "url", ChainManagerParamsURL, "err", err)
		return nil, err
	}

	return &params, nil
}

// GetNewChainParams return new chain params
func GetNewChainParams(cliCtx cliContext.CLIContext, rootChain string) (*chainManagerTypes.Params, error) {
	response, err := helper.FetchFromAPI(
		cliCtx,
		helper.GetHeimdallServerEndpoint(fmt.Sprintf(ChainNewParamsURL, rootChain)),
	)
	if err != nil {
		logger.Error("Error fetching chainmanager params", "root", rootChain, "err", err)
		return nil, err
	}

	var params chainManagerTypes.Params
	if err := json.Unmarshal(response.Result, &params); err != nil {
		logger.Error("Error unmarshalling chainmanager params", "root", rootChain, "url", ChainNewParamsURL, "err", err)
		return nil, err
	}

	return &params, nil
}

// GetCheckpointParams return params
func GetCheckpointParams(cliCtx cliContext.CLIContext) (*checkpointTypes.Params, error) {
	response, err := helper.FetchFromAPI(
		cliCtx,
		helper.GetHeimdallServerEndpoint(CheckpointParamsURL),
	)
	if err != nil {
		logger.Error("Error fetching Checkpoint params", "err", err)
		return nil, err
	}

	var params checkpointTypes.Params
	if err := json.Unmarshal(response.Result, &params); err != nil {
		logger.Error("Error unmarshalling Checkpoint params", "url", CheckpointParamsURL)
		return nil, err
	}

	return &params, nil
}

// GetBufferedCheckpoint return checkpoint from bueffer
func GetBufferedCheckpoint(cliCtx cliContext.CLIContext, rootChain string) (*hmtypes.Checkpoint, error) {
	response, err := helper.FetchFromAPI(
		cliCtx,
		helper.GetHeimdallServerEndpoint(fmt.Sprintf(BufferedCheckpointURL, rootChain)),
	)
	if err != nil {
		logger.Debug("Error fetching buffered checkpoint", "err", err)
		return nil, err
	}

	var checkpoint hmtypes.Checkpoint
	if err := json.Unmarshal(response.Result, &checkpoint); err != nil {
		logger.Error("Error unmarshalling buffered checkpoint", "url", BufferedCheckpointURL, "err", err)
		return nil, err
	}

	return &checkpoint, nil
}

// GetBufferedCheckpointSync return checkpoint sync from buffer
func GetBufferedCheckpointSync(cliCtx cliContext.CLIContext, rootChain string) (*hmtypes.Checkpoint, error) {
	response, err := helper.FetchFromAPI(
		cliCtx,
		helper.GetHeimdallServerEndpoint(fmt.Sprintf(BufferedCheckpointSyncURL, rootChain)),
	)
	if err != nil {
		logger.Debug("Error fetching buffered checkpoint sync", "root", rootChain, "err", err)
		return nil, err
	}

	var checkpoint hmtypes.Checkpoint
	if err := json.Unmarshal(response.Result, &checkpoint); err != nil {
		logger.Error("Error unmarshalling buffered checkpoint sync", "root", rootChain, "err", err)
		return nil, err
	}

	return &checkpoint, nil
}

// GetlastestCheckpoint return last successful checkpoint
func GetlastestCheckpoint(cliCtx cliContext.CLIContext, rootChain string) (*hmtypes.Checkpoint, error) {
	response, err := helper.FetchFromAPI(
		cliCtx,
		helper.GetHeimdallServerEndpoint(fmt.Sprintf(LatestCheckpointURL, rootChain)),
	)
	if err != nil {
		logger.Debug("Error fetching latest checkpoint", "err", err)
		return nil, err
	}

	var checkpoint hmtypes.Checkpoint
	if err := json.Unmarshal(response.Result, &checkpoint); err != nil {
		logger.Error("Error unmarshalling latest checkpoint", "url", LatestCheckpointURL, "err", err)
		return nil, err
	}

	return &checkpoint, nil
}

// AppendPrefix returns publickey in uncompressed format
func AppendPrefix(signerPubKey []byte) []byte {
	// append prefix - "0x04" as heimdall uses publickey in uncompressed format. Refer below link
	// https://superuser.com/questions/1465455/what-is-the-size-of-public-key-for-ecdsa-spec256r1
	prefix := make([]byte, 1)
	prefix[0] = byte(0x04)
	signerPubKey = append(prefix[:], signerPubKey[:]...)
	return signerPubKey
}

// GetNextStakingRecord return staking info from buffer
func GetNextStakingRecord(cliCtx cliContext.CLIContext, rootChain string) (*stakingTypes.StakingRecord, error) {
	response, err := helper.FetchFromAPI(
		cliCtx,
		helper.GetHeimdallServerEndpoint(fmt.Sprintf(NextStakingRecordURL, rootChain)),
	)
	if err != nil {
		logger.Debug("Error fetching next stake record", "err", err)
		return nil, err
	}

	var stakingRecord stakingTypes.StakingRecord
	if err := json.Unmarshal(response.Result, &stakingRecord); err != nil {
		logger.Error("Error unmarshalling next staking info", "url", NextStakingRecordURL,
			"root", rootChain, "err", err)
		return nil, err
	}

	return &stakingRecord, nil
}

// GetValidatorNonce fethes validator nonce and height
func GetValidatorNonce(cliCtx cliContext.CLIContext, validatorID uint64) (uint64, int64, error) {
	var validator hmtypes.Validator

	result, err := helper.FetchFromAPI(cliCtx,
		helper.GetHeimdallServerEndpoint(fmt.Sprintf(ValidatorURL, strconv.FormatUint(validatorID, 10))),
	)
	if err != nil {
		logger.Error("Error fetching validator data", "error", err)
		return 0, 0, err
	}

	err = json.Unmarshal(result.Result, &validator)
	if err != nil {
		logger.Error("error unmarshalling validator data", "error", err)
		return 0, 0, err
	}

	logger.Debug("Validator data recieved ", "validator", validator.String())

	return validator.Nonce, result.Height, nil
}

func GetWithDrawToTopics() [][]common.Hash {
	return getTopics(withDrawToTopics)
}

func getTopics(topicStr [][]string) [][]common.Hash {
	ret := make([][]common.Hash, len(topicStr))

	for i, topics := range topicStr {
		for _, topic := range topics {
			topicByte, err := hexutil.Decode(topic)
			if err != nil {
				logger.Error("Error while decoding topic", "error", err)

				return nil
			}

			var h common.Hash

			copy(h[:], topicByte)
			ret[i] = append(ret[i], h)
		}
	}

	return ret
}

func GetDynamicCheckpointFeature(cliCtx cliContext.CLIContext) (*featureManagerTypes.PlainFeatureData, error) {
	return GetTargetFeatureConfig(cliCtx, featureManagerTypes.DynamicCheckpoint)
}

// GetTargetFeatureConfig return target feature config.
func GetTargetFeatureConfig(
	cliCtx cliContext.CLIContext, feature string,
) (*featureManagerTypes.PlainFeatureData, error) {
	response, err := helper.FetchFromAPI(
		cliCtx,
		helper.GetHeimdallServerEndpoint(fmt.Sprintf(TaragetFeatureConfigURL, feature)),
	)
	if err != nil {
		logger.Error("Error fetching target feature", "feature", feature, "err", err)

		return nil, err
	}

	var params featureManagerTypes.PlainFeatureData
	if err := json.Unmarshal(response.Result, &params); err != nil {
		logger.Error("Error unmarshalling feature data", "feature", feature, "url", TaragetFeatureConfigURL, "err", err)

		return nil, err
	}

	return &params, nil
}

// GetFeatureMap return all features in a map.
func GetFeatureMap(cliCtx cliContext.CLIContext) (*featureManagerTypes.FeatureParams, error) {
	response, err := helper.FetchFromAPI(
		cliCtx,
		helper.GetHeimdallServerEndpoint(AllFeatureConfigURL),
	)
	if err != nil {
		logger.Error("Error fetching feature params", "err", err)

		return nil, err
	}

	var params featureManagerTypes.FeatureParams
	if err := json.Unmarshal(response.Result, &params); err != nil {
		logger.Error("Error unmarshalling feature params", "url", TaragetFeatureConfigURL, "err", err)

		return nil, err
	}

	return &params, nil
}
