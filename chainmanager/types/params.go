package types

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/maticnetwork/heimdall/helper"
	"github.com/maticnetwork/heimdall/params/subspace"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// Default parameter values
const (
	DefaultMainchainTxConfirmations  uint64 = 6
	DefaultMaticchainTxConfirmations uint64 = 10
	DefaultTronChainTxConfirmations  uint64 = 18
)

var (
	DefaultStateReceiverAddress hmTypes.HeimdallAddress = hmTypes.HexToHeimdallAddress("0x0000000000000000000000000000000000001001")
	DefaultValidatorSetAddress  hmTypes.HeimdallAddress = hmTypes.HexToHeimdallAddress("0x0000000000000000000000000000000000001000")
)

// Parameter keys
var (
	KeyMainchainTxConfirmations  = []byte("MainchainTxConfirmations")
	KeyMaticchainTxConfirmations = []byte("MaticchainTxConfirmations")
	KeyTronchainTxConfirmations  = []byte("TronchainTxConfirmations")
	KeyChainParams               = []byte("ChainParams")
	KeyParamsWithMultiChains     = []byte("ParamsWithMultiChains")
)

var _ subspace.ParamSet = &Params{}

// ChainParams chain related params
type ChainParams struct {
	BorChainID        string                  `json:"bor_chain_id" yaml:"bor_chain_id"`
	MaticTokenAddress hmTypes.HeimdallAddress `json:"matic_token_address" yaml:"matic_token_address"`

	// eth
	StakingManagerAddress hmTypes.HeimdallAddress `json:"staking_manager_address" yaml:"staking_manager_address"`
	SlashManagerAddress   hmTypes.HeimdallAddress `json:"slash_manager_address" yaml:"slash_manager_address"`
	RootChainAddress      hmTypes.HeimdallAddress `json:"root_chain_address" yaml:"root_chain_address"`
	StakingInfoAddress    hmTypes.HeimdallAddress `json:"staking_info_address" yaml:"staking_info_address"`
	StateSenderAddress    hmTypes.HeimdallAddress `json:"state_sender_address" yaml:"state_sender_address"`

	// tron
	TronChainAddress          string `json:"tron_chain_address" yaml:"tron_chain_address"`
	TronStateSenderAddress    string `json:"tron_state_sender_address" yaml:"tron_state_sender_address"`
	TronStakingManagerAddress string `json:"tron_staking_manager_address" yaml:"tron_staking_manager_address"`
	TronStakingInfoAddress    string `json:"tron_state_info_address" yaml:"tron_state_info_address"`

	// Bor Chain Contracts
	StateReceiverAddress hmTypes.HeimdallAddress `json:"state_receiver_address" yaml:"state_receiver_address"`
	ValidatorSetAddress  hmTypes.HeimdallAddress `json:"validator_set_address" yaml:"validator_set_address"`
}

func (cp ChainParams) String() string {
	return fmt.Sprintf(`
	BorChainID: 									%s
  MaticTokenAddress:            %s
	StakingManagerAddress:        %s
	SlashManagerAddress:        %s
	RootChainAddress:             %s
  StakingInfoAddress:           %s
	StateSenderAddress:           %s
	StateReceiverAddress: 				%s
	ValidatorSetAddress:					%s`,
		cp.BorChainID, cp.MaticTokenAddress, cp.StakingManagerAddress, cp.SlashManagerAddress, cp.RootChainAddress, cp.StakingInfoAddress, cp.StateSenderAddress, cp.StateReceiverAddress, cp.ValidatorSetAddress)
}

// Params defines the parameters for the chainmanager module.
type Params struct {
	MainchainTxConfirmations  uint64      `json:"mainchain_tx_confirmations" yaml:"mainchain_tx_confirmations"`
	MaticchainTxConfirmations uint64      `json:"maticchain_tx_confirmations" yaml:"maticchain_tx_confirmations"`
	TronchainTxConfirmations  uint64      `json:"tronchain_tx_confirmations" yaml:"tronchain_tx_confirmations"`
	ChainParams               ChainParams `json:"chain_params" yaml:"chain_params"`
}

// NewParams creates a new Params object
func NewParams(mainchainTxConfirmations, tronchainTxConfirmation, maticchainTxConfirmations uint64, chainParams ChainParams) Params {
	return Params{
		MainchainTxConfirmations:  mainchainTxConfirmations,
		MaticchainTxConfirmations: maticchainTxConfirmations,
		TronchainTxConfirmations:  tronchainTxConfirmation,
		ChainParams:               chainParams,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{KeyMainchainTxConfirmations, &p.MainchainTxConfirmations},
		{KeyMaticchainTxConfirmations, &p.MaticchainTxConfirmations},
		{KeyTronchainTxConfirmations, &p.TronchainTxConfirmations},
		{KeyChainParams, &p.ChainParams},
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("MainchainTxConfirmations: %d\n", p.MainchainTxConfirmations))
	sb.WriteString(fmt.Sprintf("MaticchainTxConfirmations: %d\n", p.MaticchainTxConfirmations))
	sb.WriteString(fmt.Sprintf("ChainParams: %s\n", p.ChainParams.String()))
	return sb.String()
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateHeimdallAddress("matic_token_address", p.ChainParams.MaticTokenAddress); err != nil {
		return err
	}

	if err := validateHeimdallAddress("staking_manager_address", p.ChainParams.StakingManagerAddress); err != nil {
		return err
	}

	if err := validateHeimdallAddress("slash_manager_address", p.ChainParams.SlashManagerAddress); err != nil {
		return err
	}

	if err := validateHeimdallAddress("root_chain_address", p.ChainParams.RootChainAddress); err != nil {
		return err
	}

	if err := validateHeimdallAddress("staking_info_address", p.ChainParams.StakingInfoAddress); err != nil {
		return err
	}

	if err := validateHeimdallAddress("state_sender_address", p.ChainParams.StateSenderAddress); err != nil {
		return err
	}

	if err := validateHeimdallAddress("state_receiver_address", p.ChainParams.StateReceiverAddress); err != nil {
		return err
	}

	if err := validateHeimdallAddress("validator_set_address", p.ChainParams.ValidatorSetAddress); err != nil {
		return err
	}

	return nil
}

func validateHeimdallAddress(key string, value hmTypes.HeimdallAddress) error {
	if value.String() == "" {
		return fmt.Errorf("Invalid value %s in chain_params", key)
	}

	return nil
}

//
// Extra functions
//

// ParamKeyTable for auth module
func ParamKeyTable() subspace.KeyTable {
	//nolint: exhaustivestruct
	return subspace.NewKeyTable().RegisterParamSet(&Params{}).RegisterParamSet(&ParamsWithMultiChains{})
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MainchainTxConfirmations:  DefaultMainchainTxConfirmations,
		MaticchainTxConfirmations: DefaultMaticchainTxConfirmations,
		TronchainTxConfirmations:  DefaultTronChainTxConfirmations,
		ChainParams: ChainParams{
			BorChainID:           helper.DefaultBttcChainID,
			StateReceiverAddress: DefaultStateReceiverAddress,
			ValidatorSetAddress:  DefaultValidatorSetAddress,
		},
	}
}

// ----------------------------------------------------------------
// ParamsWithMultiChains

//nolint: tagliatelle
type ChainData struct {
	TxConfirmations *uint64 `json:"tx_confirmations" yaml:"tx_confirmations"`
	ActivateHeight  *uint64 `json:"activate_height" yaml:"activate_height"`

	// main chain
	StakingManagerAddress *hmTypes.HeimdallAddress `json:"staking_manager_address" yaml:"staking_manager_address"`
	SlashManagerAddress   *hmTypes.HeimdallAddress `json:"slash_manager_address" yaml:"slash_manager_address"`
	RootChainAddress      *hmTypes.HeimdallAddress `json:"root_chain_address" yaml:"root_chain_address"`
	StakingInfoAddress    *hmTypes.HeimdallAddress `json:"staking_info_address" yaml:"staking_info_address"`
	StateSenderAddress    *hmTypes.HeimdallAddress `json:"state_sender_address" yaml:"state_sender_address"`
}

//nolint: tagliatelle
type ParamsWithMultiChains struct {
	ChainParameterMap map[string]ChainData `json:"chain_parameter_map" yaml:"chainParameterMap"`
}

//nolint: tagliatelle
type PlainChainData struct {
	TxConfirmations uint64 `json:"tx_confirmations" yaml:"tx_confirmations"`
	ActivateHeight  uint64 `json:"activate_height" yaml:"activate_height"`

	// main chain
	StakingManagerAddress hmTypes.HeimdallAddress `json:"staking_manager_address" yaml:"staking_manager_address"`
	SlashManagerAddress   hmTypes.HeimdallAddress `json:"slash_manager_address" yaml:"slash_manager_address"`
	RootChainAddress      hmTypes.HeimdallAddress `json:"root_chain_address" yaml:"root_chain_address"`
	StakingInfoAddress    hmTypes.HeimdallAddress `json:"staking_info_address" yaml:"staking_info_address"`
	StateSenderAddress    hmTypes.HeimdallAddress `json:"state_sender_address" yaml:"state_sender_address"`
}

func (cd *ChainData) Plainify() (pc PlainChainData) {
	if cd.TxConfirmations != nil {
		pc.TxConfirmations = *cd.TxConfirmations
	}

	if cd.ActivateHeight != nil {
		pc.ActivateHeight = *cd.ActivateHeight
	}

	if cd.StakingManagerAddress != nil {
		pc.StakingManagerAddress = *cd.StakingManagerAddress
	}

	if cd.SlashManagerAddress != nil {
		pc.SlashManagerAddress = *cd.SlashManagerAddress
	}

	if cd.RootChainAddress != nil {
		pc.RootChainAddress = *cd.RootChainAddress
	}

	if cd.StakingInfoAddress != nil {
		pc.StakingInfoAddress = *cd.StakingInfoAddress
	}

	if cd.StateSenderAddress != nil {
		pc.StateSenderAddress = *cd.StateSenderAddress
	}

	return
}

// DefaultParams returns a default set of parameters.
func DefaultParamsWithMultiChains() ParamsWithMultiChains {
	return ParamsWithMultiChains{
		ChainParameterMap: make(map[string]ChainData),
	}
}

func (pmc *ParamsWithMultiChains) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{Key: KeyParamsWithMultiChains, Value: pmc},
	}
}

func (pmc ParamsWithMultiChains) String() string {
	var ret string

	for key, val := range pmc.ChainParameterMap {
		pVal := val.Plainify()

		ret += fmt.Sprintf(`
		[chain]: %s
		TxConfirmations:					%v,
		ActivateHeight:						%v,

		StakingManagerAddress:				%s,
		SlashManagerAddress:				%s,
		RootChainAddress:					%s,
		StakingInfoAddress:					%s,
		StateSenderAddress:					%s,
		`,
			key, pVal.TxConfirmations, pVal.ActivateHeight,
			pVal.StakingManagerAddress, pVal.SlashManagerAddress, pVal.RootChainAddress,
			pVal.StakingInfoAddress, pVal.StateSenderAddress)
	}

	return ret
}
