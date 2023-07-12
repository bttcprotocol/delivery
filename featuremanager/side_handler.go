package featuremanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/featuremanager/types"
	"github.com/maticnetwork/heimdall/gov"
	govTypes "github.com/maticnetwork/heimdall/gov/types"
	hmTypes "github.com/maticnetwork/heimdall/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewSideTxHandler(keeper Keeper) hmTypes.SideTxHandler {
	return func(ctx sdk.Context, msg sdk.Msg) abci.ResponseDeliverSideTx {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgSubmitProposal:
			return SideHandleMsgSubmitProposal(ctx, keeper, msg)

		default:
			//nolint: exhaustivestruct
			return abci.ResponseDeliverSideTx{
				Code: uint32(sdk.CodeUnknownRequest),
			}
		}
	}
}

// SideHandleMsgSubmitProposal will check whether the node support the target feature.
func SideHandleMsgSubmitProposal(
	ctx sdk.Context, keeper Keeper,
	msg types.MsgSubmitProposal,
) (result abci.ResponseDeliverSideTx) {
	if err := msg.ValidateBasic(); err != nil {
		result.Result = abci.SideTxResultType_No

		return
	}

	if _, err := gov.GetValidValidator(ctx, keeper.govKeeper, msg.Proposer, msg.Validator); err != nil {
		result.Result = abci.SideTxResultType_No

		return
	}

	result.Result = abci.SideTxResultType_Yes

	switch content := msg.Content.(type) {
	case types.FeatureChangeProposal:
	Changes:
		for _, change := range content.Changes {
			chagneMapStr := change.Value

			var changeMap types.FeatureParams
			err := keeper.cdc.UnmarshalJSON([]byte(chagneMapStr), &changeMap)
			if err != nil {
				result.Result = abci.SideTxResultType_No

				break Changes
			}

			for key := range changeMap.FeatureParamMap {
				if !keeper.HasFeature(key) {
					result.Result = abci.SideTxResultType_No

					break Changes
				}
			}
		}

	default:
		result.Result = abci.SideTxResultType_No
	}

	return result
}

func NewPostTxHandler(keeper Keeper) hmTypes.PostTxHandler {
	return func(ctx sdk.Context, msg sdk.Msg, sideTxResult abci.SideTxResultType) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgSubmitProposal:
			return PostHandleMsgSubmitProposal(ctx, keeper, msg, sideTxResult)

		default:
			return sdk.ErrUnknownRequest("Unrecognized checkpoint Msg type").Result()
		}
	}
}

// PostHandleMsgSubmitProposal submit a proposal to gov to begin voting period.
func PostHandleMsgSubmitProposal(ctx sdk.Context, keeper Keeper,
	msg types.MsgSubmitProposal,
	sideTxResult abci.SideTxResultType,
) sdk.Result {
	logger := keeper.Logger(ctx)

	// skip handler if feature is not approved.
	if sideTxResult != abci.SideTxResultType_Yes {
		logger.Info("Skipping new feature proposal since side-tx didn't get yes votes")

		return common.ErrSideTxValidation(keeper.Codespace()).Result()
	}

	var err error
	// add all supported features to featuremanager.
	if content, ok := msg.Content.(types.FeatureChangeProposal); ok {
	Changes:
		for _, change := range content.Changes {
			chagneMapStr := change.Value

			var changeMap types.FeatureParams
			_ = keeper.cdc.UnmarshalJSON([]byte(chagneMapStr), &changeMap)

			for key := range changeMap.FeatureParamMap {
				if err = keeper.AddSupportedFeature(ctx, key); err != nil {
					break Changes
				}
			}
		}
	} else {
		logger.Error("Proposal is not a feature change proposal.")

		return common.ErrSideTxValidation(keeper.Codespace()).Result()
	}

	// if meets any error, cancel transfer this proposal to gov.
	if err != nil {
		logger.Error("Save supported features failed.", "error", err)

		return common.ErrSideTxValidation(keeper.Codespace()).Result()
	}

	// submit a feature-change-proposal to gov
	logger.Info("Submitting feature-change-proposal to gov", "msg", msg)

	proposal := govTypes.MsgSubmitProposal{
		Content:        msg.Content,
		InitialDeposit: msg.InitialDeposit,
		Proposer:       msg.Proposer,
		Validator:      msg.Validator,
	}

	return gov.HandleMsgSubmitProposal(ctx, keeper.govKeeper, proposal)
}
