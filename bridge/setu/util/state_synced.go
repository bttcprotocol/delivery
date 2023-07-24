package util

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	clerkTypes "github.com/maticnetwork/heimdall/clerk/types"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Storage key for token map.
const (
	TokenMapPrefixKey      = "tmp-"                 // map items' key prefix
	TokenMapHashKey        = "tm-hash"              // hash for the whole token map items
	TMapLastEventIDKey     = "tm-last-event-id"     // last event id processed for token map
	TMapCheckedEndBlockKey = "tm-checked-end-block" // end block processed for token map
)

const (
	// EventRecordURLPath - url path of heimdall for event records.
	EventRecordURLPath = "clerk/event-record/list"

	BlockDiff    = int64(25)
	DigitBase    = 10
	DigitBitSize = 64
)

// TokenMapItem - item struct.
type TokenMapItem struct {
	RootChainType string `json:"roorChainType"`
	RootToken     string `json:"rootToken"`
	ChildToken    string `json:"childToken"`
	EventID       uint64 `json:"eventId"`
}

func (tmi *TokenMapItem) validate() bool {
	if tmi.RootChainType != "" &&
		tmi.RootToken != "" &&
		tmi.ChildToken != "" &&
		tmi.EventID > 0 {
		return true
	}

	return false
}

// TokenMapProcessor - token map data holder.
type TokenMapProcessor struct {
	TokenMapLastEventID     uint64
	TokenMapCheckedEndBlock int64
	TokenMapSignagure       []byte

	cliCtx        context.CLIContext
	storageClient *leveldb.DB
	eventCache    []*TokenMapItem
	mlock         sync.Mutex
}

var (
	tokenMapProcessorOnce sync.Once
	tokenMapProcessor     *TokenMapProcessor
)

// NewTokenMapProcessor - new one.
func NewTokenMapProcessor(cliCtx context.CLIContext, storageClient *leveldb.DB) *TokenMapProcessor {
	tokenMapProcessorOnce.Do(func() {
		tokenMapProcessor = &TokenMapProcessor{
			TokenMapLastEventID:     uint64(0),
			TokenMapCheckedEndBlock: int64(0),

			cliCtx:        cliCtx,
			storageClient: storageClient,
			eventCache:    make([]*TokenMapItem, 0),
		}
	})

	if tokenMapProcessor.TokenMapLastEventID == 0 {
		if err := tokenMapProcessor.LoadLastEventIDFromDB(); err != nil {
			return nil
		}
	}

	if err := tokenMapProcessor.LoadCheckedEndBlockFromDB(); err != nil {
		return nil
	}

	return tokenMapProcessor
}

func (tmp *TokenMapProcessor) ResetParameters() error {
	tmp.TokenMapLastEventID = uint64(0)
	tmp.TokenMapCheckedEndBlock = int64(0)

	if err := tmp.UpdateTokenMapLastEventID(tmp.TokenMapLastEventID); err != nil {
		return err
	}

	if err := tmp.UpdateTokenMapCheckedEndBlock(tmp.TokenMapCheckedEndBlock); err != nil {
		return err
	}

	return tmp.deleteHash()
}

// IsInitializationDone - check if all the historical records event have been processed.
func (tmp *TokenMapProcessor) IsInitializationDone() (bool, error) {
	tmp.mlock.Lock()
	defer tmp.mlock.Unlock()

	hashBytes, err := tmp.GetHash()
	if err != nil {
		return false, err
	}

	if hashBytes != nil {
		valid, err := tmp.validateHash(hashBytes)
		if err != nil {
			return false, err
		}

		if !valid {
			if err := tmp.ResetParameters(); err != nil {
				return false, err
			}
		}

		return valid, nil
	}

	return false, nil
}

// LockForUpdate...
func (tmp *TokenMapProcessor) LockForUpdate() {
	tmp.mlock.Lock()
}

// UnLockForUpdate...
func (tmp *TokenMapProcessor) UnLockForUpdate() {
	tmp.mlock.Unlock()
}

// GetHash - query token map hash from db.
func (tmp *TokenMapProcessor) GetHash() ([]byte, error) {
	hasSig, _ := tmp.storageClient.Has([]byte(TokenMapHashKey), nil)
	if hasSig {
		hashBytes, err := tmp.storageClient.Get([]byte(TokenMapHashKey), nil)
		if err != nil {
			return nil, err
		}

		return hashBytes, nil
	}

	return nil, nil
}

// IsInitializationDoneWithBlock - check if all the historical records event have been processed.
func (tmp *TokenMapProcessor) IsInitializationDoneWithBlock(toBlock int64) (bool, error) {
	if done, err := tmp.IsInitializationDone(); err == nil {
		if done {
			if toBlock-tmp.TokenMapCheckedEndBlock > BlockDiff {
				return false, nil
			}

			return true, nil
		}
	} else {
		return false, err
	}

	return false, nil
}

func (tmp *TokenMapProcessor) validateHash(hashByte []byte) (bool, error) {
	items, err := tmp.loadTokenMap()
	if err != nil {
		return false, err
	}

	hashByteNow, err := tmp.calculateTokenMapHash(items)
	if err != nil {
		return false, err
	}

	if bytes.Equal(hashByte, hashByteNow) {
		return true, nil
	}

	return false, nil
}

func (tmp *TokenMapProcessor) deleteHash() error {
	return tmp.storageClient.Delete([]byte(TokenMapHashKey), nil)
}

// UpdateHash - recalculate hash and store it.
func (tmp *TokenMapProcessor) UpdateHash() error {
	items, err := tmp.loadTokenMap()
	if err != nil {
		return err
	}

	hashByteNow, err := tmp.calculateTokenMapHash(items)
	if err != nil {
		return err
	}

	return tmp.storageClient.Put([]byte(TokenMapHashKey), hashByteNow, nil)
}

// GetTokenMap - get token map from db.
func (tmp *TokenMapProcessor) GetTokenMap() (map[string][]*TokenMapItem, error) {
	result := make(map[string][]*TokenMapItem)

	if items, err := tmp.loadTokenMap(); err == nil {
		for _, v := range items {
			if _, ok := result[v.RootChainType]; !ok {
				result[v.RootChainType] = make([]*TokenMapItem, 0, 1)
			}

			result[v.RootChainType] = append(result[v.RootChainType], v)
		}
	} else {
		return nil, err
	}

	return result, nil
}

// GetTokenMapByRootType - get token map from db for assigned root chain.
func (tmp *TokenMapProcessor) GetTokenMapByRootType(rootChainType string) ([]*TokenMapItem, error) {
	resultMap, err := tmp.GetTokenMap()
	if err != nil {
		return nil, err
	}

	if item, ok := resultMap[rootChainType]; ok {
		return item, nil
	}

	return nil, nil
}

func (tmp *TokenMapProcessor) loadTokenMap() ([]*TokenMapItem, error) {
	errorItems := make([][]byte, 0)
	result := make([]*TokenMapItem, 0)

	iter := tmp.storageClient.NewIterator(util.BytesPrefix([]byte(TokenMapPrefixKey)), nil)
	defer iter.Release()

	for iter.Next() {
		if item, err := tmp.newTokenMapItem(iter.Key(), iter.Value()); err == nil {
			result = append(result, item)
		} else {
			errorItems = append(errorItems, iter.Key())
		}
	}

	if err := tmp.deleteKeys(errorItems); err != nil {
		return nil, err
	}

	return result, nil
}

func (tmp *TokenMapProcessor) deleteKeys(keys [][]byte) error {
	for _, v := range keys {
		if err := tmp.storageClient.Delete(v, nil); err != nil {
			return err
		}
	}

	return nil
}

func (tmp *TokenMapProcessor) newTokenMapItem(key, value []byte) (*TokenMapItem, error) {
	if strings.Index(string(key), TokenMapPrefixKey) != 0 {
		return nil, errors.New("key format is wrong")
	}

	item := &TokenMapItem{}
	if err := json.Unmarshal(value, item); err != nil {
		return nil, err
	}

	if !item.validate() {
		return nil, errors.New("key format is wrong")
	}

	return item, nil
}

// CacheTokenMapItem - cache one item.
func (tmp *TokenMapProcessor) CacheTokenMapItem(item *TokenMapItem) error {
	if !item.validate() {
		return errors.New("key format is wrong")
	}

	tmp.eventCache = append(tmp.eventCache, item)

	return nil
}

// PutTokenMapItem - cache one item.
func (tmp *TokenMapProcessor) FlushCacheTokenMapItem() error {
	for _, item := range tmp.eventCache {
		if err := tmp.PutTokenMapItem(item); err != nil {
			return err
		}
	}

	tmp.eventCache = make([]*TokenMapItem, 0)

	return nil
}

// PutTokenMapItem - store one item to db.
func (tmp *TokenMapProcessor) PutTokenMapItem(item *TokenMapItem) error {
	if !item.validate() {
		return errors.New("key format is wrong")
	}

	value, err := json.Marshal(item)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s%s%s", TokenMapPrefixKey, item.RootChainType, item.RootToken)

	return tmp.storageClient.Put([]byte(key), value, nil)
}

func (tmp *TokenMapProcessor) calculateTokenMapHash(items []*TokenMapItem) ([]byte, error) {
	if len(items) == 0 {
		return nil, nil
	}

	sort.SliceStable(items, func(i, j int) bool { return items[i].RootToken < items[j].RootToken })

	content, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	sig, err := helper.Hash(content)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (tmp *TokenMapProcessor) LoadCheckedEndBlockFromDB() error {
	hasEndBlockID, _ := tmp.storageClient.Has([]byte(TMapCheckedEndBlockKey), nil)
	if hasEndBlockID {
		blockIDBytes, err := tmp.storageClient.Get([]byte(TMapCheckedEndBlockKey), nil)
		if err != nil {
			return err
		}

		if result, err := strconv.ParseInt(string(blockIDBytes), DigitBase, DigitBitSize); err == nil {
			tmp.TokenMapCheckedEndBlock = result
		} else {
			return err
		}
	}

	return nil
}

func (tmp *TokenMapProcessor) LoadLastEventIDFromDB() error {
	hasLastEventID, _ := tmp.storageClient.Has([]byte(TMapLastEventIDKey), nil)
	if hasLastEventID {
		eventIDBytes, err := tmp.storageClient.Get([]byte(TMapLastEventIDKey), nil)
		if err != nil {
			return err
		}

		if result, err := strconv.ParseUint(string(eventIDBytes), DigitBase, DigitBitSize); err == nil {
			tmp.TokenMapLastEventID = result
		} else {
			return err
		}
	}

	return nil
}

// UpdateTokenMapLastEventID - Update LastEventID and EndBlock from DB.
func (tmp *TokenMapProcessor) UpdateTokenMapLastEventID(lastEventID uint64) error {
	return tmp.storageClient.Put(
		[]byte(TMapLastEventIDKey),
		[]byte(strconv.FormatUint(lastEventID, DigitBase)),
		nil)
}

// UpdateTokenMapCheckedEndBlock - Update LastEventID and EndBlock from DB.
func (tmp *TokenMapProcessor) UpdateTokenMapCheckedEndBlock(endblock int64) error {
	return tmp.storageClient.Put(
		[]byte(TMapCheckedEndBlockKey),
		[]byte(strconv.FormatInt(endblock, DigitBase)),
		nil)
}

// FetchStateSyncEventsWithTotal -.
func (tmp *TokenMapProcessor) FetchStateSyncEventsWithTotal(
	fromID uint64, toTime int64, stateFetchLimit int,
	total int,
) ([]*clerkTypes.EventRecord, error) {
	eventRecords := make([]*clerkTypes.EventRecord, 0)

	for {
		queryParams := fmt.Sprintf("from-id=%d&to-time=%d&limit=%d", fromID, toTime, stateFetchLimit)

		response, err := helper.FetchFromAPI(
			tmp.cliCtx,
			helper.GetHeimdallServerEndpointWithQuery(EventRecordURLPath, queryParams))
		if err != nil {
			return eventRecords, err
		}

		if response.Result == nil { // status 204
			break
		}

		var records []*clerkTypes.EventRecord
		if err := json.Unmarshal(response.Result, &records); err != nil {
			return eventRecords, err
		}

		if len(records) > 0 {
			eventRecords = append(eventRecords, records...)
		}

		if len(records) < stateFetchLimit {
			break
		}

		if len(eventRecords) >= total {
			break
		}

		for _, v := range records {
			if fromID < v.ID+1 {
				fromID = v.ID + 1
			}
		}
	}

	if len(eventRecords) > 0 {
		sort.SliceStable(eventRecords, func(i, j int) bool {
			return eventRecords[i].ID < eventRecords[j].ID
		})
	}

	return eventRecords, nil
}

// --------------------- state synced -----------------------.
var (
	byte32Type, _  = abi.NewType("bytes32", "", nil)
	bytesType, _   = abi.NewType("bytes", "", nil)
	addressType, _ = abi.NewType("address", "", nil)
	uint64Type, _  = abi.NewType("uint64", "", nil)
)

// Event type.
const (
	EventUnkonwn  = 0
	EventDeposit  = 1
	EventMaptoken = 2
)

// StateSyncedEvent - struct for state synced event.
type StateSyncedEvent struct {
	EventChainType  string `json:"eventChainType"`
	RootChainType   string `json:"rootChainType"`
	BlockHeight     uint64 `json:"blockHeight"` // block_height
	LogIndex        uint64 `json:"logIndex"`
	EventID         uint64 `json:"eventId"` // event id
	ContractAddress string `json:"contractAddresss"`
	TxHash          string `json:"txHash"` // tx_hash

	Data      string `json:"data"`      // transaction data
	EventType int8   `json:"eventType"` // event_type:  0: deposite 1: maptoken
	Address1  string `json:"address1"`  // deposite: user address maptoken: roottoken
	Address2  string `json:"address2"`  // depostite: roottoken maptoken: childtoken

	CreateTime time.Time `json:"createTime"`
}

// NewStateSyncedEvent - new one event.
func NewStateSyncedEvent(event *clerkTypes.EventRecord, chainType string) *StateSyncedEvent {
	stateSyncedEvent := &StateSyncedEvent{
		EventChainType:  chainType,
		RootChainType:   event.RootChainType,
		BlockHeight:     0,
		LogIndex:        event.LogIndex,
		EventID:         event.ID,
		ContractAddress: event.Contract.String(),
		TxHash:          event.TxHash.String(),
		Data:            strings.TrimLeft(hex.EncodeToString(event.Data), "0x"),
		CreateTime:      event.RecordTime,
		EventType:       EventUnkonwn,
	}

	return stateSyncedEvent
}

// IsMapTokenEvent - whether this event is EventMap.
func (sse *StateSyncedEvent) IsMapTokenEvent() bool {
	return sse.EventType == EventMaptoken
}

// ParshEvent - parse data to fields.
func (sse *StateSyncedEvent) ParshEvent() error {
	depositHash := crypto.Keccak256([]byte("DEPOSIT"))
	mapTokenHash := crypto.Keccak256([]byte("MAP_TOKEN"))

	arguments := abi.Arguments{
		{
			Name: "type",
			Type: byte32Type,
		},
		{
			Name: "data",
			Type: bytesType,
		},
	}

	stateData, err := arguments.Unpack(common.Hex2Bytes(sse.Data))
	if err != nil {
		return err
	}

	if len(stateData) != len(arguments) {
		return nil
	}

	btype, ok := stateData[0].([32]byte)
	if !ok {
		return nil
	}

	if bytes.Equal(btype[:], mapTokenHash) {
		if byte1, ok := stateData[1].([]byte); ok {
			sse.EventType = EventMaptoken
			if err = sse.parseTokenMap(byte1); err != nil {
				return err
			}
		}
	} else if bytes.Equal(btype[:], depositHash) {
		sse.EventType = EventDeposit
	}

	return nil
}

func (sse *StateSyncedEvent) parseTokenMap(src []byte) error {
	arguments2 := abi.Arguments{
		{
			Name: "rootToken",
			Type: addressType,
		},
		{
			Name: "childToken",
			Type: addressType,
		},
		{
			Name: "chainID",
			Type: uint64Type,
		},
		{
			Name: "data",
			Type: byte32Type,
		},
	}

	ret, err := arguments2.Unpack(src)
	if err != nil {
		return err
	}

	if len(ret) != len(arguments2) {
		return errors.New("data error")
	}

	if addr, ok := ret[0].(common.Address); ok {
		sse.Address1 = addr.String()
	}

	if addr, ok := ret[1].(common.Address); ok {
		sse.Address2 = addr.String()
	}

	return nil
}

// ConvertToTokenMap - generate a token map item from StateSyncedEvent.
func (sse *StateSyncedEvent) ConvertToTokenMap() (*TokenMapItem, error) {
	item := &TokenMapItem{
		RootChainType: sse.RootChainType,
		RootToken:     sse.Address1,
		ChildToken:    sse.Address2,
		EventID:       sse.EventID,
	}
	if !item.validate() {
		return nil, errors.New("item format is wrong")
	}

	return item, nil
}
