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

	// Get all component YAML files
	componentsDir := filepath.Join("components")
	entries, err := os.ReadDir(componentsDir)
	require.NoError(t, err, "Failed to read components directory")

	var yamlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			yamlFiles = append(yamlFiles, entry.Name())
		}
	}

	require.NotEmpty(t, yamlFiles, "No YAML files found in components directory")

	// Validate each component file
	for _, filename := range yamlFiles {
		t.Run(filename, func(t *testing.T) {
			// Read the YAML file
			yamlPath := filepath.Join(componentsDir, filename)
			yamlData, err := os.ReadFile(yamlPath)
			require.NoError(t, err, "Failed to read %s", filename)

			// Parse YAML to a map
			var component map[string]interface{}
			err = yaml.Unmarshal(yamlData, &component)
			require.NoError(t, err, "Failed to parse YAML in %s", filename)

			// Convert to JSON for validation (jsonschema library works with JSON)
			jsonData, err := json.Marshal(component)
			require.NoError(t, err, "Failed to convert to JSON for %s", filename)

			// Parse back to interface{} for validation
			var jsonComponent interface{}
			err = json.Unmarshal(jsonData, &jsonComponent)
			require.NoError(t, err, "Failed to parse JSON for %s", filename)

			// Validate against schema
			err = schema.Validate(jsonComponent)
			if err != nil {
				// Provide detailed validation error
				if ve, ok := err.(*jsonschema.ValidationError); ok {
					t.Errorf("Schema validation failed for %s:\n%s", filename, formatValidationError(ve, ""))
				} else {
					t.Errorf("Schema validation failed for %s: %v", filename, err)
				}
			}
		})
	}

	t.Logf("Successfully validated %d component files against schema", len(yamlFiles))
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
