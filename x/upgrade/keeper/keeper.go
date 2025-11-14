// Package keeper provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/upgrade/keeper imports
package keeper

import (
	"cosmossdk.io/x/upgrade/keeper"
)

// Re-export all types and functions from the new module
type (
	Keeper = keeper.Keeper
)

// Re-export functions
var (
	NewKeeper = keeper.NewKeeper
)

