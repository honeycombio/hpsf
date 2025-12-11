package config

import (
	"fmt"
	"testing"
)

func Test_getValidationRule(t *testing.T) {
	tests := []struct {
		name       string
		validation string
		propval    any // this is the value that will be passed to the validation function
		want       bool
	}{
		{"unknown validation", "unknown", "some value", false}, // test for an unknown validation type
		{"positive integer", "positive", 1, true},
		{"zero is not positive", "positive", 0, false},
		{"negative integer", "positive", -1, false},
		{"positive integer", "positive", 1.5, true},
		{"zero is not positive", "positive", 0.0, false},
		{"negative integer", "positive", -1.5, false},
		{"non-blank string", "noblanks", "hello", true},
		{"empty string", "noblanks", "", false},
		{"non-blank string slice", "noblanks", []string{"hello", "world"}, true},
		{"one empty string in the slice", "noblanks", []string{"", "world"}, false},
		{"empty string slice", "noblanks", []string{}, true},
		{"non-blank any slice", "noblanks", []any{"hello", "world"}, true},
		{"one empty string in the any slice", "noblanks", []any{"", "world"}, false},
		{"empty any slice", "noblanks", []any{}, true},
		{"non-blank map", "noblanks", map[string]any{"key1": "value1", "key2": "value2"}, true},
		{"empty map", "noblanks", map[string]any{}, true},
		{"non-empty string", "nonempty", "hello", true},
		{"empty string", "nonempty", "", false},
		{"non-empty string slice", "nonempty", []string{"hello", "world"}, true},
		{"empty string slice", "nonempty", []string{}, false},
		{"non-empty any slice", "nonempty", []any{"hello", "world"}, true},
		{"empty any slice", "nonempty", []any{}, false},
		{"non-empty map", "nonempty", map[string]any{"key1": "value1", "key2": "value2"}, true},
		{"empty map", "nonempty", map[string]any{}, false},
		{"last in a string list", "oneof(a,b, c)", "c", true},
		{"first in a string list", "oneof(a, b,c)", "a", true},
		{"middle in a string list", "oneof(a,b   ,c)", "b", true},
		{"not in a string list", "oneof(a,b,c)", "d", false},
		{"basic", "url", "http://example.com", true},
		{"invalid", "url", "not_a_url", false},
		{"basic", "duration", "5s", true},
		{"invalid", "duration", "not_a_duration", false},
		{"int fail", "atleast(5)", 4, false},
		{"int equal", "atleast(5)", 5, true},
		{"int over", "atleast(1)", 4, true},
		{"float equal", "atleast(1.5)", 1.5, true},
		{"float fail", "atleast(10.5)", 1.5, false},
		{"int fail", "atmost(5)", 14, false},
		{"int equal", "atmost(5)", 5, true},
		{"int under", "atmost(1)", 0, true},
		{"float equal", "atmost(1.5)", 1.5, true},
		{"float fail", "atmost(10.5)", 100.5, false},
		{"int range", "inrange(1,5)", 3, true},
		{"int range low end", "inrange(1,5)", 1, true},
		{"int range high end", "inrange(1,5)", 5, true},
		{"int range fail low", "inrange(1,5)", 0, false},
		{"int range fail high", "inrange(1,5)", 6, false},
		{"float range", "inrange(1.5,5.5)", 3.0, true},
		{"float range low end", "inrange(1.5,5.5)", 1.5, true},
		{"float range high end", "inrange(1.5,5.5)", 5.5, true},
		{"float range fail low", "inrange(1.5,5.5)", 1.4, false},
		{"float range fail high", "inrange(1.5,5.5)", 5.6, false},
		{"valid regex", "regex", "[a-z]+", true},
		{"invalid regex", "regex", "[a-z", false},
		{"valid regex with quantifier", "regex", "\\d{3,5}", true},
		{"invalid regex with unmatched paren", "regex", "(abc", false},
		{"valid regex with anchors", "regex", "^start.*end$", true},
		{"invalid regex with bad quantifier", "regex", "*+", false},
		{"environment variable", "regex", "$REGEX_PATTERN", true},
		{"valid regex string array", "regex", []string{"[a-z]+", "\\d+"}, true},
		{"invalid regex string array with one bad pattern", "regex", []string{"[a-z]+", "[a-z"}, false},
		{"empty regex string array", "regex", []string{}, true},
		{"single element regex string array valid", "regex", []string{"[a-zA-Z0-9]+"}, true},
		{"single element regex string array invalid", "regex", []string{"[unclosed"}, false},
		{"valid regex any array", "regex", []any{"[a-z]+", "\\d+"}, true},
		{"invalid regex any array with one bad pattern", "regex", []any{"[a-z]+", "[a-z"}, false},
		{"invalid regex any array with non-string", "regex", []any{"[a-z]+", 123}, false},
		{"regex with environment variable in array", "regex", []string{"$REGEX_PATTERN", "[a-z]+"}, true},
		{"regex with mixed env vars and patterns", "regex", []string{"$VAR1", "\\d+", "$VAR2"}, true},
		{"complex regex patterns in array", "regex", []string{"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", "\\b\\d{3}[-.]?\\d{2}[-.]?\\d{4}\\b"}, true},
		{"invalid type for regex", "regex", 123, false},
		{"invalid type for regex - map", "regex", map[string]any{"key": "value"}, false},
		{"valid Go date format", "enhancepartitionformat", "2006-01-02 15:04", true},
		{"valid C date format", "enhancepartitionformat", "%Y-%m-%d %H:%M", true},
		{"invalid non-string", "enhancepartitionformat", 123, false},
		{"invalid starts with slash", "enhancepartitionformat", "/2006-01-02", false},
		{"invalid ends with slash", "enhancepartitionformat", "2006-01-02/", false},
		{"invalid missing Go tokens", "enhancepartitionformat", "2006-01-02", false},
		{"invalid missing C tokens", "enhancepartitionformat", "%Y-%m", false},
		{"invalid mixed formats", "enhancepartitionformat", "2006-%m-%d", false},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%s_%v", tt.validation, tt.name)
		t.Run(name, func(t *testing.T) {
			rule := getValidationRule(tt.validation)

			if got := rule(tt.propval); got != tt.want {
				t.Errorf("getValidationRule() = %v, want %v", got, tt.want)
			}
		})
	}
}
