// Package upgrade provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/upgrade imports
// This package redirects to the new cosmossdk.io/x/upgrade module
package upgrade

import (
	upgrade "cosmossdk.io/x/upgrade"
	"cosmossdk.io/x/upgrade/keeper"
	"cosmossdk.io/x/upgrade/types"
)

// Re-export all types and functions from the new module
type (
	// Module types
	AppModule = upgrade.AppModule
	AppModuleBasic = upgrade.AppModuleBasic
	
	// Keeper types
	Keeper = keeper.Keeper
	
	// Plan types
	Plan = types.Plan
	SoftwareUpgradeProposal = types.SoftwareUpgradeProposal
	CancelSoftwareUpgradeProposal = types.CancelSoftwareUpgradeProposal
	
	// Message types
	MsgSoftwareUpgrade = types.MsgSoftwareUpgrade
	MsgCancelUpgrade = types.MsgCancelUpgrade
)

// Re-export functions
var (
	NewKeeper = keeper.NewKeeper
	NewAppModule = upgrade.NewAppModule
	NewPlan = types.NewPlan
	NewSoftwareUpgradeProposal = types.NewSoftwareUpgradeProposal
	NewCancelSoftwareUpgradeProposal = types.NewCancelSoftwareUpgradeProposal
	NewMsgSoftwareUpgrade = types.NewMsgSoftwareUpgrade
	NewMsgCancelUpgrade = types.NewMsgCancelUpgrade
)

// Re-export constants
const (
	ModuleName = upgrade.ModuleName
	StoreKey = upgrade.StoreKey
	RouterKey = upgrade.RouterKey
)
