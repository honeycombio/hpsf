package data

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	y "gopkg.in/yaml.v3"
)

// Reads a set of components from the local embedded filesystem (in the source, this is the
// data/components directory) and loads them into a map of TemplateComponent by name.
func LoadEmbeddedComponents() (map[string]config.TemplateComponent, error) {
	// Read the components from the filesystem
	comps, err := ComponentsFS.ReadDir("components")
	if err != nil {
		return nil, err
	}

	// Load each template
	components := make(map[string]config.TemplateComponent)
	for _, comp := range comps {
		templateData, err := ComponentsFS.ReadFile("components/" + comp.Name())
		if err != nil {
			return nil, err
		}

		var component config.TemplateComponent
		err = y.Unmarshal(templateData, &component)
		if err != nil {
			return nil, err
		}

		if _, ok := components[component.Kind]; ok {
			return nil, fmt.Errorf("duplicate component kind %s in %s and %s",
				component.Kind, components[component.Kind].Name, component.Name)
		}
		components[component.Kind] = component
	}

	return components, nil
}
