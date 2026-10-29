package signing_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// TestEIP712TypesFieldByProtobufName checks that a field whose json tag differs from its protobuf
// name (MsgUnjail's validator_addr, json tag "address") is typed under its protobuf name and that
// the signature covers its value.
func TestEIP712TypesFieldByProtobufName(t *testing.T) {
	enc := moduletestutil.MakeTestEncodingConfig()
	slashingtypes.RegisterInterfaces(enc.InterfaceRegistry)
	env := eip712Env{txConfig: enc.TxConfig, key: ethsecp256k1.GenPrivKeyFromSecret([]byte("eip712-name-fixtures"))}

	unjail := func(b byte) sdk.Msg { return slashingtypes.NewMsgUnjail(fixtureAddr(b).String()) }

	built, decoded := env.txs(t, unjail(1))
	results := hashAll(built, decoded)
	for _, r := range results {
		require.NoError(t, r.err, r.name)
		require.Equal(t, results[0].hash, r.hash, r.name)
	}

	chainID, err := sdk.ParseChainID(fixtureChainID)
	require.NoError(t, err)
	sd := signerData()
	msgTypes, _, err := authtx.GetMsgTypes(authsign.SignerData{Address: sd.Address, ChainID: sd.ChainID, AccountNumber: sd.AccountNumber, Sequence: sd.Sequence}, decoded, chainID)
	require.NoError(t, err)
	require.Contains(t, msgTypes["Msg1"], apitypes.Type{Name: "validator_addr", Type: "string"})
	require.NotContains(t, msgTypes["Msg1"], apitypes.Type{Name: "address", Type: "string"})

	sig := env.sign(t, decoded)
	require.NoError(t, env.verify(decoded, sig))

	_, otherValidator := env.txs(t, unjail(2))
	require.Error(t, env.verify(otherValidator, sig), "a signature over one validator's unjail must not verify for another's")
}
