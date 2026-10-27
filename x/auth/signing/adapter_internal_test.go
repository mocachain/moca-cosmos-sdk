package signing

import (
	"testing"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/require"
)

// These tests exercise cleanTypesAndMsgValue directly (MOCA-2186). The EIP-712
// type of a repeated struct field is built from the first element only
// (traverseFields), so a field left unset in that element is absent from the
// type. Before this change the per-element cleanup silently deleted any message
// key missing from the type, so a field populated only in a later element was
// dropped from the signed hash and a holder of the pending tx could rewrite it.
// cleanTypesAndMsgValue now reconciles the first element with allowTypeShrink
// true (it defines the element type) and every later element with it false (the
// element must match that type), returning an error on any divergence.
//
// The type sets below mirror what traverseFields produces for a permission
// Statement: `actions`/`resources` are omitempty collections (skipped from the
// type when empty in the first element); `limit_size` is a nullable sub-message
// (a nested struct type). The field list intentionally omits a field to model
// "empty/unset in the first element".

// statementTypesNoResources models the type derived from a first Statement with
// empty `resources` (skipped) and a set `limit_size` (nested struct kept).
func statementTypesNoResources() apitypes.Types {
	return apitypes.Types{
		"TypeStatement": {
			{Name: "type", Type: "string"},
			{Name: "actions", Type: "string[]"},
			{Name: "limit_size", Type: "TypeLimitSize"},
		},
		"TypeLimitSize": {
			{Name: "value", Type: "uint256"},
		},
	}
}

func TestCleanTypesAndMsgValue_RejectsUncoveredCollectionInLaterElement(t *testing.T) {
	typedData := statementTypesNoResources()
	// Later element carries a populated `resources` that the first-element type
	// omits. It must be rejected, not silently dropped.
	laterElem := map[string]interface{}{
		"type":       "/moca.permission.Statement",
		"actions":    []interface{}{"6"},
		"limit_size": map[string]interface{}{"value": "100"},
		"resources":  []interface{}{"grn:o::bucket/secret"},
	}
	err := cleanTypesAndMsgValue(typedData, "TypeStatement", laterElem, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "resources")
	require.Contains(t, err.Error(), "not covered")
	// The uncovered value must still be present (not silently removed).
	require.Contains(t, laterElem, "resources")
}

func TestCleanTypesAndMsgValue_DropsEmptyUncoveredCollection(t *testing.T) {
	typedData := statementTypesNoResources()
	// The first element legitimately carries an empty `resources` (EmitDefaults
	// emits it). It is uncovered but empty, so it is dropped exactly as before —
	// preserving the sign bytes of a uniformly shaped message.
	elem := map[string]interface{}{
		"type":       "/moca.permission.Statement",
		"actions":    []interface{}{"6"},
		"limit_size": map[string]interface{}{"value": "100"},
		"resources":  []interface{}{},
	}
	err := cleanTypesAndMsgValue(typedData, "TypeStatement", elem, true)
	require.NoError(t, err)
	require.NotContains(t, elem, "resources")
}

func TestCleanTypesAndMsgValue_RejectsMissingNilStructInLaterElement(t *testing.T) {
	typedData := statementTypesNoResources()
	// First element had `limit_size`, so the type carries it. A later element
	// whose `limit_size` is nil would, before this change, strip it from the
	// shared type (dropping the first element's quota from the hash). Reject it.
	laterElem := map[string]interface{}{
		"type":       "/moca.permission.Statement",
		"actions":    []interface{}{"6"},
		"limit_size": nil,
	}
	err := cleanTypesAndMsgValue(typedData, "TypeStatement", laterElem, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "limit_size")
	require.Contains(t, err.Error(), "not set")
	// The type must be intact for the elements already encoded against it.
	require.Contains(t, typedData, "TypeLimitSize")
	require.Len(t, typedData["TypeStatement"], 3)
}

// runStatementArray reproduces the array loop in cleanTypesAndMsgValue: the
// first element with allowTypeShrink=true, the rest with false, sharing the
// type set. It returns the first error, as the real encoder does.
func runStatementArray(typedData apitypes.Types, elems ...map[string]interface{}) error {
	for i, elem := range elems {
		if err := cleanTypesAndMsgValue(typedData, "TypeStatement", elem, i == 0); err != nil {
			return err
		}
	}
	return nil
}

func TestCleanTypesAndMsgValue_RejectsHeterogeneousBothOrderings(t *testing.T) {
	populated := func() map[string]interface{} {
		return map[string]interface{}{
			"type":       "/moca.permission.Statement",
			"actions":    []interface{}{"3"},
			"limit_size": map[string]interface{}{"value": "100"},
		}
	}
	nilQuota := func() map[string]interface{} {
		return map[string]interface{}{
			"type":       "/moca.permission.Statement",
			"actions":    []interface{}{"3"},
			"limit_size": nil,
		}
	}

	// nil-first then populated: the first element drops limit_size from the type,
	// the later populated limit_size is then uncovered and rejected.
	err := runStatementArray(statementTypesNoResources(), nilQuota(), populated())
	require.Error(t, err, "nil-then-populated limit_size must be rejected")

	// populated-first then nil: the later nil element lacks a type field, and is
	// rejected in code rather than relying on the hasher's extra-data guard.
	err = runStatementArray(statementTypesNoResources(), populated(), nilQuota())
	require.Error(t, err, "populated-then-nil limit_size must be rejected")
}

func TestCleanTypesAndMsgValue_UniformArrayAccepted(t *testing.T) {
	// Every element shares the first element's shape: no divergence, no error,
	// and the covered fields are retained.
	mk := func(res string) map[string]interface{} {
		return map[string]interface{}{
			"type":       "/moca.permission.Statement",
			"actions":    []interface{}{"3"},
			"limit_size": map[string]interface{}{"value": "100"},
			"resources":  []interface{}{res},
		}
	}
	// Here the first element populates resources, so the type covers it.
	typedData := apitypes.Types{
		"TypeStatement": {
			{Name: "type", Type: "string"},
			{Name: "actions", Type: "string[]"},
			{Name: "resources", Type: "string[]"},
			{Name: "limit_size", Type: "TypeLimitSize"},
		},
		"TypeLimitSize": {{Name: "value", Type: "uint256"}},
	}
	e0, e1 := mk("grn:o::bucket/a"), mk("grn:o::bucket/b")
	require.NoError(t, runStatementArray(typedData, e0, e1))
	require.Equal(t, []interface{}{"grn:o::bucket/a"}, e0["resources"])
	require.Equal(t, []interface{}{"grn:o::bucket/b"}, e1["resources"])
}

func TestIsEmptyEIP712Value(t *testing.T) {
	empty := []interface{}{nil, "", []interface{}{}, map[string]interface{}{}}
	for _, v := range empty {
		require.True(t, isEmptyEIP712Value(v))
	}
	nonEmpty := []interface{}{"x", []interface{}{"a"}, map[string]interface{}{"k": 1}, 0, false}
	for _, v := range nonEmpty {
		require.False(t, isEmptyEIP712Value(v))
	}
}
