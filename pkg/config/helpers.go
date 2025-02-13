package config

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"
)

const (
	RecordSeparator = "\x1e" // ASCII code 30, the record separator
)

// This file contains template helper functions, which must be listed in this
// map if they're going to be available to the template.
// The map key is the name of the function as it will be used in the template,
// and the value is the function itself.
// The function must return a string, and may take any number of arguments.
// The functions are listed below in alphabetical order; please keep them that way.
func helpers() template.FuncMap {
	return map[string]any{
		"comment":      comment,
		"firstNonZero": firstNonzero,
		"indent":       indent,
		"join":         join,
		"makeSlice":    makeSlice,
		"mapify":       mapify,
		"meta":         meta,
		"now":          now,
		"split":        split,
		"yamlf":        yamlf,
	}
}

// places a comment in the output file, even if the specified comment has multiple lines
func comment(s string) string {
	return strings.TrimRight("## "+strings.Replace(s, "\n", "\n## ", -1), " ")
}

// returns the first non-zero-valued item from the arguments
// []any is special-cased to return a comma-separated set of strings.
// If we eventually feel like the comma syntax is failing to handle some special
// cases, we can change it to use some other syntax that's less likely to occur
// in real data.
func firstNonzero(s ...any) string {
	for _, v := range s {
		if !_isZeroValue(v) {
			switch vt := v.(type) {
			case string:
				return fmt.Sprintf("%v", vt)
			case []any:
				ss := make([]string, len(vt))
				for i, vv := range vt {
					ss[i] = fmt.Sprintf("%v", vv)
				}
				return strings.Join(ss, ", ")
			case []string:
				return strings.Join(vt, ", ")
			default:
				return fmt.Sprintf("%v", vt)
			}
		}
	}
	return ""
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

// mapify takes a map and returns a string intended to be expanded
// later into a map when it's rendered to YAML.
// The result looks like "map:a:1\x1eb:2\x1e"
func mapify(a map[string]any) string {
	buf := bytes.Buffer{}
	buf.WriteString("map:")
	for k, v := range a {
		buf.WriteString(k)
		buf.WriteRune(':')
		buf.WriteString(fmt.Sprintf("%v", v))
		buf.WriteString(RecordSeparator)
	}
	return buf.String()
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
// Some of these are currently unused but will likely be used in the future.

// internal function to compare two "any" values for equivalence
func _equivalent(a, b any) bool {
	va := fmt.Sprintf("%v", a)
	vb := fmt.Sprintf("%v", b)
	return va == vb
}

// this formats an integer with underscores for readability.
// e.g. 1000000 becomes 1_000_000
func _formatIntWithUnderscores(i int) string {
	s := fmt.Sprintf("%d", i)
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

// Takes a key that may or may not be in the incoming data,
// and returns the value found, possibly doing a recursive call
// separated by dots in the key.
func _fetch(data map[string]any, key string) (any, bool) {
	if value, ok := data[key]; ok {
		return value, true
	}
	if strings.Contains(key, ".") {
		parts := strings.SplitN(key, ".", 2)
		groups := strings.Split(parts[0], "/")
		for _, g := range groups {
			if value, ok := data[g]; ok {
				if submap, ok := value.(map[string]any); ok {
					return _fetch(submap, parts[1])
				}
			}
		}
	}
	return nil, false
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
