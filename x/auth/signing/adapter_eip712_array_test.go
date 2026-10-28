package signing_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/require"

	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/query"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
)

const (
	fixtureChainID = "moca_5151-1"
	fixtureDenom   = "amoca"
	fieldAliases   = "aliases"
)

// No SDK Msg has an omitempty list or a nullable message inside a repeated struct, so the mixed-shape
// fixtures use SDK messages that do (bank Metadata, group GroupMember, tx ModeInfo); the encoder only reflects over the Go type.

type eip712Env struct {
	txConfig client.TxConfig
	key      *ethsecp256k1.PrivKey
}

func newEIP712Env(t *testing.T) eip712Env {
	t.Helper()

	enc := moduletestutil.MakeTestEncodingConfig()
	banktypes.RegisterInterfaces(enc.InterfaceRegistry)
	govv1.RegisterInterfaces(enc.InterfaceRegistry)
	consensustypes.RegisterInterfaces(enc.InterfaceRegistry)
	// registered as Msgs so the tx decoder accepts these payloads
	enc.InterfaceRegistry.RegisterImplementations((*sdk.Msg)(nil),
		&banktypes.Metadata{}, &banktypes.QueryDenomsMetadataResponse{},
		&group.QueryGroupMembersResponse{}, &txtypes.ModeInfo{})

	return eip712Env{
		txConfig: enc.TxConfig,
		key:      ethsecp256k1.GenPrivKeyFromSecret([]byte("eip712-array-fixtures")),
	}
}

// txs returns the tx as built (CLI preview input) and as decoded from its encoding (ante handler input).
func (e eip712Env) txs(t *testing.T, msgs ...sdk.Msg) (built, decoded sdk.Tx) {
	t.Helper()

	b := e.txConfig.NewTxBuilder()
	require.NoError(t, b.SetMsgs(msgs...))
	b.SetMemo("eip712 array fixture")
	b.SetGasLimit(200000)
	b.SetTimeoutHeight(1000)
	b.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin(fixtureDenom, 150)))
	b.SetFeePayer(fixtureAddr(0x11))

	bz, err := e.txConfig.TxEncoder()(b.GetTx())
	require.NoError(t, err)
	decoded, err = e.txConfig.TxDecoder()(bz)
	require.NoError(t, err)
	return b.GetTx(), decoded
}

func (e eip712Env) sign(t *testing.T, tx sdk.Tx) []byte {
	t.Helper()

	h, err := adapterHash(tx)
	require.NoError(t, err)
	sig, err := e.key.Sign(h)
	require.NoError(t, err)
	return sig
}

func (e eip712Env) verify(tx sdk.Tx, sig []byte) error {
	return authsign.EIP712VerifySignature(context.Background(), signerData(), tx, bytes.Clone(sig), e.key.PubKey())
}

func signerData() txsigning.SignerData {
	return txsigning.SignerData{Address: fixtureAddr(0x11).String(), ChainID: fixtureChainID, AccountNumber: 7, Sequence: 3}
}

func adapterHash(tx sdk.Tx) ([]byte, error) {
	return authsign.EIP712GetSignBytes(context.Background(), signerData(), tx)
}

// twinHash runs the x/auth/tx pipeline the CLI preview uses.
func twinHash(tx sdk.Tx) ([]byte, error) {
	chainID, err := sdk.ParseChainID(fixtureChainID)
	if err != nil {
		return nil, err
	}
	sd := signerData()
	asd := authsign.SignerData{Address: sd.Address, ChainID: sd.ChainID, AccountNumber: sd.AccountNumber, Sequence: sd.Sequence}
	msgTypes, signDoc, err := authtx.GetMsgTypes(asd, tx, chainID)
	if err != nil {
		return nil, err
	}
	typedData, err := authtx.WrapTxToTypedData(signDoc, msgTypes, apitypes.TypedDataDomain{
		Name: "Moca Tx", Version: "1.0.0", ChainId: ethmath.NewHexOrDecimal256(chainID.Int64()), VerifyingContract: "moca", Salt: "0",
	})
	if err != nil {
		return nil, err
	}
	return authtx.ComputeTypedDataHash(typedData)
}

type encoderResult struct {
	name string
	hash []byte
	err  error
}

// hashAll runs both encoders over both forms of the tx.
func hashAll(built, decoded sdk.Tx) []encoderResult {
	var out []encoderResult
	for _, in := range []struct {
		name string
		tx   sdk.Tx
	}{{"built", built}, {"decoded", decoded}} {
		h, err := adapterHash(in.tx)
		out = append(out, encoderResult{"adapter/" + in.name, h, err})
		h, err = twinHash(in.tx)
		out = append(out, encoderResult{"twin/" + in.name, h, err})
	}
	return out
}

func fixtureAddr(b byte) sdk.AccAddress { return sdk.AccAddress(bytes.Repeat([]byte{b}, 20)) }

func multiSendMsg() sdk.Msg {
	coins := func(a, b int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(fixtureDenom, a), sdk.NewInt64Coin("bmoca", b))
	}
	return banktypes.NewMsgMultiSend(
		banktypes.Input{Address: fixtureAddr(1).String(), Coins: coins(10, 11)},
		[]banktypes.Output{
			{Address: fixtureAddr(2).String(), Coins: coins(4, 5)},
			{Address: fixtureAddr(3).String(), Coins: coins(6, 6)},
		},
	)
}

func sendMsg() sdk.Msg {
	return banktypes.NewMsgSend(fixtureAddr(1), fixtureAddr(2), sdk.NewCoins(sdk.NewInt64Coin(fixtureDenom, 5)))
}

func submitProposalMsg(t *testing.T, inner ...sdk.Msg) sdk.Msg {
	t.Helper()

	msg, err := govv1.NewMsgSubmitProposal(inner, sdk.NewCoins(sdk.NewInt64Coin(fixtureDenom, 1)),
		fixtureAddr(1).String(), "metadata", "title", "summary", false)
	require.NoError(t, err)
	return msg
}

func consensusParamsMsg() sdk.Msg {
	return &consensustypes.MsgUpdateParams{
		Authority: fixtureAddr(1).String(),
		Block:     &cmtproto.BlockParams{MaxBytes: 1000, MaxGas: 2000},
		Validator: &cmtproto.ValidatorParams{PubKeyTypes: []string{"ed25519"}},
	}
}

// unit builds a denom unit; aliases is an omitempty list.
func unit(aliases ...string) *banktypes.DenomUnit {
	return &banktypes.DenomUnit{Denom: fixtureDenom, Aliases: aliases}
}

func metadataOf(units ...*banktypes.DenomUnit) banktypes.Metadata {
	return banktypes.Metadata{Description: "coin", DenomUnits: units, Base: fixtureDenom, Display: "moca", Name: "Moca", Symbol: "MOCA"}
}

func metadataMsg(units ...*banktypes.DenomUnit) sdk.Msg {
	md := metadataOf(units...)
	return &md
}

func denomsMetadataMsg(mds ...banktypes.Metadata) sdk.Msg {
	return &banktypes.QueryDenomsMetadataResponse{Metadatas: mds, Pagination: &query.PageResponse{Total: uint64(len(mds))}}
}

// groupMember builds an element whose nullable member is set or nil.
func groupMember(id uint64, set bool) *group.GroupMember {
	gm := &group.GroupMember{GroupId: id}
	if set {
		gm.Member = &group.Member{
			Address: fixtureAddr(byte(id)).String(), Weight: "1", Metadata: "meta", AddedAt: time.Unix(1_700_000_000, 0).UTC(),
		}
	}
	return gm
}

func membersMsg(members ...*group.GroupMember) sdk.Msg {
	return &group.QueryGroupMembersResponse{Members: members, Pagination: &query.PageResponse{Total: uint64(len(members))}}
}

func singleMode() *txtypes.ModeInfo {
	return &txtypes.ModeInfo{Sum: &txtypes.ModeInfo_Single_{Single: &txtypes.ModeInfo_Single{Mode: signingtypes.SignMode_SIGN_MODE_EIP_712}}}
}

// multiMode nests ModeInfo lists; an empty ModeInfo leaves the oneof unset.
func multiMode(infos ...*txtypes.ModeInfo) *txtypes.ModeInfo {
	return &txtypes.ModeInfo{Sum: &txtypes.ModeInfo_Multi_{Multi: &txtypes.ModeInfo_Multi{
		Bitarray:  &cryptotypes.CompactBitArray{ExtraBitsStored: 2, Elems: []byte{0xc0}},
		ModeInfos: infos,
	}}}
}

// Digests were produced on base revision 17daa50fcc (before the list-element fix) from these same fixtures.
func TestEIP712ArrayGoldenDigests(t *testing.T) {
	env := newEIP712Env(t)

	cases := []struct {
		name   string
		msgs   []sdk.Msg
		digest string
	}{
		{"bank MsgMultiSend nested coin lists", []sdk.Msg{multiSendMsg()}, "60342e9c086d67aeaf943d03ee53f6546e0cfba272f87a944e7d93dd03484bf2"},
		{"gov MsgSubmitProposal Any list of different msgs", []sdk.Msg{submitProposalMsg(t, sendMsg(), multiSendMsg())}, "b0312ee3f70cb61b469397098d9c76cc4e6c98d2dbe514e7b2ab903183184009"},
		{"gov MsgSubmitProposal empty Any list", []sdk.Msg{submitProposalMsg(t)}, "982d6d67fd01da4493a9b5d4571923bf6709dd192007ec75b479c019644141d1"},
		{"consensus MsgUpdateParams unset nullable sub-messages", []sdk.Msg{consensusParamsMsg()}, "495fed0017d93c6e7bffac3d6e9c962c6822cf6bf823ee643e16b654442015ec"},
		{"tx with two msgs", []sdk.Msg{multiSendMsg(), submitProposalMsg(t, sendMsg())}, "635f0fdea53d0e7cd64795c4ebc8b872b543c59c4e1ec17e99044993d9ea850f"},
		{"bank denoms metadata aliases on every unit", []sdk.Msg{denomsMetadataMsg(
			metadataOf(unit("wei"), unit("MOCA")), metadataOf(unit("a"), unit("b")),
		)}, "169f19b7742cace2b01ce06b531a9fdf968eaeae9e85799b192d7e6806db28d3"},
		{"bank denoms metadata no aliases", []sdk.Msg{denomsMetadataMsg(
			metadataOf(unit(), unit()), metadataOf(unit(), unit()),
		)}, "0ab03d37ddf3bc25fc6d12bcadee37d61dc72db2c74c5e0c3c33bcab3bb2d968"},
		{"bank denoms metadata later units leave aliases empty", []sdk.Msg{denomsMetadataMsg(
			metadataOf(unit("wei"), unit()), metadataOf(unit(), unit()),
		)}, "4fd49c57efa151e37b90874ffebc20d77b1011a35aac0a03d7891d227e066a3c"},
		{"group members set in every element", []sdk.Msg{membersMsg(groupMember(1, true), groupMember(2, true))}, "71f57893b8e4a5f58f50387f7a3e4a6049c0bf7b3688cf404164321c6cba0bb4"},
		{"group members unset in every element", []sdk.Msg{membersMsg(groupMember(1, false), groupMember(2, false))}, "cb4e81f93b203c44359c44a6a4047ab69098d8e1709c03ff090f74df1333b57f"},
		{"tx ModeInfo nested multisig", []sdk.Msg{multiMode(multiMode(singleMode()), multiMode(singleMode()))}, "f76ab49b76d835bd86a6425fffb0c2a58b6a81bbbba9cd412a34457ba60c336d"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			built, decoded := env.txs(t, tc.msgs...)
			for _, r := range hashAll(built, decoded) {
				require.NoErrorf(t, r.err, "%s", r.name)
				require.Equalf(t, tc.digest, hex.EncodeToString(r.hash), "%s", r.name)
			}
			require.NoError(t, env.verify(decoded, env.sign(t, decoded)))
		})
	}
}

func TestEIP712ArrayRejectsHeterogeneousElements(t *testing.T) {
	env := newEIP712Env(t)
	bare, aliased := unit(), unit("MOCA")

	const (
		notCovered = "not covered"
		notSet     = "not set"
	)
	cases := []struct {
		name    string
		bad     sdk.Msg
		uniform sdk.Msg
		field   string
		class   string
	}{
		{
			name:    "bank Metadata: later unit carries aliases the first lacks",
			bad:     metadataMsg(bare, aliased),
			uniform: metadataMsg(aliased, aliased),
			field:   fieldAliases, class: notCovered,
		},
		{
			name:    "nested: later outer element populates a field the first outer element's units omit",
			bad:     denomsMetadataMsg(metadataOf(bare, bare), metadataOf(bare, aliased)),
			uniform: denomsMetadataMsg(metadataOf(aliased, aliased), metadataOf(aliased, aliased)),
			field:   fieldAliases, class: notCovered,
		},
		{
			name:    "nested: later unit of the first outer element populates a field the first unit omits",
			bad:     denomsMetadataMsg(metadataOf(bare, aliased), metadataOf(bare, bare)),
			uniform: denomsMetadataMsg(metadataOf(aliased, aliased), metadataOf(aliased, aliased)),
			field:   fieldAliases, class: notCovered,
		},
		{
			name:    "group members: later element leaves a template field unset",
			bad:     membersMsg(groupMember(1, true), groupMember(2, false)),
			uniform: membersMsg(groupMember(1, true), groupMember(2, true)),
			field:   "member", class: notSet,
		},
		{
			name:    "group members: later element sets a field the first element left unset",
			bad:     membersMsg(groupMember(1, false), groupMember(2, true)),
			uniform: membersMsg(groupMember(1, false), groupMember(2, false)),
			field:   "member", class: notCovered,
		},
		{
			name:    "tx ModeInfo: later outer element cannot shrink the shared inner type",
			bad:     multiMode(multiMode(singleMode()), multiMode(&txtypes.ModeInfo{})),
			uniform: multiMode(multiMode(singleMode()), multiMode(singleMode())),
			field:   "single", class: notSet,
		},
		{
			name:    "tx ModeInfo: later inner element of the first outer element cannot shrink the inner type",
			bad:     multiMode(multiMode(singleMode(), &txtypes.ModeInfo{})),
			uniform: multiMode(multiMode(singleMode(), singleMode())),
			field:   "single", class: notSet,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			quoted := fmt.Sprintf("%q", tc.field)
			built, decoded := env.txs(t, tc.bad)
			for _, r := range hashAll(built, decoded) {
				require.Errorf(t, r.err, "%s must reject", r.name)
				require.ErrorContainsf(t, r.err, quoted, "%s", r.name)
				require.ErrorContainsf(t, r.err, tc.class, "%s", r.name)
			}
			require.ErrorContains(t, env.verify(decoded, make([]byte, 65)), quoted)

			built, decoded = env.txs(t, tc.uniform)
			results := hashAll(built, decoded)
			for _, r := range results {
				require.NoErrorf(t, r.err, "%s accepts the uniform variant", r.name)
				require.Equalf(t, results[0].hash, r.hash, "%s", r.name)
			}
		})
	}
}

func TestEIP712VerifySignatureBindsEveryListElement(t *testing.T) {
	env := newEIP712Env(t)
	bare := unit()

	cases := []struct {
		name                 string
		stripped, full, a, b sdk.Msg
	}{
		{
			name:     "flat",
			stripped: metadataMsg(bare, bare),
			full:     metadataMsg(bare, unit("MOCA")),
			a:        metadataMsg(unit("wei"), unit("MOCA")),
			b:        metadataMsg(unit("wei"), unit("OTHER")),
		},
		{
			name:     "nested",
			stripped: denomsMetadataMsg(metadataOf(bare, bare), metadataOf(bare, bare)),
			full:     denomsMetadataMsg(metadataOf(bare, bare), metadataOf(bare, unit("MOCA"))),
			a:        denomsMetadataMsg(metadataOf(unit("wei"), unit("MOCA")), metadataOf(unit("wei"), unit("MOCA"))),
			b:        denomsMetadataMsg(metadataOf(unit("wei"), unit("MOCA")), metadataOf(unit("wei"), unit("OTHER"))),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, stripped := env.txs(t, tc.stripped)
			_, full := env.txs(t, tc.full)
			sig := env.sign(t, stripped)
			require.NoError(t, env.verify(stripped, sig))
			// a signature over the template-shaped tx must not carry over to a tx with an extra field
			require.ErrorContains(t, env.verify(full, sig), fmt.Sprintf("%q", fieldAliases))

			_, a := env.txs(t, tc.a)
			_, b := env.txs(t, tc.b)
			sig = env.sign(t, a)
			require.NoError(t, env.verify(a, sig))
			require.Error(t, env.verify(b, sig), "a uniform list signs its later elements")
		})
	}
}
