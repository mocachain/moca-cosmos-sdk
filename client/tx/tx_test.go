package tx

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func newTestTxConfig() (client.TxConfig, codec.Codec) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig()
	return authtx.NewTxConfig(codec.NewProtoCodec(encodingConfig.InterfaceRegistry), authtx.DefaultSignModes), encodingConfig.Codec
}

// mockContext is a mock client.Context to return abitrary simulation response, used to
// unit test CalculateGas.
type mockContext struct {
	gasUsed uint64
	wantErr bool
}

func (m mockContext) Invoke(_ context.Context, _ string, _, reply any, _ ...grpc.CallOption) (err error) {
	if m.wantErr {
		return fmt.Errorf("mock err")
	}

	*(reply.(*txtypes.SimulateResponse)) = txtypes.SimulateResponse{
		GasInfo: &sdk.GasInfo{GasUsed: m.gasUsed, GasWanted: m.gasUsed},
		Result:  &sdk.Result{Data: []byte("tx data"), Log: "log"},
	}

	return nil
}

func (mockContext) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("not implemented")
}

func TestCalculateGas(t *testing.T) {
	type args struct {
		mockGasUsed uint64
		mockWantErr bool
		adjustment  float64
	}

	testCases := []struct {
		name         string
		args         args
		wantEstimate uint64
		wantAdjusted uint64
		expPass      bool
	}{
		{"error", args{0, true, 1.2}, 0, 0, false},
		{"adjusted gas", args{10, false, 1.2}, 10, 12, true},
	}

	for _, tc := range testCases {
		txCfg, _ := newTestTxConfig()
		defaultSignMode, err := signing.APISignModeToInternal(txCfg.SignModeHandler().DefaultMode())
		require.NoError(t, err)

		txf := Factory{}.
			WithChainID(sdktestutil.DefaultChainId).
			WithTxConfig(txCfg).WithSignMode(defaultSignMode)

		t.Run(tc.name, func(t *testing.T) {
			mockClientCtx := mockContext{
				gasUsed: tc.args.mockGasUsed,
				wantErr: tc.args.mockWantErr,
			}
			simRes, gotAdjusted, err := CalculateGas(mockClientCtx, txf.WithGasAdjustment(tc.args.adjustment))
			if tc.expPass {
				require.NoError(t, err)
				require.Equal(t, simRes.GasInfo.GasUsed, tc.wantEstimate)
				require.Equal(t, gotAdjusted, tc.wantAdjusted)
				require.NotNil(t, simRes.Result)
			} else {
				require.Error(t, err)
				require.Nil(t, simRes)
			}
		})
	}
}

// mockCometRPC is a mock CometBFT RPC client that answers every ABCI query with a
// simulation response carrying the given gas info and records the broadcast tx,
// used to unit test BroadcastTx.
type mockCometRPC struct {
	rpcclientmock.Client

	gasInfo sdk.GasInfo
	txBytes []byte
}

func (m *mockCometRPC) ABCIQueryWithOptions(
	context.Context, string, cmtbytes.HexBytes, rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	bz, err := (&txtypes.SimulateResponse{GasInfo: &m.gasInfo, Result: &sdk.Result{}}).Marshal()
	if err != nil {
		return nil, err
	}

	return &coretypes.ResultABCIQuery{Response: abci.ResponseQuery{Value: bz}}, nil
}

func (m *mockCometRPC) BroadcastTxSync(_ context.Context, tx cmttypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	m.txBytes = tx
	return &coretypes.ResultBroadcastTx{}, nil
}

func TestBroadcastTxGasAutoFees(t *testing.T) {
	txCfg, cdc := newTestTxConfig()
	banktypes.RegisterInterfaces(cdc.InterfaceRegistry()) // to decode the broadcast tx
	kb, err := keyring.New(t.Name(), "test", t.TempDir(), nil, cdc)
	require.NoError(t, err)

	path := hd.CreateHDPath(118, 0, 0).String()
	k, _, err := kb.NewMnemonic("test_key1", keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	from, err := k.GetAddress()
	require.NoError(t, err)

	// The node simulates 100000 gas, which the 1.5 gas adjustment turns into a
	// 150000 gas limit. As upstream, the min gas price the node reports never
	// sets the fee.
	const nodeMinGasPrice = "25.000000000000000000stake"

	testCases := []struct {
		name        string
		fees        string
		gasPrices   string
		minGasPrice string
		expFees     string
	}{
		{
			name:        "fees are kept",
			fees:        "7stake",
			minGasPrice: nodeMinGasPrice,
			expFees:     "7stake",
		},
		{
			name:        "gas prices apply to the adjusted gas limit",
			gasPrices:   "2stake",
			minGasPrice: nodeMinGasPrice,
			expFees:     "300000stake",
		},
		{
			name:        "no fees or gas prices sets no fee",
			minGasPrice: nodeMinGasPrice,
		},
		{
			name:    "fees with an empty node min gas price",
			fees:    "7stake",
			expFees: "7stake",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := &mockCometRPC{gasInfo: sdk.GasInfo{GasUsed: 100000, MinGasPrice: tc.minGasPrice}}
			clientCtx := client.Context{}.
				WithClient(node).
				WithCodec(cdc).
				WithTxConfig(txCfg).
				WithKeyring(kb).
				WithFromName("test_key1").
				WithFromAddress(from).
				WithChainID(sdktestutil.DefaultChainId).
				WithBroadcastMode(flags.BroadcastSync).
				WithSkipConfirmation(true).
				WithOutput(io.Discard)

			txf := Factory{}.
				WithTxConfig(txCfg).
				WithAccountRetriever(client.MockAccountRetriever{ReturnAccNum: 50, ReturnAccSeq: 23}).
				WithKeybase(kb).
				WithFromName("test_key1").
				WithChainID(sdktestutil.DefaultChainId).
				WithSignMode(signingtypes.SignMode_SIGN_MODE_DIRECT).
				WithSimulateAndExecute(true).
				WithGasAdjustment(1.5).
				WithFees(tc.fees).
				WithGasPrices(tc.gasPrices)

			err := BroadcastTx(clientCtx, txf, banktypes.NewMsgSend(from, sdk.AccAddress("to"), nil))
			require.NoError(t, err)

			tx, err := txCfg.TxDecoder()(node.txBytes)
			require.NoError(t, err)
			require.Equal(t, tc.expFees, tx.(sdk.FeeTx).GetFee().String())
			require.Equal(t, uint64(150000), tx.(sdk.FeeTx).GetGas())
		})
	}
}

func mockTxFactory(txCfg client.TxConfig) Factory {
	return Factory{}.
		WithTxConfig(txCfg).
		WithAccountNumber(50).
		WithSequence(23).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID("test-chain")
}

func TestBuildSimTx(t *testing.T) {
	txCfg, cdc := newTestTxConfig()
	defaultSignMode, err := signing.APISignModeToInternal(txCfg.SignModeHandler().DefaultMode())
	require.NoError(t, err)

	kb, err := keyring.New(t.Name(), "test", t.TempDir(), nil, cdc)
	require.NoError(t, err)

	path := hd.CreateHDPath(118, 0, 0).String()
	_, _, err = kb.NewMnemonic("test_key1", keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)

	txf := mockTxFactory(txCfg).WithSignMode(defaultSignMode).WithKeybase(kb).WithChainID(sdktestutil.DefaultChainId)
	msg := banktypes.NewMsgSend(sdk.AccAddress("from"), sdk.AccAddress("to"), nil)
	bz, err := txf.BuildSimTx(msg)
	require.NoError(t, err)
	require.NotNil(t, bz)
}

func TestBuildUnsignedTx(t *testing.T) {
	txConfig, cdc := newTestTxConfig()
	kb, err := keyring.New(t.Name(), "test", t.TempDir(), nil, cdc)
	require.NoError(t, err)

	path := hd.CreateHDPath(118, 0, 0).String()

	_, _, err = kb.NewMnemonic("test_key1", keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)

	txf := mockTxFactory(txConfig).WithKeybase(kb).WithChainID(sdktestutil.DefaultChainId)
	msg := banktypes.NewMsgSend(sdk.AccAddress("from"), sdk.AccAddress("to"), nil)
	tx, err := txf.BuildUnsignedTx(msg)
	require.NoError(t, err)
	require.NotNil(t, tx)

	sigs, err := tx.GetTx().(signing.SigVerifiableTx).GetSignaturesV2()
	require.NoError(t, err)
	require.Empty(t, sigs)
}

func TestBuildUnsignedTxWithWithExtensionOptions(t *testing.T) {
	txCfg := moduletestutil.MakeBuilderTestTxConfig()
	extOpts := []*codectypes.Any{
		{
			TypeUrl: "/test",
			Value:   []byte("test"),
		},
	}
	txf := mockTxFactory(txCfg).WithExtensionOptions(extOpts...)
	msg := banktypes.NewMsgSend(sdk.AccAddress("from"), sdk.AccAddress("to"), nil)
	tx, err := txf.BuildUnsignedTx(msg)
	require.NoError(t, err)
	require.NotNil(t, tx)
	txb := tx.(*moduletestutil.TestTxBuilder)
	require.Equal(t, extOpts, txb.ExtOptions)
}

func TestMnemonicInMemo(t *testing.T) {
	txConfig, cdc := newTestTxConfig()
	kb, err := keyring.New(t.Name(), "test", t.TempDir(), nil, cdc)
	require.NoError(t, err)

	path := hd.CreateHDPath(118, 0, 0).String()

	_, seed, err := kb.NewMnemonic("test_key1", keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)

	testCases := []struct {
		name  string
		memo  string
		error bool
	}{
		{name: "bare seed", memo: seed, error: true},
		{name: "padding bare seed", memo: fmt.Sprintf("   %s", seed), error: true},
		{name: "prefixed", memo: fmt.Sprintf("%s: %s", "prefixed: ", seed), error: false},
		{name: "normal memo", memo: "this is a memo", error: false},
		{name: "empty memo", memo: "", error: false},
		{name: "invalid mnemonic", memo: strings.Repeat("egg", 24), error: false},
		{name: "caps", memo: strings.ToUpper(seed), error: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txf := Factory{}.
				WithTxConfig(txConfig).
				WithAccountNumber(50).
				WithSequence(23).
				WithFees("50stake").
				WithMemo(tc.memo).
				WithChainID(sdktestutil.DefaultChainId).
				WithKeybase(kb)

			msg := banktypes.NewMsgSend(sdk.AccAddress("from"), sdk.AccAddress("to"), nil)
			tx, err := txf.BuildUnsignedTx(msg)
			if tc.error {
				require.Error(t, err)
				require.ErrorContains(t, err, "mnemonic")
				require.Nil(t, tx)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tx)
			}
		})
	}
}

func TestSign(t *testing.T) {
	txConfig, cdc := newTestTxConfig()
	requireT := require.New(t)
	path := hd.CreateHDPath(118, 0, 0).String()
	kb, err := keyring.New(t.Name(), "test", t.TempDir(), nil, cdc)
	requireT.NoError(err)

	from1 := "test_key1"
	from2 := "test_key2"

	// create a new key using a mnemonic generator and test if we can reuse seed to recreate that account
	_, seed, err := kb.NewMnemonic(from1, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	requireT.NoError(err)
	requireT.NoError(kb.Delete(from1))
	k1, _, err := kb.NewMnemonic(from1, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	requireT.NoError(err)

	k2, err := kb.NewAccount(from2, seed, "", path, hd.Secp256k1)
	requireT.NoError(err)

	pubKey1, err := k1.GetPubKey()
	requireT.NoError(err)
	pubKey2, err := k2.GetPubKey()
	requireT.NoError(err)
	requireT.NotEqual(pubKey1.Bytes(), pubKey2.Bytes())
	t.Log("Pub keys:", pubKey1, pubKey2)

	txfNoKeybase := mockTxFactory(txConfig).WithChainID(sdktestutil.DefaultChainId)
	txfDirect := txfNoKeybase.
		WithKeybase(kb).
		WithSignMode(signingtypes.SignMode_SIGN_MODE_DIRECT)
	txfAmino := txfDirect.
		WithSignMode(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	addr1, err := k1.GetAddress()
	requireT.NoError(err)
	addr2, err := k2.GetAddress()
	requireT.NoError(err)
	msg1 := banktypes.NewMsgSend(addr1, sdk.AccAddress("to"), nil)
	msg2 := banktypes.NewMsgSend(addr2, sdk.AccAddress("to"), nil)
	txb, err := txfNoKeybase.BuildUnsignedTx(msg1, msg2)
	requireT.NoError(err)
	txb2, err := txfNoKeybase.BuildUnsignedTx(msg1, msg2)
	requireT.NoError(err)
	txbSimple, err := txfNoKeybase.BuildUnsignedTx(msg2)
	requireT.NoError(err)

	testCases := []struct {
		name         string
		txf          Factory
		txb          client.TxBuilder
		from         string
		overwrite    bool
		expectedPKs  []cryptotypes.PubKey
		matchingSigs []int // if not nil, check matching signature against old ones.
	}{
		{
			"should fail if txf without keyring",
			txfNoKeybase, txb, from1, true, nil, nil,
		},
		{
			"should fail for non existing key",
			txfAmino, txb, "unknown", true, nil, nil,
		},
		{
			"amino: should succeed with keyring",
			txfAmino, txbSimple, from1, true,
			[]cryptotypes.PubKey{pubKey1},
			nil,
		},
		{
			"direct: should succeed with keyring",
			txfDirect, txbSimple, from1, true,
			[]cryptotypes.PubKey{pubKey1},
			nil,
		},

		/**** test double sign Amino mode ****/
		{
			"amino: should sign multi-signers tx",
			txfAmino, txb, from1, true,
			[]cryptotypes.PubKey{pubKey1},
			nil,
		},
		{
			"amino: should append a second signature and not overwrite",
			txfAmino, txb, from2, false,
			[]cryptotypes.PubKey{pubKey1, pubKey2},
			[]int{0, 0},
		},
		{
			"amino: should overwrite a signature",
			txfAmino, txb, from2, true,
			[]cryptotypes.PubKey{pubKey2},
			[]int{1, 0},
		},

		/**** test double sign Direct mode
		  signing transaction with 2 or more DIRECT signers should fail in DIRECT mode ****/
		{
			"direct: should  append a DIRECT signature with existing AMINO",
			// txb already has 1 AMINO signature
			txfDirect, txb, from1, false,
			[]cryptotypes.PubKey{pubKey2, pubKey1},
			nil,
		},
		{
			"direct: should add single DIRECT sig in multi-signers tx",
			txfDirect, txb2, from1, false,
			[]cryptotypes.PubKey{pubKey1},
			nil,
		},
		{
			"direct: should fail to append 2nd DIRECT sig in multi-signers tx",
			txfDirect, txb2, from2, false,
			[]cryptotypes.PubKey{},
			nil,
		},
		{
			"amino: should append 2nd AMINO sig in multi-signers tx with 1 DIRECT sig",
			// txb2 already has 1 DIRECT signature
			txfAmino, txb2, from2, false,
			[]cryptotypes.PubKey{},
			nil,
		},
		{
			"direct: should overwrite multi-signers tx with DIRECT sig",
			txfDirect, txb2, from1, true,
			[]cryptotypes.PubKey{pubKey1},
			nil,
		},
	}

	var prevSigs []signingtypes.SignatureV2
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err = Sign(context.TODO(), tc.txf, tc.from, tc.txb, tc.overwrite)
			if len(tc.expectedPKs) == 0 {
				requireT.Error(err)
			} else {
				requireT.NoError(err)
				sigs := testSigners(requireT, tc.txb.GetTx(), tc.expectedPKs...)
				if tc.matchingSigs != nil {
					requireT.Equal(prevSigs[tc.matchingSigs[0]], sigs[tc.matchingSigs[1]])
				}
				prevSigs = sigs
			}
		})
	}
}

func TestPreprocessHook(t *testing.T) {
	_, _, addr2 := testdata.KeyTestPubAddr()

	txConfig, cdc := newTestTxConfig()
	requireT := require.New(t)
	path := hd.CreateHDPath(118, 0, 0).String()
	kb, err := keyring.New(t.Name(), "test", t.TempDir(), nil, cdc)
	requireT.NoError(err)

	from := "test_key"
	kr, _, err := kb.NewMnemonic(from, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	requireT.NoError(err)

	extVal := &testdata.Cat{
		Moniker: "einstein",
		Lives:   9,
	}
	extAny, err := codectypes.NewAnyWithValue(extVal)
	requireT.NoError(err)

	preprocessHook := client.PreprocessTxFn(func(chainID string, key keyring.KeyType, tx client.TxBuilder) error {
		extensionBuilder, ok := tx.(authtx.ExtensionOptionsTxBuilder)
		requireT.True(ok)

		// Set new extension and tip
		extensionBuilder.SetExtensionOptions(extAny)

		return nil
	})

	txfDirect := mockTxFactory(txConfig).
		WithKeybase(kb).
		WithSignMode(signingtypes.SignMode_SIGN_MODE_DIRECT).
		WithPreprocessTxHook(preprocessHook).
		WithChainID(sdktestutil.DefaultChainId)

	addr1, err := kr.GetAddress()
	requireT.NoError(err)
	msg1 := banktypes.NewMsgSend(addr1, sdk.AccAddress("to"), nil)
	msg2 := banktypes.NewMsgSend(addr2, sdk.AccAddress("to"), nil)
	txb, err := txfDirect.BuildUnsignedTx(msg1, msg2)
	requireT.NoError(err)

	err = Sign(context.TODO(), txfDirect, from, txb, false)
	requireT.NoError(err)

	// Run preprocessing
	err = txfDirect.PreprocessTx(from, txb)
	requireT.NoError(err)

	hasExtOptsTx, ok := txb.(ante.HasExtensionOptionsTx)
	requireT.True(ok)

	hasOneExt := len(hasExtOptsTx.GetExtensionOptions()) == 1
	requireT.True(hasOneExt)

	opt := hasExtOptsTx.GetExtensionOptions()[0]
	requireT.Equal(opt, extAny)
}

func testSigners(require *require.Assertions, tr signing.Tx, pks ...cryptotypes.PubKey) []signingtypes.SignatureV2 {
	sigs, err := tr.GetSignaturesV2()
	require.Len(sigs, len(pks))
	require.NoError(err)
	require.Len(sigs, len(pks))
	for i := range pks {
		require.True(sigs[i].PubKey.Equals(pks[i]), "Signature is signed with a wrong pubkey. Got: %s, expected: %s", sigs[i].PubKey, pks[i])
	}
	return sigs
}
