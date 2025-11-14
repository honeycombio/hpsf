package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// A TemplateProperty describes a property of a component. A property is a
// user-settable value that can be used to configure the component. Properties
// have a name, a type (which can be used to validate the value), and a default
// value. The advanced flag can be used to indicate that the property should be
// suppressed by default in the UI (only shown if the user selects an "advanced"
// option). We also allow for validations, which can be used to constrain the
// value of the property. The property can also have a summary and a
// description, which are used to document the property. The display field
// provides a human-readable label for the property in the UI.
// It is explicitly intended that the value of the display field can change 
// without affecting any existing use of the component (and without a version bump)
type TemplateProperty struct {
	Name        string        `yaml:"name"`
	Display     string        `yaml:"display"`
	Summary     string        `yaml:"summary,omitempty"`
	Description string        `yaml:"description,omitempty"`
	Type        hpsf.PropType `yaml:"type"`
	Subtype     string        `yaml:"subtype,omitempty"`
	Advanced    bool          `yaml:"advanced,omitempty"`
	Validations []string      `yaml:"validations,omitempty"`
	Default     any           `yaml:"default,omitempty"`
}

// returns a validation function based on the validation string provided.
// the validation string can include a set of comma-separated arguments in parens,
// which will be bound to the validation function when appropriate.
func getValidationRule(validation string) func(val any) bool {
	// parse off any arguments
	valpat := regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?$`)
	argpat := regexp.MustCompile(`[\t ]*,[\t ]*`)
	// Extract the base validation name and any arguments
	matches := valpat.FindStringSubmatch(validation)
	if len(matches) == 0 {
		// this is a bug in the validation string
		return alwaysFail
	}

	// matches[1] is the validation name
	name := matches[1]
	args := []string{}
	// matches[2] contains the arguments
	if len(matches) == 3 {
		if len(matches[2]) > 0 {
			args = argpat.Split(matches[2], -1) // split on commas, remove empty strings
		}
	}

	switch name {
	case "positive":
		// positive will check if a numeric value is greater than 0
		// nonumeric types will return false
		return positive
	case "noblanks":
		// noblanks will check that no strings in the value are blank;
		// this is useful for strings and slices
		return noBlankStrings
	case "nonempty":
		// nonempty will check if the value has at least one element; this is
		// useful for strings, slices, maps
		return nonempty
	case "oneof":
		// oneof will check if the value is exactly one of the provided options
		// (using only string comparison)
		return oneof(args...)
	case "url":
		// url will check if the value can be parsed as a URL
		return isURL
	case "hostorip":
		// hostorip will check if the value is a valid host or IP address
		// without a scheme or port.
		return isHostOrIP
	case "duration":
		// duration will check if the value can be parsed as a duration string
		return isDuration
	case "atleast":
		// atLeast will check if the value is greater than or equal to the provided argument
		return atLeast(args...)
	case "atmost":
		// atMost will check if the value is less than or equal to the provided argument
		return atMost(args...)
	case "inrange":
		// inRange will check if the value is within a specified range
		return inRange(args...)
	case "regex":
		// regex will check if the value is a valid regular expression
		return isValidRegex
	default:
		// If no match, always return false
		return alwaysFail
	}
}

// Validate validates the given HPSF property against the TemplateProperty.
// It runs the property's own validation, plus any additional validations specified in the Validations field.
// Returns an hpsf.Error if the validation fails, or nil if the validation passes.
func (tp *TemplateProperty) Validate(prop hpsf.Property) error {
	var value any
	// this tries to coerce the value of the property to the type specified in the TemplateProperty
	// and stores it in the `value` variable
	if err := prop.Type.ValueCoerce(prop.Value, &value); err != nil {
		// if the type of the property does not match the type of the template property, return an error
		return hpsf.NewError("value cannot be converted to expected type " + tp.Type.String()).
			WithProperty(tp.Name).
			WithCause(err)
	}

	// If there are additional validations, run them against "value" which
	// has been coerced to the expected type. This allows for additional
	// validations to be applied beyond just type checking.
	for _, validation := range tp.Validations {
		// get the validation function based on the validation string
		validationFunc := getValidationRule(validation)
		// run the validation function against the coerced value
		if !validationFunc(value) {
			// if the validation fails, return an error
			return hpsf.NewError("validation failed for property: " + validation).
				WithProperty(tp.Name)
		}
	}
	return nil
}

// always returns false
func alwaysFail(_ any) bool {
	return false
}

func positive(val any) bool {
	switch v := val.(type) {
	case int:
		return v > 0
	case float64:
		return v > 0
	default:
		return false
	}
}

// ensures that no strings are blank, even in string slices
// use nonempty to ensure that a slace has at least one value
func noBlankStrings(val any) bool {
	// nil happens when there is no default value and no value is supplied
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case string:
		return len(v) > 0
	case []string:
		for _, s := range v {
			if len(s) == 0 {
				return false
			}
		}
	case []any:
		for _, a := range v {
			if s, ok := a.(string); ok {
				if len(s) == 0 {
					return false
				}
			}
		}
	}
	return true
}

// nonempty checks if a value that has a length (like a string, slice, or map)
// is non-empty. Other types are always considered nonempty.
func nonempty(val any) bool {
	switch v := val.(type) {
	case string:
		return len(v) > 0
	case []any:
		return len(v) > 0
	case []string:
		return len(v) > 0
	case []int:
		return len(v) > 0
	case []float64:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	default:
		return true // for other types, we consider them non-empty
	}
}

// checks if a string is one of the provided options (case sensitive)
// this returns a function that checks if the parameter, when converted to a string,
// is exactly equal to one of the options provided.
func oneof(options ...string) func(val any) bool {
	return func(val any) bool {
		strVal := fmt.Sprint(val) // ensure we have a string to compare against
		for _, option := range options {
			if strVal == option {
				return true
			}
		}
		return false
	}
}

// checks that the value can be parsed as a URL and contains a non-empty scheme and host.
// it also should not contain a port
// we special-case environment variable expansions
func isURL(val any) bool {
	s := fmt.Sprint(val)
	// If the value is an environment variable, we don't validate it here
	if len(s) > 0 && s[0] == '$' {
		return true // environment variables are always valid
	}

	u, ok := url.Parse(s)
	if ok != nil {
		return false
	}
	// Check if the URL has a scheme and host
	if u.Scheme == "" || u.Host == "" {
		// If the scheme or host is empty, it's not a valid URL for our purposes
		return false
	}
	// Ensure the URL does not contain a port
	if u.Port() != "" {
		return false
	}
	return true
}

// Matches a valid hostname without checking it on the net. This pattern taken from
// https://github.com/asaskevich/govalidator/blob/master/patterns.go#L33
// The govalidator library is big, intended to validate structs, and
// uses reflect, so we don't want to depend on it just for this pattern.
var dnsNamePat = regexp.MustCompile(`^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`)

// checks if the value is a valid host or IP address without a scheme or port
// we special-case environment variable expansions
func isHostOrIP(val any) bool {
	s := fmt.Sprint(val)
	// If the value is an environment variable, we don't validate it here
	if len(s) > 0 && s[0] == '$' {
		return true // environment variables are always valid
	}

	ip := net.ParseIP(s)
	if ip != nil {
		return true
	}

	if dnsNamePat.MatchString(s) {
		// If it matches the DNS name pattern, we're going to say it's a valid host
		return true
	}

	return false
}

// isDuration checks if the value can be parsed as a duration string
func isDuration(val any) bool {
	_, err := time.ParseDuration(fmt.Sprint(val))
	return err == nil
}

// this returns a function that checks if the value is at least
// the minimum value provided.
func atLeast(options ...string) func(val any) bool {
	// if no options provided, always fail
	if len(options) == 0 {
		return alwaysFail
	}
	minval, err := strconv.ParseFloat(options[0], 64)
	if err != nil {
		return alwaysFail
	}

	return func(val any) bool {
		switch v := val.(type) {
		case int:
			return v >= int(minval)
		case float64:
			return v >= minval
		default:
			return false // for other types, return false
		}
	}
}

// this returns a function that checks if the value is at least
// the minimum value provided.
func atMost(options ...string) func(val any) bool {
	// if no options provided, always fail
	if len(options) == 0 {
		return alwaysFail
	}
	maxval, err := strconv.ParseFloat(options[0], 64)
	if err != nil {
		return alwaysFail
	}

	return func(val any) bool {
		switch v := val.(type) {
		case int:
			return v <= int(maxval)
		case float64:
			return v <= maxval
		default:
			return false // for other types, return false
		}
	}
}

// rangeCheck checks if the value falls within a specified range;
// it expects 2 arguments (min and max) to be passed in as strings
func inRange(options ...string) func(val any) bool {
	// if no options provided, always fail
	if len(options) < 2 {
		return alwaysFail // need at least two options for a range
	}
	// parse the minimum and maximum values
	minval, err := strconv.ParseFloat(options[0], 64)
	if err != nil {
		return alwaysFail // if we can't parse the minimum value, fail
	}
	maxval, err := strconv.ParseFloat(options[1], 64)
	if err != nil {
		return alwaysFail // if we can't parse the maximum value, fail
	}

	if minval > maxval {
		maxval, minval = minval, maxval // ensure min is less than max
	}

	return func(val any) bool {
		switch v := val.(type) {
		case int:
			return v >= int(minval) && v <= int(maxval)
		case float64:
			return v >= minval && v <= maxval
		default:
			return false // for other types, return false
		}
	}
}

// isValidRegex checks if the value is a valid regular expression
func isValidRegex(val any) bool {
	s := fmt.Sprint(val)
	// If the value is an environment variable, we don't validate it here
	if len(s) > 0 && s[0] == '$' {
		return true // environment variables are always valid
	}

	_, err := regexp.Compile(s)
	return err == nil
}
