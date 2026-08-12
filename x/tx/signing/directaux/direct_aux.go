package directaux

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-proto/anyutil"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/x/tx/signing"
)

// SignModeHandler is the SIGN_MODE_DIRECT_AUX implementation of signing.SignModeHandler.
type SignModeHandler struct {
	signersContext *signing.Context
	fileResolver   signing.ProtoFileResolver
	typeResolver   protoregistry.MessageTypeResolver
}

// SignModeHandlerOptions are the options for the SignModeHandler.
type SignModeHandlerOptions struct {
	// TypeResolver is the protoregistry.MessageTypeResolver to use for resolving protobuf types when unpacking any messages.
	TypeResolver protoregistry.MessageTypeResolver

	// SignersContext is the signing.Context to use for getting signers.
	SignersContext *signing.Context
}

// NewSignModeHandler returns a new SignModeHandler.
func NewSignModeHandler(options SignModeHandlerOptions) (SignModeHandler, error) {
	h := SignModeHandler{}

	if options.SignersContext == nil {
		return h, errors.New("signers context is required")
	}
	h.signersContext = options.SignersContext

	h.fileResolver = h.signersContext.FileResolver()

	if options.TypeResolver == nil {
		h.typeResolver = protoregistry.GlobalTypes
	} else {
		h.typeResolver = options.TypeResolver
	}

	return h, nil
}

var _ signing.SignModeHandler = SignModeHandler{}

// Mode implements signing.SignModeHandler.Mode.
func (h SignModeHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX
}

// getFirstSigner returns the first signer from the first message in the tx. It replicates behavior in
// https://github.com/cosmos/cosmos-sdk/blob/4a6a1e3cb8de459891cb0495052589673d14ef51/x/auth/tx/builder.go#L142
func (h SignModeHandler) getFirstSigner(txData signing.TxData) ([]byte, error) {
	if len(txData.Body.Messages) == 0 {
		return nil, errors.New("no signer found")
	}

	msg, err := anyutil.Unpack(txData.Body.Messages[0], h.fileResolver, h.typeResolver)
	if err != nil {
		return nil, err
	}
	signer, err := h.signersContext.GetSigners(msg)
	if err != nil {
		return nil, err
	}
	return signer[0], nil
}

// GetSignBytes implements signing.SignModeHandler.GetSignBytes.
func (h SignModeHandler) GetSignBytes(
	_ context.Context, signerData signing.SignerData, txData signing.TxData,
) ([]byte, error) {
	feePayer := txData.AuthInfo.Fee.Payer
	if feePayer == "" {
		fp, err := h.getFirstSigner(txData)
		if err != nil {
			return nil, err
		}
		feePayer = hex.EncodeToString(fp)
	}
	if sameHexAddress(feePayer, signerData.Address) {
		return nil, fmt.Errorf("fee payer %s cannot sign with %s: unauthorized",
			feePayer, signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX)
	}

	signDocDirectAux := &txv1beta1.SignDocDirectAux{
		BodyBytes:     txData.BodyBytes,
		PublicKey:     signerData.PubKey,
		ChainId:       signerData.ChainID,
		AccountNumber: signerData.AccountNumber,
		Sequence:      signerData.Sequence,
	}

	protov2MarshalOpts := proto.MarshalOptions{Deterministic: true}
	return protov2MarshalOpts.Marshal(signDocDirectAux)
}

// sameHexAddress reports whether a and b are the same 20-byte hex-encoded
// address, tolerating an optional "0x"/"0X" prefix and any letter case on
// either side. If either side isn't valid hex, it falls back to an exact
// string match rather than declaring them unequal — the fee-payer guard
// this feeds should stay at least as strict as a plain string comparison
// for input it can't normalize, not more permissive.
func sameHexAddress(a, b string) bool {
	aBytes, aErr := hex.DecodeString(strings.TrimPrefix(strings.TrimPrefix(a, "0x"), "0X"))
	bBytes, bErr := hex.DecodeString(strings.TrimPrefix(strings.TrimPrefix(b, "0x"), "0X"))
	if aErr != nil || bErr != nil {
		return a == b
	}
	return bytes.Equal(aBytes, bBytes)
}
