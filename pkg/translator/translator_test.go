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
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	"github.com/honeycombio/hpsf/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThatEachTestFileHasAMatchingComponent(t *testing.T) {
	deleteExtras := false

	allComponents := make(map[string]struct{})
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	for _, comp := range comps {
		allComponents[strings.ToLower(comp.Kind)] = struct{}{}
	}

	subdirs := []string{"collector_config", "refinery_config", "refinery_rules", "hpsf"}
	for _, subdir := range subdirs {
		testFiles, err := os.ReadDir("testdata/" + subdir)
		require.NoError(t, err)

		// for every file in our subdir, we expect to find a component in the hpsf package
		// that has the same name as portion of the filename before the _.
		// look it up in the components map
		for _, file := range testFiles {
			if file.Name() == "default.yaml" {
				// don't mess with the default.yaml file
				continue
			}
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
				fullname := path.Join("testdata", subdir, file.Name())
				t.Run(fullname, func(t *testing.T) {
					// get the component name from the file name by splitting on the underscore
					// and taking the first part
					parts := strings.Split(file.Name(), "_")
					componentName := strings.ToLower(parts[0])

					// check if the component exists in the map
					if _, ok := allComponents[componentName]; !ok {
						t.Errorf("No matching component found for test file %s", file.Name())

						if deleteExtras {
							// if deleteExtras is true, delete the file
							err := os.Remove(fullname)
							require.NoError(t, err)
							t.Logf("Deleted test file %s because no matching component was found", file.Name())
						}
					}
				})
			}
		}
	}
}

func TestGenerateConfigForAllComponents(t *testing.T) {
	// set this to true to overwrite the testdata files with the generated
	// config files if they are different
	var overwrite bool

	// this allows for the make target regenerate_translator_testdata to work instead of editing
	if os.Getenv("OVERWRITE_TESTDATA") == "1" {
		overwrite = true
	}

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
					configType := template.Kind
					h, err := hpsf.FromYAML(strings.NewReader(inputData))
					require.NoError(t, err)

					cfg, err := tlater.GenerateConfig(&h, configType, nil)
					require.NoError(t, err)
					if cfg == nil {
						continue // skip if no config is generated for this component
					}

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
		ct                     hpsftypes.Type
		inputHPSFTestData      string
		expectedConfigTestData string
	}{
		{
			desc:                   "Refinery Config",
			ct:                     hpsftypes.RefineryConfig,
			expectedConfigTestData: "testdata/refinery_config/default.yaml",
		},
		{
			desc:                   "Refinery Rules",
			ct:                     hpsftypes.RefineryRules,
			expectedConfigTestData: "testdata/refinery_rules/default.yaml",
		},
		{
			desc:                   "Collector Config",
			ct:                     hpsftypes.CollectorConfig,
			expectedConfigTestData: "testdata/collector_config/default.yaml",
		},
	}

	// set this to true to overwrite the testdata files with the generated
	// config files if they are different
	var overwrite bool

	// this allows for the make target regenerate_translator_testdata to work instead of editing
	if os.Getenv("OVERWRITE_TESTDATA") == "1" {
		overwrite = true
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			b, err := os.ReadFile(tC.expectedConfigTestData)
			require.NoError(t, err)
			var expectedConfig = string(b)

			templates, err := data.LoadEmbeddedTemplates()
			require.NoError(t, err)

			h, ok := templates[data.DefaultConfigurationKind]
			require.True(t, ok)

			tlater := NewEmptyTranslator()
			comps, err := data.LoadEmbeddedComponents()
			require.NoError(t, err)
			tlater.InstallComponents(comps)

			cfg, err := tlater.GenerateConfig(&h, tC.ct, nil)
			require.NoError(t, err)

			got, err := cfg.RenderYAML()
			require.NoError(t, err)

			if overwrite && !reflect.DeepEqual(expectedConfig, string(got)) {
				// overwrite the testdata file with the generated config
				err = os.WriteFile(tC.expectedConfigTestData, got, 0644)
				require.NoError(t, err)
				t.Logf("Overwrote %s with generated config", tC.expectedConfigTestData)
			} else {
				assert.Equal(t, expectedConfig, string(got), "in file %s", tC.expectedConfigTestData)
			}
		})
	}
	if overwrite {
		t.Fail()
		t.Log("Some testdata files were overwritten. Please review the changes and commit them if they are correct.")
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

	cfg, err := tlater.GenerateConfig(&hpsf, hpsftypes.RefineryRules, nil)
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

			h, err := hpsf.FromYAML(strings.NewReader(inputData))
			require.NoError(t, err)

			err = tlater.ValidateConfig(&h)
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

func TestTranslator_ValidateBadConfigs(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		reasons string // comma-separated list of expected error contents, one for each error in the Details of the result
	}{
		{"duplicate names", "testdata/bad_hpsf/dup_names.yaml", "duplicate component name"},
		{"missing component", "testdata/bad_hpsf/missing_comp.yaml", "destination component not found,source component not found"},
		{"missing StartSampling", "testdata/bad_hpsf/missing_startsampling.yaml", "no samplers are allowed,exactly one input connection"},
		{"missing property", "testdata/bad_hpsf/missing_property.yaml", "property not found"},
		{"missing port", "testdata/bad_hpsf/missing_port.yaml", "source component does not have a port,destination component does not have a port"},
		{"missing condition on lower index", "testdata/bad_hpsf/missing_condition_on_lower_index.yaml", "Every path on a startsampler except the one with the highest index must connect to a condition"},
		{"missing component for specified version", "testdata/bad_hpsf/invalid_component_version.yaml", "failed to locate corresponding template component for HoneycombExporter@v999999.1.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := os.ReadFile(tt.file)
			require.NoError(t, err)
			var inputData = string(b)

			h, err := hpsf.FromYAML(strings.NewReader(inputData))
			require.NoError(t, err)

			trans := NewEmptyTranslator()
			comps, err := data.LoadEmbeddedComponents()
			require.NoError(t, err)
			trans.InstallComponents(comps)

			err = trans.ValidateConfig(&h)
			if err == nil {
				t.Errorf("Translator.ValidateConfig() did not error when it should have")
			}
			result, ok := err.(validator.Result)
			if !ok {
				t.Errorf("Translator.ValidateConfig() did not return a validator.Result, got %T", err)
				t.Fail()
			}
			if result.IsEmpty() {
				t.Errorf("Translator.ValidateConfig() returned empty result, expected errors")
			}
			contents := strings.Split(tt.reasons, ",")
			if len(contents) != len(result.Details) {
				t.Errorf("Translator.ValidateConfig() returned %d errors, expected %d", len(result.Details), len(contents))
				t.FailNow()
			}
			for i, detail := range result.Details {
				if !strings.Contains(detail.Error(), strings.TrimSpace(contents[i])) {
					t.Errorf("Translator.ValidateConfig() error %d did not contain expected text: %q, got: %s",
						i, strings.TrimSpace(contents[i]), detail.Error())
				}
			}
		})
	}
}

func TestTranslator_ValidateValidConfigs(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"valid condition on lower index", "testdata/bad_hpsf/valid_condition_on_lower_index.yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := os.ReadFile(tt.file)
			require.NoError(t, err)
			var inputData = string(b)

			h, err := hpsf.FromYAML(strings.NewReader(inputData))
			require.NoError(t, err)

			trans := NewEmptyTranslator()
			comps, err := data.LoadEmbeddedComponents()
			require.NoError(t, err)
			trans.InstallComponents(comps)

			err = trans.ValidateConfig(&h)
			if err != nil {
				t.Errorf("Translator.ValidateConfig() should not error for valid config, got: %v", err)
			}
		})
	}
}

func TestSampling(t *testing.T) {
	c := `
components:
  - name: Receive OTel_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Check for Errors_1
    kind: ErrorExistsCondition
  - name: Drop_1
    kind: Dropper
  - name: Sample by Events per Second_1
    kind: EMAThroughputSampler
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
connections:
  - source:
      component: Receive OTel_1
      port: Traces
      type: OTelTraces
    destination:
      component: Start Sampling_1
      port: Traces
      type: OTelTraces
  - source:
      component: Check for Errors_1
      port: And
      type: SampleData
    destination:
      component: Drop_1
      port: Sample
      type: SampleData
  - source:
      component: Start Sampling_1
      port: Rule 1
      type: SampleData
    destination:
      component: Check for Errors_1
      port: Match
      type: SampleData
  - source:
      component: Start Sampling_1
      port: Rule 2
      type: SampleData
    destination:
      component: Sample by Events per Second_1
      port: Sample
      type: SampleData
  - source:
      component: Sample by Events per Second_1
      port: Events
      type: HoneycombEvents
    destination:
      component: Send to Honeycomb_1
      port: Events
      type: HoneycombEvents
layout:
  components:
    - name: Receive OTel_1
      position:
        x: 50
        y: 0
    - name: Start Sampling_1
      position:
        x: 277
        y: 0
    - name: Check for Errors_1
      position:
        x: 680
        y: 0
    - name: Drop_1
      position:
        x: 875
        y: 0
    - name: Sample by Events per Second_1
      position:
        x: 660
        y: 160
    - name: Send to Honeycomb_1
      position:
        x: 1060
        y: 160
`

	h, err := hpsf.FromYAML(strings.NewReader(c))
	require.NoError(t, err)

	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)

	x, err := tlater.GenerateConfig(&h, hpsftypes.RefineryRules, nil)
	require.NoError(t, err)
	require.NotNil(t, x)
}
