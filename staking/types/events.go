package types

// Checkpoint tags
var (
	EventTypeNewProposer    = "new-proposer"
	EventTypeValidatorJoin  = "validator-join"
	EventTypeSignerUpdate   = "signer-update"
	EventTypeStakeUpdate    = "stake-update"
	EventTypeValidatorExit  = "validator-exit"
	EventTypeStakingSync    = "staking-sync"
	EventTypeStakingSyncAck = "staking-ack"

	AttributeKeySigner            = "signer"
	AttributeKeyDeactivationEpoch = "deactivation-epoch"
	AttributeKeyActivationEpoch   = "activation-epoch"
	AttributeKeyValidatorID       = "validator-id"
	AttributeKeyValidatorNonce    = "validator-nonce"
	AttributeKeyUpdatedAt         = "updated-at"
	AttributeKeyRootChain         = "root-chain"

	AttributeValueCategory = ModuleName
)
