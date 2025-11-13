// Package errors provides compatibility for legacy cosmos-sdk errors.
//
// Deprecated: This package provides backward compatibility for code that uses
// github.com/cosmos/cosmos-sdk/types/errors. New code should use cosmossdk.io/errors directly.
// Migration guide: https://docs.cosmos.network/v0.50/migrations
//
// Migration examples:
//   - sdkerrors.Register → errors.Register
//   - sdkerrors.Wrap → errors.Wrap
//   - sdkerrors.ErrUnknownAddress → errors.Register("sdk", 7, "unknown address")
package errors

import (
	"cosmossdk.io/errors"
)

// Deprecated: Use cosmossdk.io/errors.Register directly.
// Legacy error functions for backward compatibility
var (
	Register = errors.Register
	Wrap     = errors.Wrap
	Wrapf    = errors.Wrapf
	New      = errors.New
)

// Deprecated: Use cosmossdk.io/errors.Error directly.
// Legacy error types
type Error = errors.Error

// Deprecated: Use cosmossdk.io/errors.Register to create custom error constants.
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
	ErrJSONUnmarshal     = Register("sdk", 8, "failed to unmarshal JSON bytes")
	ErrUnauthorized      = Register("sdk", 4, "unauthorized")
	ErrorInvalidSigner   = Register("sdk", 6, "invalid signer")
	ErrWrongPassword     = Register("sdk", 2, "invalid account password")
	ErrLogic             = Register("sdk", 11, "internal logic error")
	ErrInvalidCoins      = Register("sdk", 10, "invalid coins")
	ErrKeyNotFound       = Register("sdk", 9, "key not found")
	ErrTxInMempoolCache  = Register("sdk", 19, "tx already in mempool")
	ErrMempoolIsFull     = Register("sdk", 20, "mempool is full")
	ErrTxTooLarge        = Register("sdk", 21, "tx too large")
	ErrPanic             = Register("sdk", 111222, "panic")
	ErrUnknownRequest    = Register("sdk", 1, "unknown request")
	ErrTxDecode          = Register("sdk", 2, "tx parse error")
	ErrInvalidHeight     = Register("sdk", 22, "invalid height")
	ErrOutOfGas          = Register("sdk", 11, "out of gas")
	ErrInvalidChainID    = Register("sdk", 28, "invalid chain-id")
	ErrInvalidVersion      = Register("sdk", 42, "invalid version")
	ErrInsufficientFunds   = Register("sdk", 5, "insufficient funds")
	ErrMemoTooLarge        = Register("sdk", 14, "memo too large")
	ErrTxTimeoutHeight     = Register("sdk", 15, "tx timeout height")
	ErrUnknownExtensionOptions = Register("sdk", 16, "unknown extension options")
	ErrInvalidGasLimit     = Register("sdk", 12, "invalid gas limit")
	ErrWrongSequence       = Register("sdk", 3, "wrong sequence")
	ErrTooManySignatures   = Register("sdk", 17, "too many signatures")
	ErrNotSupported        = Register("sdk", 39, "not supported")
)

// Common constants
const (
	RootCodespace = "sdk"
)