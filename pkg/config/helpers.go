package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
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

// This file contains template helper functions, which must be listed in this
// map if they're going to be available to the template.
// The map key is the name of the function as it will be used in the template,
// and the value is the function itself.
// The function can return a value of any type, and may take any number of arguments.
// The functions are listed below in alphabetical order; please keep them that way.
func helpers() template.FuncMap {
	return map[string]any{
		"buildurl":           buildurl,
		"comment":            comment,
		"encodeAsArray":      encodeAsArray,
		"encodeAsBool":       encodeAsBool,
		"encodeAsInt":        encodeAsInt,
		"encodeAsFloat":      encodeAsFloat,
		"encodeAsMap":        encodeAsMap,
		"encodeAsMapWithKey": encodeAsMapWithKey,
		"indent":             indent,
		"join":               join,
		"makeSlice":          makeSlice,
		"meta":               meta,
		"nonempty":           nonempty,
		"now":                now,
		"split":              split,
		"upper":              strings.ToUpper,
		"yamlf":              yamlf,
	}
}

// buildurl constructs a URL based on the provided parameters. A path is optional.
func buildurl(args ...any) string {
	var insecure bool
	var port int
	var host, path string
	switch len(args) {
	case 4:
		path = args[3].(string)
	case 3:
		path = ""
	default:
		return ""
	}

	insecure = args[0].(bool)
	host = args[1].(string)
	port = _asInt(args[2])
	scheme := "https"
	if insecure {
		scheme = "http"
	}

	url := fmt.Sprintf("%s://%s:%d", scheme, host, port)
	if path != "" {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		url += path
	}
	return url
}

// places a comment in the output file, even if the specified comment has multiple lines
func comment(s string) string {
	return strings.TrimRight("## "+strings.Replace(s, "\n", "\n## ", -1), " ")
}

// encodeAsArray takes a slice and returns a string intended to be expanded
// later into an array when it's rendered to YAML.
// The result looks like "arr\x1dA:1\x1fB:2"
func encodeAsArray(arr any) string {
	switch a := arr.(type) {
	case []string:
		return ArrPrefix + strings.Join(a, FieldSeparator)
	case []any:
		return ArrPrefix + strings.Join(_getStringsFrom(arr), FieldSeparator)
	default:
		return ""
	}
}

// encodeAsBool takes any value and returns a string with the appropriate marker
// so that it will be expanded later into a bool when it's rendered to YAML.
// Numbers are interpreted as true if they are not zero.
func encodeAsBool(a any) string {
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

// encodeAsFloat takes a string and returns a string with the appropriate marker
// so that it will be expanded later into a float when it's rendered to YAML.
func encodeAsFloat(a any) string {
	value := "0"
	switch v := a.(type) {
	case int:
		value = strconv.Itoa(v)
	case float64:
		value = fmt.Sprintf("%f", v)
	case string:
		value = v
	case bool:
		if v {
			value = "1"
		}
	}
	return FloatPrefix + value
}

// encodeAsInt takes a string and returns a string with the appropriate marker
// so that it will be expanded later into an integer when it's rendered to YAML.
func encodeAsInt(a any) string {
	value := "0"
	switch v := a.(type) {
	case int:
		value = strconv.Itoa(v)
	case float64:
		value = fmt.Sprintf("%f", v)
	case string:
		value = v
	case bool:
		if v {
			value = "1"
		}
	}
	return IntPrefix + value
}

// encodeAsMap takes a map (which may contain nested maps) and returns a string
// intended to be expanded later into the same map when it's rendered to YAML.
// We encode to JSON because it's fast and easy.
func encodeAsMap(a map[string]any) string {
	buf := bytes.Buffer{}
	j := json.NewEncoder(&buf)
	// There's no model for returning an error, but also...
	// we know the data we're encoding was valid YAML and we're writing
	// to a buffer, so there doesn't seem to be an error we
	// could encounter that would be meaningful.
	_ = j.Encode(a)
	return MapPrefix + buf.String()
}

// encodeAsMapWithKey takes a key and a value (which must be a map) and returns a string
// intended to be expanded later into the same map when it's rendered to YAML.
// The key is used as the key for the outer map.
func encodeAsMapWithKey(customKey string, v any) string {
	if v.(map[string]any) == nil {
		return ""
	}

	m := v.(map[string]any)
	for key, value := range m {
		m[key] = map[string]any{customKey: value}
	}

	return encodeAsMap(m)
}

// indents a string by the specified number of spaces
func indent(count int, s string) string {
	return strings.Repeat(" ", count) + _indentRest(count, s)
}

// joins a slice of strings with the specified separator
func join(a []string, sep string) string {
	return strings.Join(a, sep)
}

// creates a slice of strings from the arguments
func makeSlice(a ...string) []string {
	return a
}

// wraps a string in "{{" and "}}" to indicate that it's a template variable
func meta(s string) string {
	return "{{ " + s + " }}"
}

// returns the current date and time in UTC
func now() string {
	t := time.Now().UTC()
	return fmt.Sprintf("on %s at %s UTC", t.Format("2006-01-02"), t.Format("15:04:05"))
}

// splits a string into a slice of strings using the specified separator
func split(s, sep string) []string {
	return strings.Split(s, sep)
}

// simplistic YAML formatting of a value
func yamlf(a any) string {
	switch v := a.(type) {
	case string:
		// if it's a plain string, return it as a string
		pat := regexp.MustCompile("^[a-zA-z0-9]+$")
		if pat.MatchString(v) {
			return v
		}
		// play some games with quotes to make it look better
		hasSingleQuote := strings.Contains(v, "'")
		hasDoubleQuote := strings.Contains(v, `"`)
		switch {
		case hasDoubleQuote && !hasSingleQuote:
			return fmt.Sprintf(`'%s'`, v)
		default:
			return fmt.Sprintf("%#v", v)
		}
	case int:
		return _formatIntWithUnderscores(v)
	case float64:
		return fmt.Sprintf("%f", v)
	case time.Duration:
		return v.String()
	default:
		return fmt.Sprintf("%v", a)
	}
}

// The functions below are internal to this file hence the leading underscore.

// this formats an integer with underscores for readability.
// e.g. 1000000 becomes 1_000_000
func _formatIntWithUnderscores(i int) string {
	s := strconv.Itoa(i)
	var output []string
	for len(s) > 3 {
		output = append([]string{s[len(s)-3:]}, output...)
		s = s[:len(s)-3]
	}
	output = append([]string{s}, output...)
	return strings.Join(output, "_")
}

// indents a string by the specified number of spaces, but only after newlines (used by indent)
func _indentRest(count int, s string) string {
	eolpat := regexp.MustCompile(`[ \t]*\n[ \t]*`)
	return eolpat.ReplaceAllString(s, "\n"+strings.Repeat(" ", count))
}

func _isZeroValue(value any) bool {
	switch v := value.(type) {
	case string:
		return v == ""
	case int:
		return v == 0
	case int64:
		return v == 0
	case float64:
		return v == 0.0
	case bool:
		return !v
	case []string:
		return len(v) == 0
	case []any:
		return len(v) == 0
	case map[string]string:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	case nil:
		return true
	default:
		return false
	}
}

// Takes a value that is a slice of strings or a slice of any and returns a
// slice of strings.
func _getStringsFrom(value any) []string {
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

// Converts a value to an int, handling various types.
func _asInt(a any) int {
	switch v := a.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}
