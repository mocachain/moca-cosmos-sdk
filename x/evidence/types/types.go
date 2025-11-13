package types

// Compatibility types for legacy x/evidence imports
// In Cosmos SDK v0.50, x/evidence has been moved to cosmossdk.io/x/evidence
// This provides minimal compatibility for existing code

import (
	"time"
)

// Evidence represents evidence of misbehavior
type Evidence interface {
	Route() string
	Type() string
	String() string
	Hash() []byte
	ValidateBasic() error
	GetHeight() int64
}

// Equivocation represents equivocation evidence
type Equivocation struct {
	Height           int64     `json:"height"`
	Time             time.Time `json:"time"`
	Power            int64     `json:"power"`
	ConsensusAddress string    `json:"consensus_address"`
}

// Route returns the evidence route
func (e Equivocation) Route() string { return "evidence" }

// Type returns the evidence type
func (e Equivocation) Type() string { return "equivocation" }

// String returns string representation
func (e Equivocation) String() string { return "equivocation" }

// Hash returns evidence hash
func (e Equivocation) Hash() []byte { return []byte{} }

// ValidateBasic validates the evidence
func (e Equivocation) ValidateBasic() error { return nil }

// GetHeight returns the evidence height
func (e Equivocation) GetHeight() int64 { return e.Height }
