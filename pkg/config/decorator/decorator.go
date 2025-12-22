package decorator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// We use GroupSeparator, FieldSeparator, and RecordSeparator as delimiters so
// that we are unlikely to see a false positive with user data. We probably
// could have done this with a more sophisticated encoding/decoding but we don't
// think it matters in the context of templated configuration files.
const (
	GroupSeparator  = "\x1d" // ASCII code 29, the group separator
	RecordSeparator = "\x1e" // ASCII code 30, the record separator
	FieldSeparator  = "\x1f" // ASCII code 31, the unit (field) separator
)

// Prefixes we support:
const (
	IntPrefix   = "int" + GroupSeparator
	BoolPrefix  = "bool" + GroupSeparator
	FloatPrefix = "float" + GroupSeparator
	ArrPrefix   = "arr" + GroupSeparator
	MapPrefix   = "map" + GroupSeparator
)

// EncodeAsArray takes a slice and returns a string intended to be expanded
// later into an array when it's rendered to YAML.
// The result looks like "arr\x1dA:1\x1fB:2"
func EncodeAsArray(arr any) string {
	switch a := arr.(type) {
	case []string:
		return ArrPrefix + strings.Join(a, FieldSeparator)
	case []any:
		return ArrPrefix + strings.Join(getStringsFrom(arr), FieldSeparator)
	default:
		return ""
	}
}

// EncodeAsBool takes any value and returns a string with the appropriate marker
// so that it will be expanded later into a bool when it's rendered to YAML.
// Numbers are interpreted as true if they are not zero.
func EncodeAsBool(a any) string {
	value := "false"
	switch v := a.(type) {
	case bool:
		if v {
			value = "true"
		}
	case int:
		if v != 0 {
			value = "true"
		}
	case float64:
		if v != 0 {
			value = "true"
		}
	case string:
		if v == "true" {
			value = "true"
		}
	}
	return BoolPrefix + value
}

// EncodeAsFloat takes any value and returns a string with the appropriate marker
// so that it will be expanded later into a float when it's rendered to YAML.
// If the value cannot be parsed as a float, it returns a 0.
func EncodeAsFloat(a any) string {
	value := "0"
	switch v := a.(type) {
	case int:
		value = strconv.Itoa(v)
	case float64:
		value = fmt.Sprintf("%f", v)
	case string:
		// find the first thing that looks like a (possibly signed) float in the string
		pat := regexp.MustCompile(`[+-]?\d+(\.\d+)?`)
		match := pat.FindString(v)
		if match != "" {
			value = match
		}
	case bool:
		if v {
			value = "1"
		}
	}
	return FloatPrefix + value
}

// EncodeAsInt takes an "any", tries to convert it to an integer, and then
// returns a string with the appropriate marker so that it will be expanded
// later into an integer when it's rendered to YAML.
// Floats are truncated to integers (towards 0).
// If the value cannot be parsed as an integer, it returns a 0.
func EncodeAsInt(a any) string {
	value := "0"
	switch v := a.(type) {
	case int:
		value = strconv.Itoa(v)
	case float64:
		value = fmt.Sprintf("%d", int(v))
	case string:
		// find the first thing that looks like an integer (possibly signed) in the string
		// This is specifically to cope with something like "1.50000", which won't parse
		// when we use Atoi on it later -- we want to extract only the part that will parse.
		pat := regexp.MustCompile(`[+-]?[\d]+`)
		match := pat.FindString(v)
		if match != "" {
			value = match
		}
	case bool:
		if v {
			value = "1"
		}
	}
	return IntPrefix + value
}

// EncodeAsMap takes a map (which may contain nested maps) and returns a string
// intended to be expanded later into the same map when it's rendered to YAML.
// We encode to JSON because it's fast and easy.
func EncodeAsMap(a map[string]any) string {
	buf := bytes.Buffer{}
	j := json.NewEncoder(&buf)
	// There's no model for returning an error, but also...
	// we know the data we're encoding was valid YAML and we're writing
	// to a buffer, so there doesn't seem to be an error we
	// could encounter that would be meaningful.
	_ = j.Encode(a)
	return MapPrefix + buf.String()
}

// Undecorate removes type decorations from strings and returns the desired type.
// These decorations were placed there by the encoding functions in this package.
// Since everything that comes out of a Go template is a string, for things that
// needed to not be strings, we flagged them with a decoration indicating the
// desired type. Now we need to do some extra work to make sure that we return
// the indicated type. If it can't be converted to the desired type, we return
// the string as is.
func Undecorate(s string) any {
	switch {
	case strings.HasPrefix(s, IntPrefix):
		s = strings.TrimPrefix(s, IntPrefix)
		i, err := strconv.Atoi(s)
		if err == nil {
			return i
		}
	case strings.HasPrefix(s, BoolPrefix):
		s = strings.TrimPrefix(s, BoolPrefix)
		b, err := strconv.ParseBool(s)
		if err == nil {
			return b
		}
	case strings.HasPrefix(s, FloatPrefix):
		s = strings.TrimPrefix(s, FloatPrefix)
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f
		}
	case strings.HasPrefix(s, ArrPrefix):
		s = strings.TrimPrefix(s, ArrPrefix)
		items := strings.Split(s, FieldSeparator)
		// we need to trim the spaces from the items and we don't want blanks
		// in the array
		var arr []string
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item != "" {
				arr = append(arr, item)
			}
		}
		return arr
	case strings.HasPrefix(s, MapPrefix):
		s = strings.TrimPrefix(s, MapPrefix)
		// s is encoded as a JSON map, so we need to decode it
		var m map[string]any
		// we ignore the error here because the input string
		// was marshaled by us and we know it's valid JSON,
		// and there's nothing we can do with it anyway.
		json.Unmarshal([]byte(s), &m)
		return m
	}
	return s
}

// getStringsFrom converts various types to a slice of strings
func getStringsFrom(value any) []string {
	result := make([]string, 0)

	if ary, ok := value.([]string); ok {
		return ary
	}

	if ary, ok := value.([]any); ok {
		for _, elt := range ary {
			if v, ok := elt.(string); ok {
				result = append(result, v)
			}
		}
	}
	return result
}
