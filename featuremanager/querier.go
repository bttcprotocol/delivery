package featuremanager

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/maticnetwork/heimdall/featuremanager/types"
)

// NewQuerier creates a querier for auth REST endpoints.
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryTargetFeature:
			return queryTargetFeature(ctx, req, keeper)

		case types.QueryFeatureMap:
			return queryPropsoalChainParamMap(ctx, keeper)

		default:
			return nil, sdk.ErrUnknownRequest("unknown featuremanager query endpoint")
		}
	}
}

func queryTargetFeature(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryTargetFeatureParam
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil && len(req.Data) != 0 {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to parse params", err.Error()))
	}

	res := keeper.GetFeatureParams(ctx).FeatureParamMap

	rawData, ok := res[params.TargetFeature]
	if !ok {
		rawData = types.FeatureData{} //nolint: exhaustivestruct
	}

	feature := rawData.Plainify()

	bz, err := json.Marshal(feature)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

func queryPropsoalChainParamMap(ctx sdk.Context, keeper Keeper) ([]byte, sdk.Error) {
	data, err := json.Marshal(keeper.GetFeatureParams(ctx))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return data, nil
}
