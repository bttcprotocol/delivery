package rollback

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/maticnetwork/heimdall/cmd/deliveryd/rollback/rootmulti"

	cosmosTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/maticnetwork/heimdall/auth/types"
	bankTypes "github.com/maticnetwork/heimdall/bank/types"
	borTypes "github.com/maticnetwork/heimdall/bor/types"
	chainmanagerTypes "github.com/maticnetwork/heimdall/chainmanager/types"
	checkpointTypes "github.com/maticnetwork/heimdall/checkpoint/types"
	clerkTypes "github.com/maticnetwork/heimdall/clerk/types"
	govTypes "github.com/maticnetwork/heimdall/gov/types"
	paramsTypes "github.com/maticnetwork/heimdall/params/types"
	sidechannelTypes "github.com/maticnetwork/heimdall/sidechannel/types"
	slashingTypes "github.com/maticnetwork/heimdall/slashing/types"
	stakingTypes "github.com/maticnetwork/heimdall/staking/types"
	supplyTypes "github.com/maticnetwork/heimdall/supply/types"
	topupTypes "github.com/maticnetwork/heimdall/topup/types"

	dbm "github.com/tendermint/tm-db"
)

type MutiStore struct {
	rootmulti.Store
	mountKeys map[string]*cosmosTypes.KVStoreKey
}

func NewMultiStore(db dbm.DB) *MutiStore {
	return &MutiStore{
		*rootmulti.NewStore(db),
		make(map[string]*cosmosTypes.KVStoreKey),
	}
}

func (rs *MutiStore) MountKeys() {
	rs.mountKeys = sdk.NewKVStoreKeys(
		bam.MainStoreKey,
		sidechannelTypes.StoreKey,
		authTypes.StoreKey,
		bankTypes.StoreKey,
		supplyTypes.StoreKey,
		govTypes.StoreKey,
		chainmanagerTypes.StoreKey,
		stakingTypes.StoreKey,
		slashingTypes.StoreKey,
		checkpointTypes.StoreKey,
		borTypes.StoreKey,
		clerkTypes.StoreKey,
		topupTypes.StoreKey,
		paramsTypes.StoreKey,
	)

	for _, key := range rs.mountKeys {
		rs.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}
}
