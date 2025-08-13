package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// verify interface at compile time
var (
	_ sdk.Msg = &MsgUnjail{}
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgImpeach{}
)

// NewMsgUnjail creates a new MsgUnjail instance
func NewMsgUnjail(validatorAddr string) *MsgUnjail {
	return &MsgUnjail{
		ValidatorAddr: validatorAddr,
	}
}

// NewMsgImpeach creates a new MsgImpeach instance
func NewMsgImpeach(valAddr, from sdk.AccAddress) *MsgImpeach {
	return &MsgImpeach{
		ValidatorAddress: valAddr.String(),
		From:             from.String(),
	}
}

// GetSigners implements the sdk.Msg interface.
func (msg MsgImpeach) GetSigners() []sdk.AccAddress {
	fromAddr, _ := sdk.AccAddressFromHexUnsafe(msg.From)
	return []sdk.AccAddress{fromAddr}
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgImpeach) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(msg.From); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid account address: %s", err)
	}

	if _, err := sdk.AccAddressFromHexUnsafe(msg.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	return nil
}
