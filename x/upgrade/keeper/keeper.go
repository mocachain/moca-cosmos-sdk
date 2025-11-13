package keeper

// Compatibility keeper for legacy x/upgrade imports
// In Cosmos SDK v0.50, x/upgrade has been moved to cosmossdk.io/x/upgrade
// This provides minimal compatibility for existing code

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper defines the upgrade module keeper
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewKeeper creates a new upgrade Keeper instance
func NewKeeper(cdc codec.BinaryCodec, key storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: key,
	}
}

// ScheduleUpgrade schedules an upgrade
func (k Keeper) ScheduleUpgrade(ctx sdk.Context, plan interface{}) error {
	// Compatibility stub - always succeeds
	return nil
}

// GetUpgradePlan gets the current upgrade plan
func (k Keeper) GetUpgradePlan(ctx sdk.Context) (interface{}, bool) {
	// Compatibility stub - returns nil
	return nil, false
}

// ClearUpgradePlan clears the upgrade plan
func (k Keeper) ClearUpgradePlan(ctx sdk.Context) {
	// Compatibility stub - does nothing
}