// Package evidence provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/evidence imports
package evidence

import (
	evidence "cosmossdk.io/x/evidence"
	evidencetypes "cosmossdk.io/x/evidence/types"
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

// Re-export types functions
var (
	RegisterInterfaces = evidencetypes.RegisterInterfaces
)

// Re-export constants
const (
	ModuleName = evidencetypes.ModuleName
	StoreKey   = evidencetypes.StoreKey
	// RouterKey is deprecated in v0.50, routing is handled via module manager
)

