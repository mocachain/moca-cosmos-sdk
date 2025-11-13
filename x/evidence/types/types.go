// Package types provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/evidence/types imports.
//
// Deprecated: This package provides backward compatibility for code that uses
// github.com/cosmos/cosmos-sdk/x/evidence/types. New code should use cosmossdk.io/x/evidence/types directly.
// Migration guide: https://docs.cosmos.network/v0.50/migrations
//
// Migration examples:
//   - github.com/cosmos/cosmos-sdk/x/evidence/types → cosmossdk.io/x/evidence/types
package types

import (
	"cosmossdk.io/x/evidence/types"
)

// Deprecated: Use cosmossdk.io/x/evidence/types.Evidence directly.
// Re-export all types and functions from the new module
type (
	Evidence = types.Evidence
	Equivocation = types.Equivocation
	MsgSubmitEvidence = types.MsgSubmitEvidence
)

// Deprecated: Use cosmossdk.io/x/evidence/types.NewMsgSubmitEvidence directly.
// Re-export functions
var (
	NewMsgSubmitEvidence = types.NewMsgSubmitEvidence
	NewEquivocation = types.NewEquivocation
)
