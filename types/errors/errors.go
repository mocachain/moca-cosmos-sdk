// Package errors provides compatibility for legacy cosmos-sdk errors
package errors

import (
	"cosmossdk.io/errors"
)

// Legacy error functions for backward compatibility
var (
	Register = errors.Register
	Wrap     = errors.Wrap
	Wrapf    = errors.Wrapf
	New      = errors.New
)

// Legacy error types
type Error = errors.Error

// Common error constants for backward compatibility
var (
	ErrUnknownAddress  = Register("sdk", 7, "unknown address")
	ErrPackAny         = Register("sdk", 15, "failed packing protobuf message to Any")
	ErrInvalidPubKey   = Register("sdk", 5, "invalid pubkey")
	ErrInvalidType     = Register("sdk", 17, "invalid type")
	ErrAppConfig       = Register("sdk", 1, "error in app.toml")
	ErrNotFound        = Register("sdk", 38, "not found")
	ErrInvalidRequest  = Register("sdk", 18, "invalid request")
	ErrNoSignatures    = Register("sdk", 16, "no signatures supplied")
	ErrInsufficientFee = Register("sdk", 13, "insufficient fee")
	ErrInvalidAddress  = Register("sdk", 20, "invalid address")
	ErrJSONUnmarshal   = Register("sdk", 8, "failed to unmarshal JSON bytes")
	ErrUnauthorized    = Register("sdk", 4, "unauthorized")
)

// Common constants
const (
	RootCodespace = "sdk"
)