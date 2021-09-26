package types

import (
	"encoding/json"
)

//
// Gensis state
//

// GenesisState - all chainmanager state that must be provided at genesis
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`

	ChainInfos []ChainInfo `json:"chain_infos" yaml:"chain_infos"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params, chainInfos []ChainInfo) GenesisState {
	return GenesisState{
		Params:     params,
		ChainInfos: chainInfos,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultParams(), []ChainInfo{})
}

// ValidateGenesis performs basic validation of auth genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return nil
}

// GetGenesisStateFromAppState returns staking GenesisState given raw application genesis state
func GetGenesisStateFromAppState(appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		ModuleCdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}
	return genesisState
}
