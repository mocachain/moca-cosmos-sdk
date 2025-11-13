package feegrant

// Compatibility redirect for legacy github.com/cosmos/cosmos-sdk/x/feegrant imports
// This package redirects to the new cosmossdk.io/x/feegrant module

import (
	"cosmossdk.io/x/feegrant"
)

// Re-export all types and functions from the new module
type (
	// Module types
	AppModule = feegrant.AppModule
	AppModuleBasic = feegrant.AppModuleBasic
	
	// Keeper types
	Keeper = feegrant.Keeper
	
	// Basic allowance types
	BasicAllowance = feegrant.BasicAllowance
	PeriodicAllowance = feegrant.PeriodicAllowance
	AllowedMsgAllowance = feegrant.AllowedMsgAllowance
	
	// Grant types
	Grant = feegrant.Grant
	
	// Message types
	MsgGrantAllowance = feegrant.MsgGrantAllowance
	MsgRevokeAllowance = feegrant.MsgRevokeAllowance
	
	// Query types
	QueryAllowanceRequest = feegrant.QueryAllowanceRequest
	QueryAllowanceResponse = feegrant.QueryAllowanceResponse
	QueryAllowancesRequest = feegrant.QueryAllowancesRequest
	QueryAllowancesResponse = feegrant.QueryAllowancesResponse
)

// Re-export functions
var (
	NewKeeper = feegrant.NewKeeper
	NewAppModule = feegrant.NewAppModule
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
