// Package types provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/evidence/types imports
package types

import (
	"cosmossdk.io/x/evidence/types"
)

// Re-export all types and functions from the new module
type (
	Evidence = types.Evidence
	Equivocation = types.Equivocation
	MsgSubmitEvidence = types.MsgSubmitEvidence
)

// Re-export functions
var (
	NewMsgSubmitEvidence = types.NewMsgSubmitEvidence
	NewEquivocation = types.NewEquivocation
)
