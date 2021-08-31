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

// GetStakingQueue
func (k *Keeper) GetStakingQueue(ctx sdk.Context, rootID byte) ([]stakingTypes.StakingRecord, error) {
	key := getStakingQueueKey(rootID)
	store := ctx.KVStore(k.storeKey)

	var records []stakingTypes.StakingRecord
	if store.Has(key) {
		err := k.cdc.UnmarshalBinaryBare(store.Get(key), &records)
		if err != nil {
			k.Logger(ctx).Error("Error unmarshalling staking queue record", "root", rootID, "error", err)
			return nil, err
		}
		return records, nil
	}
	return nil, nil
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
			if index+1 == len(records) {
				store.Delete(key)
				return
			}
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

// UpdateStakingRecordTimestamp update staking record timestamp
func (k *Keeper) UpdateStakingRecordTimestamp(ctx sdk.Context, rootID byte, validatorID hmTypes.ValidatorID, nonce uint64, timestamp uint64) {
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
			records[index].TimeStamp = timestamp
			out, err := k.cdc.MarshalBinaryBare(records)
			if err != nil {
				k.Logger(ctx).Error("Error marshalling staking queue record", "error", err)
				return
			}
			store.Set(key, out)
			return
		}
	}
}
