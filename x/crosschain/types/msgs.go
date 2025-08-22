package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgUpdateChannelPermissions{}
	_ sdk.Msg = &MsgMintModuleTokens{}
)

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateChannelPermissions) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateChannelPermissions) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	for _, per := range m.ChannelPermissions {
		if per == nil {
			return fmt.Errorf("channel permission is nil")
		}
	}

	return nil
}

// GetSigners returns the expected signers for a MsgMintModuleTokens message.
func (m *MsgMintModuleTokens) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgMintModuleTokens) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if m.Amount.LTE(math.ZeroInt()) {
		return fmt.Errorf("amount must be positive, is %s", m.Amount)
	}

	return nil
}
