package config

import (
	"embed"
	"fmt"

	y "gopkg.in/yaml.v3"
)

//go:embed components/*.yaml
var componentsFS embed.FS

// This reads the components from the filesystem (in the source, this is the
// components directory) and loads them into a map of TemplateComponent by name.
func LoadTemplateComponents() (map[string]TemplateComponent, error) {
	// Read the components from the filesystem
	comps, err := componentsFS.ReadDir("components")
	if err != nil {
		return nil, err
	}

	// Load each template
	components := make(map[string]TemplateComponent)
	for _, comp := range comps {
		templateData, err := componentsFS.ReadFile("components/" + comp.Name())
		if err != nil {
			return nil, err
		}

		var component TemplateComponent
		err = y.Unmarshal(templateData, &component)
		if err != nil {
			fmt.Println(comp.Name(), err)
			return nil, err
		}

		if _, ok := components[component.Kind]; ok {
			return nil, fmt.Errorf("duplicate component kind %s in %s and %s", component.Kind, components[component.Kind].Name, component.Name)
		}
		components[component.Kind] = component
	}

	return components, nil
}
