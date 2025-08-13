package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg                            = &MsgCreateValidator{}
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateValidator)(nil)
	_ sdk.Msg                            = &MsgEditValidator{}
	_ sdk.Msg                            = &MsgDelegate{}
	_ sdk.Msg                            = &MsgUndelegate{}
	_ sdk.Msg                            = &MsgBeginRedelegate{}
	_ sdk.Msg                            = &MsgCancelUnbondingDelegation{}
	_ sdk.Msg                            = &MsgUpdateParams{}
)

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
// Delegator address and validator address are the same.
func NewMsgCreateValidator(
	valAddr string, pubKey cryptotypes.PubKey,
	selfDelegation sdk.Coin, description Description, commission CommissionRates, minSelfDelegation math.Int,
	from, selfDelAddr, relayerAddr, challengerAddr sdk.AccAddress, blsKey, blsProof string,
) (*MsgCreateValidator, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}
	return &MsgCreateValidator{
		Description:       description,
		ValidatorAddress:  valAddr,
		DelegatorAddress:  selfDelAddr.String(),
		Pubkey:            pkAny,
		Value:             selfDelegation,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
		From:              from.String(),
		RelayerAddress:    relayerAddr.String(),
		ChallengerAddress: challengerAddr.String(),
		BlsKey:            blsKey,
		BlsProof:          blsProof,
	}, nil
}

// GetSigners implements the sdk.Msg interface. It returns the address(es) that
// must sign over msg.GetSignBytes().
// If the validator address is not same as delegator's, then the validator must
// sign the msg as well.
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromHexUnsafe(msg.From)
	return []sdk.AccAddress{from}
}

// Validate validates the MsgCreateValidator sdk msg.
func (msg MsgCreateValidator) Validate() error {
	// note that unmarshaling from bech32 ensures both non-empty and valid
	_, err := sdk.AccAddressFromHexUnsafe(msg.ValidatorAddress)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if msg.Pubkey == nil {
		return ErrEmptyValidatorPubKey
	}

	if !msg.Value.IsValid() || !msg.Value.Amount.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid delegation amount")
	}

	if _, err := sdk.AccAddressFromHexUnsafe(msg.RelayerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid relayer address: %s", err)
	}

	if _, err := sdk.AccAddressFromHexUnsafe(msg.ChallengerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid challenger address: %s", err)
	}

	if len(msg.BlsKey) != 2*sdk.BLSPubKeyLength {
		return ErrValidatorInvalidBlsKey
	}

	if len(msg.BlsProof) != 2*sdk.BLSSignatureLength {
		return ErrValidatorInvalidBlsProof.Wrapf("proof length is invalid %d", len(msg.BlsProof))
	}

	if msg.Description == (Description{}) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if msg.Commission == (CommissionRates{}) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty commission")
	}

	if err := msg.Commission.Validate(); err != nil {
		return err
	}

	if !msg.MinSelfDelegation.IsPositive() {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"minimum self delegation must be a positive integer",
		)
	}

	if msg.Value.Amount.LT(msg.MinSelfDelegation) {
		return ErrSelfDelegationBelowMinimum
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.Pubkey, &pubKey)
}

// NewMsgEditValidator creates a new MsgEditValidator instance
func NewMsgEditValidator(
	valAddr string, description Description, newRate *math.LegacyDec, newMinSelfDelegation *math.Int,
	newRelayerAddr, newChallengerAddr string, newBlsKey, newBlsProof string,
) *MsgEditValidator {
	return &MsgEditValidator{
		Description:       description,
		CommissionRate:    newRate,
		ValidatorAddress:  valAddr,
		MinSelfDelegation: newMinSelfDelegation,
		RelayerAddress:    newRelayerAddr,
		ChallengerAddress: newChallengerAddr,
		BlsKey:            newBlsKey,
		BlsProof:          newBlsProof,
	}
}

// NewMsgDelegate creates a new MsgDelegate instance.
func NewMsgDelegate(delAddr, valAddr string, amount sdk.Coin) *MsgDelegate {
	return &MsgDelegate{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           amount,
	}
}

// NewMsgBeginRedelegate creates a new MsgBeginRedelegate instance.
func NewMsgBeginRedelegate(
	delAddr, valSrcAddr, valDstAddr string, amount sdk.Coin,
) *MsgBeginRedelegate {
	return &MsgBeginRedelegate{
		DelegatorAddress:    delAddr,
		ValidatorSrcAddress: valSrcAddr,
		ValidatorDstAddress: valDstAddr,
		Amount:              amount,
	}
}

// NewMsgUndelegate creates a new MsgUndelegate instance.
func NewMsgUndelegate(delAddr, valAddr string, amount sdk.Coin) *MsgUndelegate {
	return &MsgUndelegate{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           amount,
	}
}

// NewMsgCancelUnbondingDelegation creates a new MsgCancelUnbondingDelegation instance.
func NewMsgCancelUnbondingDelegation(delAddr, valAddr string, creationHeight int64, amount sdk.Coin) *MsgCancelUnbondingDelegation {
	return &MsgCancelUnbondingDelegation{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           amount,
		CreationHeight:   creationHeight,
	}
}
