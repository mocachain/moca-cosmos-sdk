package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zkMeLabs/moca-cosmos-sdk/x/capability/types"
)

// Keeper defines the capability module keeper
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewKeeper creates a new capability Keeper instance
func NewKeeper(cdc codec.BinaryCodec, key storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: key,
	}
}

// GetCapability returns a capability for a given name
func (k Keeper) GetCapability(ctx sdk.Context, name string) (*types.Capability, bool) {
	// Compatibility stub - returns a dummy capability
	return types.NewCapability(1), true
}

// ClaimCapability claims a capability
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *types.Capability, name string) error {
	// Compatibility stub - always succeeds
	return nil
}

// AuthenticateCapability authenticates a capability
func (k Keeper) AuthenticateCapability(ctx sdk.Context, cap *types.Capability, name string) bool {
	// Compatibility stub - always returns true
	return true
}

