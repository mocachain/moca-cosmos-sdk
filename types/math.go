package types

import (
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
)

// Deprecated: Legacy type aliases for backward compatibility.
// These types are provided for compatibility with code that uses sdk.Dec, sdk.Int, etc.
// New code should use cosmossdk.io/math and cosmossdk.io/store/types directly.
// Migration guide: https://docs.cosmos.network/v0.50/migrations
//
// Migration examples:
//   - sdk.Dec → math.LegacyDec (or math.Dec for new code)
//   - sdk.Int → math.Int
//   - sdk.KVStore → storetypes.KVStore
//   - sdk.CommitMultiStore → storetypes.CommitMultiStore
type Dec = math.LegacyDec
type Int = math.Int
type KVStore = storetypes.KVStore
type CommitMultiStore = storetypes.CommitMultiStore

// Deprecated: Legacy Dec constructor functions for backward compatibility.
// New code should use cosmossdk.io/math functions directly.
// Migration guide: https://docs.cosmos.network/v0.50/migrations
//
// Migration examples:
//   - sdk.NewDec → math.LegacyNewDec
//   - sdk.NewInt → math.NewInt
var (
	NewDec               = math.LegacyNewDec
	NewDecWithPrec       = math.LegacyNewDecWithPrec
	NewDecFromStr        = math.LegacyNewDecFromStr
	NewDecFromIntWithPrec = math.LegacyNewDecFromIntWithPrec
	NewInt               = math.NewInt
	NewIntFromString     = math.NewIntFromString
	OneDec               = math.LegacyOneDec
	ZeroInt              = math.ZeroInt
)

func (ip IntProto) String() string {
	return ip.Int.String()
}

func (dp DecProto) String() string {
	return dp.Dec.String()
}
