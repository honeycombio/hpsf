// contains the validation logic for the HPSF datatype.
package hpsf

import (
	"errors"

	"github.com/honeycombio/hpsf/pkg/validator"
)

// validateNames checks that all component names are unique.
func (h *HPSF) validateNames() error {
	result := validator.NewResult("hpsf name validation errors")

	// check that all components have unique names
	nameSet := make(map[string]struct{})
	for _, c := range h.Components {
		if _, exists := nameSet[c.GetSafeName()]; exists {
			result.Add(NewError("duplicate component name").WithComponent(c.Name))
		} else {
			nameSet[c.GetSafeName()] = struct{}{}
		}
	}

	return result.ErrOrNil()
}

// validateConnectionSources checks that all connections have valid source and
// destination components. We can't check port names because they are only available
// after instantiating the real components.
func (h *HPSF) validateConnectionSources() error {
	result := validator.NewResult("hpsf connection source validation errors")
	for _, c := range h.Connections {
		src := h.getComponent(c.Source.Component)
		if src == nil {
			result.Add(NewError("Connection source component not found").WithComponent(c.Source.Component))
		}

		dst := h.getComponent(c.Destination.Component)
		if dst == nil {
			result.Add(NewError("Connection destination component not found").WithComponent(c.Destination.Component))
		}
	}

	return result.ErrOrNil()
}

// Validate checks that the HPSF is valid, returning a list of errors if it is not.
// If it detects minor issues that can be corrected, it will fix them and return.
// For example, if a property specifies that it requires an integer but the value
// is a string that can be parsed as an integer, it will parse it and store the
// result as an integer in the value.
func (h *HPSF) Validate() error {
	result := validator.NewResult("hpsf validation errors")

	// if the HPSF is empty, it's invalid
	if len(h.Components) == 0 && len(h.Containers) == 0 {
		result.Add(errors.New("empty HPSF is not valid"))
	}

	for _, c := range h.Components {
		e := c.Validate()
		result.Add(e)
	}
	for _, c := range h.Connections {
		e := c.Validate()
		result.Add(e)
	}
	for _, c := range h.Containers {
		e := c.Validate()
		result.Add(e)
	}

	result.Add(h.validateConnectionSources())
	result.Add(h.validateNames())

	return result.ErrOrNil()
}
