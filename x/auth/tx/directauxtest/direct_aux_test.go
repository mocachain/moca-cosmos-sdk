// Package directauxtest holds the regression test for x/auth/tx's
// SIGN_MODE_DIRECT_AUX fee-payer guard. It lives in its own package because
// x/auth/tx's own test binary does not currently build (a pre-existing,
// unrelated internal-test import cycle), so the guard is exercised through the
// package's exported surface instead.
package directauxtest

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// TestSignModeDirectAuxHandler_FeePayerNonCanonicalSpelling pins the fee-payer
// guard in x/auth/tx.SignModeDirectAuxHandler.GetSignBytes: a fee payer and
// signer that decode to the same address must be rejected regardless of which
// valid spelling (letter case, optional "0x" prefix) each side uses. This is
// the x/auth/tx twin of the x/tx/signing/directaux end-to-end test.
func TestSignModeDirectAuxHandler_FeePayerNonCanonicalSpelling(t *testing.T) {
	cdc := codectestutil.CodecOptions{}.NewCodec()
	testdata.RegisterInterfaces(cdc.InterfaceRegistry())
	_, pubKey, _ := testdata.KeyTestPubAddr()

	feePayer := sdk.AccAddress(bytes.Repeat([]byte{0xAB}, 20))
	otherAddr := sdk.AccAddress(bytes.Repeat([]byte{0xCD}, 20))

	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	buildTx := func() sdk.Tx {
		b := txConfig.NewTxBuilder()
		require.NoError(t, b.SetMsgs(testdata.NewTestMsg(feePayer)))
		b.SetGasLimit(20000)
		b.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("atom", 150)))
		b.SetFeePayer(feePayer)
		return b.GetTx()
	}

	handler := authtx.SignModeDirectAuxHandler{}

	// Same account as the fee payer, spelled lowercase and without the "0x"
	// prefix — must still be rejected. A raw string / EqualFold comparison
	// would miss the prefix difference and let the fee payer sign.
	nonCanonical := strings.ToLower(strings.TrimPrefix(feePayer.String(), "0x"))
	require.NotEqual(t, feePayer.String(), nonCanonical, "sanity check: fixture must differ in spelling from the canonical form")

	_, err := handler.GetSignBytes(
		signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
		authsigning.SignerData{
			Address:       nonCanonical,
			ChainID:       "moca_2288-1",
			AccountNumber: 1,
			Sequence:      2,
			PubKey:        pubKey,
		},
		buildTx(),
	)
	require.ErrorContains(t, err, "cannot sign with")

	// A different account must not trip the guard — sign bytes are produced.
	signBytes, err := handler.GetSignBytes(
		signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
		authsigning.SignerData{
			Address:       otherAddr.String(),
			ChainID:       "moca_2288-1",
			AccountNumber: 1,
			Sequence:      2,
			PubKey:        pubKey,
		},
		buildTx(),
	)
	require.NoError(t, err)
	require.NotEmpty(t, signBytes)
}
