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

// Re-export functions
var (
	NewPlan = types.NewPlan
	NewSoftwareUpgradeProposal = types.NewSoftwareUpgradeProposal
	NewCancelSoftwareUpgradeProposal = types.NewCancelSoftwareUpgradeProposal
	NewMsgSoftwareUpgrade = types.NewMsgSoftwareUpgrade
	NewMsgCancelUpgrade = types.NewMsgCancelUpgrade
)
