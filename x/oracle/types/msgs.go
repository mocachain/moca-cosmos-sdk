package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ValidatorBitSetLength = 4 // 256 bits
	BLSPublicKeyLength    = 128
	BLSSignatureLength    = 64
)

type (
	BLSPublicKey [BLSPublicKeyLength]byte
	BLSSignature [BLSSignatureLength]byte
)

func NewMsgClaim(fromAddr string, srcShainId, destChainId uint32, sequence, timestamp uint64, payload []byte, voteAddrSet []uint64, aggSignature []byte) *MsgClaim {
	return &MsgClaim{
		FromAddress:    fromAddr,
		SrcChainId:     srcShainId,
		DestChainId:    destChainId,
		Sequence:       sequence,
		Timestamp:      timestamp,
		Payload:        payload,
		VoteAddressSet: voteAddrSet,
		AggSignature:   aggSignature,
	}
}

// Route implements the LegacyMsg interface.
func (m MsgClaim) Route() string { return sdk.MsgTypeURL(&m) }

// Type implements the LegacyMsg interface.
func (m MsgClaim) Type() string { return sdk.MsgTypeURL(&m) }

// ValidateBasic does a sanity check on the provided data.
func (m *MsgClaim) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(m.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}

	// if m.SrcChainId > math.MaxUint16 {
	// 	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest,
	// 		fmt.Sprintf("chain id should not be larger than %d", math.MaxUint16))
	// }

	// if m.DestChainId > math.MaxUint16 {
	// 	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest,
	// 		fmt.Sprintf("chain id should not be larger than %d", math.MaxUint16))
	// }

	if len(m.Payload) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "payload should not be empty")
	}

	if len(m.VoteAddressSet) != ValidatorBitSetLength {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"length of vote address set should be %d", ValidatorBitSetLength)
	}

	if len(m.AggSignature) != BLSSignatureLength {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"length of signature should be %d", BLSSignatureLength)
	}

	if m.Timestamp == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "timestamp should not be 0")
	}

	return nil
}

// GetSigners returns the expected signers for MsgCancelUpgrade.
func (m *MsgClaim) GetSigners() []sdk.AccAddress {
	fromAddress := sdk.MustAccAddressFromHex(m.FromAddress)
	return []sdk.AccAddress{fromAddress}
}

// GetBlsSignBytes returns the sign bytes of bls signature
func (m *MsgClaim) GetBlsSignBytes() [32]byte {
	blsClaim := &BlsClaim{
		SrcChainId:  m.SrcChainId,
		DestChainId: m.DestChainId,
		Timestamp:   m.Timestamp,
		Sequence:    m.Sequence,
		Payload:     m.Payload,
	}
	return blsClaim.GetSignBytes()
}

type BlsClaim struct {
	SrcChainId  uint32
	DestChainId uint32
	Timestamp   uint64
	Sequence    uint64
	Payload     []byte
}

func (c *BlsClaim) GetSignBytes() [32]byte {
	bts, err := rlp.EncodeToBytes(c)
	if err != nil {
		panic("encode bls claim error")
	}

	btsHash := sdk.Keccak256Hash(bts)
	return btsHash
}

type Packages []Package

type Package struct {
	ChannelId sdk.ChannelID
	Sequence  uint64
	Payload   []byte
}

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
