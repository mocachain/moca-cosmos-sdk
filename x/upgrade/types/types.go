package types

// Compatibility types for legacy x/upgrade imports
// In Cosmos SDK v0.50, x/upgrade has been moved to cosmossdk.io/x/upgrade
// This provides minimal compatibility for existing code

import (
	"time"
)

// Plan represents an upgrade plan
type Plan struct {
	Name   string    `json:"name"`
	Height int64     `json:"height"`
	Time   time.Time `json:"time"`
	Info   string    `json:"info"`
}

// NewPlan creates a new upgrade plan
func NewPlan(name string, height int64, info string) Plan {
	return Plan{
		Name:   name,
		Height: height,
		Info:   info,
	}
}

// String returns string representation
func (p Plan) String() string {
	return p.Name
}

// ShouldExecute returns whether the plan should execute
func (p Plan) ShouldExecute(ctx interface{}) bool {
	// Compatibility stub - always returns false
	return false
}
