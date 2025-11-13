package keeper

// Compatibility keeper for legacy x/feegrant imports
// In Cosmos SDK v0.50, x/feegrant has been moved to cosmossdk.io/x/feegrant
// This provides minimal compatibility for existing code

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper defines the feegrant module keeper
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewKeeper creates a new feegrant Keeper instance
func NewKeeper(cdc codec.BinaryCodec, key storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: key,
	}
}

// GrantAllowance grants fee allowance
func (k Keeper) GrantAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress, allowance interface{}) error {
	// Compatibility stub - always succeeds
	return nil
}

// RevokeAllowance revokes fee allowance
func (k Keeper) RevokeAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress) error {
	// Compatibility stub - always succeeds
	return nil
}

// GetAllowance gets fee allowance
func (k Keeper) GetAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress) (interface{}, error) {
	// Compatibility stub - returns nil
	return nil, nil
}