package data

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
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

// Reads a set of templates from the local embedded filesystem (in the source, this is the
// data/templates directory) and loads them into a map of TemplateComponent by name.
func LoadEmbeddedTemplates() (map[string]hpsf.HPSF, error) {
	// Read the components from the filesystem
	temps, err := TemplatesFS.ReadDir("templates")
	if err != nil {
		return nil, err
	}

	// Load each template
	templates := make(map[string]hpsf.HPSF)
	for _, comp := range temps {
		templateData, err := TemplatesFS.ReadFile("templates/" + comp.Name())
		if err != nil {
			return nil, err
		}

		var template hpsf.HPSF
		err = y.Unmarshal(templateData, &template)
		if err != nil {
			return nil, err
		}

		if _, ok := templates[template.Kind]; ok {
			return nil, fmt.Errorf("duplicate template kind %s in %s and %s",
				template.Kind, templates[template.Kind].Name, template.Name)
		}
		templates[template.Kind] = template
	}

	return templates, nil
}
