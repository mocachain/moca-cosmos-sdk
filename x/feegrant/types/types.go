package types

// Compatibility types for legacy x/feegrant imports
// In Cosmos SDK v0.50, x/feegrant has been moved to cosmossdk.io/x/feegrant
// This provides minimal compatibility for existing code

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FeeAllowanceI represents a fee allowance interface
type FeeAllowanceI interface {
	Accept(ctx sdk.Context, fee sdk.Coins, msgs []sdk.Msg) (bool, error)
	ValidateBasic() error
}

// BasicAllowance represents basic fee allowance
type BasicAllowance struct {
	SpendLimit sdk.Coins `json:"spend_limit"`
	Expiration *int64    `json:"expiration,omitempty"`
}

// Accept accepts fee payment
func (a BasicAllowance) Accept(ctx sdk.Context, fee sdk.Coins, msgs []sdk.Msg) (bool, error) {
	// Compatibility stub - always accepts
	return true, nil
}

// ValidateBasic validates the allowance
func (a BasicAllowance) ValidateBasic() error {
	// Compatibility stub - always valid
	return nil
}

// PeriodicAllowance represents periodic fee allowance
type PeriodicAllowance struct {
	Basic            BasicAllowance `json:"basic"`
	Period           int64          `json:"period"`
	PeriodSpendLimit sdk.Coins      `json:"period_spend_limit"`
	PeriodCanSpend   sdk.Coins      `json:"period_can_spend"`
	PeriodReset      int64          `json:"period_reset"`
}

// Accept accepts fee payment
func (a PeriodicAllowance) Accept(ctx sdk.Context, fee sdk.Coins, msgs []sdk.Msg) (bool, error) {
	// Compatibility stub - always accepts
	return true, nil
}

// ValidateBasic validates the allowance
func (a PeriodicAllowance) ValidateBasic() error {
	// Compatibility stub - always valid
	return nil
}
