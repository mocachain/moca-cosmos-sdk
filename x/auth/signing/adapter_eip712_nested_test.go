package signing_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/signing/internal/nestedpb"
)

// These fixtures use nestedpb.MsgNestedFields, whose repeated `items` each carry
// an omitempty repeated `labels` and a nullable `limit` sub-message. That is the
// MOCA-2186 shape: the EIP-712 encoder derives the item type from the first
// item, so a later item shaped differently used to be reshaped to the first
// item's type rather than rejected. The merged adapter_eip712_array_test.go
// proves the same property on stand-in SDK messages; this file pins it on a
// single self-contained message built, encoded and decoded through the real tx
// pipeline. Both encoders (the x/auth/signing adapter and the x/auth/tx twin)
// and both the built and decoded forms are checked, via the shared hashAll
// helper.

const (
	labelAll      = "label-*"
	labelPrefixed = "label-report-*"
)

func newNestedEnv(t *testing.T) eip712Env {
	t.Helper()

	enc := moduletestutil.MakeTestEncodingConfig()
	nestedpb.RegisterInterfaces(enc.InterfaceRegistry)

	return eip712Env{
		txConfig: enc.TxConfig,
		key:      ethsecp256k1.GenPrivKeyFromSecret([]byte("eip712-nested-fixtures")),
	}
}

func u64(v uint64) *nestedpb.UInt64Value { return &nestedpb.UInt64Value{Value: v} }

// item builds a NestedItem. Every item carries the same mode and tag, so only
// `labels` and `limit` can make an items list non-uniform -- the fields that
// MOCA-2186 is about.
func item(labels []string, limit *nestedpb.UInt64Value) *nestedpb.NestedItem {
	return &nestedpb.NestedItem{
		Mode:   nestedpb.MODE_ALLOW,
		Tags:   []nestedpb.Tag{nestedpb.TAG_READ},
		Labels: labels,
		Limit:  limit,
	}
}

// nestedMsg builds a MsgNestedFields with a fixed signer, owner and scope.
func nestedMsg(items ...*nestedpb.NestedItem) sdk.Msg {
	return &nestedpb.MsgNestedFields{
		Signer: fixtureAddr(0x11).String(),
		Owner: &nestedpb.Owner{
			Type:  nestedpb.OWNER_TYPE_ACCOUNT,
			Value: fixtureAddr(0x22).String(),
		},
		Scope: "scope-a",
		Items: items,
	}
}

// Digests were produced on base revision 17daa50fcc (before the list-element
// fix) from these same fixtures; uniformly shaped messages hash identically
// before and after the fix.
func TestEIP712NestedGoldenDigests(t *testing.T) {
	env := newNestedEnv(t)

	cases := []struct {
		name   string
		msgs   []sdk.Msg
		digest string
	}{
		{"one item, no labels or limit", []sdk.Msg{nestedMsg(item(nil, nil))},
			"5bd558b35395107d2e4ee035aedf79825d3167481ad447b615ead1ecc7406265"},
		{"two bare items", []sdk.Msg{nestedMsg(item(nil, nil), item(nil, nil))},
			"be4e910bc2d39a954528caa92d304e47b603e93624904bd77f84c3cf3ffd13d2"},
		{"two items with labels", []sdk.Msg{nestedMsg(
			item([]string{labelAll, labelPrefixed}, nil),
			item([]string{labelAll}, nil),
		)}, "88984a29ee1f361c805409047af12bf2c6cf97c450e3dd4ade9e7f169e1c4ab0"},
		{"two items with limit", []sdk.Msg{nestedMsg(
			item(nil, u64(1024)), item(nil, u64(2048)),
		)}, "89544a4df1bdc6627149d05b11a9fd848e7d8ee813bcf2522e2bdf3fe10e2c0b"},
		{"two items with labels and limit", []sdk.Msg{nestedMsg(
			item([]string{labelAll}, u64(1024)),
			item([]string{labelPrefixed}, u64(2048)),
		)}, "d8472218a8f252735821886869aa1547c5d2379c2b5a02666287001001ccc243"},
		{"labels populated first, empty later", []sdk.Msg{nestedMsg(
			item([]string{labelAll}, nil), item(nil, nil),
		)}, "c0e2822f5786db0223ecc6d5eb8917b8c0cfa6abf47b9b27d54d70f2c62a17bc"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			built, decoded := env.txs(t, tc.msgs...)
			results := hashAll(built, decoded)
			for _, r := range results {
				require.NoErrorf(t, r.err, "%s", r.name)
				require.Equalf(t, results[0].hash, r.hash, "%s differs from %s", r.name, results[0].name)
			}
			require.Equal(t, tc.digest, hex.EncodeToString(results[0].hash))
			require.NoError(t, env.verify(decoded, env.sign(t, decoded)))
		})
	}
}

func TestEIP712NestedRejectsHeterogeneousItems(t *testing.T) {
	env := newNestedEnv(t)

	const (
		fieldLabels = "labels"
		fieldLimit  = "limit"
		notCovered  = "not covered"
		notSet      = "not set"
	)
	cases := []struct {
		name    string
		bad     sdk.Msg
		uniform sdk.Msg
		field   string
		class   string
	}{
		{
			name:    "later item sets labels the first leaves empty",
			bad:     nestedMsg(item(nil, nil), item([]string{labelAll}, nil)),
			uniform: nestedMsg(item([]string{labelAll}, nil), item([]string{labelAll}, nil)),
			field:   fieldLabels, class: notCovered,
		},
		{
			name:    "later item sets limit the first leaves nil",
			bad:     nestedMsg(item(nil, nil), item(nil, u64(1024))),
			uniform: nestedMsg(item(nil, u64(1024)), item(nil, u64(1024))),
			field:   fieldLimit, class: notCovered,
		},
		{
			name:    "later item drops the limit the first sets",
			bad:     nestedMsg(item(nil, u64(1024)), item(nil, nil)),
			uniform: nestedMsg(item(nil, u64(1024)), item(nil, u64(2048))),
			field:   fieldLimit, class: notSet,
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

// TestEIP712NestedShrinkIsAccepted documents the precise boundary: a later item
// whose labels list is empty is still fully represented by the first item's
// type, so it is accepted rather than rejected (an empty repeated field encodes
// to nothing). Only a later item that carries a field the type omits, or drops a
// sub-message the type sets, is rejected.
func TestEIP712NestedShrinkIsAccepted(t *testing.T) {
	env := newNestedEnv(t)

	msg := nestedMsg(item([]string{labelAll, labelPrefixed}, nil), item(nil, nil))
	built, decoded := env.txs(t, msg)
	results := hashAll(built, decoded)
	for _, r := range results {
		require.NoErrorf(t, r.err, "%s", r.name)
		require.Equalf(t, results[0].hash, r.hash, "%s", r.name)
	}
	require.NoError(t, env.verify(decoded, env.sign(t, decoded)))
}

func TestEIP712NestedSignatureBindsEveryItem(t *testing.T) {
	env := newNestedEnv(t)

	t.Run("a stripped item must not accept an injected label", func(t *testing.T) {
		// Signer approves a message whose items carry no labels.
		_, stripped := env.txs(t, nestedMsg(item(nil, nil), item(nil, nil)))
		sig := env.sign(t, stripped)
		require.NoError(t, env.verify(stripped, sig))

		// A holder of the signed tx scopes a later item to a label. The
		// signature must not carry over to the rewritten message.
		_, injected := env.txs(t, nestedMsg(item(nil, nil), item([]string{labelAll}, nil)))
		require.ErrorContains(t, env.verify(injected, sig), `"labels"`)
	})

	t.Run("a covered later label value is bound by the signature", func(t *testing.T) {
		_, a := env.txs(t, nestedMsg(
			item([]string{labelAll}, nil), item([]string{labelPrefixed}, nil),
		))
		sig := env.sign(t, a)
		require.NoError(t, env.verify(a, sig))

		// Same shape, but the second item's covered label value changed.
		_, b := env.txs(t, nestedMsg(
			item([]string{labelAll}, nil), item([]string{"label-other-*"}, nil),
		))
		require.Error(t, env.verify(b, sig), "a uniform items list signs its later elements")
	})
}
