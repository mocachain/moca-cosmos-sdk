// Package evidence provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/evidence imports
package evidence

import (
	evidence "cosmossdk.io/x/evidence"
)

// Re-export all types and functions from the new module
type (
	AppModule = evidence.AppModule
	AppModuleBasic = evidence.AppModuleBasic
)

// Re-export functions
var (
	NewAppModule = evidence.NewAppModule
)

// Re-export constants
const (
	ModuleName = evidence.ModuleName
	StoreKey = evidence.StoreKey
	RouterKey = evidence.RouterKey
)
