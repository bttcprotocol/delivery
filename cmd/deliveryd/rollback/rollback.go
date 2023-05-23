package rollback

import (
	"fmt"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/maticnetwork/heimdall/helper"
	stakingcli "github.com/maticnetwork/heimdall/staking/client/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/version"
	db "github.com/tendermint/tm-db"
)

// nolint: revive
func RollbackCmd(ctx *server.Context) *cobra.Command {
	// nolint: exhaustivestruct
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "rollback cosmos-sdk and tendermint state by one height",
		Long: `
A state rollback is performed to recover from an incorrect application state transition,
when Tendermint has persisted an incorrect app hash and is thus unable to make
progress. Rollback overwrites a state at height n with the state at height n - 1.
The application also roll back to height n - 1. No blocks are removed, so upon
restarting Tendermint the transactions in block n will be re-executed against the
application.
`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			height, hash, err := RollbackState(config)
			if err != nil {
				return fmt.Errorf("failed to rollback tendermint state: %w", err)
			}
			// nolint: forbidigo
			fmt.Printf("Rolled back state to height %d and hash %X\n", height, hash)

			return nil
		},
	}

	cmd.Flags().String(cli.HomeFlag, helper.DefaultNodeHome, "node's home directory")
	cmd.Flags().String(helper.FlagClientHome, helper.DefaultCLIHome, "client's home directory")
	cmd.Flags().String(client.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().Int(stakingcli.FlagValidatorID, 1, "--id=<validator ID here>, if left blank will be assigned 1")

	return cmd
}

// RollbackState takes the state at the current height n and overwrites it with the state
// at height n - 1. Note state here refers to tendermint state not application state.
// Returns the latest state height and app hash alongside an error if there was one.
// nolint: revive
func RollbackState(config *cfg.Config) (int64, []byte, error) {
	// use the parsed config to load the block and state store
	blocksdb, statedb, multistore, evidencedb, err := loadStore(config)
	if err != nil {
		return -1, nil, err
	}

	defer func() {
		blocksdb.Close()
		statedb.Close()
		multistore.Close()
		evidencedb.Close()
	}()

	// rollback the last block and state
	height, hash, err := rollbackState(blocksdb, statedb)
	if err != nil {
		return height, hash, err
	}

	// rollback the multistore
	cms := NewMultiStore(multistore)
	cms.MountKeys()
	_ = cms.LoadLatestVersion()
	cms.RollbackToVersion(height)

	// rollback evidencestore
	es := NewEvidenceStore(evidencedb)
	es.RollbackToVersion(height)

	return height, hash, nil
}

// load blockstore, statestore, multistore and evidencestore
// nolint: goerr113
func loadStore(cfg *cfg.Config) (db.DB, db.DB, db.DB, db.DB, error) {
	dbType := db.DBBackendType(cfg.DBBackend)

	// Get BlockStore
	if !os.FileExists(filepath.Join(cfg.DBDir(), "blockstore.db")) {
		return nil, nil, nil, nil, fmt.Errorf("no blockstore found in %v", cfg.DBDir())
	}

	blockStoreDB := db.NewDB("blockstore", dbType, cfg.DBDir())

	// Get StateStore
	if !os.FileExists(filepath.Join(cfg.DBDir(), "state.db")) {
		return nil, nil, nil, nil, fmt.Errorf("no blockstore found in %v", cfg.DBDir())
	}

	stateDB := db.NewDB("state", dbType, cfg.DBDir())

	// Get MultiStore
	if !os.FileExists(filepath.Join(cfg.DBDir(), "application.db")) {
		return nil, nil, nil, nil, fmt.Errorf("no blockstore found in %v", cfg.DBDir())
	}

	multiStoreDB := db.NewDB("application", dbType, cfg.DBDir())

	// Get EvidenceStore
	if !os.FileExists(filepath.Join(cfg.DBDir(), "evidence.db")) {
		return nil, nil, nil, nil, fmt.Errorf("no blockstore found in %v", cfg.DBDir())
	}

	evidenceStoreDB := db.NewDB("evidence", dbType, cfg.DBDir())

	return blockStoreDB, stateDB, multiStoreDB, evidenceStoreDB, nil
}

// invalidState.LastBlockHeight - 1
// nolint: goerr113
func rollbackState(blockStoreDB, stateDB db.DB) (int64, []byte, error) {
	blockStore := NewBlockStore(blockStoreDB)
	invalidState := state.LoadState(stateDB)

	// rollback to target height
	rollbackHeight := invalidState.LastBlockHeight - 1

	rollbackBlock := blockStore.LoadBlockMeta(rollbackHeight)
	if rollbackBlock == nil {
		return -1, nil, fmt.Errorf("block at height %d not found", rollbackHeight)
	}

	// we also need to retrieve the latest block because the app hash and
	// last results hash is only agreed upon in the following block
	latestBlock := blockStore.LoadBlockMeta(invalidState.LastBlockHeight)
	if latestBlock == nil {
		return -1, nil, fmt.Errorf("block at height %d not found", rollbackHeight)
	}

	// previous validators
	previousLastValidatorSet, err := state.LoadValidators(stateDB, rollbackHeight)
	if err != nil {
		return -1, nil, err
	}

	previousParams, err := state.LoadConsensusParams(stateDB, rollbackHeight)
	if err != nil {
		return -1, nil, err
	}

	valChangeHeight := invalidState.LastHeightValidatorsChanged
	// this can only happen if the validator set changed since the last block
	if valChangeHeight > rollbackHeight {
		valChangeHeight = rollbackHeight + 1
	}

	paramsChangeHeight := invalidState.LastHeightConsensusParamsChanged
	// this can only happen if params changed from the last block
	if paramsChangeHeight > rollbackHeight {
		paramsChangeHeight = rollbackHeight + 1
	}

	// build the new state from the old state and the prior block
	// nolint: exhaustivestruct
	rolledBackState := state.State{
		Version: state.Version{
			Consensus: version.Consensus{
				Block: version.BlockProtocol,
				App:   0,
			},
			Software: version.TMCoreSemVer,
		},
		// immutable fields
		ChainID: invalidState.ChainID,

		LastBlockHeight:  rollbackBlock.Header.Height,
		LastBlockTotalTx: rollbackBlock.Header.TotalTxs,
		LastBlockID:      rollbackBlock.BlockID,
		LastBlockTime:    rollbackBlock.Header.Time,

		NextValidators:              invalidState.Validators,
		Validators:                  invalidState.LastValidators,
		LastValidators:              previousLastValidatorSet,
		LastHeightValidatorsChanged: valChangeHeight,

		ConsensusParams:                  previousParams,
		LastHeightConsensusParamsChanged: paramsChangeHeight,

		LastResultsHash: latestBlock.Header.LastResultsHash,
		AppHash:         latestBlock.Header.AppHash,
	}

	// saving the state
	state.SaveState(stateDB, rolledBackState)

	// prune blocks
	_, _ = blockStore.PruneBlocks(rollbackHeight, blockStore.Height())

	// saving blockStore block height
	bsj := LoadBlockStoreStateJSON(blockStoreDB)
	bsj.Height = rollbackHeight
	bsj.Save(blockStoreDB)

	return rolledBackState.LastBlockHeight, rolledBackState.AppHash, nil
}
