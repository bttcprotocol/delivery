package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	featuremanagerTypes "github.com/maticnetwork/heimdall/featuremanager/types"
	govtypes "github.com/maticnetwork/heimdall/gov/types"
	"github.com/maticnetwork/heimdall/params/types"
)

// NewParamChangeProposalHandler new param changes proposal handler
func NewParamChangeProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Error {
		switch c := content.(type) {
		case types.ParameterChangeProposal:
			return handleParameterChangeProposal(ctx, k, c)

		case featuremanagerTypes.FeatureChangeProposal:
			return handleFeatureChangeProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized param proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}

func handleParameterChangeProposal(ctx sdk.Context, k Keeper, p types.ParameterChangeProposal) sdk.Error {
	for _, c := range p.Changes {
		// featuremanagerModule should be changed via handleFeatureChangeProposal.
		if c.Subspace == featuremanagerTypes.ModuleName {
			continue
		}

		ss, ok := k.GetSubspace(c.Subspace)
		if !ok {
			return types.ErrUnknownSubspace(k.codespace, c.Subspace)
		}

		k.Logger(ctx).Info(
			fmt.Sprintf("setting new parameter; key: %s, value: %s", c.Key, c.Value),
		)

		if err := ss.Update(ctx, []byte(c.Key), []byte(c.Value)); err != nil {
			return types.ErrSettingParameter(k.codespace, c.Key, c.Value, err.Error())
		}
	}

	return nil
}

func handleFeatureChangeProposal(ctx sdk.Context, keeper Keeper,
	p featuremanagerTypes.FeatureChangeProposal,
) sdk.Error {
	for _, change := range p.Changes {
		subspace, ok := keeper.GetSubspace(featuremanagerTypes.ModuleName)
		if !ok {
			return types.ErrUnknownSubspace(keeper.codespace, featuremanagerTypes.ModuleName)
		}

		keeper.Logger(ctx).Info(
			fmt.Sprintf("update feature; value: %s", change.Value),
		)

		if err := subspace.Update(ctx, featuremanagerTypes.KeyFeatureParams, []byte(change.Value)); err != nil {
			return types.ErrSettingParameter(keeper.codespace,
				string(featuremanagerTypes.KeyFeatureParams), change.Value, err.Error())
		}
	}

	return nil
}
