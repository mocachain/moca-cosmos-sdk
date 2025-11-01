package signing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

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
