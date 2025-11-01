package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cosmos-sdk/x/gashub/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgSetMsgGasParams{}, "cosmos-sdk/MsgSetMsgGasParams")

	cdc.RegisterInterface((*isMsgGasParams_GasParams)(nil), nil)
	cdc.RegisterConcrete(&MsgGasParams_FixedType{}, "cosmos-sdk/MsgGasParams/FixedType", nil)
	cdc.RegisterConcrete(&MsgGasParams_GrantType{}, "cosmos-sdk/MsgGasParams/GrantType", nil)
	cdc.RegisterConcrete(&MsgGasParams_MultiSendType{}, "cosmos-sdk/MsgGasParams/MultiSendType", nil)
	cdc.RegisterConcrete(&MsgGasParams_GrantAllowanceType{}, "cosmos-sdk/MsgGasParams/GrantAllowanceType", nil)

	cdc.RegisterConcrete(&Params{}, "cosmos-sdk/x/gashub/Params", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgSetMsgGasParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
