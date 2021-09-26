package chainmanager

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/chainmanager/types"
	"github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/maticnetwork/heimdall/params/subspace"
	hmTypes "github.com/maticnetwork/heimdall/types"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	NewChainParamsKey = []byte{0x11} // prefix key for when storing state
)

// Keeper stores all related data
type Keeper struct {
	cdc *codec.Codec
	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey
	// codespace
	codespace sdk.CodespaceType
	// param space
	paramSpace subspace.Subspace
	// contract caller
	contractCaller helper.ContractCaller
}

// NewKeeper create new keeper
func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	paramSpace subspace.Subspace,
	codespace sdk.CodespaceType,
	caller helper.ContractCaller,
) Keeper {
	// create keeper
	keeper := Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		paramSpace:     paramSpace.WithKeyTable(types.ParamKeyTable()),
		codespace:      codespace,
		contractCaller: caller,
	}
	return keeper
}

// Codespace returns the codespace
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// GetChainParams get new chain Params
func (k *Keeper) GetChainParams(ctx sdk.Context, rootChain string) (types.ChainInfo, error) {
	var chainInfo types.ChainInfo
	store := ctx.KVStore(k.storeKey)
	key := append(NewChainParamsKey, hmTypes.GetRootChainID(rootChain))
	if store.Has(key) {
		err := k.cdc.UnmarshalBinaryBare(store.Get(key), &chainInfo)
		if err != nil {
			k.Logger(ctx).Error("Error marshalling chain params from store value",
				"root", rootChain, "error", err)
			return chainInfo, err
		} else {
			return chainInfo, nil
		}
	}
	return chainInfo, common.ErrNoChainParamsFound(k.Codespace())
}

// AddNewChainParams adds new chain into chain list
func (k *Keeper) AddNewChainParams(ctx sdk.Context, chainInfo types.ChainInfo) error {
	key := append(NewChainParamsKey, hmTypes.GetRootChainID(chainInfo.RootChainType))
	value, err := k.cdc.MarshalBinaryBare(chainInfo)
	if err != nil {
		k.Logger(ctx).Error("Error marshalling chain info", "root", chainInfo.RootChainType, "error", err)
		return err
	}
	if err := k.setChainInfoStore(ctx, key, value); err != nil {
		return err
	}
	k.Logger(ctx).Info("Adding new chain info to state", "chainInfo", chainInfo)
	return nil
}

// GetNewChainParamsList get new chain into chain list
func (k *Keeper) GetNewChainParamsList(ctx sdk.Context) []types.ChainInfo {
	store := ctx.KVStore(k.storeKey)
	// get checkpoint header iterator
	iterator := sdk.KVStorePrefixIterator(store, NewChainParamsKey)
	defer iterator.Close()

	// create headers
	var chainInfos []types.ChainInfo

	// loop through validators to get valid validators
	for ; iterator.Valid(); iterator.Next() {
		var chainInfo types.ChainInfo
		if err := k.cdc.UnmarshalBinaryBare(iterator.Value(), &chainInfo); err == nil {
			chainInfos = append(chainInfos, chainInfo)
		}
	}
	return chainInfos
}

// setEventRecordStore adds value to store by key
func (k *Keeper) setChainInfoStore(ctx sdk.Context, key, value []byte) error {
	store := ctx.KVStore(k.storeKey)
	// check if already set
	if store.Has(key) {
		return errors.New("key already exists")
	}

	// store value in provided key
	store.Set(key, value)
	// return
	return nil
}

// GetChainActivationHeight get new chain Params
func (k *Keeper) GetChainActivationHeight(ctx sdk.Context, rootChain string) uint64 {
	chainParams, err := k.GetChainParams(ctx, rootChain)
	if err != nil {
		return 0
	}
	res := chainParams.ActivationHeight
	return res
}

// -----------------------------------------------------------------------------
// Params

// SetParams sets the chainmanager module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetParams gets the chainmanager module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}
