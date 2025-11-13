// Package upgrade provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/upgrade imports
package upgrade

import (
	upgrade "cosmossdk.io/x/upgrade"
)

// Re-export all types and functions from the new module
type (
	AppModule = upgrade.AppModule
	AppModuleBasic = upgrade.AppModuleBasic
)

// Re-export functions
var (
	NewAppModule = upgrade.NewAppModule
)

// Re-export constants
const (
	ModuleName = upgrade.ModuleName
	StoreKey = upgrade.StoreKey
	RouterKey = upgrade.RouterKey
)
