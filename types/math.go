package types

import (
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
)

// Legacy type aliases for backward compatibility
type Dec = math.LegacyDec
type Int = math.Int
type KVStore = storetypes.KVStore
type CommitMultiStore = storetypes.CommitMultiStore

// Legacy Dec constructor functions for backward compatibility
var (
	NewDec         = math.LegacyNewDec
	NewDecWithPrec = math.LegacyNewDecWithPrec
	NewDecFromStr  = math.LegacyNewDecFromStr
	NewIntFromString = math.NewIntFromString
)

func (ip IntProto) String() string {
	return ip.Int.String()
}

func (dp DecProto) String() string {
	return dp.Dec.String()
}
