package types

// Checkpoint tags
var (
	EventTypeCheckpoint           = "checkpoint"
	EventTypeCheckpointAck        = "checkpoint-ack"
	EventTypeCheckpointNoAck      = "checkpoint-noack"
	EventTypeCheckpointSync       = "checkpoint-sync"
	EventTypeCheckpointSyncAck    = "checkpoint-sync-ack"
	EventTypeRepairCheckpoint     = "repair-checkpoint"
	EventTypeRepairCheckpointTest = "repair-checkpoint-test"

	AttributeKeyProposer         = "proposer"
	AttributeKeyStartBlock       = "start-block"
	AttributeKeyEndBlock         = "end-block"
	AttributeKeyHeaderIndex      = "header-index"
	AttributeKeyNewProposer      = "new-proposer"
	AttributeKeyRootHash         = "root-hash"
	AttributeKeyAccountHash      = "account-hash"
	AttributeKeyRootChain        = "root-chain"
	AttributeKeyTestMessage      = "test_message"
	AttributeKeyCheckpointNumber = "checkpoint_number"

	AttributeValueCategory = ModuleName
)
