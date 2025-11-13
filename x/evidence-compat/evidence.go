// Package evidence provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/evidence imports
// This package redirects to the new cosmossdk.io/x/evidence module
package evidence

import (
	evidence "cosmossdk.io/x/evidence"
	"cosmossdk.io/x/evidence/keeper"
	"cosmossdk.io/x/evidence/types"
)

// Re-export all types and functions from the new module
type (
	// Module types
	AppModule = evidence.AppModule
	AppModuleBasic = evidence.AppModuleBasic
	
	// Keeper types
	Keeper = keeper.Keeper
	
	// Evidence types
	Evidence = types.Evidence
	Equivocation = types.Equivocation
	
	// Message types
	MsgSubmitEvidence = types.MsgSubmitEvidence
)

// Re-export functions
var (
	NewKeeper = keeper.NewKeeper
	NewAppModule = evidence.NewAppModule
	NewMsgSubmitEvidence = types.NewMsgSubmitEvidence
	NewEquivocation = types.NewEquivocation
)

// Re-export constants
const (
	ModuleName = evidence.ModuleName
	StoreKey = evidence.StoreKey
	RouterKey = evidence.RouterKey
)
