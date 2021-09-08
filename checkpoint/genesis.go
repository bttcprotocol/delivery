package checkpoint

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/maticnetwork/heimdall/checkpoint/types"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	keeper.SetParams(ctx, data.Params)

	// Set last no-ack
	if data.LastNoACK > 0 {
		keeper.SetLastNoAck(ctx, data.LastNoACK)
	}

	// Add finalised checkpoints to state
	if len(data.Checkpoints) != 0 {
		// check if we are provided all the headers
		if int(data.AckCount) != len(data.Checkpoints) {
			panic(errors.New("Incorrect state in state-dump , Please Check "))
		}
		// sort headers before loading to state
		data.Checkpoints = hmTypes.SortHeaders(data.Checkpoints)
		// load checkpoints to state
		for i, checkpoint := range data.Checkpoints {
			checkpointIndex := uint64(i) + 1
			if err := keeper.AddCheckpoint(ctx, checkpointIndex, checkpoint, hmTypes.RootChainTypeEth); err != nil {
				keeper.Logger(ctx).Error("InitGenesis | AddCheckpoint", "error", err)
			}
		}
	}

	// Add checkpoint in buffer
	if data.BufferedCheckpoint != nil {
		if err := keeper.SetCheckpointBuffer(ctx, *data.BufferedCheckpoint, hmTypes.RootChainTypeEth); err != nil {
			keeper.Logger(ctx).Error("InitGenesis | SetCheckpointBuffer", "error", err)
		}
	}

	// Set initial ack count
	keeper.UpdateACKCountWithValue(ctx, data.AckCount, hmTypes.RootChainTypeEth)

	// Add finalised checkpoints to state
	if len(data.TronCheckpoints) != 0 {
		// check if we are provided all the headers
		if int(data.TronAckCount) != len(data.TronCheckpoints) {
			panic(errors.New("Incorrect state in state-dump , Please Check "))
		}
		// sort headers before loading to state
		data.TronCheckpoints = hmTypes.SortHeaders(data.TronCheckpoints)
		// load checkpoints to state
		for i, checkpoint := range data.TronCheckpoints {
			checkpointIndex := uint64(i) + 1
			if err := keeper.AddCheckpoint(ctx, checkpointIndex, checkpoint, hmTypes.RootChainTypeTron); err != nil {
				keeper.Logger(ctx).Error("InitGenesis | TronAddCheckpoint", "error", err)
			}
		}
	}
	keeper.UpdateACKCountWithValue(ctx, data.TronAckCount, hmTypes.RootChainTypeTron)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	params := keeper.GetParams(ctx)

	bufferedCheckpoint, _ := keeper.GetCheckpointFromBuffer(ctx, hmTypes.RootChainTypeEth)
	return types.NewGenesisState(
		params,
		bufferedCheckpoint,
		keeper.GetLastNoAck(ctx),
		keeper.GetACKCount(ctx, hmTypes.RootChainTypeEth),
		hmTypes.SortHeaders(keeper.GetCheckpoints(ctx)),
		keeper.GetACKCount(ctx, hmTypes.RootChainTypeTron),
		hmTypes.SortHeaders(keeper.GetOtherCheckpoints(ctx, hmTypes.RootChainTypeTron)),
	)
}
