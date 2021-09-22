package chainmanager

import (
	"encoding/json"

	hmTpyes "github.com/maticnetwork/heimdall/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/maticnetwork/heimdall/chainmanager/types"
)

// NewQuerier creates a querier for auth REST endpoints
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, req, keeper)
		case types.QueryNewChainParam:
			return queryNewChainParams(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown chainmanager query endpoint")
		}
	}
}

func queryParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	bz, err := json.Marshal(keeper.GetParams(ctx))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// query for new chain params, replace address for new eth fork
func queryNewChainParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryChainParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil && len(req.Data) != 0 {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to parse params", err.Error()))
	}
	response := keeper.GetParams(ctx)
	newChainParams, _ := keeper.GetChainParams(ctx, params.RootChain)
	if params.RootChain == hmTpyes.RootChainTypeBsc {
		response.MainchainTxConfirmations = newChainParams.TxConfirmations
		response.ChainParams.RootChainAddress = newChainParams.RootChainAddress
		response.ChainParams.StateSenderAddress = newChainParams.StateSenderAddress
		response.ChainParams.StakingInfoAddress = newChainParams.StakingInfoAddress
		response.ChainParams.StakingManagerAddress = newChainParams.StakingManagerAddress
	}
	bz, err := json.Marshal(response)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
