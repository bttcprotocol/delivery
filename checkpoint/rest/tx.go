package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/maticnetwork/heimdall/checkpoint"
	"github.com/maticnetwork/heimdall/helper"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc(
		"/checkpoint/new",
		newCheckpointHandler(cdc, cliCtx),
	).Methods("POST")
	r.HandleFunc("/checkpoint/ack", NewCheckpointACKHandler(cdc, cliCtx)).Methods("POST")
	r.HandleFunc("/checkpoint/no-ack", NewCheckpointNoACKHandler(cdc, cliCtx)).Methods("POST")
}

type (
	// HeaderBlock struct for incoming checkpoint
	HeaderBlock struct {
		Proposer   common.Address `json:"proposer"`
		RootHash   common.Hash    `json:"rootHash"`
		StartBlock uint64         `json:"startBlock"`
		EndBlock   uint64         `json:"endBlock"`
	}
	// HeaderACK struct for sending ACK for a new headers
	// by providing the header index assigned my mainchain contract
	HeaderACK struct {
		HeaderBlock uint64 `json:"headerBlock"`
	}
)

func newCheckpointHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m HeaderBlock

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		err = json.Unmarshal(body, &m)
		if err != nil {
			RestLogger.Error("Error unmarshalling json epoch checkpoint", "error", err)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := checkpoint.NewMsgCheckpointBlock(
			m.Proposer,
			m.StartBlock,
			m.EndBlock,
			m.RootHash,
			uint64(time.Now().Unix()),
		)

		txBytes, err := helper.CreateTxBytes(msg)
		if err != nil {
			RestLogger.Error("Unable to create txBytes", "proposer", m.Proposer.Hex(), "endBlock", m.EndBlock, "startBlock", m.StartBlock, "rootHash", m.RootHash.Hex())
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := helper.SendTendermintRequest(cliCtx, txBytes, helper.BroadcastAsync)
		if err != nil {
			RestLogger.Error("Error while sending request to Tendermint", "error", err)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		result, err := json.Marshal(&resp)
		if err != nil {
			RestLogger.Error("Error while marshalling tendermint response", "error", err)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, result, cliCtx.Indent)
	}
}

func NewCheckpointACKHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}

		var m HeaderACK
		err = json.Unmarshal(body, &m)
		if err != nil {
			RestLogger.Error("Error unmarshalling Header ACK", "error", err)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}

		// create new msg checkpoint ack
		msg := checkpoint.NewMsgCheckpointAck(m.HeaderBlock, uint64(time.Now().Unix()))
		txBytes, err := helper.CreateTxBytes(msg)
		if err != nil {
			RestLogger.Error("Unable to create txBytes", "error", err, "headerBlock", m.HeaderBlock)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}

		resp, err := helper.SendTendermintRequest(cliCtx, txBytes, helper.BroadcastAsync)
		if err != nil {
			RestLogger.Error("Error while sending request to Tendermint", "error", err)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}

		result, err := json.Marshal(&resp)
		if err != nil {
			RestLogger.Error("Error while marshalling tendermint response", "error", err)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}
		rest.PostProcessResponse(w, cdc, result, cliCtx.Indent)
	}
}

func NewCheckpointNoACKHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// create new msg checkpoint ack
		msg := checkpoint.NewMsgCheckpointNoAck(uint64(time.Now().Unix()))

		txBytes, err := helper.CreateTxBytes(msg)
		if err != nil {
			RestLogger.Error("Unable to create txBytes", "error", err, "timestamp", time.Now().Unix())
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}

		resp, err := helper.SendTendermintRequest(cliCtx, txBytes, helper.BroadcastAsync)
		if err != nil {
			RestLogger.Error("Error while sending request to Tendermint", "error", err)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}

		result, err := json.Marshal(&resp)
		if err != nil {
			RestLogger.Error("Error while marshalling tendermint response", "error", err)
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}
		rest.PostProcessResponse(w, cdc, result, cliCtx.Indent)
	}
}
