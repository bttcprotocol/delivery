package featuremanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/featuremanager/types"
)

func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, keeper, msg)

		default:
			return sdk.ErrTxDecode("Invalid message in checkpoint module").Result()
		}
	}
}

// handleMsgSubmitProposal will check the proposal format,
// the proposal content should be checked in side_handler.
func handleMsgSubmitProposal(ctx sdk.Context, _ Keeper, msg types.MsgSubmitProposal) sdk.Result {
	// check content format
	if err := msg.ValidateBasic(); err != nil {
		return err.Result()
	}

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
