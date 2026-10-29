package signing_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	slashingv1beta1 "cosmossdk.io/api/cosmos/slashing/v1beta1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// TestEIP712DynamicMsgSignsAsGogoproto checks that a dynamicpb message, as autocli builds it, gets
// the same EIP-712 sign bytes as its gogoproto type, and that a signature over the built tx
// verifies against the decoded tx.
func TestEIP712DynamicMsgSignsAsGogoproto(t *testing.T) {
	enc := moduletestutil.MakeTestEncodingConfig()
	slashingtypes.RegisterInterfaces(enc.InterfaceRegistry)
	env := eip712Env{txConfig: enc.TxConfig, key: ethsecp256k1.GenPrivKeyFromSecret([]byte("eip712-dynamicpb-fixtures"))}

	validator := fixtureAddr(1).String()
	dyn := dynamicpb.NewMessage((&slashingv1beta1.MsgUnjail{}).ProtoReflect().Descriptor())
	dyn.Set(dyn.Descriptor().Fields().ByName("validator_addr"), protoreflect.ValueOfString(validator))

	gogoBuilt, _ := env.txs(t, slashingtypes.NewMsgUnjail(validator))
	want, err := adapterHash(gogoBuilt)
	require.NoError(t, err)

	built, decoded := env.txs(t, dyn)
	for _, r := range hashAll(built, decoded) {
		require.NoError(t, r.err, r.name)
		require.Equal(t, want, r.hash, r.name)
	}

	sig := env.sign(t, built)
	require.NoError(t, env.verify(decoded, sig))
}
