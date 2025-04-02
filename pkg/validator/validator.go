package validator

import (
	"errors"

	y "gopkg.in/yaml.v3"
)

// Result is the value returned by the Validate method; it conforms to the error interface
// and also contains a list of errors and a message.
type Result struct {
	Details []error
	Msg     string
}

func (e Result) Error() string {
	return e.Msg
}

func (e Result) Unwrap() error {
	output := []error{}
	for _, d := range e.Details {
		// if any of the details are themselves a Result, unpack them recursively and add to the output.
		if other, ok := d.(Result); ok {
			// recursively unpack the details of the other Result
			output = append(output, other.Unwrap())
		} else {
			// otherwise just append the error to the output
			output = append(output, d)
		}
	}
	return errors.Join(output...)
}

func NewResult(msg string) Result {
	return Result{
		Msg:     msg,
		Details: nil,
	}
}

// Add adds an error to the list of results; it's a no-op if the error is nil
// If the error is a validator.Error, it will be flattened into the current list
func (e *Result) Add(err error) {
	if e == nil {
		return
	}
	if err == nil {
		return
	}
	if other, ok := err.(Result); ok {
		e.Details = append(e.Details, other.Details...)
		// we always want to keep our own message as it provides the outer context
		if e.Msg == "" {
			e.Msg = other.Msg
		}
		return
	}
	e.Details = append(e.Details, err)
}

// Len returns the number of errors in the Result.
func (e Result) Len() int {
	return len(e.Details)
}

// IsEmpty returns true if there are no errors in the Result
// This is a convenience method to check if the Result has any errors
func (e Result) IsEmpty() bool {
	return e.Len() == 0
}

func (e Result) ErrOrNil() error {
	if e.IsEmpty() {
		return nil
	}
	return e
}

// Validator is an interface that can be implemented by any struct that needs to be validated.
// It returns an error that may be a simple error or a Result, in case it's useful to provide multiple
// errors.
type Validator interface {
	Validate() error
}

// this is a brain-dead validator that just tries to unmarshal the input into appropriate forms
func EnsureYAML(input []byte) (map[string]any, error) {
	// validate the input is parseable YAML (parses into a map)

	// try unmarshaling into map
	var hpsfMap map[string]any
	err := y.Unmarshal(input, &hpsfMap)
	if err != nil {
		return nil, err
	}

	return hpsfMap, nil
}
