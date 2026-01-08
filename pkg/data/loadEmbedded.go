package data

import (
	"crypto/sha1"
	"fmt"
	"path"
	"strings"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	y "gopkg.in/yaml.v3"
)

const DefaultConfigurationKind = "TemplateDefault"

// LoadEmbeddedComponents reads a set of components from the local embedded filesystem (in the source, this is the
// data/components directory) and loads them into a map of TemplateComponent by name.
// Components are organized in a 2-level structure: components/{style}/{component_name}/component.yaml
// where style is receivers/processors/exporters/samplers/conditions/startsampling
func LoadEmbeddedComponents() (map[string]config.TemplateComponent, error) {
	// Read the components from the filesystem
	entries, err := EmbeddedFS.ReadDir("components")
	if err != nil {
		return nil, err
	}

	components := make(map[string]config.TemplateComponent)

	for _, entry := range entries {
		// Skip non-directories and special directories (starting with _ or .)
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), "_") || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Level 1: style directory (receivers, processors, exporters, samplers, conditions, startsampling)
		styleDir := entry.Name()
		stylePath := path.Join("components", styleDir)

		// Read component directories within style
		componentEntries, err := EmbeddedFS.ReadDir(stylePath)
		if err != nil {
			continue
		}

		for _, componentEntry := range componentEntries {
			if !componentEntry.IsDir() {
				continue
			}

			// Level 2: component directory
			// Component YAML file matches directory name (e.g., my_component/my_component.yaml)
			componentName := componentEntry.Name()
			componentPath := path.Join(stylePath, componentName, componentName+".yaml")
			componentData, err := EmbeddedFS.ReadFile(componentPath)
			if err != nil {
				continue // Skip if component yaml doesn't exist
			}

			var component config.TemplateComponent
			if err := y.Unmarshal(componentData, &component); err != nil {
				return nil, fmt.Errorf("failed to unmarshal %s: %w", componentPath, err)
			}

			// Check for duplicate Kind
			if _, ok := components[component.Kind]; ok {
				return nil, fmt.Errorf("duplicate component kind %s in %s and %s",
					component.Kind, components[component.Kind].Name, component.Name)
			}

			components[component.Kind] = component
		}
	}

	return components, nil
}

// LoadEmbeddedTemplates reads a set of templates from the local embedded filesystem (in the source, this is the
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
		// skip non-yaml files
		if !strings.HasSuffix(comp.Name(), ".yaml") {
			continue
		}
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

// CalculateChecksums reads the components and templates in the non-test
// subdirectories in the embedded filesystem and returns all the checksums in a
// map.
func CalculateChecksums() (map[string]string, error) {
	dirs, err := EmbeddedFS.ReadDir(".")
	if err != nil {
		return nil, err
	}
	results := make(map[string]string)
	for _, dir := range dirs {
		if dir.IsDir() && !strings.HasPrefix(dir.Name(), "test") {
			checksums, err := calculateChecksums(dir.Name())
			if err != nil {
				return nil, err
			}
			for k, v := range checksums {
				results[k] = v
			}
		}
	}
	return results, nil
}

// calculateChecksums reads the templates in a directory in the embedded
// filesystem and calculates a sha1 checksum of the contents. It will generate
// the same results as doing `sha1sum subdir/*` from the data directory. This
// is used to verify that the templates have not changed since the last release.
// Yes, we know that sha1 is not a secure hash, but we're not using it for security,
// and compatibility with a well-known command line tool is a feature.
func calculateChecksums(subdir string) (map[string]string, error) {
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
