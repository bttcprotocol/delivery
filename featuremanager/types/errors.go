package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Param module codespace constants.
const (
	DefaultCodespace sdk.CodespaceType = "feature"

	CodeUnknownSubspace  sdk.CodeType = 1
	CodeSettingParameter sdk.CodeType = 2
	CodeEmptyData        sdk.CodeType = 3

	CodeUnknownProposal          sdk.CodeType = 4
	CodeInactiveProposal         sdk.CodeType = 5
	CodeAlreadyActiveProposal    sdk.CodeType = 6
	CodeAlreadyFinishedProposal  sdk.CodeType = 7
	CodeAddressNotStaked         sdk.CodeType = 8
	CodeInvalidContent           sdk.CodeType = 9
	CodeInvalidProposalType      sdk.CodeType = 10
	CodeInvalidVote              sdk.CodeType = 11
	CodeInvalidGenesis           sdk.CodeType = 12
	CodeInvalidProposalStatus    sdk.CodeType = 13
	CodeProposalHandlerNotExists sdk.CodeType = 14
)

// ErrUnknownSubspace returns an unknown subspace error.
func ErrUnknownSubspace(codespace sdk.CodespaceType, space string) sdk.Error {
	return sdk.NewError(codespace, CodeUnknownSubspace, fmt.Sprintf("unknown subspace %s", space))
}

// ErrSettingParameter returns an error for failing to set a parameter.
func ErrSettingParameter(codespace sdk.CodespaceType, key, value, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeSettingParameter, fmt.Sprintf(
		"error setting parameter %s on %s: %s", value, key, msg))
}

// ErrEmptyChanges returns an error for empty parameter changes.
func ErrEmptyChanges(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "submitted parameter changes are empty")
}

// ErrEmptySubspace returns an error for an empty subspace.
func ErrEmptySubspace(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "parameter subspace is empty")
}

// ErrEmptyKey returns an error for when an empty key is given.
func ErrEmptyKey(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "parameter key is empty")
}

// ErrEmptyValue returns an error for when an empty key is given.
func ErrEmptyValue(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "parameter value is empty")
}

func ErrInvalidProposalContent(cs sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(cs, CodeInvalidContent, fmt.Sprintf("invalid proposal content: %s", msg))
}

func ErrInvalidProposalType(codespace sdk.CodespaceType, proposalType string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidProposalType, fmt.Sprintf(
		"proposal type '%s' is not valid", proposalType))
}

func ErrNoProposalHandlerExists(codespace sdk.CodespaceType, content interface{}) sdk.Error {
	return sdk.NewError(codespace, CodeProposalHandlerNotExists, fmt.Sprintf(
		"'%T' does not have a corresponding handler", content))
}
