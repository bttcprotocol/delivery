package checkpoint

import (
	"errors"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/maticnetwork/heimdall/chainmanager"
	"github.com/maticnetwork/heimdall/checkpoint/types"
	cmn "github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/params/subspace"
	"github.com/maticnetwork/heimdall/staking"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

var (
	DefaultValue = []byte{0x01} // Value to store in CacheCheckpoint and CacheCheckpointACK & ValidatorSetChange Flag

	BufferCheckpointSyncKey = []byte{0x02} // Key to store checkpoint in buffer

	ACKCountKey         = []byte{0x11} // key to store ACK count
	BufferCheckpointKey = []byte{0x12} // Key to store checkpoint in buffer
	EthCheckpointKey    = []byte{0x13} // prefix key for when storing checkpoint after ACK
	LastNoACKKey        = []byte{0x14} // key to store last no-ack

	TronCheckpointKey = []byte{0x21} // prefix key for when storing checkpoint after ACK
	BscCheckpointKey  = []byte{0x22} // prefix key for when storing checkpoint after ACK

)

// ModuleCommunicator manages different module interaction
type ModuleCommunicator interface {
	GetAllDividendAccounts(ctx sdk.Context) []hmTypes.DividendAccount
}

// Keeper stores all related data
type Keeper struct {
	cdc *codec.Codec
	// staking keeper
	sk staking.Keeper
	ck chainmanager.Keeper
	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey
	// codespace
	codespace sdk.CodespaceType
	// param space
	paramSpace subspace.Subspace

	// module communicator
	moduleCommunicator ModuleCommunicator
}

// NewKeeper create new keeper
func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	paramSpace subspace.Subspace,
	codespace sdk.CodespaceType,
	stakingKeeper staking.Keeper,
	chainKeeper chainmanager.Keeper,
	moduleCommunicator ModuleCommunicator,
) Keeper {
	keeper := Keeper{
		cdc:                cdc,
		storeKey:           storeKey,
		paramSpace:         paramSpace.WithKeyTable(types.ParamKeyTable()),
		codespace:          codespace,
		sk:                 stakingKeeper,
		ck:                 chainKeeper,
		moduleCommunicator: moduleCommunicator,
	}
	return keeper
}

// Codespace returns the codespace
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// AddCheckpoint adds checkpoint into final blocks
func (k *Keeper) AddCheckpoint(ctx sdk.Context, checkpointNumber uint64, checkpoint hmTypes.Checkpoint, rootChain string) error {
	key := GetCheckpointKey(checkpointNumber, rootChain)
	err := k.addCheckpoint(ctx, key, checkpoint)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info("Adding good checkpoint to state",
		"root", rootChain, "checkpoint", checkpoint, "checkpointNumber", checkpointNumber)
	return nil
}

func getCheckpointBufferKey(rootID byte) []byte {
	return append(BufferCheckpointKey, rootID)
}

// SetCheckpointBuffer set Checkpoint Buffer
func (k *Keeper) SetCheckpointBuffer(ctx sdk.Context, checkpoint hmTypes.Checkpoint, rootChain string) error {
	key := getCheckpointBufferKey(hmTypes.GetRootChainID(rootChain))
	err := k.addCheckpoint(ctx, key, checkpoint)
	if err != nil {
		return err
	}
	return nil
}

// addCheckpoint adds checkpoint to store
func (k *Keeper) addCheckpoint(ctx sdk.Context, key []byte, checkpoint hmTypes.Checkpoint) error {
	store := ctx.KVStore(k.storeKey)

	// create Checkpoint block and marshall
	out, err := k.cdc.MarshalBinaryBare(checkpoint)
	if err != nil {
		k.Logger(ctx).Error("Error marshalling checkpoint", "error", err)
		return err
	}

	// store in key provided
	store.Set(key, out)

	return nil
}

// GetCheckpointByNumber to get checkpoint by checkpoint number
func (k *Keeper) GetCheckpointByNumber(ctx sdk.Context, number uint64, rootChain string) (hmTypes.Checkpoint, error) {
	store := ctx.KVStore(k.storeKey)
	var _checkpoint hmTypes.Checkpoint
	checkpointKey := GetCheckpointKey(number, rootChain)

	if store.Has(checkpointKey) {
		err := k.cdc.UnmarshalBinaryBare(store.Get(checkpointKey), &_checkpoint)
		if err != nil {
			return _checkpoint, err
		} else {
			return _checkpoint, nil
		}
	}
	return _checkpoint, errors.New("Invalid checkpoint Index")
}

// GetCheckpointList returns all checkpoints with params like page and limit
func (k *Keeper) GetCheckpointList(ctx sdk.Context, page uint64, limit uint64, rootChain string) ([]hmTypes.Checkpoint, error) {
	store := ctx.KVStore(k.storeKey)

	// create headers
	var checkpoints []hmTypes.Checkpoint

	// have max limit
	if limit > 20 {
		limit = 20
	}

	// get paginated iterator
	iterator := hmTypes.KVStorePrefixIteratorPaginated(store, EthCheckpointKey, uint(page), uint(limit))
	switch rootChain {
	case hmTypes.RootChainTypeTron:
		iterator = hmTypes.KVStorePrefixIteratorPaginated(store, TronCheckpointKey, uint(page), uint(limit))
	case hmTypes.RootChainTypeBsc:
		iterator = hmTypes.KVStorePrefixIteratorPaginated(store, BscCheckpointKey, uint(page), uint(limit))
	}

	// loop through validators to get valid validators
	for ; iterator.Valid(); iterator.Next() {
		var checkpoint hmTypes.Checkpoint
		if err := k.cdc.UnmarshalBinaryBare(iterator.Value(), &checkpoint); err == nil {
			checkpoints = append(checkpoints, checkpoint)
		}
	}

	return checkpoints, nil
}

// GetLastCheckpoint gets last checkpoint, checkpoint number = TotalACKs
func (k *Keeper) GetLastCheckpoint(ctx sdk.Context, rootChain string) (hmTypes.Checkpoint, error) {
	store := ctx.KVStore(k.storeKey)
	acksCount := k.GetACKCount(ctx, rootChain)

	lastCheckpointKey := acksCount

	// fetch checkpoint and unmarshall
	var _checkpoint hmTypes.Checkpoint

	// no checkpoint received
	// header key
	headerKey := GetCheckpointKey(lastCheckpointKey, rootChain)
	if store.Has(headerKey) {
		err := k.cdc.UnmarshalBinaryBare(store.Get(headerKey), &_checkpoint)
		if err != nil {
			k.Logger(ctx).Error("Unable to fetch last checkpoint from store",
				"root", rootChain, "key", lastCheckpointKey, "acksCount", acksCount)
			return _checkpoint, err
		} else {
			return _checkpoint, nil
		}
	}
	return _checkpoint, cmn.ErrNoCheckpointFound(k.Codespace())
}

// GetCheckpointKey appends prefix to checkpointNumber
func GetCheckpointKey(checkpointNumber uint64, rootChain string) []byte {
	var key []byte
	switch rootChain {
	case hmTypes.RootChainTypeEth:
		key = EthCheckpointKey
	case hmTypes.RootChainTypeTron:
		key = TronCheckpointKey
	case hmTypes.RootChainTypeBsc:
		key = BscCheckpointKey
	}
	checkpointNumberBytes := []byte(strconv.FormatUint(checkpointNumber, 10))
	return append(key, checkpointNumberBytes...)
}

// HasStoreValue check if value exists in store or not
func (k *Keeper) HasStoreValue(ctx sdk.Context, key []byte) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(key)
}

// FlushCheckpointBuffer flushes Checkpoint Buffer
func (k *Keeper) FlushCheckpointBuffer(ctx sdk.Context, rootChain string) {
	store := ctx.KVStore(k.storeKey)
	key := getCheckpointBufferKey(hmTypes.GetRootChainID(rootChain))
	store.Delete(key)
}

// GetCheckpointFromBuffer gets checkpoint in buffer
func (k *Keeper) GetCheckpointFromBuffer(ctx sdk.Context, rootChain string) (*hmTypes.Checkpoint, error) {
	store := ctx.KVStore(k.storeKey)

	// checkpoint block header
	var checkpoint hmTypes.Checkpoint
	key := getCheckpointBufferKey(hmTypes.GetRootChainID(rootChain))

	if store.Has(key) {
		// Get checkpoint and unmarshall
		err := k.cdc.UnmarshalBinaryBare(store.Get(key), &checkpoint)
		return &checkpoint, err
	}

	return nil, errors.New("No checkpoint found in buffer")
}

func getCheckpointSyncKey(rootID byte) []byte {
	return append(BufferCheckpointSyncKey, rootID)
}

// CheckpointSyncBuffer set Checkpoint sync Buffer
func (k *Keeper) SetCheckpointSyncBuffer(ctx sdk.Context, checkpoint hmTypes.Checkpoint, rootChain string) error {
	store := ctx.KVStore(k.storeKey)

	key := getCheckpointSyncKey(hmTypes.GetRootChainID(rootChain))

	// create Checkpoint sync and marshall
	out, err := k.cdc.MarshalBinaryBare(checkpoint)
	if err != nil {
		k.Logger(ctx).Error("Error marshalling checkpoint", "error", err)
		return err
	}

	// store in key provided
	store.Set(key, out)
	return nil
}

// GetCheckpointSyncFromBuffer gets checkpoint sync in buffer
func (k *Keeper) GetCheckpointSyncFromBuffer(ctx sdk.Context, rootChain string) (*hmTypes.Checkpoint, error) {
	store := ctx.KVStore(k.storeKey)

	key := getCheckpointSyncKey(hmTypes.GetRootChainID(rootChain))
	// checkpoint block header
	if store.Has(key) {
		var checkpoint hmTypes.Checkpoint
		// Get checkpoint and unmarshall
		err := k.cdc.UnmarshalBinaryBare(store.Get(key), &checkpoint)
		return &checkpoint, err
	}

	return nil, errors.New("no checkpoint sync found in buffer")
}

// FlushCheckpointSyncBuffer flushes Checkpoint sync Buffer
func (k *Keeper) FlushCheckpointSyncBuffer(ctx sdk.Context, rootChain string) {
	store := ctx.KVStore(k.storeKey)

	key := getCheckpointSyncKey(hmTypes.GetRootChainID(rootChain))
	store.Delete(key)
}

// SetLastNoAck set last no-ack object
func (k *Keeper) SetLastNoAck(ctx sdk.Context, timestamp uint64) {
	store := ctx.KVStore(k.storeKey)
	// convert timestamp to bytes
	value := []byte(strconv.FormatUint(timestamp, 10))
	// set no-ack
	store.Set(LastNoACKKey, value)
}

// GetLastNoAck returns last no ack
func (k *Keeper) GetLastNoAck(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	// check if ack count is there
	if store.Has(LastNoACKKey) {
		// get current ACK count
		result, err := strconv.ParseUint(string(store.Get(LastNoACKKey)), 10, 64)
		if err == nil {
			return uint64(result)
		}
	}
	return 0
}

// GetCheckpoints get checkpoint all checkpoints
func (k *Keeper) GetCheckpoints(ctx sdk.Context) []hmTypes.Checkpoint {
	store := ctx.KVStore(k.storeKey)
	// get checkpoint header iterator
	iterator := sdk.KVStorePrefixIterator(store, TronCheckpointKey)
	defer iterator.Close()

	// create headers
	var headers []hmTypes.Checkpoint

	// loop through validators to get valid validators
	for ; iterator.Valid(); iterator.Next() {
		var checkpoint hmTypes.Checkpoint
		if err := k.cdc.UnmarshalBinaryBare(iterator.Value(), &checkpoint); err == nil {
			headers = append(headers, checkpoint)
		}
	}
	return headers
}

// GetOtherCheckpoints get checkpoint all checkpoints
func (k *Keeper) GetOtherCheckpoints(ctx sdk.Context, rootChain string) []hmTypes.Checkpoint {
	store := ctx.KVStore(k.storeKey)
	// get checkpoint header iterator
	var iterator sdk.Iterator
	if hmTypes.RootChainTypeTron == rootChain {
		iterator = sdk.KVStorePrefixIterator(store, TronCheckpointKey)
		defer iterator.Close()
	}
	// create headers
	var headers []hmTypes.Checkpoint

	// loop through validators to get valid validators
	for ; iterator.Valid(); iterator.Next() {
		var checkpoint hmTypes.Checkpoint
		if err := k.cdc.UnmarshalBinaryBare(iterator.Value(), &checkpoint); err == nil {
			headers = append(headers, checkpoint)
		}
	}
	return headers
}

//
// Ack count
//

func GetAckCountKey(rootID byte) []byte {
	return append(ACKCountKey, rootID)
}

// GetACKCount returns current ACK count
func (k Keeper) GetACKCount(ctx sdk.Context, rootChain string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := GetAckCountKey(hmTypes.GetRootChainID(rootChain))
	// checkpoint block header
	if store.Has(key) {
		// check if ack count is there
		ackCount, err := strconv.ParseUint(string(store.Get(key)), 10, 64)
		if err != nil {
			k.Logger(ctx).Error("Unable to convert key to int")
		} else {
			return ackCount
		}
	}

	return 0
}

// UpdateACKCountWithValue updates ACK with value
func (k Keeper) UpdateACKCountWithValue(ctx sdk.Context, value uint64, rootChain string) {
	store := ctx.KVStore(k.storeKey)

	// convert
	ackCount := []byte(strconv.FormatUint(value, 10))

	// update
	key := GetAckCountKey(hmTypes.GetRootChainID(rootChain))
	store.Set(key, ackCount)
}

// UpdateACKCount updates ACK count by 1
func (k Keeper) UpdateACKCount(ctx sdk.Context, rootChain string) {
	store := ctx.KVStore(k.storeKey)

	// get current ACK Count
	ACKCount := k.GetACKCount(ctx, rootChain)

	// increment by 1
	ACKs := []byte(strconv.FormatUint(ACKCount+1, 10))
	// update
	key := GetAckCountKey(hmTypes.GetRootChainID(rootChain))
	store.Set(key, ACKs)

}

// -----------------------------------------------------------------------------
// Params

// SetParams sets the auth module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetParams gets the auth module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}
