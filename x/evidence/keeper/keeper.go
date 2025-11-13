package keeper

// Compatibility keeper for legacy x/evidence imports
// In Cosmos SDK v0.50, x/evidence has been moved to cosmossdk.io/x/evidence
// This provides minimal compatibility for existing code

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper defines the evidence module keeper
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewKeeper creates a new evidence Keeper instance
func NewKeeper(cdc codec.BinaryCodec, key storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: key,
	}
}

// SubmitEvidence submits evidence
func (k Keeper) SubmitEvidence(ctx sdk.Context, evidence interface{}) error {
	// Compatibility stub - always succeeds
	return nil
}

// GetEvidence gets evidence by hash
func (k Keeper) GetEvidence(ctx sdk.Context, evidenceHash []byte) (interface{}, bool) {
	// Compatibility stub - returns nil
	return nil, false
}

// GetAllEvidence gets all evidence
func (k Keeper) GetAllEvidence(ctx sdk.Context) []interface{} {
	// Compatibility stub - returns empty slice
	return []interface{}{}
}