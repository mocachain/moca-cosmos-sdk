// Package upgrade provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/upgrade imports
package upgrade

import (
	upgrade "cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
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

// Re-export types package types and functions
type (
	QueryClient = upgradetypes.QueryClient
)

var (
	NewQueryClient = upgradetypes.NewQueryClient
)

// Re-export constants
const (
	ModuleName = upgradetypes.ModuleName
	StoreKey   = upgradetypes.StoreKey
	RouterKey  = upgradetypes.RouterKey
)

