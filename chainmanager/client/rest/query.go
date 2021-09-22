package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	chainTypes "github.com/maticnetwork/heimdall/chainmanager/types"
)

// HTTP request handler to query the auth params values
func paramsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", chainTypes.QuerierRoute, chainTypes.QueryParams)
		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryNewParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		// get checkpoint number
		vars := mux.Vars(r)
		rootChain, ok := vars["root"]
		if !ok {
			err := fmt.Errorf("'%s' is not a valid rootChain", vars["root"])
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// get query params
		queryParams, err := cliCtx.Codec.MarshalJSON(chainTypes.NewQueryChainParams(rootChain))
		if err != nil {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", chainTypes.QuerierRoute, chainTypes.QueryNewChainParam)
		res, height, err := cliCtx.QueryWithData(route, queryParams)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
