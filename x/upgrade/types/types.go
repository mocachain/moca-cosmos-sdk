// Package types provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/upgrade/types imports
package types

import (
	"cosmossdk.io/x/upgrade/types"
)

// Re-export all types and functions from the new module
type (
	Plan = types.Plan
	SoftwareUpgradeProposal = types.SoftwareUpgradeProposal
	CancelSoftwareUpgradeProposal = types.CancelSoftwareUpgradeProposal
	MsgSoftwareUpgrade = types.MsgSoftwareUpgrade
	MsgCancelUpgrade = types.MsgCancelUpgrade
)

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
		Plan:      (*types.Plan)(plan),
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
)
