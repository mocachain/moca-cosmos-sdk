// Package module provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/feegrant/module imports
package module

import (
	"cosmossdk.io/x/feegrant/module"
)

// Re-export all types and functions from the new module
type (
	AppModule = module.AppModule
	AppModuleBasic = module.AppModuleBasic
)

// Re-export functions
var (
	NewAppModule = module.NewAppModule
)
