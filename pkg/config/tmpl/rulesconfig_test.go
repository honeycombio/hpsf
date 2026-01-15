package tmpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupKeysByNumericSuffix(t *testing.T) {
	// Test with numeric suffixes
	kvs := map[string]any{
		"Fields.1":   []string{"field1"},
		"Operator.1": ">=",
		"Value.1":    "400",
		"Datatype.1": "int",
		"Fields.2":   []string{"field2"},
		"Operator.2": "<=",
		"Value.2":    "499",
		"Datatype.2": "int",
	}

	groups := groupKeysByNumericSuffix(kvs)
	require.Len(t, groups, 2)

	// Check first group
	assert.Equal(t, []string{"field1"}, groups[0]["Fields"])
	assert.Equal(t, ">=", groups[0]["Operator"])
	assert.Equal(t, "400", groups[0]["Value"])
	assert.Equal(t, "int", groups[0]["Datatype"])

	// Check second group
	assert.Equal(t, []string{"field2"}, groups[1]["Fields"])
	assert.Equal(t, "<=", groups[1]["Operator"])
	assert.Equal(t, "499", groups[1]["Value"])
	assert.Equal(t, "int", groups[1]["Datatype"])
}

func TestGroupKeysByNumericSuffixNoSuffixes(t *testing.T) {
	// Test without numeric suffixes
	kvs := map[string]any{
		"Fields":   []string{"field1"},
		"Operator": ">=",
		"Value":    "400",
		"Datatype": "int",
	}

	groups := groupKeysByNumericSuffix(kvs)
	assert.Nil(t, groups)
}
