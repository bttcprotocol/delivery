package staking

//
// Staking sync
//

import (
	hmTypes "github.com/maticnetwork/heimdall/types"

	stakingTypes "github.com/maticnetwork/heimdall/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//
// staking queue
//

func getStakingQueueKey(rootID byte) []byte {
	return append(stakingSendingQueueKey, rootID)
}

// AddStakingRecordToQueue adds staking record to root cueue
func (k *Keeper) AddStakingRecordToQueue(ctx sdk.Context, rootID byte, stakingRecord stakingTypes.StakingRecord) {
	key := getStakingQueueKey(rootID)
	store := ctx.KVStore(k.storeKey)

	var records []stakingTypes.StakingRecord
	if store.Has(key) {
		err := k.cdc.UnmarshalBinaryBare(store.Get(key), &records)
		if err != nil {
			k.Logger(ctx).Error("Error unmarshalling staking queue record", "root", rootID, "error", err)
			return
		}
	}
	records = append(records, stakingRecord)
	out, err := k.cdc.MarshalBinaryBare(records)
	if err != nil {
		k.Logger(ctx).Error("Error marshalling staking queue record", "error", err)
		return
	}
	store.Set(key, out)
}

// GetNextStakingRecordFromQueue
func (k *Keeper) GetNextStakingRecordFromQueue(ctx sdk.Context, rootID byte) (*stakingTypes.StakingRecord, error) {
	key := getStakingQueueKey(rootID)
	store := ctx.KVStore(k.storeKey)

	var records []stakingTypes.StakingRecord
	if store.Has(key) {
		err := k.cdc.UnmarshalBinaryBare(store.Get(key), &records)
		if err != nil {
			k.Logger(ctx).Error("Error unmarshalling staking queue record", "root", rootID, "error", err)
			return nil, err
		}
		return &records[0], nil
	}
	return nil, nil
}

// findStakingRecordFromQueue
func (k *Keeper) findStakingRecordFromQueue(ctx sdk.Context, rootID byte, stakingRecord *stakingTypes.StakingRecord) bool {
	key := getStakingQueueKey(rootID)
	store := ctx.KVStore(k.storeKey)

	var records []stakingTypes.StakingRecord
	if store.Has(key) {
		err := k.cdc.UnmarshalBinaryBare(store.Get(key), &records)
		if err != nil {
			k.Logger(ctx).Error("Error unmarshalling staking queue record", "root", rootID, "error", err)
			return false
		}
		for _, record := range records {
			if record.ValidatorID == stakingRecord.ValidatorID &&
				record.Nonce == stakingRecord.Nonce &&
				record.TxHash == stakingRecord.TxHash {
				return true
			}
		}
		return false
	}
	return false
}

// removeStakingRecordFromQueue
func (k *Keeper) removeStakingRecordFromQueue(ctx sdk.Context, rootID byte, validatorID hmTypes.ValidatorID, nonce uint64) {
	key := getStakingQueueKey(rootID)
	store := ctx.KVStore(k.storeKey)

	var records []stakingTypes.StakingRecord
	if !store.Has(key) {
		return
	}
	err := k.cdc.UnmarshalBinaryBare(store.Get(key), &records)
	if err != nil {
		k.Logger(ctx).Error("Error unmarshalling staking queue record", "root", rootID, "error", err)
		return
	}

	for index, record := range records {
		if record.ValidatorID == validatorID && record.Nonce == nonce {
			results := records[index+1:]
			out, err := k.cdc.MarshalBinaryBare(results)
			if err != nil {
				k.Logger(ctx).Error("Error marshalling staking queue record", "error", err)
				return
			}
			store.Set(key, out)
			return
		}
	}
}

//
// staking buffer
//

func getBufferStakingKey(rootChain string) []byte {
	rootChainID := hmTypes.GetRootChainID(rootChain)
	key := append(bufferStakeEventKey, rootChainID)
	return key
}

// SetStakingRecordBuffer store StakeEvent Buffer
func (k *Keeper) SetStakingRecordBuffer(ctx sdk.Context, stakingInfo stakingTypes.StakingRecord, rootChain string) error {
	store := ctx.KVStore(k.storeKey)

	// create staking event and marshall
	out, err := k.cdc.MarshalBinaryBare(stakingInfo)
	if err != nil {
		k.Logger(ctx).Error("Error marshalling staking info", "error", err)
		return err
	}
	key := getBufferStakingKey(rootChain)
	// store in key provided
	store.Set(key, out)

	return nil
}

// GetStakingRecordFromBuffer returns staking buffer
func (k *Keeper) GetStakingRecordFromBuffer(ctx sdk.Context, rootChain string) (*stakingTypes.StakingRecord, error) {
	store := ctx.KVStore(k.storeKey)
	var stakingInfo stakingTypes.StakingRecord
	key := getBufferStakingKey(rootChain)
	if store.Has(key) {
		// Get staking info and unmarshall
		err := k.cdc.UnmarshalBinaryBare(store.Get(key), &stakingInfo)
		return &stakingInfo, err
	}
	return nil, nil
}

// FlushStakingBuffer flush staking buffer
func (k *Keeper) FlushStakingBuffer(ctx sdk.Context, rootChain string) {
	store := ctx.KVStore(k.storeKey)
	key := getBufferStakingKey(rootChain)
	store.Delete(key)
}