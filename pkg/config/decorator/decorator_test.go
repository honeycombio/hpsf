package decorator

import (
	"strings"
	"testing"
)

func Test_EncodeAsInt(t *testing.T) {
	tests := []struct {
		name string
		arg  any
		want string
	}{
		// Integer inputs
		{"positive int", 42, IntPrefix + "42"},
		{"negative int", -17, IntPrefix + "-17"},
		{"zero int", 0, IntPrefix + "0"},
		{"large positive int", 2147483647, IntPrefix + "2147483647"},
		{"large negative int", -2147483648, IntPrefix + "-2147483648"},

		// Float inputs (truncated towards zero)
		{"positive float", 42.7, IntPrefix + "42"},
		{"negative float", -17.9, IntPrefix + "-17"},
		{"zero float", 0.0, IntPrefix + "0"},
		{"small positive float", 0.9, IntPrefix + "0"},
		{"small negative float", -0.9, IntPrefix + "0"},

		// String inputs (first integer found)
		{"string with positive int", "value is 123", IntPrefix + "123"},
		{"string with negative int", "temp -45 degrees", IntPrefix + "-45"},
		{"string with no int", "no numbers here", IntPrefix + "0"},
		{"empty string", "", IntPrefix + "0"},
		{"string starting with int", "42abc", IntPrefix + "42"},
		{"string with multiple ints", "first 10 second 20", IntPrefix + "10"},

		// Boolean inputs
		{"bool true", true, IntPrefix + "1"},
		{"bool false", false, IntPrefix + "0"},

		// Other types (should default to 0)
		{"nil", nil, IntPrefix + "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeAsInt(tt.arg); got != tt.want {
				t.Errorf("EncodeAsInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_EncodeAsFloat(t *testing.T) {
	tests := []struct {
		name string
		arg  any
		want string
	}{
		// Integer inputs (converted to float string)
		{"positive int", 42, FloatPrefix + "42"},
		{"negative int", -17, FloatPrefix + "-17"},
		{"zero int", 0, FloatPrefix + "0"},

		// Float inputs
		{"positive float", 42.75, FloatPrefix + "42.750000"},
		{"negative float", -17.25, FloatPrefix + "-17.250000"},
		{"zero float", 0.0, FloatPrefix + "0.000000"},
		{"small positive float", 0.001, FloatPrefix + "0.001000"},
		{"small negative float", -0.001, FloatPrefix + "-0.001000"},
		{"large float", 3.14159265359, FloatPrefix + "3.141593"},

		// String inputs (first float found)
		{"string with positive float", "temp 23.5 degrees", FloatPrefix + "23.5"},
		{"string with negative float", "change -4.2%", FloatPrefix + "-4.2"},
		{"string with int", "count 42", FloatPrefix + "42"},
		{"string with no number", "no numbers here", FloatPrefix + "0"},
		{"empty string", "", FloatPrefix + "0"},
		{"string starting with float", "3.14abc", FloatPrefix + "3.14"},

		// Boolean inputs
		{"bool true", true, FloatPrefix + "1"},
		{"bool false", false, FloatPrefix + "0"},

		// Other types (should default to 0)
		{"nil", nil, FloatPrefix + "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeAsFloat(tt.arg); got != tt.want {
				t.Errorf("EncodeAsFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_EncodeAsBool(t *testing.T) {
	tests := []struct {
		name string
		arg  any
		want string
	}{
		// Boolean inputs
		{"bool true", true, BoolPrefix + "true"},
		{"bool false", false, BoolPrefix + "false"},

		// Integer inputs (non-zero is true)
		{"positive int", 42, BoolPrefix + "true"},
		{"negative int", -17, BoolPrefix + "true"},
		{"zero int", 0, BoolPrefix + "false"},
		{"one int", 1, BoolPrefix + "true"},

		// Float inputs (non-zero is true)
		{"positive float", 42.7, BoolPrefix + "true"},
		{"negative float", -17.9, BoolPrefix + "true"},
		{"zero float", 0.0, BoolPrefix + "false"},
		{"small positive float", 0.001, BoolPrefix + "true"},
		{"small negative float", -0.001, BoolPrefix + "true"},

		// String inputs (only "true" string is true)
		{"string true", "true", BoolPrefix + "true"},
		{"string false", "false", BoolPrefix + "false"},
		{"string TRUE", "TRUE", BoolPrefix + "false"}, // case sensitive
		{"string yes", "yes", BoolPrefix + "false"},
		{"string 1", "1", BoolPrefix + "false"},
		{"empty string", "", BoolPrefix + "false"},
		{"string with true", "this is true", BoolPrefix + "false"},

		// Other types (should default to false)
		{"nil", nil, BoolPrefix + "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeAsBool(tt.arg); got != tt.want {
				t.Errorf("EncodeAsBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_EncodeAsArray(t *testing.T) {
	tests := []struct {
		name string
		arg  any
		want string
	}{
		// String slice inputs
		{"empty string slice", []string{}, ArrPrefix},
		{"single string", []string{"hello"}, ArrPrefix + "hello"},
		{"multiple strings", []string{"hello", "world"}, ArrPrefix + "hello" + FieldSeparator + "world"},
		{"strings with special chars", []string{"hello world", "test-123"}, ArrPrefix + "hello world" + FieldSeparator + "test-123"},

		// Any slice inputs (only string elements are extracted)
		{"empty any slice", []any{}, ArrPrefix},
		{"any slice with strings only", []any{"hello", "world"}, ArrPrefix + "hello" + FieldSeparator + "world"},
		{"mixed any slice - strings extracted", []any{"hello", 42, "world", true}, ArrPrefix + "hello" + FieldSeparator + "world"},
		{"any slice with no strings", []any{1, 2, 3}, ArrPrefix},
		{"any slice with bool and nums", []any{true, false, 42}, ArrPrefix},

		// Other types (should return empty string)
		{"single string", "not an array", ""},
		{"single int", 42, ""},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeAsArray(tt.arg); got != tt.want {
				t.Errorf("EncodeAsArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_EncodeAsMap(t *testing.T) {
	tests := []struct {
		name string
		arg  map[string]any
		want string
	}{
		// Basic map tests (JSON encoder adds trailing newline)
		{"empty map", map[string]any{}, MapPrefix + "{}\n"},
		{"single key string", map[string]any{"key": "value"}, MapPrefix + `{"key":"value"}` + "\n"},
		{"single key int", map[string]any{"count": 42}, MapPrefix + `{"count":42}` + "\n"},
		{"single key bool", map[string]any{"enabled": true}, MapPrefix + `{"enabled":true}` + "\n"},
		{"single key float", map[string]any{"temperature": 23.5}, MapPrefix + `{"temperature":23.5}` + "\n"},

		// Nested structures
		{"nested map", map[string]any{"outer": map[string]any{"inner": "value"}}, MapPrefix + `{"outer":{"inner":"value"}}` + "\n"},
		{"map with array", map[string]any{"items": []any{"a", "b", "c"}}, MapPrefix + `{"items":["a","b","c"]}` + "\n"},

		// Special values
		{"map with nil", map[string]any{"null_value": nil}, MapPrefix + `{"null_value":null}` + "\n"},
		{"map with empty string", map[string]any{"empty": ""}, MapPrefix + `{"empty":""}` + "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeAsMap(tt.arg)
			if got != tt.want {
				t.Errorf("EncodeAsMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_EncodeAsMap_MultipleKeys(t *testing.T) {
	// Test multiple keys separately since JSON key ordering is not deterministic
	testMap := map[string]any{"name": "test", "count": 42, "enabled": true}
	got := EncodeAsMap(testMap)

	// Check that it starts with the correct prefix
	if !strings.HasPrefix(got, MapPrefix) {
		t.Errorf("EncodeAsMap() = %v, expected to start with %v", got, MapPrefix)
	}

	// Check that all expected key-value pairs are present
	expectedSubstrings := []string{`"count":42`, `"enabled":true`, `"name":"test"`}
	for _, expected := range expectedSubstrings {
		if !strings.Contains(got, expected) {
			t.Errorf("EncodeAsMap() = %v, expected to contain %v", got, expected)
		}
	}

	// Check that it ends with newline
	if got[len(got)-1] != '\n' {
		t.Errorf("EncodeAsMap() = %v, expected to end with newline", got)
	}
}
