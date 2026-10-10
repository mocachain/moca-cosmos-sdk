package signing_test

import (
	"context"
	"math/big"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	txsigning "cosmossdk.io/x/tx/signing"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

func TestGetSignBytesAdapterNoPublicKey(t *testing.T) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig()
	txConfig := encodingConfig.TxConfig
	_, _, addr := testdata.KeyTestPubAddrEthSecp256k1(t)
	signerData := authsign.SignerData{
		Address:       addr.String(),
		ChainID:       sdktestutil.DefaultChainId,
		AccountNumber: 11,
		Sequence:      15,
	}
	w := txConfig.NewTxBuilder()
	w.SetFeePayer(addr)
	_, err := authsign.GetSignBytesAdapter(
		context.Background(),
		txConfig.SignModeHandler(),
		signing.SignMode_SIGN_MODE_DIRECT,
		signerData,
		w.GetTx())
	require.NoError(t, err)
}

// TestEIP712VerifySignatureRejectsMalleatedSignature proves that
// EIP712VerifySignature accepts only the canonical low-S form of a
// signature. A signature (r, s, v) and its malleated twin (r, N-s, v^1)
// both recover to the same pubkey; only one of the two must verify.
func TestEIP712VerifySignatureRejectsMalleatedSignature(t *testing.T) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig()
	txConfig := encodingConfig.TxConfig
	privKey, pubKey, addr := testdata.KeyTestPubAddrEthSecp256k1(t)

	signerData := txsigning.SignerData{
		Address:       addr.String(),
		ChainID:       sdktestutil.DefaultChainId,
		AccountNumber: 11,
		Sequence:      15,
	}
	w := txConfig.NewTxBuilder()
	w.SetFeePayer(addr)
	tx := w.GetTx()

	sigHash, err := authsign.EIP712GetSignBytes(context.Background(), signerData, tx)
	require.NoError(t, err)

	sig, err := privKey.Sign(sigHash)
	require.NoError(t, err)
	require.Len(t, sig, 65)

	// Sanity check: the honestly produced (low-S) signature verifies.
	honest := append([]byte(nil), sig...)
	err = authsign.EIP712VerifySignature(context.Background(), signerData, tx, honest, pubKey)
	require.NoError(t, err, "the original low-S signature must verify")

	// Malleate: s' = N - s, v' = v ^ 1. (r, s', v') recovers to the same
	// pubkey as (r, s, v), so it is a second valid signature for the same tx.
	malleated := append([]byte(nil), sig...)
	curveOrder := ethcrypto.S256().Params().N
	s := new(big.Int).SetBytes(malleated[32:64])
	sPrime := new(big.Int).Sub(curveOrder, s)
	sPrimeBytes := sPrime.Bytes()
	for i := 32; i < 64; i++ {
		malleated[i] = 0
	}
	copy(malleated[64-len(sPrimeBytes):64], sPrimeBytes)
	malleated[64] ^= 1

	err = authsign.EIP712VerifySignature(context.Background(), signerData, tx, malleated, pubKey)
	require.Error(t, err, "the malleated high-S signature must be rejected")
}
