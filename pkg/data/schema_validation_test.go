package data

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestComponentsValidateAgainstSchema ensures all component YAML files
// conform to the component-schema.json specification.
func TestComponentsValidateAgainstSchema(t *testing.T) {
	// Load the JSON schema
	schemaPath := filepath.Join("..", "..", "component-schema.json")
	schemaData, err := os.ReadFile(schemaPath)
	require.NoError(t, err, "Failed to read component-schema.json")

	// Compile the schema
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	// Add the schema to the compiler
	err = compiler.AddResource("component-schema.json", bytes.NewReader(schemaData))
	require.NoError(t, err, "Failed to add schema to compiler")

	schema, err := compiler.Compile("component-schema.json")
	require.NoError(t, err, "Failed to compile JSON schema")

	// Get all component YAML files (3-level structure: target/style/component)
	componentsDir := filepath.Join("components")
	entries, err := os.ReadDir(componentsDir)
	require.NoError(t, err, "Failed to read components directory")

	var componentPaths []struct {
		name string
		path string
	}

	for _, entry := range entries {
		// Skip special directories
		if entry.Name() == "_templates" || entry.Name()[0] == '.' {
			continue
		}

		if !entry.IsDir() {
			continue
		}

		// Level 1: target (collector/refinery)
		level1Path := filepath.Join(componentsDir, entry.Name())
		level1Entries, err := os.ReadDir(level1Path)
		if err != nil {
			continue
		}

		for _, level1Entry := range level1Entries {
			if !level1Entry.IsDir() {
				continue
			}

			// Level 2: style (receivers/processors/exporters/samplers/conditions/startsampling)
			level2Path := filepath.Join(level1Path, level1Entry.Name())
			level2Entries, err := os.ReadDir(level2Path)
			if err != nil {
				continue
			}

			for _, level2Entry := range level2Entries {
				if !level2Entry.IsDir() {
					continue
				}

				// Level 3: component directory
				componentYaml := filepath.Join(level2Path, level2Entry.Name(), "component.yaml")
				if _, err := os.Stat(componentYaml); err == nil {
					componentPaths = append(componentPaths, struct {
						name string
						path string
					}{entry.Name() + "/" + level1Entry.Name() + "/" + level2Entry.Name(), componentYaml})
				}
			}
		}
	}

	require.NotEmpty(t, componentPaths, "No component files found in components directory")

	// Validate each component file
	for _, comp := range componentPaths {
		t.Run(comp.name, func(t *testing.T) {
			// Read the YAML file
			yamlData, err := os.ReadFile(comp.path)
			require.NoError(t, err, "Failed to read %s", comp.name)

			// Parse YAML to a map
			var component map[string]interface{}
			err = yaml.Unmarshal(yamlData, &component)
			require.NoError(t, err, "Failed to parse YAML in %s", comp.name)

			// Convert to JSON for validation (jsonschema library works with JSON)
			jsonData, err := json.Marshal(component)
			require.NoError(t, err, "Failed to convert to JSON for %s", comp.name)

			// Parse back to interface{} for validation
			var jsonComponent interface{}
			err = json.Unmarshal(jsonData, &jsonComponent)
			require.NoError(t, err, "Failed to parse JSON for %s", comp.name)

			// Validate against schema
			err = schema.Validate(jsonComponent)
			if err != nil {
				// Provide detailed validation error
				if ve, ok := err.(*jsonschema.ValidationError); ok {
					t.Errorf("Schema validation failed for %s:\n%s", comp.name, formatValidationError(ve, ""))
				} else {
					t.Errorf("Schema validation failed for %s: %v", comp.name, err)
				}
			}
		})
	}

	t.Logf("Successfully validated %d component files against schema", len(componentPaths))
}

// formatValidationError recursively formats validation errors for better readability
func formatValidationError(ve *jsonschema.ValidationError, indent string) string {
	result := indent + "- " + ve.Message + "\n"
	result += indent + "  Instance: " + ve.InstanceLocation + "\n"
	result += indent + "  Schema: " + ve.KeywordLocation + "\n"

	for _, cause := range ve.Causes {
		result += formatValidationError(cause, indent+"  ")
	}

	return result
}
