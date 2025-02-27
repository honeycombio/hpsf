package data

import (
	"crypto/sha1"
	"fmt"
	"path"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	y "gopkg.in/yaml.v3"
)

// Reads a set of components from the local embedded filesystem (in the source, this is the
// data/components directory) and loads them into a map of TemplateComponent by name.
func LoadEmbeddedComponents() (map[string]config.TemplateComponent, error) {
	// Read the components from the filesystem
	comps, err := EmbeddedFS.ReadDir("components")
	if err != nil {
		return nil, err
	}

	// Load each template
	components := make(map[string]config.TemplateComponent)
	for _, comp := range comps {
		templateData, err := EmbeddedFS.ReadFile(path.Join("components", comp.Name()))
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
	temps, err := EmbeddedFS.ReadDir("templates")
	if err != nil {
		return nil, err
	}

	// Load each template
	templates := make(map[string]hpsf.HPSF)
	for _, comp := range temps {
		templateData, err := EmbeddedFS.ReadFile(path.Join("templates", comp.Name()))
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

// CalculateChecksums reads the templates in a directory in the embedded
// filesystem and calculates a sha1 checksum of the contents. It will generate
// the same results as doing `sha1sum subdir/*` from the data directory. This
// is used to verify that the templates have not changed since the last release.
// Yes, we know that sha1 is not a secure hash, but we're not using it for security,
// and compatibility with a well-known command line tool is a feature.
func CalculateChecksums(subdir string) (map[string]string, error) {
	// Read the components from the filesystem
	temps, err := EmbeddedFS.ReadDir(subdir)
	if err != nil {
		return nil, err
	}

	results := make(map[string]string)

	// Read and sha1 sum each file
	for _, comp := range temps {
		checksum := sha1.New()
		name := path.Join(subdir, comp.Name())
		templateData, err := EmbeddedFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		_, err = checksum.Write(templateData)
		if err != nil {
			return nil, err
		}
		result := checksum.Sum(nil)
		results[name] = fmt.Sprintf("%x", result)
	}

	return results, nil
}
