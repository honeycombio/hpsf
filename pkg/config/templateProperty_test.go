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
