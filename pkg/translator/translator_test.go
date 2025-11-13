package translator

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"slices"
	"strings"
	"testing"
	"text/template"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	"github.com/honeycombio/hpsf/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
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
			// ignore some files that don't correspond to components
			if slices.Contains([]string{"default.yaml", "empty.yaml"}, file.Name()) {
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
	overwrite := false

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
					h, err := hpsf.FromYAML(inputData)
					require.NoError(t, err)

					cfg, err := tlater.GenerateConfig(&h, configType, LatestVersion, nil)
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

			cfg, err := tlater.GenerateConfig(&h, tC.ct, LatestVersion, nil)
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
	b, err := os.ReadFile("testdata/refinery_rules/empty.yaml")
	require.NoError(t, err)
	var expectedConfig = string(b)

	hpsf := hpsf.HPSF{}
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)

	cfg, err := tlater.GenerateConfig(&hpsf, hpsftypes.RefineryRules, LatestVersion, nil)
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

			h, err := hpsf.FromYAML(inputData)
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

			h, err := hpsf.FromYAML(inputData)
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

			h, err := hpsf.FromYAML(inputData)
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

	h, err := hpsf.FromYAML(c)
	require.NoError(t, err)

	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)

	x, err := tlater.GenerateConfig(&h, hpsftypes.RefineryRules, LatestVersion, nil)
	require.NoError(t, err)
	require.NotNil(t, x)
}

func TestConditions(t *testing.T) {
	c := `
components:
  - name: Receive OTel_1
    kind: OTelReceiver
    version: v0.1.0
  - name: Start Sampling_1
    kind: SamplingSequencer
    version: v0.1.0
  - name: {{ .ConditionName }}
    kind: {{ .ConditionKind }}
    version: v0.1.0{{ if .Properties }}
    properties:{{ range .Properties }}
      - name: {{ .Name }}
        value: {{ .Value }}{{ end }}{{ end }}
  - name: Sample at a Fixed Rate_1
    kind: DeterministicSampler
    version: v0.1.0
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
    version: v0.1.0
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
      component: Sample at a Fixed Rate_1
      port: Events
      type: HoneycombEvents
    destination:
      component: Send to Honeycomb_1
      port: Events
      type: HoneycombEvents
  - source:
      component: Start Sampling_1
      port: Rule 1
      type: SampleData
    destination:
      component: {{ .ConditionName }}
      port: Match
      type: SampleData
  - source:
      component: {{ .ConditionName }}
      port: And
      type: SampleData
    destination:
      component: Sample at a Fixed Rate_1
      port: Sample
      type: SampleData
`
	tests := []struct {
		conditionName string
		conditionKind string
		properties    []struct {
			Name  string
			Value string
		}
	}{
		{
			conditionName: "ErrorExistsCondition_1",
			conditionKind: "ErrorExistsCondition",
		},
		{
			conditionName: "FieldStartsWithCondition_1",
			conditionKind: "FieldStartsWithCondition",
		},
		{
			conditionName: "LongDurationCondition_1",
			conditionKind: "LongDurationCondition",
		},
		{
			conditionName: "FieldContainsCondition_1",
			conditionKind: "FieldContainsCondition",
		},
		{
			conditionName: "RootSpanCondition_1",
			conditionKind: "RootSpanCondition",
		},
		{
			conditionName: "CompareIntegerField_1",
			conditionKind: "CompareIntegerFieldCondition",
			properties: []struct {
				Name  string
				Value string
			}{
				{Name: "Fields", Value: `["status_code"]`},
				{Name: "Operator", Value: "="},
				{Name: "Value", Value: "500"},
			},
		},
		{
			conditionName: "ForceSpanScope_1",
			conditionKind: "ForceSpanScope",
		},
		{
			conditionName: "ForceSpanScope_1",
			conditionKind: "ForceSpanScope",
		},
		{
			conditionName: "CompareStringField_1",
			conditionKind: "CompareStringFieldCondition",
			properties: []struct {
				Name  string
				Value string
			}{
				{Name: "Fields", Value: `["status_code"]`},
				{Name: "Operator", Value: "="},
				{Name: "Value", Value: "error"},
			},
		},
		{
			conditionName: "CompareIntegerField_1",
			conditionKind: "CompareIntegerFieldCondition",
			properties: []struct {
				Name  string
				Value string
			}{
				{Name: "Fields", Value: `["status_code"]`},
				{Name: "Operator", Value: "="},
				{Name: "Value", Value: "500"},
			},
		},
		{
			conditionName: "CompareDecimalField_1",
			conditionKind: "CompareDecimalFieldCondition",
			properties: []struct {
				Name  string
				Value string
			}{
				{Name: "Fields", Value: `["duration_ms"]`},
				{Name: "Operator", Value: "="},
				{Name: "Value", Value: "500"},
			},
		},
		{
			conditionName: "MatchRegularExpression_1",
			conditionKind: "MatchRegularExpression",
		},
	}
	for _, tt := range tests {
		t.Run(tt.conditionName, func(t *testing.T) {
			tmpl, err := template.New("test").Parse(c)
			require.NoError(t, err)

			testdata := map[string]interface{}{
				"ConditionName": tt.conditionName,
				"ConditionKind": tt.conditionKind,
				"Properties":    tt.properties,
			}

			// Execute template into a buffer
			var buf bytes.Buffer
			err = tmpl.Execute(&buf, testdata)
			require.NoError(t, err)

			// Decode YAML from buffer
			h, err := hpsf.FromYAML(buf.String())
			require.NoError(t, err)

			// Generate config
			tlater := NewEmptyTranslator()
			comps, err := data.LoadEmbeddedComponents()
			require.NoError(t, err)
			tlater.InstallComponents(comps)

			cfg, err := tlater.GenerateConfig(&h, hpsftypes.RefineryRules, LatestVersion, nil)
			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}

func TestCompareIntegerFieldScope(t *testing.T) {
	// we just have to replace the operator in the config template
	// to test the scope logic, so we use a format string for sprintf
	configFormat := `
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Compare Integer Field_1
    kind: CompareIntegerFieldCondition
    properties:
      - name: Fields
        value: ["status_code"]
      - name: Operator
        value: "%s"
      - name: Value
        value: 500
  - name: Keep All_1
    kind: KeepAllSampler
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
connections:
  - source:
      component: OTel Receiver_1
      port: Traces
      type: OTelTraces
    destination:
      component: Start Sampling_1
      port: Traces
      type: OTelTraces
  - source:
      component: Start Sampling_1
      port: Rule 1
      type: SampleData
    destination:
      component: Compare Integer Field_1
      port: Match
      type: SampleData
  - source:
      component: Compare Integer Field_1
      port: And
      type: SampleData
    destination:
      component: Keep All_1
      port: Sample
      type: SampleData
  - source:
      component: Keep All_1
      port: Events
      type: HoneycombEvents
    destination:
      component: Send to Honeycomb_1
      port: Events
      type: HoneycombEvents
`

	// Test that the scope field is set correctly based on the operator
	testCases := []struct {
		name          string
		operator      string
		expectedScope string
	}{
		{
			name:          "not_equals_operator_should_set_span_scope",
			operator:      "!=",
			expectedScope: "span",
		},
		{
			name:          "equals_operator_should_set_trace_scope",
			operator:      "==",
			expectedScope: "trace",
		},
		{
			name:          "greater_than_operator_should_set_trace_scope",
			operator:      ">",
			expectedScope: "trace",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a simple HPSF configuration with CompareIntegerField
			hpsfConfig := fmt.Sprintf(configFormat, tc.operator)
			h, err := hpsf.FromYAML(hpsfConfig)
			require.NoError(t, err)

			// Generate the refinery rules configuration
			tlater := NewEmptyTranslator()
			comps, err := data.LoadEmbeddedComponents()
			require.NoError(t, err)
			tlater.InstallComponents(comps)

			cfg, err := tlater.GenerateConfig(&h, hpsftypes.RefineryRules, LatestVersion, nil)
			require.NoError(t, err)
			require.NotNil(t, cfg)

			// Render the configuration to YAML
			got, err := cfg.RenderYAML()
			require.NoError(t, err)

			// Parse the generated YAML to check the scope
			var rulesConfig map[string]interface{}
			err = yamlv3.Unmarshal(got, &rulesConfig)
			require.NoError(t, err)

			// Navigate to the rule and check the scope
			samplers := rulesConfig["Samplers"].(map[string]interface{})
			defaultSampler := samplers["__default__"].(map[string]interface{})
			rulesBasedSampler := defaultSampler["RulesBasedSampler"].(map[string]interface{})
			rules := rulesBasedSampler["Rules"].([]interface{})
			rule := rules[0].(map[string]interface{})

			// Check that the scope is set correctly
			scope, exists := rule["Scope"]
			require.True(t, exists, "Scope field should be present")
			require.Equal(t, tc.expectedScope, scope, "Scope should be %s for operator %s", tc.expectedScope, tc.operator)
		})
	}
}

func TestScopeMergingLogic(t *testing.T) {
	testCases := []struct {
		name          string
		hpsfConfig    string
		expectedScope string
		description   string
	}{
		{
			name: "single_force_span_scope_should_promote_without_scope",
			hpsfConfig: `
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Force Span Scope_1
    kind: ForceSpanScope
  - name: Keep All_1
    kind: KeepAllSampler
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
connections:
  - source:
      component: OTel Receiver_1
      port: Traces
      type: OTelTraces
    destination:
      component: Start Sampling_1
      port: Traces
      type: OTelTraces
  - source:
      component: Start Sampling_1
      port: Rule 1
      type: SampleData
    destination:
      component: Force Span Scope_1
      port: Match
      type: SampleData
  - source:
      component: Force Span Scope_1
      port: And
      type: SampleData
    destination:
      component: Keep All_1
      port: Sample
      type: SampleData
  - source:
      component: Keep All_1
      port: Events
      type: HoneycombEvents
    destination:
      component: Send to Honeycomb_1
      port: Events
      type: HoneycombEvents
`,
			expectedScope: "span",
			description:   "Single ForceSpanScope should promote to deterministic sampler without scope",
		},
		{
			name: "multiple_conditions_without_force_span_scope_should_not_set_scope",
			hpsfConfig: `
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Error Exists_1
    kind: ErrorExistsCondition
  - name: Long Duration_1
    kind: LongDurationCondition
  - name: Keep All_1
    kind: KeepAllSampler
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
connections:
  - source:
      component: OTel Receiver_1
      port: Traces
      type: OTelTraces
    destination:
      component: Start Sampling_1
      port: Traces
      type: OTelTraces
  - source:
      component: Start Sampling_1
      port: Rule 1
      type: SampleData
    destination:
      component: Error Exists_1
      port: Match
      type: SampleData
  - source:
      component: Error Exists_1
      port: And
      type: SampleData
    destination:
      component: Long Duration_1
      port: Match
      type: SampleData
  - source:
      component: Long Duration_1
      port: And
      type: SampleData
    destination:
      component: Keep All_1
      port: Sample
      type: SampleData
  - source:
      component: Keep All_1
      port: Events
      type: HoneycombEvents
    destination:
      component: Send to Honeycomb_1
      port: Events
      type: HoneycombEvents
`,
			expectedScope: "trace",
			description:   "Multiple conditions without ForceSpanScope should not set scope (defaults to trace)",
		},
		{
			name: "force_span_scope_with_other_conditions_should_set_span",
			hpsfConfig: `
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Error Exists_1
    kind: ErrorExistsCondition
  - name: Force Span Scope_1
    kind: ForceSpanScope
  - name: Long Duration_1
    kind: LongDurationCondition
  - name: Keep All_1
    kind: KeepAllSampler
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
connections:
  - source:
      component: OTel Receiver_1
      port: Traces
      type: OTelTraces
    destination:
      component: Start Sampling_1
      port: Traces
      type: OTelTraces
  - source:
      component: Start Sampling_1
      port: Rule 1
      type: SampleData
    destination:
      component: Error Exists_1
      port: Match
      type: SampleData
  - source:
      component: Error Exists_1
      port: And
      type: SampleData
    destination:
      component: Force Span Scope_1
      port: Match
      type: SampleData
  - source:
      component: Force Span Scope_1
      port: And
      type: SampleData
    destination:
      component: Long Duration_1
      port: Match
      type: SampleData
  - source:
      component: Long Duration_1
      port: And
      type: SampleData
    destination:
      component: Keep All_1
      port: Sample
      type: SampleData
  - source:
      component: Keep All_1
      port: Events
      type: HoneycombEvents
    destination:
      component: Send to Honeycomb_1
      port: Events
      type: HoneycombEvents
`,
			expectedScope: "span",
			description:   "ForceSpanScope with other conditions should set scope to span",
		},
		{
			name: "compare_integer_field_with_force_span_scope_should_set_span",
			hpsfConfig: `
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Compare Integer Field_1
    kind: CompareIntegerFieldCondition
    properties:
      - name: Fields
        value: ["status_code"]
      - name: Operator
        value: "=="
      - name: Value
        value: 500
  - name: Force Span Scope_1
    kind: ForceSpanScope
  - name: Keep All_1
    kind: KeepAllSampler
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
connections:
  - source:
      component: OTel Receiver_1
      port: Traces
      type: OTelTraces
    destination:
      component: Start Sampling_1
      port: Traces
      type: OTelTraces
  - source:
      component: Start Sampling_1
      port: Rule 1
      type: SampleData
    destination:
      component: Compare Integer Field_1
      port: Match
      type: SampleData
  - source:
      component: Compare Integer Field_1
      port: And
      type: SampleData
    destination:
      component: Force Span Scope_1
      port: Match
      type: SampleData
  - source:
      component: Force Span Scope_1
      port: And
      type: SampleData
    destination:
      component: Keep All_1
      port: Sample
      type: SampleData
  - source:
      component: Keep All_1
      port: Events
      type: HoneycombEvents
    destination:
      component: Send to Honeycomb_1
      port: Events
      type: HoneycombEvents
`,
			expectedScope: "span",
			description:   "CompareIntegerField with ForceSpanScope should set scope to span",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the HPSF configuration
			h, err := hpsf.FromYAML(tc.hpsfConfig)
			require.NoError(t, err)

			// Generate the refinery rules configuration
			tlater := NewEmptyTranslator()
			comps, err := data.LoadEmbeddedComponents()
			require.NoError(t, err)
			tlater.InstallComponents(comps)

			cfg, err := tlater.GenerateConfig(&h, hpsftypes.RefineryRules, LatestVersion, nil)
			require.NoError(t, err)
			require.NotNil(t, cfg)

			// Render the configuration to YAML
			got, err := cfg.RenderYAML()
			require.NoError(t, err)

			// Parse the generated YAML to check the scope
			var rulesConfig map[string]interface{}
			err = yamlv3.Unmarshal(got, &rulesConfig)
			require.NoError(t, err)

			// Navigate to the rule and check the scope
			samplers := rulesConfig["Samplers"].(map[string]interface{})
			defaultSampler := samplers["__default__"].(map[string]interface{})

			// Check if we have a RulesBasedSampler or a promoted sampler
			var scope interface{}
			var exists bool

			if rulesBasedSampler, ok := defaultSampler["RulesBasedSampler"]; ok {
				// We have a RulesBasedSampler, check the first rule
				rbs := rulesBasedSampler.(map[string]interface{})
				rules := rbs["Rules"].([]interface{})
				rule := rules[0].(map[string]interface{})
				scope, exists = rule["Scope"]

				if tc.expectedScope == "trace" && !exists {
					// For trace scope, it's acceptable to not have the scope field set (defaults to trace)
					// This is the case for multiple conditions without ForceSpanScope
				} else {
					// For span scope or when we expect scope to be explicitly set
					require.True(t, exists, "Scope field should be present in rules-based sampler")
					require.Equal(t, tc.expectedScope, scope, "Scope should be %s: %s", tc.expectedScope, tc.description)
				}
			} else {
				// We have a promoted sampler (like DeterministicSampler), scope should not be present
				_, exists = defaultSampler["Scope"]
				require.False(t, exists, "Scope field should not be present in promoted sampler")
			}
		})
	}
}

func TestArtifactVersionSupported(t *testing.T) {
	for _, tc := range []struct {
		name            string
		artifactVersion string
		component       config.TemplateComponent
		wantErr         string
	}{
		{
			name:    "no artifact version",
			wantErr: "",
		},
		{
			name:            "latest artifact version",
			wantErr:         "",
			artifactVersion: LatestVersion,
			component: config.TemplateComponent{
				Minimum: map[hpsftypes.Type]string{
					hpsftypes.CollectorConfig: "v0.100.0",
				},
				Maximum: map[hpsftypes.Type]string{
					hpsftypes.CollectorConfig: "v0.200.0",
				},
			},
		},
		{
			name:            "no minimum version",
			wantErr:         "",
			artifactVersion: "v0.1.0",
			component: config.TemplateComponent{
				Maximum: map[hpsftypes.Type]string{
					hpsftypes.CollectorConfig: "v0.200.0",
				},
			},
		},
		{
			name:            "no maximum version",
			wantErr:         "",
			artifactVersion: "v0.101.0",
			component: config.TemplateComponent{
				Minimum: map[hpsftypes.Type]string{
					hpsftypes.CollectorConfig: "v0.100.0",
				},
			},
		},
		{
			name:            "not supported below min",
			wantErr:         "requirement minimum version of v0.101.0",
			artifactVersion: "v0.100.0",
			component: config.TemplateComponent{
				Minimum: map[hpsftypes.Type]string{
					hpsftypes.CollectorConfig: "v0.101.0",
				},
			},
		},
		{
			name:            "not supported beyond max",
			wantErr:         "requirement maximum version of v0.120.0",
			artifactVersion: "v0.150.0",
			component: config.TemplateComponent{
				Minimum: map[hpsftypes.Type]string{
					hpsftypes.CollectorConfig: "v0.100.0",
				},
				Maximum: map[hpsftypes.Type]string{
					hpsftypes.CollectorConfig: "v0.120.0",
				},
			},
		},
		{
			name:            "specific agent type supported",
			wantErr:   "",
			artifactVersion: "v0.150.0",
			component: config.TemplateComponent{
				Minimum: map[hpsftypes.Type]string{
					hpsftypes.CollectorConfig: "v0.100.0",
					hpsftypes.RefineryRules:   "v0.200.0",
				},
				Maximum: map[hpsftypes.Type]string{
					hpsftypes.CollectorConfig: "v0.151.0",
					hpsftypes.RefineryRules:   "v0.300.0",
				},
			},
		},
		{
			name:            "expected agent type not specified",
			wantErr:   "",
			artifactVersion: "v0.150.0",
			component: config.TemplateComponent{
				Minimum: map[hpsftypes.Type]string{
					hpsftypes.RefineryRules:   "v0.151.0",
				},
				Maximum: map[hpsftypes.Type]string{
					hpsftypes.RefineryRules:   "v0.152.0",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := artifactVersionSupported(tc.component, hpsftypes.CollectorConfig, tc.artifactVersion)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestIndexedPortSequences ensures that for every component which specifies one or more
// indexed ports (Index > 0), the index values:
//  1. Start at 1 (no zero or negative values allowed)
//  2. Are contiguous with no gaps (i.e. if the largest index is N there are exactly N indexed ports)
//  3. Contain no duplicates
//
// This enforces deterministic ordering semantics relied upon by path and rules generation.
func TestIndexedPortSequences(t *testing.T) {
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)

	for _, comp := range comps {
		var indexed []int
		indexCount := map[int]int{}
		for _, p := range comp.Ports {
			if p.Index > 0 { // only consider explicitly indexed ports
				indexed = append(indexed, p.Index)
				indexCount[p.Index]++
			} else if p.Index < 0 { // defensive: negative should never happen
				t.Errorf("component %s has negative port index %d on port %s", comp.Kind, p.Index, p.Name)
			}
		}
		if len(indexed) == 0 {
			continue // nothing to validate for this component
		}

		// Sort indices to check ordering & gaps
		slices.Sort(indexed)

		// Must start at 1
		if indexed[0] != 1 {
			t.Errorf("component %s indexed ports must start at 1; got first index %d (all: %v)", comp.Kind, indexed[0], indexed)
			continue
		}

		// Highest index determines expected count
		highest := indexed[len(indexed)-1]
		if highest != len(indexed) {
			// Could be due to a gap or duplicates; build a list of missing indices for clarity
			missing := []int{}
			for i := 1; i <= highest; i++ {
				if indexCount[i] == 0 {
					missing = append(missing, i)
				}
			}
			t.Errorf("component %s expected contiguous indices 1..%d with no gaps; found %v (missing %v)", comp.Kind, highest, indexed, missing)
			continue
		}

		// Check for duplicates explicitly (any count > 1)
		for idx, count := range indexCount {
			if count > 1 {
				t.Errorf("component %s has duplicate index %d (occurrences=%d)", comp.Kind, idx, count)
			}
		}
	}
}

func TestComponentVersionSupported(t *testing.T) {
	tests := []struct {
		name             string
		templateVersion  string
		requestedVersion string
		expected         bool
	}{
		{
			name:             "empty requested version matches any template version",
			templateVersion:  "v0.1.0",
			requestedVersion: "",
			expected:         true,
		},
		{
			name:             "empty template version only matches empty requested",
			templateVersion:  "",
			requestedVersion: "",
			expected:         true,
		},
		{
			name:             "empty template version rejects non-empty requested",
			templateVersion:  "",
			requestedVersion: "v0.1.0",
			expected:         false,
		},
		{
			name:             "exact version match",
			templateVersion:  "v0.1.0",
			requestedVersion: "v0.1.0",
			expected:         true,
		},
		{
			name:             "patch upgrade allowed - v0.1.0 -> v0.1.1",
			templateVersion:  "v0.1.1",
			requestedVersion: "v0.1.0",
			expected:         true,
		},
		{
			name:             "patch downgrade not allowed - v0.1.1 -> v0.1.0",
			templateVersion:  "v0.1.0",
			requestedVersion: "v0.1.1",
			expected:         false,
		},
		{
			name:             "minor upgrade allowed - v0.1.0 -> v0.2.0",
			templateVersion:  "v0.2.0",
			requestedVersion: "v0.1.0",
			expected:         true,
		},
		{
			name:             "minor downgrade not allowed - v0.2.0 -> v0.1.0",
			templateVersion:  "v0.1.0",
			requestedVersion: "v0.2.0",
			expected:         false,
		},
		{
			name:             "major version upgrade not allowed - v0.1.0 -> v1.0.0",
			templateVersion:  "v1.0.0",
			requestedVersion: "v0.1.0",
			expected:         false,
		},
		{
			name:             "major version downgrade not allowed - v1.0.0 -> v0.1.0",
			templateVersion:  "v0.1.0",
			requestedVersion: "v1.0.0",
			expected:         false,
		},
		{
			name:             "same major version different minor/patch allowed",
			templateVersion:  "v1.2.3",
			requestedVersion: "v1.1.0",
			expected:         true,
		},
		{
			name:             "invalid template version falls back to string equality",
			templateVersion:  "invalid",
			requestedVersion: "invalid",
			expected:         true,
		},
		{
			name:             "invalid requested version falls back to string equality",
			templateVersion:  "invalid",
			requestedVersion: "different",
			expected:         false,
		},
		{
			name:             "mixed invalid/valid semver falls back to string equality",
			templateVersion:  "v0.1.0",
			requestedVersion: "invalid",
			expected:         false,
		},
		{
			name:             "complex version with build metadata",
			templateVersion:  "v1.0.1-alpha+001",
			requestedVersion: "v1.0.0",
			expected:         true,
		},
		{
			name:             "prerelease versions",
			templateVersion:  "v1.0.0-alpha.2",
			requestedVersion: "v1.0.0-alpha.1",
			expected:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := componentVersionSupported(tt.templateVersion, tt.requestedVersion)
			assert.Equal(t, tt.expected, result,
				"componentVersionSupported(%q, %q) = %t, expected %t",
				tt.templateVersion, tt.requestedVersion, result, tt.expected)
		})
	}
}

func TestInspect_HoneycombExporter(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My Honeycomb Exporter
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-api-key
      - name: APIEndpoint
        value: api.honeycomb.io
      - name: APIPort
        value: 443
      - name: MetricsDataset
        value: my-metrics
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "My Honeycomb Exporter", exp.Name)
	assert.Equal(t, "HoneycombExporter", exp.Kind)
	// Verify version falls back to template version when not specified in HPSF
	assert.NotEmpty(t, exp.Version, "Version should be populated from template component")

	// Verify properties contain actual component values
	assert.NotNil(t, exp.Properties)
}

func TestInspect_S3ArchiveExporter(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My S3 Archive
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-telemetry-bucket
      - name: Region
        value: us-west-2
      - name: Prefix
        value: telemetry/
      - name: PartitionFormat
        value: year=%Y/month=%m/day=%d
      - name: Marshaler
        value: otlp_json
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "My S3 Archive", exp.Name)
	assert.Equal(t, "S3ArchiveExporter", exp.Kind)

	// Verify properties is accessible without casting
	assert.Equal(t, "us-west-2", exp.Properties["Region"])
	assert.Equal(t, "my-telemetry-bucket", exp.Properties["Bucket"])
	assert.Equal(t, "telemetry/", exp.Properties["Prefix"])
}

func TestInspect_EnhanceIndexingS3Exporter(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My Enhanced S3
    kind: EnhanceIndexingS3Exporter
    properties:
      - name: Bucket
        value: my-indexed-bucket
      - name: Region
        value: eu-west-1
      - name: APIKey
        value: test-key
      - name: APISecret
        value: test-secret
      - name: APIEndpoint
        value: https://api.honeycomb.io
      - name: IndexedFields
        value:
          - custom.field1
          - custom.field2
      - name: PartitionFormat
        value: year=%Y/month=%m
      - name: Marshaler
        value: otlp_proto
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "My Enhanced S3", exp.Name)
	assert.Equal(t, "EnhanceIndexingS3Exporter", exp.Kind)

	// Verify properties is accessible without casting
	assert.Equal(t, "eu-west-1", exp.Properties["Region"])
	assert.Equal(t, "my-indexed-bucket", exp.Properties["Bucket"])
	assert.Nil(t, exp.Properties["Prefix"]) // Not set in config
}

func TestInspect_OTelGRPCExporter(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My OTLP gRPC
    kind: OTelGRPCExporter
    properties:
      - name: Host
        value: otel-collector.example.com
      - name: Port
        value: 4317
      - name: Insecure
        value: false
      - name: Headers
        value:
          authorization: Bearer token123
          x-custom-header: value
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "OTelGRPCExporter", exp.Kind)

	// Verify properties map exists (even if empty)
	assert.NotNil(t, exp.Properties)
}

func TestInspect_OTelHTTPExporter(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My OTLP HTTP
    kind: OTelHTTPExporter
    properties:
      - name: Host
        value: otel-collector.example.com
      - name: Port
        value: 4318
      - name: Insecure
        value: true
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "OTelHTTPExporter", exp.Kind)

	// Verify properties map exists (even if empty)
	assert.NotNil(t, exp.Properties)
}

func TestInspect_DebugExporter(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My Debug
    kind: DebugExporter
    properties:
      - name: Verbosity
        value: detailed
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "DebugExporter", exp.Kind)

	// Verify properties map exists (even if empty)
	assert.NotNil(t, exp.Properties)
}

func TestInspect_NopExporter(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My Nop
    kind: NopExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "NopExporter", exp.Kind)

	// Verify properties map exists (even if empty)
	assert.NotNil(t, exp.Properties)
}

func TestInspect_MultipleExporters(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Receiver1
    kind: OTelGRPCReceiver
  - name: Honeycomb Export
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-key
      - name: APIEndpoint
        value: api.honeycomb.io
  - name: S3 Archive
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-bucket
      - name: Region
        value: us-east-1
  - name: Processor1
    kind: BatchProcessor
  - name: Debug Export
    kind: DebugExporter
    properties:
      - name: Verbosity
        value: basic
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 3, "should extract only the 3 exporters")

	// Check that we have all the expected exporter types
	exporterTypes := make(map[string]bool)
	exporterNames := make(map[string]bool)
	for _, exp := range exporters {
		exporterTypes[exp.Kind] = true
		exporterNames[exp.Name] = true
	}

	assert.True(t, exporterTypes["HoneycombExporter"])
	assert.True(t, exporterTypes["S3ArchiveExporter"])
	assert.True(t, exporterTypes["DebugExporter"])

	// Verify names are captured
	assert.True(t, exporterNames["Honeycomb Export"])
	assert.True(t, exporterNames["S3 Archive"])
	assert.True(t, exporterNames["Debug Export"])
}

func TestInspect_InvalidYAML(t *testing.T) {
	hpsfConfig := `this is not valid yaml: {[`

	_, err := hpsf.FromYAML(hpsfConfig)
	assert.Error(t, err, "should return error for invalid YAML")
}

func TestInspect_EmptyConfig(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components: []
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	assert.Empty(t, exporters, "should return empty slice for config with no components")
}

func TestInspect_UnknownComponentKind(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Unknown Component
    kind: NonExistentExporter
    properties:
      - name: SomeProp
        value: somevalue
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	assert.Empty(t, exporters, "should skip unknown component kinds")
}

func TestInspect_MissingOptionalProperties(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Minimal Honeycomb
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-key
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "HoneycombExporter", exp.Kind)

	// Properties should be extracted even when minimal config provided
	assert.NotNil(t, exp.Properties)
}

func TestInspect_S3ArchiveExporter_UsesDefaultRegion(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Minimal S3
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-bucket
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "S3ArchiveExporter", exp.Kind)

	// Region should use default from template
	assert.Equal(t, "us-east-1", exp.Properties["Region"])
	assert.Equal(t, "my-bucket", exp.Properties["Bucket"])
	assert.Nil(t, exp.Properties["Prefix"]) // Not set in config
}

func TestInspect_EnhanceIndexingS3Exporter_UsesDefaultRegion(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Minimal Enhanced S3
    kind: EnhanceIndexingS3Exporter
    properties:
      - name: Bucket
        value: my-indexed-bucket
      - name: APIKey
        value: test-key
      - name: APISecret
        value: test-secret
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "EnhanceIndexingS3Exporter", exp.Kind)

	// Region should use default from template
	assert.Equal(t, "us-east-1", exp.Properties["Region"])
	assert.Equal(t, "my-indexed-bucket", exp.Properties["Bucket"])
	assert.Nil(t, exp.Properties["Prefix"]) // Not set in config
}

func TestInspect_AllComponentTypes(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
    properties:
      - name: GRPCPort
        value: 4317
      - name: HTTPPort
        value: 4318
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
    properties:
      - name: CheckInterval
        value: 1s
      - name: LimitPercentage
        value: 50
  - name: Honeycomb Export
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-key
  - name: S3 Archive
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-bucket
      - name: Region
        value: us-west-2
  - name: Another Receiver
    kind: NopReceiver
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	result := tlater.Inspect(h)

	// Verify receivers
	receivers := result.Filter(Receivers).Components
	require.Len(t, receivers, 2)
	assert.Equal(t, "OTLP Receiver", receivers[0].Name)
	assert.Equal(t, "OTelReceiver", receivers[0].Kind)
	assert.Equal(t, 4317, receivers[0].Properties["GRPCPort"])
	assert.Equal(t, 4318, receivers[0].Properties["HTTPPort"])
	assert.Equal(t, "Another Receiver", receivers[1].Name)
	assert.Equal(t, "NopReceiver", receivers[1].Kind)

	// Verify processors
	processors := result.Filter(Processors).Components
	require.Len(t, processors, 1)
	assert.Equal(t, "Memory Limiter", processors[0].Name)
	assert.Equal(t, "MemoryLimiterProcessor", processors[0].Kind)
	assert.Equal(t, "1s", processors[0].Properties["CheckInterval"])
	assert.Equal(t, 50, processors[0].Properties["LimitPercentage"])
	assert.Equal(t, 20, processors[0].Properties["SpikeLimitPercentage"])

	// Verify exporters
	exporters := result.Filter(Exporters).Components
	require.Len(t, exporters, 2)
	assert.Equal(t, "Honeycomb Export", exporters[0].Name)
	assert.Equal(t, "HoneycombExporter", exporters[0].Kind)
	assert.Equal(t, "S3 Archive", exporters[1].Name)
	assert.Equal(t, "S3ArchiveExporter", exporters[1].Kind)
	assert.Equal(t, "us-west-2", exporters[1].Properties["Region"])
	assert.Equal(t, "my-bucket", exporters[1].Properties["Bucket"])
}

func TestInspect_PropertiesAccessWithoutCasting(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Test S3
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: test-bucket
      - name: Region
        value: eu-central-1
      - name: Prefix
        value: data/
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	exporters := tlater.Inspect(h).Filter(Exporters).Components
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "Test S3", exp.Name)

	// Demonstrate accessing properties without type casting
	region := exp.Properties["Region"]
	assert.Equal(t, "eu-central-1", region)

	bucket := exp.Properties["Bucket"]
	assert.Equal(t, "test-bucket", bucket)

	prefix := exp.Properties["Prefix"]
	assert.Equal(t, "data/", prefix)

	// Non-existent keys return nil
	assert.Nil(t, exp.Properties["NonExistentKey"])
}

func TestInspect_DeterministicSampler(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Sample at Fixed Rate
    kind: DeterministicSampler
    properties:
      - name: SampleRate
        value: 10
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	samplers := tlater.Inspect(h).Filter(Samplers).Components
	require.Len(t, samplers, 2) // SamplingSequencer + DeterministicSampler

	// Find the DeterministicSampler
	var sampler ComponentInfo
	for _, s := range samplers {
		if s.Kind == "DeterministicSampler" {
			sampler = s
			break
		}
	}

	assert.Equal(t, "Sample at Fixed Rate", sampler.Name)
	assert.Equal(t, "DeterministicSampler", sampler.Kind)
	assert.Equal(t, "sampler", sampler.Style)
	assert.NotNil(t, sampler.Properties)
	assert.Equal(t, 10, sampler.Properties["SampleRate"])
}

func TestInspect_EMAThroughputSampler(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Throughput Sampler
    kind: EMAThroughputSampler
    properties:
      - name: GoalThroughputPerSec
        value: 100
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	samplers := tlater.Inspect(h).Filter(Samplers).Components
	require.Len(t, samplers, 2)

	var sampler ComponentInfo
	for _, s := range samplers {
		if s.Kind == "EMAThroughputSampler" {
			sampler = s
			break
		}
	}

	assert.Equal(t, "Throughput Sampler", sampler.Name)
	assert.Equal(t, "EMAThroughputSampler", sampler.Kind)
	assert.Equal(t, "sampler", sampler.Style)
	assert.NotNil(t, sampler.Properties)
	assert.Equal(t, 100, sampler.Properties["GoalThroughputPerSec"])
}

func TestInspect_Dropper(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Drop These
    kind: Dropper
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	samplers := tlater.Inspect(h).Filter(Samplers).Components
	require.Len(t, samplers, 2) // SamplingSequencer + Dropper

	var dropper ComponentInfo
	for _, s := range samplers {
		if s.Kind == "Dropper" {
			dropper = s
			break
		}
	}

	assert.Equal(t, "Drop These", dropper.Name)
	assert.Equal(t, "Dropper", dropper.Kind)
	assert.Equal(t, "dropper", dropper.Style)
	assert.NotNil(t, dropper.Properties)
}

func TestInspect_ErrorExistsCondition(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Check for Errors
    kind: ErrorExistsCondition
  - name: Keep All_1
    kind: KeepAllSampler
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	samplers := tlater.Inspect(h).Filter(Samplers).Components
	require.Len(t, samplers, 3) // SamplingSequencer + ErrorExistsCondition + KeepAllSampler

	var condition ComponentInfo
	for _, s := range samplers {
		if s.Kind == "ErrorExistsCondition" {
			condition = s
			break
		}
	}

	assert.Equal(t, "Check for Errors", condition.Name)
	assert.Equal(t, "ErrorExistsCondition", condition.Kind)
	assert.Equal(t, "condition", condition.Style)
	assert.NotNil(t, condition.Properties)
}

func TestInspect_VersionHandling(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Receiver with Version
    kind: OTelReceiver
    version: v0.1.0
  - name: Exporter without Version
    kind: HoneycombExporter
connections:
  - source:
      component: Receiver with Version
      port: Traces
      type: OTelTraces
    destination:
      component: Exporter without Version
      port: Traces
      type: OTelTraces
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	result := tlater.Inspect(h)
	require.Len(t, result.Components, 2)

	// Find the receiver with specified version
	var receiver ComponentInfo
	var exporter ComponentInfo
	for _, comp := range result.Components {
		if comp.Name == "Receiver with Version" {
			receiver = comp
		} else if comp.Name == "Exporter without Version" {
			exporter = comp
		}
	}

	// Verify receiver returns the HPSF-specified version
	assert.Equal(t, "v0.1.0", receiver.Version, "Should return version specified in HPSF")

	// Verify exporter returns the template version when not specified in HPSF
	assert.NotEmpty(t, exporter.Version, "Should fallback to template version")
	// Get the actual template component to verify it matches
	tc := tlater.GetComponents()["HoneycombExporter"]
	assert.Equal(t, tc.Version, exporter.Version, "Should match template component version")
}

func TestInspect_CompareIntegerFieldCondition(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Check Status Code
    kind: CompareIntegerFieldCondition
    properties:
      - name: Fields
        value: ["http.status_code"]
      - name: Operator
        value: "=="
      - name: Value
        value: 500
  - name: Keep All_1
    kind: KeepAllSampler
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	samplers := tlater.Inspect(h).Filter(Samplers).Components
	require.Len(t, samplers, 3)

	var condition ComponentInfo
	for _, s := range samplers {
		if s.Kind == "CompareIntegerFieldCondition" {
			condition = s
			break
		}
	}

	assert.Equal(t, "Check Status Code", condition.Name)
	assert.Equal(t, "CompareIntegerFieldCondition", condition.Kind)
	assert.Equal(t, "condition", condition.Style)
	assert.NotNil(t, condition.Properties)
	assert.Equal(t, "==", condition.Properties["Operator"])
	assert.Equal(t, 500, condition.Properties["Value"])
}

func TestInspect_SamplingSequencer(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling Here
    kind: SamplingSequencer
  - name: Sample_1
    kind: DeterministicSampler
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	samplers := tlater.Inspect(h).Filter(Samplers).Components
	require.Len(t, samplers, 2)

	var sequencer ComponentInfo
	for _, s := range samplers {
		if s.Kind == "SamplingSequencer" {
			sequencer = s
			break
		}
	}

	assert.Equal(t, "Start Sampling Here", sequencer.Name)
	assert.Equal(t, "SamplingSequencer", sequencer.Kind)
	assert.Equal(t, "startsampling", sequencer.Style)
	assert.NotNil(t, sequencer.Properties)
}

func TestInspect_MultipleSamplingComponents(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTel Receiver_1
    kind: OTelReceiver
  - name: Start Sampling_1
    kind: SamplingSequencer
  - name: Check Errors
    kind: ErrorExistsCondition
  - name: Drop Bad Traffic
    kind: Dropper
  - name: Sample Good Traffic
    kind: DeterministicSampler
  - name: Send to Honeycomb_1
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	// Test getting all sampling-related components
	samplers := tlater.Inspect(h).Filter(Samplers).Components
	require.Len(t, samplers, 4, "should have SamplingSequencer + ErrorExistsCondition + Dropper + DeterministicSampler")

	// Verify we have all the expected types
	styles := make(map[string]int)
	kinds := make(map[string]bool)
	for _, s := range samplers {
		styles[s.Style]++
		kinds[s.Kind] = true
	}

	assert.Equal(t, 1, styles["startsampling"])
	assert.Equal(t, 1, styles["condition"])
	assert.Equal(t, 1, styles["dropper"])
	assert.Equal(t, 1, styles["sampler"])

	assert.True(t, kinds["SamplingSequencer"])
	assert.True(t, kinds["ErrorExistsCondition"])
	assert.True(t, kinds["Dropper"])
	assert.True(t, kinds["DeterministicSampler"])
}

func TestInspectionResult_Exporters(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
  - name: Honeycomb Export
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-key
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
  - name: S3 Archive
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-bucket
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	result := tlater.Inspect(h)

	// Exporters() should return only exporters
	exporters := result.Filter(Exporters).Components
	require.Len(t, exporters, 2)
	assert.Equal(t, "Honeycomb Export", exporters[0].Name)
	assert.Equal(t, "exporter", exporters[0].Style)
	assert.Equal(t, "HoneycombExporter", exporters[0].Kind)
	assert.Equal(t, "S3 Archive", exporters[1].Name)
	assert.Equal(t, "exporter", exporters[1].Style)
	assert.Equal(t, "S3ArchiveExporter", exporters[1].Kind)
}

func TestInspectionResult_Receivers(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
    properties:
      - name: GRPCPort
        value: 4317
  - name: Honeycomb Export
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-key
  - name: Nop Receiver
    kind: NopReceiver
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	result := tlater.Inspect(h)

	// Receivers() should return only receivers
	receivers := result.Filter(Receivers).Components
	require.Len(t, receivers, 2)
	assert.Equal(t, "OTLP Receiver", receivers[0].Name)
	assert.Equal(t, "receiver", receivers[0].Style)
	assert.Equal(t, "OTelReceiver", receivers[0].Kind)
	assert.Equal(t, 4317, receivers[0].Properties["GRPCPort"])
	assert.Equal(t, "Nop Receiver", receivers[1].Name)
	assert.Equal(t, "receiver", receivers[1].Style)
	assert.Equal(t, "NopReceiver", receivers[1].Kind)
}

func TestInspectionResult_Processors(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
    properties:
      - name: CheckInterval
        value: 1s
  - name: Honeycomb Export
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	result := tlater.Inspect(h)

	// Processors() should return only processors
	processors := result.Filter(Processors).Components
	require.Len(t, processors, 1)
	assert.Equal(t, "Memory Limiter", processors[0].Name)
	assert.Equal(t, "processor", processors[0].Style)
	assert.Equal(t, "MemoryLimiterProcessor", processors[0].Kind)
	assert.Equal(t, "1s", processors[0].Properties["CheckInterval"])
}

func TestInspectionResult_DirectComponentsAccess(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
  - name: Honeycomb Export
    kind: HoneycombExporter
  - name: S3 Archive
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-bucket
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	result := tlater.Inspect(h)

	// Can access all components directly
	require.Len(t, result.Components, 4)

	// Verify all components are present with correct styles
	assert.Equal(t, "OTLP Receiver", result.Components[0].Name)
	assert.Equal(t, "receiver", result.Components[0].Style)

	assert.Equal(t, "Memory Limiter", result.Components[1].Name)
	assert.Equal(t, "processor", result.Components[1].Style)

	assert.Equal(t, "Honeycomb Export", result.Components[2].Name)
	assert.Equal(t, "exporter", result.Components[2].Style)

	assert.Equal(t, "S3 Archive", result.Components[3].Name)
	assert.Equal(t, "exporter", result.Components[3].Style)

	// Can iterate through all components
	styleCount := make(map[string]int)
	for _, comp := range result.Components {
		styleCount[comp.Style]++
	}
	assert.Equal(t, 1, styleCount["receiver"])
	assert.Equal(t, 1, styleCount["processor"])
	assert.Equal(t, 2, styleCount["exporter"])
}

func TestInspectionResult_EmptyFilters(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Honeycomb Export
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	result := tlater.Inspect(h)

	// Only has exporters, so receivers and processors should be empty
	assert.Len(t, result.Filter(Exporters).Components, 1)
	assert.Len(t, result.Filter(Receivers).Components, 0)
	assert.Len(t, result.Filter(Processors).Components, 0)
}

func TestInspectionResult_Samplers(t *testing.T) {
	tlater := NewEmptyTranslator()
	comps, err := data.LoadEmbeddedComponents()
	require.NoError(t, err)
	tlater.InstallComponents(comps)
	require.Equal(t, comps, tlater.GetComponents())

	templates, err := data.LoadEmbeddedTemplates()
	require.NoError(t, err)
	tlater.InstallTemplates(templates)
	require.Equal(t, templates, tlater.GetTemplates())

	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
  - name: Start Sampling
    kind: SamplingSequencer
  - name: Check Errors
    kind: ErrorExistsCondition
  - name: Drop Errors
    kind: Dropper
  - name: Sample Traffic
    kind: DeterministicSampler
  - name: Honeycomb Export
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	result := tlater.Inspect(h)

	// Samplers() should return only sampling-related components (condition, dropper, sampler, startsampling)
	samplers := result.Filter(Samplers).Components
	require.Len(t, samplers, 4)

	// Verify each has the correct style
	for _, sampler := range samplers {
		assert.Contains(t, []string{"condition", "dropper", "sampler", "startsampling"}, sampler.Style)

		switch sampler.Kind {
		case "SamplingSequencer":
			assert.Equal(t, "startsampling", sampler.Style)
			assert.Equal(t, "Start Sampling", sampler.Name)
		case "ErrorExistsCondition":
			assert.Equal(t, "condition", sampler.Style)
			assert.Equal(t, "Check Errors", sampler.Name)
		case "Dropper":
			assert.Equal(t, "dropper", sampler.Style)
			assert.Equal(t, "Drop Errors", sampler.Name)
		case "DeterministicSampler":
			assert.Equal(t, "sampler", sampler.Style)
			assert.Equal(t, "Sample Traffic", sampler.Name)
		}
	}

	// Verify non-sampling components are excluded
	allComponents := result.Components
	assert.Len(t, allComponents, 7)

	// Count styles to ensure all 4 sampling styles are represented
	styleCount := make(map[string]int)
	for _, s := range samplers {
		styleCount[s.Style]++
	}
	assert.Equal(t, 1, styleCount["startsampling"])
	assert.Equal(t, 1, styleCount["condition"])
	assert.Equal(t, 1, styleCount["dropper"])
	assert.Equal(t, 1, styleCount["sampler"])
}

func TestVersionError_Error(t *testing.T) {
	msg := "version mismatch error"
	err := NewVersionError(msg)
	assert.Equal(t, msg, err.Error())
}

func TestVersionError_Is(t *testing.T) {
	versionErr := NewVersionError("test error")

	// Test that errors.Is works with VersionError
	assert.True(t, errors.Is(versionErr, &VersionError{}))

	// Test with another VersionError instance
	otherVersionErr := NewVersionError("other error")
	assert.True(t, errors.Is(versionErr, otherVersionErr))

	// Test with a different error type
	genericErr := errors.New("generic error")
	assert.False(t, errors.Is(versionErr, genericErr))

	// Test with wrapped VersionError
	wrappedErr := fmt.Errorf("wrapped: %w", versionErr)
	assert.True(t, errors.Is(wrappedErr, &VersionError{}))
}

func TestVersionError_As(t *testing.T) {
	originalMsg := "test version error"
	versionErr := NewVersionError(originalMsg)

	// Test that errors.As works with VersionError
	var targetErr *VersionError
	assert.True(t, errors.As(versionErr, &targetErr))
	assert.Equal(t, originalMsg, targetErr.Error())
	assert.Equal(t, versionErr, targetErr)

	// Test with wrapped VersionError
	wrappedErr := fmt.Errorf("wrapped: %w", versionErr)
	var wrappedTargetErr *VersionError
	assert.True(t, errors.As(wrappedErr, &wrappedTargetErr))
	assert.Equal(t, originalMsg, wrappedTargetErr.Error())

	// Test with incompatible error type
	genericErr := errors.New("generic error")
	var genericTargetErr *VersionError
	assert.False(t, errors.As(genericErr, &genericTargetErr))
	assert.Nil(t, genericTargetErr)
}

func TestVersionError_IsAndAs_Integration(t *testing.T) {
	// Test integration of Is and As methods with real usage scenarios
	originalErr := NewVersionError("component version v2.0.0 is not supported, minimum required v3.0.0")

	// Simulate error being wrapped in the application
	applicationErr := fmt.Errorf("failed to create component: %w", originalErr)

	// Check if the error is a VersionError using errors.Is
	if errors.Is(applicationErr, &VersionError{}) {
		// Extract the VersionError using errors.As
		var versionErr *VersionError
		if errors.As(applicationErr, &versionErr) {
			assert.Contains(t, versionErr.Error(), "component version")
			assert.Contains(t, versionErr.Error(), "minimum required")
		} else {
			t.Fatal("errors.As should have succeeded for VersionError")
		}
	} else {
		t.Fatal("errors.Is should have identified VersionError")
	}
}
