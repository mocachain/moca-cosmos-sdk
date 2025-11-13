// Package module provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/feegrant/module imports.
//
// Deprecated: This package provides backward compatibility for code that uses
// github.com/cosmos/cosmos-sdk/x/feegrant/module. New code should use cosmossdk.io/x/feegrant/module directly.
// Migration guide: https://docs.cosmos.network/v0.50/migrations
//
// Migration examples:
//   - github.com/cosmos/cosmos-sdk/x/feegrant/module → cosmossdk.io/x/feegrant/module
package module

import (
	"cosmossdk.io/x/feegrant/module"
)

// Deprecated: Use cosmossdk.io/x/feegrant/module.AppModule directly.
// Re-export all types and functions from the new module
type (
	AppModule = module.AppModule
	AppModuleBasic = module.AppModuleBasic
)

// Deprecated: Use cosmossdk.io/x/feegrant/module.NewAppModule directly.
// Re-export functions
var (
	NewAppModule = module.NewAppModule
)
