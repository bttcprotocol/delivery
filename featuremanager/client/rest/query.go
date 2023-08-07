package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/maticnetwork/heimdall/featuremanager/types"
)

// HTTP request handler to query the featuremanager params values.
func queryTargetFeatureHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		cliCtx, isSuccess := rest.ParseQueryHeightOrReturnBadRequest(writer, cliCtx, req)
		if !isSuccess {
			return
		}

		vars := mux.Vars(req)
		targetFeature, ok := vars["target-feature"]

		if !ok {
			err := fmt.Errorf("'%s' is not a valid feature", vars["target-feature"])
			rest.WriteErrorResponse(writer, http.StatusBadRequest, err.Error())

			return
		}

		// get query params
		queryParams, err := cliCtx.Codec.MarshalJSON(types.QueryTargetFeatureParam{
			TargetFeature: targetFeature,
		})
		if err != nil {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryTargetFeature)

		res, height, err := cliCtx.QueryWithData(route, queryParams)
		if err != nil {
			rest.WriteErrorResponse(writer, http.StatusInternalServerError, err.Error())

			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(writer, cliCtx, res)
	}
}

func queryFeatureMapHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(writer, cliCtx, req)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryFeatureMap)

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(writer, http.StatusInternalServerError, err.Error())

			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(writer, cliCtx, res)
	}
}
