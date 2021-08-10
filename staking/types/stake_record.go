package types

import (
	"fmt"

	hmTypes "github.com/maticnetwork/heimdall/types"
	hmtypes "github.com/maticnetwork/heimdall/types"
)

// StakingRecord struct
type StakingRecord struct {
	ValidatorID hmTypes.ValidatorID  `json:"id"`
	TxHash      hmtypes.HeimdallHash `json:"tx_hash"`
	Nonce       uint64               `json:"nonce"`
	TimeStamp   uint64               `json:"timestamp"`
}

// String returns human readable string
func (s StakingRecord) String() string {
	return fmt.Sprintf(
		"StakingRecord {%v %v %v %v}",
		s.ValidatorID,
		s.Nonce,
		s.TxHash.Hex(),
		s.TimeStamp,
	)
}
