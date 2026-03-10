package types

// Compatibility types for legacy x/capability imports
// In Cosmos SDK v0.50, x/capability has been removed
// This provides minimal compatibility for existing code

// Capability represents a capability
type Capability struct {
	Index uint64 `json:"index"`
}

// NewCapability creates a new Capability
func NewCapability(index uint64) *Capability {
	return &Capability{Index: index}
}

// GetIndex returns the index of the capability
func (c *Capability) GetIndex() uint64 {
	return c.Index
}


