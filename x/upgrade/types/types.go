// Package types provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/upgrade/types imports.
//
// Deprecated: This package provides backward compatibility for code that uses
// github.com/cosmos/cosmos-sdk/x/upgrade/types. New code should use cosmossdk.io/x/upgrade/types directly.
// Migration guide: https://docs.cosmos.network/v0.50/migrations
//
// Migration examples:
//   - github.com/cosmos/cosmos-sdk/x/upgrade/types → cosmossdk.io/x/upgrade/types
//   - types.NewPlan → types.NewPlan (same function name, but from new package)
package types

import (
	"cosmossdk.io/x/upgrade/types"
)

// Deprecated: Use cosmossdk.io/x/upgrade/types.Plan directly.
// Re-export all types and functions from the new module
type (
	Plan = types.Plan
	SoftwareUpgradeProposal = types.SoftwareUpgradeProposal
	CancelSoftwareUpgradeProposal = types.CancelSoftwareUpgradeProposal
	MsgSoftwareUpgrade = types.MsgSoftwareUpgrade
	MsgCancelUpgrade = types.MsgCancelUpgrade
	QueryClient = types.QueryClient
)

// Deprecated: Use cosmossdk.io/x/upgrade/types.NewPlan directly.
// Compatibility functions for legacy API
// These functions create new instances of the types
func NewPlan(name string, height int64, info string) *Plan {
	p := types.Plan{
		Name:   name,
		Height: height,
		Info:   info,
	}
	return (*Plan)(&p)
}

func NewMsgSoftwareUpgrade(authority string, plan *Plan) *MsgSoftwareUpgrade {
	msg := types.MsgSoftwareUpgrade{
		Authority: authority,
		Plan:      types.Plan(*plan),
	}
	return (*MsgSoftwareUpgrade)(&msg)
}

func NewMsgCancelUpgrade(authority string) *MsgCancelUpgrade {
	msg := types.MsgCancelUpgrade{
		Authority: authority,
	}
	return (*MsgCancelUpgrade)(&msg)
}

// Re-export functions that exist in the new module
var (
	NewSoftwareUpgradeProposal = types.NewSoftwareUpgradeProposal
	NewCancelSoftwareUpgradeProposal = types.NewCancelSoftwareUpgradeProposal
	NewQueryClient = types.NewQueryClient
)

// Re-export constants
const (
	StoreKey   = types.StoreKey
	ModuleName = types.ModuleName
)

// Re-export UpgradeStoreLoader
var (
	UpgradeStoreLoader = types.UpgradeStoreLoader
)
