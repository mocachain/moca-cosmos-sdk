package signing

import (
	"testing"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/require"
)

// type sets mirror traverseFields output for a permission Statement
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
	require.Contains(t, laterElem, "resources")
}

func TestCleanTypesAndMsgValue_DropsEmptyUncoveredCollection(t *testing.T) {
	typedData := statementTypesNoResources()
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
	laterElem := map[string]interface{}{
		"type":       "/moca.permission.Statement",
		"actions":    []interface{}{"6"},
		"limit_size": nil,
	}
	err := cleanTypesAndMsgValue(typedData, "TypeStatement", laterElem, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "limit_size")
	require.Contains(t, err.Error(), "not set")
	require.Contains(t, typedData, "TypeLimitSize")
	require.Len(t, typedData["TypeStatement"], 3)
}

// runStatementArray mirrors the encoder list loop.
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

	err := runStatementArray(statementTypesNoResources(), nilQuota(), populated())
	require.Error(t, err, "nil-then-populated limit_size must be rejected")

	err = runStatementArray(statementTypesNoResources(), populated(), nilQuota())
	require.Error(t, err, "populated-then-nil limit_size must be rejected")
}

func TestCleanTypesAndMsgValue_UniformArrayAccepted(t *testing.T) {
	mk := func(res string) map[string]interface{} {
		return map[string]interface{}{
			"type":       "/moca.permission.Statement",
			"actions":    []interface{}{"3"},
			"limit_size": map[string]interface{}{"value": "100"},
			"resources":  []interface{}{res},
		}
	}
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
