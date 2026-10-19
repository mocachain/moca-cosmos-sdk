package ante_test

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// hexAddressCodec resolves the 0x-hex addresses that moca accounts render as;
// the suite's default encoding config can only decode bech32 signers.
type hexAddressCodec struct{}

func (hexAddressCodec) StringToBytes(text string) ([]byte, error) {
	return sdk.AccAddressFromHexUnsafe(text)
}

func (hexAddressCodec) BytesToString(bz []byte) (string, error) {
	return sdk.AccAddress(bz).String(), nil
}

// setupSigDecoratorsAtHeightZero returns a suite at block height 0 whose tx
// config signs with EIP-712, and an eth_secp256k1 signer with a non-zero
// account number and no public key set. The tx builder holds a message from
// that signer.
func setupSigDecoratorsAtHeightZero(t *testing.T, isCheckTx bool) (*AnteTestSuite, TestAccount) {
	t.Helper()

	suite := SetupTestSuite(t, isCheckTx)
	suite.ctx = suite.ctx.WithBlockHeight(0)

	registry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: txsigning.Options{
			AddressCodec:          hexAddressCodec{},
			ValidatorAddressCodec: hexAddressCodec{},
		},
	})
	require.NoError(t, err)
	std.RegisterInterfaces(registry)
	testdata.RegisterInterfaces(registry)
	suite.clientCtx.TxConfig, err = authtx.NewTxConfigWithOptions(codec.NewProtoCodec(registry), authtx.ConfigOptions{
		EnabledSignModes: []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712},
		SigningContext:   registry.SigningContext(),
	})
	require.NoError(t, err)
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	priv, _, addr := testdata.KeyTestPubAddrEthSecp256k1(t)
	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
	require.NoError(t, acc.SetAccountNumber(1000))
	suite.accountKeeper.SetAccount(suite.ctx, acc)

	require.NoError(t, suite.txBuilder.SetMsgs(testdata.NewTestMsg(addr)))
	suite.txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
	suite.txBuilder.SetGasLimit(testdata.NewTestGasLimit())

	return suite, TestAccount{acc, priv}
}

// sigDecorators chains the decorators that skip their checks while the genesis
// block is executed.
func sigDecorators(suite *AnteTestSuite, sigGasConsumer ante.SignatureVerificationGasConsumer) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		ante.NewSetPubKeyDecorator(suite.accountKeeper),
		ante.NewSigGasConsumeDecorator(suite.accountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(suite.accountKeeper, suite.clientCtx.TxConfig.SignModeHandler()),
	)
}

// signEIP712Tx signs the suite's tx builder with priv over accNum and accSeq.
func signEIP712Tx(t *testing.T, suite *AnteTestSuite, priv cryptotypes.PrivKey, accNum, accSeq uint64) sdk.Tx {
	t.Helper()
	tx, err := suite.CreateTestTx(suite.ctx, []cryptotypes.PrivKey{priv}, []uint64{accNum}, []uint64{accSeq}, suite.ctx.ChainID(), signing.SignMode_SIGN_MODE_EIP_712)
	require.NoError(t, err)
	return tx
}

type heightZeroTxCase struct {
	name   string
	tx     func(t *testing.T, suite *AnteTestSuite, signer TestAccount) sdk.Tx
	expErr error
}

// heightZeroTxCases are txs that the signature decorators reject at any height
// outside genesis block execution.
func heightZeroTxCases(t *testing.T) []heightZeroTxCase {
	t.Helper()
	otherPriv, _, _ := testdata.KeyTestPubAddrEthSecp256k1(t)

	return []heightZeroTxCase{
		{
			"signature without signer info",
			func(t *testing.T, suite *AnteTestSuite, _ TestAccount) sdk.Tx {
				t.Helper()
				bz, err := suite.clientCtx.TxConfig.TxEncoder()(suite.txBuilder.GetTx())
				require.NoError(t, err)
				var raw txtypes.TxRaw
				require.NoError(t, raw.Unmarshal(bz))
				raw.Signatures = [][]byte{[]byte("placeholder")}
				bz, err = raw.Marshal()
				require.NoError(t, err)
				tx, err := suite.clientCtx.TxConfig.TxDecoder()(bz)
				require.NoError(t, err)
				return tx
			},
			sdkerrors.ErrTxDecode,
		},
		{
			"empty signature",
			func(t *testing.T, suite *AnteTestSuite, signer TestAccount) sdk.Tx {
				t.Helper()
				require.NoError(t, suite.txBuilder.SetSignatures(signing.SignatureV2{
					PubKey: signer.priv.PubKey(),
					Data:   &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_EIP_712},
				}))
				return suite.txBuilder.GetTx()
			},
			sdkerrors.ErrUnauthorized,
		},
		{
			"signed with account number 0",
			func(t *testing.T, suite *AnteTestSuite, signer TestAccount) sdk.Tx {
				t.Helper()
				return signEIP712Tx(t, suite, signer.priv, 0, 0)
			},
			sdkerrors.ErrUnauthorized,
		},
		{
			"signer info for another key",
			func(t *testing.T, suite *AnteTestSuite, signer TestAccount) sdk.Tx {
				t.Helper()
				return signEIP712Tx(t, suite, otherPriv, signer.acc.GetAccountNumber(), 0)
			},
			sdkerrors.ErrInvalidPubKey,
		},
		{
			"wrong sequence",
			func(t *testing.T, suite *AnteTestSuite, signer TestAccount) sdk.Tx {
				t.Helper()
				return signEIP712Tx(t, suite, signer.priv, signer.acc.GetAccountNumber(), 1)
			},
			sdkerrors.ErrWrongSequence,
		},
	}
}

func TestSigDecoratorsAtHeightZero_CheckTxRejectsInvalidTx(t *testing.T) {
	for _, tc := range heightZeroTxCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			suite, signer := setupSigDecoratorsAtHeightZero(t, true)
			_, err := sigDecorators(suite, ante.DefaultSigVerificationGasConsumer)(suite.ctx, tc.tx(t, suite, signer), false)
			require.ErrorIs(t, err, tc.expErr)
		})
	}

	// ReCheckTx and simulate do not verify the signature itself, but still
	// check the sequence.
	for _, mode := range []struct {
		name     string
		recheck  bool
		simulate bool
	}{
		{"recheck", true, false},
		{"simulate", false, true},
	} {
		t.Run("wrong sequence on "+mode.name, func(t *testing.T) {
			suite, signer := setupSigDecoratorsAtHeightZero(t, true)
			ctx := suite.ctx
			if mode.recheck {
				ctx = ctx.WithIsReCheckTx(true)
			}
			tx := signEIP712Tx(t, suite, signer.priv, signer.acc.GetAccountNumber(), 1)
			_, err := sigDecorators(suite, ante.DefaultSigVerificationGasConsumer)(ctx, tx, mode.simulate)
			require.ErrorIs(t, err, sdkerrors.ErrWrongSequence)
		})
	}
}

func TestSigDecoratorsAtHeightZero_CheckTxAcceptsSignedTx(t *testing.T) {
	suite, signer := setupSigDecoratorsAtHeightZero(t, true)
	require.NotZero(t, signer.acc.GetAccountNumber())
	tx := signEIP712Tx(t, suite, signer.priv, signer.acc.GetAccountNumber(), 0)

	var charged []cryptotypes.PubKey
	sigGasConsumer := func(meter storetypes.GasMeter, sig signing.SignatureV2, params types.Params) error {
		charged = append(charged, sig.PubKey)
		return ante.DefaultSigVerificationGasConsumer(meter, sig, params)
	}
	_, err := sigDecorators(suite, sigGasConsumer)(suite.ctx, tx, false)
	require.NoError(t, err)

	pk, err := suite.accountKeeper.GetPubKey(suite.ctx, signer.acc.GetAddress())
	require.NoError(t, err)
	require.NotNil(t, pk, "the public key must be set on the account")
	require.True(t, signer.priv.PubKey().Equals(pk))
	require.Len(t, charged, 1, "the signature verification gas must be charged")
	require.True(t, pk.Equals(charged[0]))
}

func TestSigDecoratorsAtHeightZero_GenesisBlockExecSkipsChecks(t *testing.T) {
	// InitChain delivers the genesis txs at height 0 outside CheckTx. The
	// decorators pass them through without reading or writing any state.
	for _, tc := range heightZeroTxCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			suite, signer := setupSigDecoratorsAtHeightZero(t, false)
			ctx := suite.ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
			newCtx, err := sigDecorators(suite, ante.DefaultSigVerificationGasConsumer)(ctx, tc.tx(t, suite, signer), false)
			require.NoError(t, err)
			require.Zero(t, newCtx.GasMeter().GasConsumed())

			pk, err := suite.accountKeeper.GetPubKey(suite.ctx, signer.acc.GetAddress())
			require.NoError(t, err)
			require.Nil(t, pk)
		})
	}
}
