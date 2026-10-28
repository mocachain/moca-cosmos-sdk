package nestedpb

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInterfaces registers MsgNestedFields as an sdk.Msg implementation so a
// transaction carrying it can be built, encoded and decoded through the real tx
// pipeline, the way the EIP-712 encoder sees it.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgNestedFields{})
}
