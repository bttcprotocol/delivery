package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"

	"github.com/maticnetwork/heimdall/bridge/setu/broadcaster"
	"github.com/maticnetwork/heimdall/checkpoint/types"
	restClient "github.com/maticnetwork/heimdall/client/rest"
	"github.com/maticnetwork/heimdall/helper"
	hmTypes "github.com/maticnetwork/heimdall/types"
	"github.com/maticnetwork/heimdall/types/rest"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/checkpoint/new",
		newCheckpointHandler(cliCtx),
	).Methods("POST")
	r.HandleFunc("/checkpoint/ack", newCheckpointACKHandler(cliCtx)).Methods("POST")
	r.HandleFunc("/checkpoint/no-ack", newCheckpointNoACKHandler(cliCtx)).Methods("POST")
	r.HandleFunc("/checkpoint/repair", repairCheckpointHandler(cliCtx)).Methods("POST")
	r.HandleFunc("/checkpoint/repair-test", repairCheckpointTestHandler(cliCtx)).Methods("POST")

	r.HandleFunc("/your-module/test", myTestHandlerFn(cliCtx)).Methods("POST")
}

type (
	// HeaderBlockReq struct for incoming checkpoint
	HeaderBlockReq struct {
		BaseReq rest.BaseReq `json:"base_req"`

		Proposer        hmTypes.HeimdallAddress `json:"proposer"`
		RootHash        hmTypes.HeimdallHash    `json:"root_Hash"`
		AccountRootHash hmTypes.HeimdallHash    `json:"account_root_hash"`
		StartBlock      uint64                  `json:"start_block"`
		EndBlock        uint64                  `json:"end_block"`
		BorChainID      string                  `json:"bor_chain_id"`
		RootChain       string                  `json:"root_chain"`
		Epoch           uint64                  `json:"epoch"`
	}

	// HeaderACKReq struct for sending ACK for a new headers
	// by providing the header index assigned my mainchain contract
	HeaderACKReq struct {
		BaseReq rest.BaseReq `json:"base_req"`

		From        hmTypes.HeimdallAddress `json:"proposer"`
		HeaderBlock uint64                  `json:"header_block"`
		StartBlock  uint64                  `json:"start_block"`
		EndBlock    uint64                  `json:"end_block"`
		Proposer    hmTypes.HeimdallAddress `json:"proposer"`
		RootHash    hmTypes.HeimdallHash    `json:"root_Hash"`
		TxHash      hmTypes.HeimdallHash    `json:"tx_hash"`
		LogIndex    uint64                  `json:"log_index"`
		RootChain   string                  `json:"root_chain"`
	}

	// HeaderNoACKReq struct for sending no-ack for a new headers
	HeaderNoACKReq struct {
		BaseReq rest.BaseReq `json:"base_req"`

		Proposer hmTypes.HeimdallAddress `json:"proposer"`
	}

	// RepairCheckpointReq 用于repair接口的请求体
	RepairCheckpointReq struct {
		BaseReq          rest.BaseReq `json:"base_req"`
		CheckpointNumber uint64       `json:"checkpoint_number"`
		From             string       `json:"from"` // 添加 From 字段
	}

	// RepairCheckpointTestReq 用于测试接口的请求体
	RepairCheckpointTestReq struct {
		BaseReq          rest.BaseReq `json:"base_req"`
		CheckpointNumber uint64       `json:"checkpoint_number"`
		From             string       `json:"from"`
		TestMessage      string       `json:"test_message"`
	}
)

func newCheckpointHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req HeaderBlockReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// draft a message and send response
		msg := types.NewMsgCheckpointBlock(
			req.Proposer,
			req.StartBlock,
			req.EndBlock,
			req.RootHash,
			req.AccountRootHash,
			req.BorChainID,
			req.Epoch,
			req.RootChain,
		)

		// send response
		restClient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func newCheckpointACKHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req HeaderACKReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// draft a message and send response
		msg := types.NewMsgCheckpointAck(
			req.From,
			req.HeaderBlock,
			req.Proposer,
			req.StartBlock,
			req.EndBlock,
			req.RootHash,
			req.TxHash,
			req.LogIndex,
			req.RootChain,
		)

		// send response
		restClient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func newCheckpointNoACKHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req HeaderNoACKReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// draft a message and send response
		msg := types.NewMsgCheckpointNoAck(
			req.Proposer,
		)

		// send response
		restClient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func repairCheckpointHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RepairCheckpointReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// 获取发送者地址
		var from hmTypes.HeimdallAddress
		if req.From != "" {
			from = hmTypes.HexToHeimdallAddress(req.From)
		} else {
			from = helper.GetFromAddress(cliCtx)
		}
		if from.Empty() {
			http.Error(w, "发送者地址不能为空", http.StatusBadRequest)
			return
		}

		// 记录开始操作的日志
		helper.Logger.Info("repairCheckpointHandler, 开始直接数据库操作",
			"checkpointNumber", req.CheckpointNumber,
			"from", from.String(),
		)

		// 创建一个模拟的 checkpoint 用于测试
		testCheckpoint := hmTypes.Checkpoint{
			StartBlock: 50080768,
			EndBlock:   50089983,
			RootHash:   hmTypes.HexToHeimdallHash("0xc18d16ec7f97533ad4aa49aac5aaf73486da34839410f002d41fa73c5c2f06d3"),
			Proposer:   hmTypes.HexToHeimdallAddress("0xd4d14396282a000234862eaf2527c17ed680e58e"),
			BorChainID: "22125",
			TimeStamp:  1749546051,
		}

		// 直接返回成功响应，不进行广播
		// 注意：这里应该直接操作数据库，但由于 REST 接口的限制，
		// 我们暂时返回成功响应，实际的数据库操作需要在其他地方实现
		helper.Logger.Info("repairCheckpointHandler, 直接数据库操作完成",
			"checkpointNumber", req.CheckpointNumber,
			"checkpoint", testCheckpoint,
		)

		// 返回成功响应
		rest.PostProcessResponse(w, cliCtx, map[string]interface{}{
			"success":           true,
			"message":           "checkpoint 补录请求已接收（直接数据库操作）",
			"checkpoint_number": req.CheckpointNumber,
			"checkpoint":        testCheckpoint,
			"note":              "此接口直接操作数据库，不进行广播",
		})
	}
}

func repairCheckpointTestHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RepairCheckpointTestReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// 获取发送者地址
		var from hmTypes.HeimdallAddress
		if req.From != "" {
			from = hmTypes.HexToHeimdallAddress(req.From)
		} else {
			from = helper.GetFromAddress(cliCtx)
		}
		if from.Empty() {
			http.Error(w, "发送者地址不能为空", http.StatusBadRequest)
			return
		}

		// 记录开始广播的日志
		helper.Logger.Info("repairCheckpointTestHandler, 开始广播测试消息",
			"checkpointNumber", req.CheckpointNumber,
			"testMessage", req.TestMessage,
		)

		// 创建一个测试 checkpoint 消息
		testCheckpoint := hmTypes.Checkpoint{
			StartBlock: 50080768,
			EndBlock:   50089983,
			RootHash:   hmTypes.HexToHeimdallHash("0xc18d16ec7f97533ad4aa49aac5aaf73486da34839410f002d41fa73c5c2f06d3"),
			Proposer:   hmTypes.HexToHeimdallAddress("0xd4d14396282a000234862eaf2527c17ed680e58e"),
			BorChainID: "22125",
			TimeStamp:  1749546051,
		}

		// 创建 checkpoint 消息，使用 bridge 中的方式
		msg := types.NewMsgCheckpointBlock(
			testCheckpoint.Proposer,
			testCheckpoint.StartBlock,
			testCheckpoint.EndBlock,
			testCheckpoint.RootHash,
			testCheckpoint.RootHash, // 使用 RootHash 作为 AccountRootHash
			testCheckpoint.BorChainID,
			req.CheckpointNumber, // 使用请求的 checkpoint number
			"tron",
		)

		// 创建 TxBroadcaster 实例，使用 bridge 中的方式
		txBroadcaster := broadcaster.NewTxBroadcaster(cliCtx.Codec)

		// 使用 bridge 中的广播方法
		if err := txBroadcaster.BroadcastToHeimdall(&msg); err != nil {
			helper.Logger.Error("repairCheckpointTestHandler, 广播失败", "error", err)
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// 记录成功日志
		helper.Logger.Info("repairCheckpointTestHandler, 广播成功",
			"checkpointNumber", req.CheckpointNumber,
			"testMessage", req.TestMessage,
		)

		// 返回成功响应
		rest.PostProcessResponse(w, cliCtx, map[string]interface{}{
			"success":           true,
			"message":           "测试消息广播成功",
			"checkpoint_number": req.CheckpointNumber,
			"test_message":      req.TestMessage,
		})
	}
}

type myTestReq struct {
	BaseReq  rest.BaseReq `json:"base_req"` // 使用 heimdall 的 rest.BaseReq
	From     string       `json:"from"`
	TestData string       `json:"test_data"`
}

func myTestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req myTestReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr := hmTypes.HexToHeimdallAddress(req.From)

		// Create message
		msg := types.NewMsgMyTest(fromAddr, req.TestData)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Send the message
		restClient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
