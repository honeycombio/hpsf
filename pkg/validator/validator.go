package validator

import (
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

func (e Result) Unwrap() []error {
	return e.Details
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

func (e Result) Len() int {
	return len(e.Details)
}

func (e Result) ErrOrNil() error {
	if e.Len() == 0 {
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
