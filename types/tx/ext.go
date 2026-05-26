package tx

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// ExtensionOptionI defines the interface for tx extension options.
type ExtensionOptionI interface{} //nolint:revive // to avoid breaking change

// TxExtensionOptionI is the upstream cosmos-sdk name for the same interface,
// retained as an alias so downstream consumers that import the upstream
// identifier (e.g. cosmos/evm at codec setup) compile against this fork
// without modification.
type TxExtensionOptionI = ExtensionOptionI

// unpackTxExtensionOptionsI unpacks Any's to TxExtensionOptionI's.
func unpackTxExtensionOptionsI(unpacker types.AnyUnpacker, anys []*types.Any) error {
	for _, any := range anys {
		var opt ExtensionOptionI
		err := unpacker.UnpackAny(any, &opt)
		if err != nil {
			return err
		}
	}

	return nil
}
