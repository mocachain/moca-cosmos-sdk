// Package errors provides compatibility for legacy cosmos-sdk errors.
//
// Deprecated: This package provides backward compatibility for code that uses
// github.com/cosmos/cosmos-sdk/types/errors. New code should use cosmossdk.io/errors directly.
// Migration guide: https://docs.cosmos.network/v0.50/migrations
//
// Migration examples:
//   - sdkerrors.Register → errors.Register
//   - sdkerrors.Wrap → errors.Wrap
//   - sdkerrors.ErrUnknownAddress → errors.Register("sdk", 9, "unknown address")
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
// Error codes match the official Cosmos SDK v0.50.13 definitions
var (
	// ErrTxDecode is returned if we cannot parse a transaction
	ErrTxDecode = Register("sdk", 2, "tx parse error")
	// ErrInvalidSequence is used the sequence number (nonce) is incorrect for the signature
	ErrInvalidSequence = Register("sdk", 3, "invalid sequence")
	// ErrUnauthorized is used whenever a request without sufficient authorization is handled
	ErrUnauthorized = Register("sdk", 4, "unauthorized")
	// ErrInsufficientFunds is used when the account cannot pay requested amount
	ErrInsufficientFunds = Register("sdk", 5, "insufficient funds")
	// ErrUnknownRequest to doc
	ErrUnknownRequest = Register("sdk", 6, "unknown request")
	// ErrInvalidAddress to doc
	ErrInvalidAddress = Register("sdk", 7, "invalid address")
	// ErrInvalidPubKey to doc
	ErrInvalidPubKey = Register("sdk", 8, "invalid pubkey")
	// ErrUnknownAddress to doc
	ErrUnknownAddress = Register("sdk", 9, "unknown address")
	// ErrInvalidCoins to doc
	ErrInvalidCoins = Register("sdk", 10, "invalid coins")
	// ErrOutOfGas to doc
	ErrOutOfGas = Register("sdk", 11, "out of gas")
	// ErrMemoTooLarge to doc
	ErrMemoTooLarge = Register("sdk", 12, "memo too large")
	// ErrInsufficientFee to doc
	ErrInsufficientFee = Register("sdk", 13, "insufficient fee")
	// ErrTooManySignatures to doc
	ErrTooManySignatures = Register("sdk", 14, "maximum number of signatures exceeded")
	// ErrNoSignatures to doc
	ErrNoSignatures = Register("sdk", 15, "no signatures supplied")
	// ErrJSONMarshal defines an ABCI typed JSON marshaling error
	ErrJSONMarshal = Register("sdk", 16, "failed to marshal JSON bytes")
	// ErrJSONUnmarshal defines an ABCI typed JSON unmarshalling error
	ErrJSONUnmarshal = Register("sdk", 17, "failed to unmarshal JSON bytes")
	// ErrInvalidRequest defines an ABCI typed error where the request contains invalid data
	ErrInvalidRequest = Register("sdk", 18, "invalid request")
	// ErrTxInMempoolCache defines an ABCI typed error where a tx already exists in the mempool
	ErrTxInMempoolCache = Register("sdk", 19, "tx already in mempool")
	// ErrMempoolIsFull defines an ABCI typed error where the mempool is full
	ErrMempoolIsFull = Register("sdk", 20, "mempool is full")
	// ErrTxTooLarge defines an ABCI typed error where tx is too large
	ErrTxTooLarge = Register("sdk", 21, "tx too large")
	// ErrKeyNotFound defines an error when the key doesn't exist
	ErrKeyNotFound = Register("sdk", 22, "key not found")
	// ErrWrongPassword defines an error when the key password is invalid
	ErrWrongPassword = Register("sdk", 23, "invalid account password")
	// ErrorInvalidSigner defines an error when the tx intended signer does not match the given signer
	ErrorInvalidSigner = Register("sdk", 24, "tx intended signer does not match the given signer")
	// ErrorInvalidGasAdjustment defines an error for an invalid gas adjustment
	ErrorInvalidGasAdjustment = Register("sdk", 25, "invalid gas adjustment")
	// ErrInvalidHeight defines an error for an invalid height
	ErrInvalidHeight = Register("sdk", 26, "invalid height")
	// ErrInvalidVersion defines a general error for an invalid version
	ErrInvalidVersion = Register("sdk", 27, "invalid version")
	// ErrInvalidChainID defines an error when the chain-id is invalid
	ErrInvalidChainID = Register("sdk", 28, "invalid chain-id")
	// ErrInvalidType defines an error an invalid type
	ErrInvalidType = Register("sdk", 29, "invalid type")
	// ErrTxTimeoutHeight defines an error for when a tx is rejected out due to an explicitly set timeout height
	ErrTxTimeoutHeight = Register("sdk", 30, "tx timeout height")
	// ErrUnknownExtensionOptions defines an error for unknown extension options
	ErrUnknownExtensionOptions = Register("sdk", 31, "unknown extension options")
	// ErrWrongSequence defines an error where the account sequence defined in the signer info doesn't match the account's actual sequence number
	ErrWrongSequence = Register("sdk", 32, "incorrect account sequence")
	// ErrPackAny defines an error when packing a protobuf message to Any fails
	ErrPackAny = Register("sdk", 33, "failed packing protobuf message to Any")
	// ErrUnpackAny defines an error when unpacking a protobuf message from Any fails
	ErrUnpackAny = Register("sdk", 34, "failed unpacking protobuf message from Any")
	// ErrLogic defines an internal logic error, e.g. an invariant or assertion that is violated
	ErrLogic = Register("sdk", 35, "internal logic error")
	// ErrConflict defines a conflict error, e.g. when two goroutines try to access the same resource and one of them fails
	ErrConflict = Register("sdk", 36, "conflict")
	// ErrNotSupported is returned when we call a branch of a code which is currently not supported
	ErrNotSupported = Register("sdk", 37, "feature not supported")
	// ErrNotFound defines an error when requested entity doesn't exist in the state
	ErrNotFound = Register("sdk", 38, "not found")
	// ErrIO should be used to wrap internal errors caused by external operation
	ErrIO = Register("sdk", 39, "Internal IO error")
	// ErrAppConfig defines an error occurred if application configuration is misconfigured
	ErrAppConfig = Register("sdk", 40, "error in app.toml")
	// ErrInvalidGasLimit defines an error when an invalid GasWanted value is supplied
	ErrInvalidGasLimit = Register("sdk", 41, "invalid gas limit")
	// ErrPanic should only be set when we recovering from a panic
	ErrPanic = errors.ErrPanic
)

// Common constants
const (
	RootCodespace = "sdk"
)
