// Package feegrant provides compatibility for legacy github.com/cosmos/cosmos-sdk/x/feegrant imports
// This package redirects to the new cosmossdk.io/x/feegrant module
package feegrant

import (
	feegrant "cosmossdk.io/x/feegrant"
)

// Re-export all types and functions from the new module
type (
	// Basic allowance types
	BasicAllowance = feegrant.BasicAllowance
	PeriodicAllowance = feegrant.PeriodicAllowance
	AllowedMsgAllowance = feegrant.AllowedMsgAllowance
	
	// Grant types
	Grant = feegrant.Grant
	
	// Message types
	MsgGrantAllowance = feegrant.MsgGrantAllowance
	MsgRevokeAllowance = feegrant.MsgRevokeAllowance
)

// Re-export functions
var (
	NewBasicAllowance = feegrant.NewBasicAllowance
	NewPeriodicAllowance = feegrant.NewPeriodicAllowance
	NewAllowedMsgAllowance = feegrant.NewAllowedMsgAllowance
	NewGrant = feegrant.NewGrant
	NewMsgGrantAllowance = feegrant.NewMsgGrantAllowance
	NewMsgRevokeAllowance = feegrant.NewMsgRevokeAllowance
)

// Re-export constants
const (
	ModuleName = feegrant.ModuleName
	StoreKey = feegrant.StoreKey
	RouterKey = feegrant.RouterKey
)
