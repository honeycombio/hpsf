package translator

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
)

func TestGenerateConfigForAllComponents(t *testing.T) {
	// set this to true to overwrite the testdata files with the generated
	// config files if they are different
	overwrite := true

	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	for _, component := range comps {
		for _, properties := range []string{"all", "defaults"} {
			testData := fmt.Sprintf("%s_%s.yaml", strings.ToLower(component.Kind), properties)
			t.Run(testData, func(t *testing.T) {
				// testdata source config is an hpsf file that lives in testdata/hpsf
				// we can't automatically generate the testdata (we can generate the expected config)
				b, err := os.ReadFile(path.Join("testdata", "hpsf", testData))
				require.NoError(t, err)
				var inputData = string(b)

				for _, template := range component.Templates {
					configType := config.Type(template.Kind)
					var hpsf *hpsf.HPSF
					dec := yamlv3.NewDecoder(strings.NewReader(inputData))
					err = dec.Decode(&hpsf)
					require.NoError(t, err)

					cfg, err := tlater.GenerateConfig(hpsf, configType, nil)
					require.NoError(t, err)

					got, err := cfg.RenderYAML()
					require.NoError(t, err)

					var expectedConfig = ""
					if !overwrite {
						b, err = os.ReadFile(path.Join("testdata", string(configType), testData))
						require.NoError(t, err)
						expectedConfig = string(b)
					}

					if overwrite && !reflect.DeepEqual(expectedConfig, string(got)) {
						// overwrite the testdata file with the generated config
						err = os.WriteFile(path.Join("testdata", string(configType), testData), got, 0644)
						require.NoError(t, err)
						t.Logf("Overwrote %s with generated config", path.Join("testdata", string(configType), testData))
					} else {
						assert.Equal(t, expectedConfig, string(got))
					}
				}
			})
		}
	}
	if overwrite {
		t.Fail()
		t.Log("Some testdata files were overwritten. Please review the changes and commit them if they are correct.")
	}
}

func TestDefaultHPSF(t *testing.T) {
	testCases := []struct {
		desc                   string
		ct                     config.Type
		inputHPSFTestData      string
		expectedConfigTestData string
	}{
		{
			desc:                   "Refinery Config",
			ct:                     config.RefineryConfigType,
			expectedConfigTestData: "testdata/refinery_config/default.yaml",
		},
		{
			desc:                   "Refinery Rules",
			ct:                     config.RefineryRulesType,
			expectedConfigTestData: "testdata/refinery_rules/default.yaml",
		},
		{
			desc:                   "Collector Config",
			ct:                     config.CollectorConfigType,
			expectedConfigTestData: "testdata/collector_config/default.yaml",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			b, err := os.ReadFile(tC.expectedConfigTestData)
			require.NoError(t, err)
			var expectedConfig = string(b)

			h, err := data.LoadEmbeddedDefaultTemplate()
			require.NoError(t, err)

			tlater := NewEmptyTranslator()
			comps, err := data.LoadEmbeddedComponents()
			require.NoError(t, err)
			tlater.InstallComponents(comps)

			cfg, err := tlater.GenerateConfig(&h, tC.ct, nil)
			require.NoError(t, err)

			got, err := cfg.RenderYAML()
			require.NoError(t, err)

			assert.Equal(t, expectedConfig, string(got))
		})
	}
}

func TestHPSFWithoutSamplerComponentGeneratesValidRefineryRules(t *testing.T) {
	b, err := os.ReadFile("testdata/refinery_rules/default.yaml")
	require.NoError(t, err)
	var expectedConfig = string(b)

	hpsf := hpsf.HPSF{}
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)

	cfg, err := tlater.GenerateConfig(&hpsf, config.RefineryRulesType, nil)
	require.NoError(t, err)

	got, err := cfg.RenderYAML()
	require.NoError(t, err)

	assert.Equal(t, expectedConfig, string(got))
}

func TestTranslatorValidation(t *testing.T) {
	// read all yaml files in testdata
	entries, err := os.ReadDir(path.Join("testdata", "hpsf"))
	require.NoError(t, err)
	// Filter for YAML files
	var yamlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			// Construct the full path to the file
			filePath := path.Join("testdata", "hpsf", entry.Name())
			yamlFiles = append(yamlFiles, filePath)
		}
	}

	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)

	// Iterate over each YAML file and test for validation
	for _, filePath := range yamlFiles {
		t.Run("Validate test for "+filePath, func(t *testing.T) {
			// Read the file content
			b, err := os.ReadFile(filePath)
			require.NoError(t, err)
			var inputData = string(b)

			var hpsf *hpsf.HPSF
			dec := yamlv3.NewDecoder(strings.NewReader(inputData))
			err = dec.Decode(&hpsf)
			require.NoError(t, err)

			err = tlater.ValidateConfig(hpsf)
			require.NoError(t, errors.Unwrap(err))
		})
	}
}

func TestOrderedComponentMap(t *testing.T) {
	// Create a mock component for testing
	mockComponent := func(name string) config.Component {
		comp := &hpsf.Component{Name: name, Kind: "test"}
		tc := config.GenericBaseComponent{Component: *comp}
		return &tc
	}

	t.Run("NewOrderedComponentMap creates empty map", func(t *testing.T) {
		ocm := NewOrderedComponentMap()
		require.NotNil(t, ocm)
		require.Empty(t, ocm.Keys)
		require.Empty(t, ocm.Values)
	})

	t.Run("Set adds key-value pairs in order", func(t *testing.T) {
		ocm := NewOrderedComponentMap()
		comp1 := mockComponent("comp1")
		comp2 := mockComponent("comp2")
		comp3 := mockComponent("comp3")

		ocm.Set("key1", comp1)
		ocm.Set("key2", comp2)
		ocm.Set("key3", comp3)

		require.Len(t, ocm.Keys, 3)
		require.Len(t, ocm.Values, 3)
		require.Equal(t, []string{"key1", "key2", "key3"}, ocm.Keys)

		// Check values are stored correctly
		val, ok := ocm.Values["key1"]
		require.True(t, ok)
		require.Equal(t, comp1, val)
	})

	t.Run("Set overwrites existing value without duplicating key", func(t *testing.T) {
		ocm := NewOrderedComponentMap()
		comp1 := mockComponent("comp1")
		comp2 := mockComponent("comp2")
		updatedComp := mockComponent("updated")

		ocm.Set("key1", comp1)
		ocm.Set("key2", comp2)
		ocm.Set("key1", updatedComp) // Overwrite key1

		require.Len(t, ocm.Keys, 2)
		require.Len(t, ocm.Values, 2)
		require.Equal(t, []string{"key1", "key2"}, ocm.Keys)

		// Check the value was updated
		val, ok := ocm.Values["key1"]
		require.True(t, ok)
		require.Equal(t, updatedComp, val)
	})

	t.Run("Get retrieves values correctly", func(t *testing.T) {
		ocm := NewOrderedComponentMap()
		comp1 := mockComponent("comp1")

		ocm.Set("key1", comp1)

		// Get existing key
		val, ok := ocm.Get("key1")
		require.True(t, ok)
		require.Equal(t, comp1, val)

		// Get non-existent key
		val, ok = ocm.Get("nonexistent")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("Items returns components in insertion order", func(t *testing.T) {
		ocm := NewOrderedComponentMap()
		comp1 := mockComponent("comp1")
		comp2 := mockComponent("comp2")
		comp3 := mockComponent("comp3")

		ocm.Set("key2", comp2) // Deliberately not inserting in key order
		ocm.Set("key1", comp1)
		ocm.Set("key3", comp3)

		// Collect items from iterator
		var items []config.Component
		for comp := range ocm.Items() {
			items = append(items, comp)
		}

		// Check order matches insertion order, not key order
		require.Len(t, items, 3)
		require.Equal(t, comp2, items[0])
		require.Equal(t, comp1, items[1])
		require.Equal(t, comp3, items[2])
	})

	t.Run("Items with early exit", func(t *testing.T) {
		ocm := NewOrderedComponentMap()
		comp1 := mockComponent("comp1")
		comp2 := mockComponent("comp2")
		comp3 := mockComponent("comp3")

		ocm.Set("key1", comp1)
		ocm.Set("key2", comp2)
		ocm.Set("key3", comp3)

		// Counter to track number of iterations
		count := 0

		// Use Items iterator but return false after first item to stop iteration
		for comp := range ocm.Items() {
			count++
			require.Equal(t, comp1, comp)
			// Exit after first item
			break
		}

		require.Equal(t, 1, count, "Iterator should have stopped after first item")
	})
}
