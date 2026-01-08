package signing

import (
	"context"
	"fmt"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	// errorsmod "cosmossdk.io/errors"
	txsigning "cosmossdk.io/x/tx/signing"
	// ethcrypto "github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/crypto/secp256k1"

	// "github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// APISignModesToInternal converts a protobuf SignMode array to a signing.SignMode array.
func APISignModesToInternal(modes []signingv1beta1.SignMode) ([]signing.SignMode, error) {
	internalModes := make([]signing.SignMode, len(modes))
	for i, mode := range modes {
		internalMode, err := APISignModeToInternal(mode)
		if err != nil {
			return nil, err
		}
		internalModes[i] = internalMode
	}
	return internalModes, nil
}

// APISignModeToInternal converts a protobuf SignMode to a signing.SignMode.
func APISignModeToInternal(mode signingv1beta1.SignMode) (signing.SignMode, error) {
	switch mode {
	case signingv1beta1.SignMode_SIGN_MODE_DIRECT:
		return signing.SignMode_SIGN_MODE_DIRECT, nil
	case signingv1beta1.SignMode_SIGN_MODE_LEGACY_AMINO_JSON:
		return signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, nil
	case signingv1beta1.SignMode_SIGN_MODE_TEXTUAL:
		return signing.SignMode_SIGN_MODE_TEXTUAL, nil
	case signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX:
		return signing.SignMode_SIGN_MODE_DIRECT_AUX, nil
	case signingv1beta1.SignMode_SIGN_MODE_EIP_712:
		return signing.SignMode_SIGN_MODE_EIP_712, nil
	default:
		return signing.SignMode_SIGN_MODE_UNSPECIFIED, fmt.Errorf("unsupported sign mode %s", mode)
	}
}

// internalSignModeToAPI converts a signing.SignMode to a protobuf SignMode.
func internalSignModeToAPI(mode signing.SignMode) (signingv1beta1.SignMode, error) {
	switch mode {
	case signing.SignMode_SIGN_MODE_DIRECT:
		return signingv1beta1.SignMode_SIGN_MODE_DIRECT, nil
	case signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON:
		return signingv1beta1.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, nil
	case signing.SignMode_SIGN_MODE_TEXTUAL:
		return signingv1beta1.SignMode_SIGN_MODE_TEXTUAL, nil
	case signing.SignMode_SIGN_MODE_DIRECT_AUX:
		return signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX, nil
	case signing.SignMode_SIGN_MODE_EIP_712:
		return signingv1beta1.SignMode_SIGN_MODE_EIP_712, nil
	default:
		return signingv1beta1.SignMode_SIGN_MODE_UNSPECIFIED, fmt.Errorf("unsupported sign mode %s", mode)
	}
}

// VerifySignature verifies a transaction signature contained in SignatureData abstracting over different signing
// modes. It differs from VerifySignature in that it uses the new txsigning.TxData interface in x/tx.
func VerifySignature(
	ctx context.Context,
	pubKey cryptotypes.PubKey,
	signerData txsigning.SignerData,
	signatureData signing.SignatureData,
	handler *txsigning.HandlerMap,
	txData txsigning.TxData,
) error {
	switch data := signatureData.(type) {
	case *signing.SingleSignatureData:
		signMode, err := internalSignModeToAPI(data.SignMode)
		if err != nil {
			return err
		}
		if data.SignMode == signing.SignMode_SIGN_MODE_EIP_712 {
			// sig := data.Signature
			// sigHash, err := handler.GetSignBytes(ctx, signMode, signerData, txData)
			// if err != nil {
			// 	return err
			// }

			// // check signature length
			// if len(sig) != ethcrypto.SignatureLength {
			// 	return errorsmod.Wrap(errorsmod.ErrorInvalidSigner, "signature length doesn't match typical [R||S||V] signature 65 bytes")
			// }

			// // remove the recovery offset if needed (ie. Metamask eip712 signature)
			// if sig[ethcrypto.RecoveryIDOffset] == 27 || sig[ethcrypto.RecoveryIDOffset] == 28 {
			// 	sig[ethcrypto.RecoveryIDOffset] -= 27
			// }

			// // recover the pubkey from the signature
			// feePayerPubkey, err := secp256k1.RecoverPubkey(sigHash, sig)
			// if err != nil {
			// 	return errorsmod.Wrap(err, "failed to recover fee payer from sig")
			// }
			// ecPubKey, err := ethcrypto.UnmarshalPubkey(feePayerPubkey)
			// if err != nil {
			// 	return errorsmod.Wrap(err, "failed to unmarshal recovered fee payer pubkey")
			// }

			// // check that the recovered pubkey matches the one in the signerData data
			// pk := &ethsecp256k1.PubKey{
			// 	Key: ethcrypto.CompressPubkey(ecPubKey),
			// }
			// if !pubKey.Equals(pk) {
			// 	return errorsmod.Wrapf(errorsmod.ErrorInvalidSigner, "feePayer's pubkey %s is different from signature's pubkey %s", pubKey, pk)
			// }
			// return nil
		}
		// This should never happen, but we add it just for test cases.
		signBytes, err := handler.GetSignBytes(ctx, signMode, signerData, txData)
		if err != nil {
			return err
		}
		if !pubKey.VerifySignature(signBytes, data.Signature) {
			return fmt.Errorf("unable to verify single signer signature")
		}
		return nil

	case *signing.MultiSignatureData:
		return fmt.Errorf("multi signature is not allowed")
	default:
		return fmt.Errorf("unexpected SignatureData %T", signatureData)
	}
}
