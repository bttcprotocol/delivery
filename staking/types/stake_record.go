package types

import (
	"fmt"

	hmTypes "github.com/maticnetwork/heimdall/types"
	hmtypes "github.com/maticnetwork/heimdall/types"
)

// StakingRecord struct
type StakingRecord struct {
	Type        string               `json:"type"`
	ValidatorID hmTypes.ValidatorID  `json:"id"`
	Nonce       uint64               `json:"nonce"`
	Height      int64                `json:"height"`
	TxHash      hmtypes.HeimdallHash `json:"tx_hash"`
	TimeStamp   uint64               `json:"timestamp"`
}

// String returns human readable string
func (s StakingRecord) String() string {
	return fmt.Sprintf(
		"StakingRecord {%v %v %v %v %v %v}",
		s.Type,
		s.ValidatorID,
		s.Nonce,
		s.Height,
		s.TxHash.Hex(),
		s.TimeStamp,
	)
}
