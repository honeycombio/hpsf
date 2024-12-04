package validator

import (
	"fmt"

	y "gopkg.in/yaml.v3"
)

type Error struct {
	Msg string
}

func (e Error) Error() string {
	return e.Msg
}

func NewError(msg string) error {
	return Error{Msg: msg}
}

func NewErrorf(format string, args ...interface{}) error {
	return Error{Msg: fmt.Sprintf(format, args...)}
}

// Validator is an interface that can be implemented by any struct that needs to be validated
// It returns a list of errors that are encountered during validation; this list may be empty or nil.
type Validator interface {
	Validate() error
}

// this is a brain-dead validator that just tries to unmarshal the input into appropriate forms
func EnsureYAML(input []byte) error {
	// validate the input is parseable YAML

	// try unmarshaling into map
	var hpsfMap map[string]any
	err := y.Unmarshal(input, &hpsfMap)
	if err != nil {
		return err
	}

	return nil
}
