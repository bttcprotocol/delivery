package chainmanager

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/chainmanager/types"
	hmCommon "github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/helper"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// NewHandler new handler
func NewHandler(k Keeper, contractCaller helper.IContractCaller) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgNewChain:
			return HandleMsgNewChain(ctx, msg, k, contractCaller)

		default:
			return sdk.ErrTxDecode("Invalid message in staking module").Result()
		}
	}
}

// HandleMsgNewChain msg validator join
func HandleMsgNewChain(ctx sdk.Context, msg types.MsgNewChain, k Keeper, contractCaller helper.IContractCaller) sdk.Result {

	k.Logger(ctx).Debug("✅ Validating new chain msg", "msg", msg)

	if hmTypes.GetRootChainID(msg.RootChainType) == 0 {
		k.Logger(ctx).Error("Wrong root chain type", "root", msg.RootChainType)
		return hmCommon.ErrWrongRootChain(k.Codespace()).Result()
	}
	// get chain params from store
	_, err := k.GetChainParams(ctx, msg.RootChainType)
	if err == nil {
		return hmCommon.ErrChainPamramsExist(k.Codespace()).Result()
	}

	// Emit event join
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeNewChain,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyRootChain, msg.RootChainType),
			sdk.NewAttribute(types.AttributeKeyActivationHeight, strconv.FormatUint(msg.ActivationHeight, 10)),
		),
	})

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
