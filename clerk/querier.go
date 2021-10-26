package clerk

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethTypes "github.com/maticnetwork/bor/core/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/maticnetwork/heimdall/clerk/types"
	"github.com/maticnetwork/heimdall/helper"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// NewQuerier creates a querier for auth REST endpoints
func NewQuerier(keeper Keeper, contractCaller helper.IContractCaller) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryRecord:
			return handleQueryRecord(ctx, req, keeper)
		case types.QueryRecordList:
			return handleQueryRecordList(ctx, req, keeper)
		case types.QueryRecordListWithTime:
			return handleQueryRecordListWithTime(ctx, req, keeper)
		case types.QueryRecordSequence:
			return handleQueryRecordSequence(ctx, req, keeper, contractCaller)
		default:
			return nil, sdk.ErrUnknownRequest("unknown auth query endpoint")
		}
	}
}

// root chain: query root chain record with root chain record id
func handleQueryRecord(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryRecordParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	// get state record by record id
	record, err := keeper.GetEventRecord(ctx, params.RecordID)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not get state record", err.Error()))
	}

	// json record
	bz, err := json.Marshal(record)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// bor chain use only with rearrange id
func handleQueryRecordList(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params hmTypes.QueryPaginationParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	res, err := keeper.GetEventRecordList(ctx, params.Page, params.Limit)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr(fmt.Sprintf("could not fetch record list with page %v and limit %v", params.Page, params.Limit), err.Error()))
	}

	bz, err := json.Marshal(res)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// bor only with rearrange id
func handleQueryRecordListWithTime(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryRecordTimePaginationParams
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	res, err := keeper.GetEventRecordListWithTime(ctx, params.FromTime, params.ToTime, params.Page, params.Limit)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr(fmt.Sprintf("could not fetch record list with fromTime %v and toTime %v", params.FromTime, params.ToTime), err.Error()))
	}

	bz, err := json.Marshal(res)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// root chain
func handleQueryRecordSequence(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, contractCallerObj helper.IContractCaller) ([]byte, sdk.Error) {
	var params types.QueryRecordSequenceParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	chainParams := keeper.chainKeeper.GetParams(ctx)
	var receipt *ethTypes.Receipt
	var err error
	// get main tx receipt
	switch params.RootChainType {
	case hmTypes.RootChainTypeEth:
		receipt, err = contractCallerObj.GetConfirmedTxReceipt(hmTypes.HexToHeimdallHash(params.TxHash).EthHash(),
			chainParams.MainchainTxConfirmations, hmTypes.RootChainTypeEth)
	case hmTypes.RootChainTypeBsc:
		bscChain, err := keeper.chainKeeper.GetChainParams(ctx, hmTypes.RootChainTypeBsc)
		if err != nil {
			return nil, sdk.ErrInternal(fmt.Sprintf("wrong chain type = " + params.RootChainType + "plealse pass correct chainType like bsc"))
		}
		receipt, err = contractCallerObj.GetConfirmedTxReceipt(hmTypes.HexToHeimdallHash(params.TxHash).EthHash(),
			bscChain.TxConfirmations, hmTypes.RootChainTypeBsc)
	case hmTypes.RootChainTypeTron:
		receipt, err = contractCallerObj.GetTronTransactionReceipt(hmTypes.HexToHeimdallHash(params.TxHash).TronHash().Hex())
	default:
		return nil, sdk.ErrInternal(fmt.Sprintf("wrong chain type = " + params.RootChainType + "please pass correct chainType like eth or tron"))
	}
	if err != nil || receipt == nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("Transaction is not confirmed yet. Please wait for sometime and try again"))
	}

	// sequence id
	sequence := helper.CalculateSequence(receipt.BlockNumber, params.LogIndex, params.RootChainType)
	// check if incoming tx already exists
	if !keeper.HasRecordSequence(ctx, sequence.String()) {
		return nil, nil
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, sequence)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
