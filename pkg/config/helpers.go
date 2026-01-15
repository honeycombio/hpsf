package config

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/honeycombio/hpsf/pkg/config/decorator"
)

// This file contains template helper functions, which must be listed in this
// map if they're going to be available to the template.
// The map key is the name of the function as it will be used in the template,
// and the value is the function itself.
// The function can return a value of any type, and may take any number of arguments.
// The functions are listed below in alphabetical order; please keep them that way.
func helpers() template.FuncMap {
	return map[string]any{
		"appendSlices":  appendSlices,
		"buildurl":      buildurl,
		"comment":       comment,
		"encodeAsArray": decorator.EncodeAsArray,
		"encodeAsBool":  decorator.EncodeAsBool,
		"encodeAsInt":   decorator.EncodeAsInt,
		"encodeAsFloat": decorator.EncodeAsFloat,
		"encodeAsMap":   decorator.EncodeAsMap,
		"getChecked":    getChecked,
		"indent":        indent,
		"join":          join,
		"lower":         strings.ToLower,
		"makeSlice":     makeSlice,
		"mapValues":     mapValues,
		"meta":          meta,
		"nonempty":      nonempty,
		"now":           now,
		"processOTTL":   processOTTL,
		"split":         split,
		"upper":         strings.ToUpper,
		"yamlf":         yamlf,
	}
}

// appendSlices combines two slices into one.
func appendSlices(slice1 []any, slice2 []any) []any {
	return append(slice1, slice2...)
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

// mapValues extracts values from a map and returns them as a slice
// Values are returned in sorted order by key for deterministic output
func mapValues(m any) []any {
	result := make([]any, 0)

	if mapVal, ok := m.(map[string]any); ok {
		keys := slices.Collect(maps.Keys(mapVal))
		slices.Sort(keys)
		for _, k := range keys {
			result = append(result, mapVal[k])
		}
	} else if mapVal, ok := m.(map[string]string); ok {
		keys := slices.Collect(maps.Keys(mapVal))
		slices.Sort(keys)
		for _, k := range keys {
			result = append(result, mapVal[k])
		}
	}

	return result
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

// This accepts a block of text as a string, and splits it into an array of individual lines,
// eliminating comments and blank lines, and also leading "-" characters that are used in
// YAML lists that might be copy-pasted into the component. (These get put back on later.)
func processOTTL(statements any) []string {
	var lines, result []string
	switch s := statements.(type) {
	case string:
		lines = strings.Split(s, "\n")
	case []string:
		lines = s
	case []any:
		lines = _getStringsFrom(s)
	default:
		return []string{fmt.Sprintf("expected a string in processOTTL, got %T", statements)}
	}
	for _, line := range lines {
		line = strings.Trim(line, " \t\r-")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		result = append(result, line)
	}
	return result
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

// ChecklistItem represents an item in a checklist property definition
type ChecklistItem struct {
	ID          string `yaml:"id"`
	DisplayName string `yaml:"displayName"`
	Value       string `yaml:"value"`
	TooltipText string `yaml:"tooltipText"`
}

// getChecked takes a TemplateProperty (checklist type) and a list of checked IDs,
// and returns the values from the subtype definition for the checked items.
// This is used in templates to get the actual regex patterns for selected checklist items.
func getChecked(prop TemplateProperty, checkedIDs any) []any {
	if prop.Type.String() != "checklist" {
		return []any{}
	}

	// Handle subtype as []any containing checklist items
	subtypeSlice, ok := prop.Subtype.([]any)
	if !ok {
		return []any{}
	}

	// Convert []any to []ChecklistItem
	var items []ChecklistItem
	for _, item := range subtypeSlice {
		if itemMap, ok := item.(map[string]any); ok {
			checklistItem := ChecklistItem{}
			if id, ok := itemMap["id"].(string); ok {
				checklistItem.ID = id
			}
			if displayName, ok := itemMap["displayName"].(string); ok {
				checklistItem.DisplayName = displayName
			}
			if value, ok := itemMap["value"].(string); ok {
				checklistItem.Value = value
			}
			if tooltipText, ok := itemMap["tooltipText"].(string); ok {
				checklistItem.TooltipText = tooltipText
			}
			items = append(items, checklistItem)
		}
	}

	// Convert checkedIDs to a map for quick lookup
	checkedMap := make(map[string]bool)
	for _, id := range _getStringsFrom(checkedIDs) {
		checkedMap[id] = true
	}

	// Extract values for checked items
	var result []any
	for _, item := range items {
		if checkedMap[item.ID] {
			result = append(result, item.Value)
		}
	}

	return result
}
