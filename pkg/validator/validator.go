package validator

import (
	y "gopkg.in/yaml.v3"
)

type Error struct {
	Details []error
	Msg     string
}

func (e Error) Error() string {
	return e.Msg
}

func (e Error) Unwrap() []error {
	return e.Details
}

func NewError(msg string) Error {
	return Error{
		Msg:     msg,
		Details: nil,
	}
}

func NewErrorWith(msg string, err error) Error {
	return Error{
		Msg:     msg,
		Details: []error{err},
	}
}

// Add adds an error to the list of errors; it's a no-op if the error is nil
// If the error is a validator.Error, it will be flattened into the current list
func (e *Error) Add(err error) {
	if e == nil {
		return
	}
	if err == nil {
		return
	}
	if other, ok := err.(Error); ok {
		e.Details = append(e.Details, other.Details...)
		// we always want to keep our own message as it provides the outer context
		if e.Msg == "" {
			e.Msg = other.Msg
		}
		return
	}
	e.Details = append(e.Details, err)
}

func (e Error) Len() int {
	return len(e.Details)
}

func (e Error) ErrOrNil() error {
	if e.Len() == 0 {
		return nil
	}
	return e
}

// Validator is an interface that can be implemented by any struct that needs to be validated
// It returns a list of errors that are encountered during validation; this list may be empty or nil.
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
