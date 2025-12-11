package config

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	y "gopkg.in/yaml.v3"
)

// TestValidateAtLeastOneOf tests the at_least_one_of validation type
func TestValidateAtLeastOneOf(t *testing.T) {
	// Create a test template component
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "PropA", Type: hpsf.PTYPE_STRING},
			{Name: "PropB", Type: hpsf.PTYPE_STRING},
			{Name: "PropC", Type: hpsf.PTYPE_STRING},
		},
		Validations: []string{
			"at_least_one_of(PropA, PropB, PropC)",
		},
	}

	// Test case 1: All properties empty - should fail
	t.Run("AllEmpty", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "PropA", Value: ""},
				{Name: "PropB", Value: ""},
				{Name: "PropC", Value: ""},
			},
		}
		err := tc.Validate(component)
		if err == nil {
			t.Error("Expected validation to fail when all properties are empty")
		}
	})

	// Test case 2: One property set - should pass
	t.Run("OneSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "PropA", Value: "value"},
				{Name: "PropB", Value: ""},
				{Name: "PropC", Value: ""},
			},
		}
		err := tc.Validate(component)
		if err != nil {
			t.Errorf("Expected validation to pass when one property is set, got: %v", err)
		}
	})

	// Test case 3: Multiple properties set - should pass
	t.Run("MultipleSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "PropA", Value: "value1"},
				{Name: "PropB", Value: "value2"},
				{Name: "PropC", Value: ""},
			},
		}
		err := tc.Validate(component)
		if err != nil {
			t.Errorf("Expected validation to pass when multiple properties are set, got: %v", err)
		}
	})
}

// TestValidateExactlyOneOf tests the exactly_one_of validation type
func TestValidateExactlyOneOf(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "APIKey", Type: hpsf.PTYPE_STRING},
			{Name: "BearerToken", Type: hpsf.PTYPE_STRING},
			{Name: "BasicAuth", Type: hpsf.PTYPE_STRING},
		},
		Validations: []string{
			"exactly_one_of(APIKey, BearerToken, BasicAuth)",
		},
	}

	// Test case 1: No properties set - should fail
	t.Run("NoneSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "APIKey", Value: ""},
				{Name: "BearerToken", Value: ""},
				{Name: "BasicAuth", Value: ""},
			},
		}
		err := tc.Validate(component)
		if err == nil {
			t.Error("Expected validation to fail when no properties are set")
		}
	})

	// Test case 2: Exactly one property set - should pass
	t.Run("ExactlyOneSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "APIKey", Value: "key123"},
				{Name: "BearerToken", Value: ""},
				{Name: "BasicAuth", Value: ""},
			},
		}
		err := tc.Validate(component)
		if err != nil {
			t.Errorf("Expected validation to pass when exactly one property is set, got: %v", err)
		}
	})

	// Test case 3: Multiple properties set - should fail
	t.Run("MultipleSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "APIKey", Value: "key123"},
				{Name: "BearerToken", Value: "token456"},
				{Name: "BasicAuth", Value: ""},
			},
		}
		err := tc.Validate(component)
		if err == nil {
			t.Error("Expected validation to fail when multiple properties are set")
		}
	})
}

// TestValidateMutuallyExclusive tests the mutually_exclusive validation type
func TestValidateMutuallyExclusive(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "GzipCompression", Type: hpsf.PTYPE_BOOL, Default: false},
			{Name: "LZ4Compression", Type: hpsf.PTYPE_BOOL, Default: false},
		},
		Validations: []string{
			"mutually_exclusive(GzipCompression, LZ4Compression)",
		},
	}

	// Test case 1: Both properties false - should pass
	t.Run("BothFalse", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "GzipCompression", Value: false},
				{Name: "LZ4Compression", Value: false},
			},
		}
		err := tc.Validate(component)
		if err != nil {
			t.Errorf("Expected validation to pass when both properties are false, got: %v", err)
		}
	})

	// Test case 2: Only one property true - should pass
	t.Run("OnlyOneTrue", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "GzipCompression", Value: true},
				{Name: "LZ4Compression", Value: false},
			},
		}
		err := tc.Validate(component)
		if err != nil {
			t.Errorf("Expected validation to pass when only one property is true, got: %v", err)
		}
	})

	// Test case 3: Both properties true - should fail
	t.Run("BothTrue", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "GzipCompression", Value: true},
				{Name: "LZ4Compression", Value: true},
			},
		}
		err := tc.Validate(component)
		if err == nil {
			t.Error("Expected validation to fail when both properties are true")
		}
	})
}

// TestValidateRequireTogether tests the require_together validation type
func TestValidateRequireTogether(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "Username", Type: hpsf.PTYPE_STRING},
			{Name: "Password", Type: hpsf.PTYPE_STRING},
		},
		Validations: []string{
			"require_together(Username, Password)",
		},
	}

	// Test case 1: Both properties empty - should pass
	t.Run("BothEmpty", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "Username", Value: ""},
				{Name: "Password", Value: ""},
			},
		}
		err := tc.Validate(component)
		if err != nil {
			t.Errorf("Expected validation to pass when both properties are empty, got: %v", err)
		}
	})

	// Test case 2: Both properties set - should pass
	t.Run("BothSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "Username", Value: "user123"},
				{Name: "Password", Value: "pass456"},
			},
		}
		err := tc.Validate(component)
		if err != nil {
			t.Errorf("Expected validation to pass when both properties are set, got: %v", err)
		}
	})

	// Test case 3: Only one property set - should fail
	t.Run("OnlyOneSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "Username", Value: "user123"},
				{Name: "Password", Value: ""},
			},
		}
		err := tc.Validate(component)
		if err == nil {
			t.Error("Expected validation to fail when only one property is set")
		}
	})
}

// TestValidateConditionalRequireTogether tests the conditional_require_together validation type
func TestValidateConditionalRequireTogether(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "EnableTLS", Type: hpsf.PTYPE_BOOL, Default: false},
			{Name: "TLSCertPath", Type: hpsf.PTYPE_STRING},
			{Name: "TLSKeyPath", Type: hpsf.PTYPE_STRING},
		},
		Validations: []string{
			"conditional_require_together(TLSCertPath, TLSKeyPath | when EnableTLS=true)",
		},
	}

	// Test case 1: Condition false - should pass regardless of other properties
	t.Run("ConditionFalse", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "EnableTLS", Value: false},
				{Name: "TLSCertPath", Value: ""},
				{Name: "TLSKeyPath", Value: ""},
			},
		}
		err := tc.Validate(component)
		if err != nil {
			t.Errorf("Expected validation to pass when condition is false, got: %v", err)
		}
	})

	// Test case 2: Condition true and all required properties set - should pass
	t.Run("ConditionTrueAllSet", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "EnableTLS", Value: true},
				{Name: "TLSCertPath", Value: "/path/to/cert"},
				{Name: "TLSKeyPath", Value: "/path/to/key"},
			},
		}
		err := tc.Validate(component)
		if err != nil {
			t.Errorf("Expected validation to pass when condition is true and all properties are set, got: %v", err)
		}
	})

	// Test case 3: Condition true but required properties missing - should fail
	t.Run("ConditionTrueMissingProperties", func(t *testing.T) {
		component := &hpsf.Component{
			Name: "TestComponent",
			Properties: []hpsf.Property{
				{Name: "EnableTLS", Value: true},
				{Name: "TLSCertPath", Value: "/path/to/cert"},
				{Name: "TLSKeyPath", Value: ""},
			},
		}
		err := tc.Validate(component)
		if err == nil {
			t.Error("Expected validation to fail when condition is true but required properties are missing")
		}
	})
}

// TestUnknownValidationType tests that unknown validation types are handled correctly
func TestUnknownValidationType(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "PropA", Type: hpsf.PTYPE_STRING},
		},
		Validations: []string{
			"unknown_validation_type(PropA)",
		},
	}

	component := &hpsf.Component{
		Name: "TestComponent",
		Properties: []hpsf.Property{
			{Name: "PropA", Value: "value"},
		},
	}

	err := tc.Validate(component)
	if err == nil {
		t.Error("Expected validation to fail for unknown validation type")
	}
}

// TestNonExistentProperty tests that referencing non-existent properties is handled correctly
func TestNonExistentProperty(t *testing.T) {
	tc := &TemplateComponent{
		Properties: []TemplateProperty{
			{Name: "PropA", Type: hpsf.PTYPE_STRING},
		},
		Validations: []string{
			"at_least_one_of(PropA, NonExistentProp)",
		},
	}

	component := &hpsf.Component{
		Name: "TestComponent",
		Properties: []hpsf.Property{
			{Name: "PropA", Value: "value"},
		},
	}

	err := tc.Validate(component)
	if err == nil {
		t.Error("Expected validation to fail when referencing non-existent property")
	}
}

// TestParseComponentValidation tests the validation string parsing
func TestParseComponentValidation(t *testing.T) {
	tests := []struct {
		name           string
		validationStr  string
		expectType     string
		expectProps    []string
		expectCondProp string
		expectCondVal  any
		expectError    bool
	}{
		{
			name:          "simple at_least_one_of",
			validationStr: "at_least_one_of(PropA, PropB, PropC)",
			expectType:    "at_least_one_of",
			expectProps:   []string{"PropA", "PropB", "PropC"},
		},
		{
			name:          "exactly_one_of with spaces",
			validationStr: "exactly_one_of( APIKey , BearerToken )",
			expectType:    "exactly_one_of",
			expectProps:   []string{"APIKey", "BearerToken"},
		},
		{
			name:           "conditional with boolean",
			validationStr:  "conditional_require_together(TLSCertPath, TLSKeyPath | when EnableTLS=true)",
			expectType:     "conditional_require_together",
			expectProps:    []string{"TLSCertPath", "TLSKeyPath"},
			expectCondProp: "EnableTLS",
			expectCondVal:  true,
		},
		{
			name:           "conditional with string",
			validationStr:  "conditional_require_together(PropA, PropB | when Mode=production)",
			expectType:     "conditional_require_together",
			expectProps:    []string{"PropA", "PropB"},
			expectCondProp: "Mode",
			expectCondVal:  "production",
		},
		{
			name:          "invalid format",
			validationStr: "invalid_format",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationType, properties, conditionProperty, conditionValue, err := parseComponentValidation(tt.validationStr)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if validationType != tt.expectType {
				t.Errorf("Expected type %s, got %s", tt.expectType, validationType)
			}

			if len(properties) != len(tt.expectProps) {
				t.Errorf("Expected %d properties, got %d", len(tt.expectProps), len(properties))
			} else {
				for i, expected := range tt.expectProps {
					if properties[i] != expected {
						t.Errorf("Expected property[%d] to be %s, got %s", i, expected, properties[i])
					}
				}
			}

			if conditionProperty != tt.expectCondProp {
				t.Errorf("Expected condition property %s, got %s", tt.expectCondProp, conditionProperty)
			}

			if conditionValue != tt.expectCondVal {
				t.Errorf("Expected condition value %v, got %v", tt.expectCondVal, conditionValue)
			}
		})
	}
}

func TestTemplateComponentUnmarshal(t *testing.T) {
	templateData := `
kind: TestExporter
version: v0.1.0
minimum:
  collector_config: v0.120.0
  refinery_config: v2.0.0
maximum:
  collector_config: v0.1.0.0
  refinery_config: v4.0.0
name: test name
logo: honeycomb
summary: testing.
description: |-
  testing component config unmarshalling
comment: test comment
tags:
  - category:output
  - service:refinery
type: base
style: exporter
status: alpha
metadata:
  foo: bar
ports:
  # inputs
  - name: Events
    direction: input
    type: HoneycombEvents
properties:
  - name: APIKey
    summary: The API key to use to authenticate with Honeycomb.
    description: |
      The API key to use to authenticate with Honeycomb.
    type: string
    validations:
      - noblanks
    default: ${HTP_EXPORTER_APIKEY}
    advanced: true
validations:
  - "test_validation"
templates:
  - kind: refinery_config
    name: HoneycombExporter_RefineryConfig
    format: dotted
    data:
      - key: AccessKeys.SendKey
        value: "{{ .Values.APIKey }}"
        suppress_if: '{{ eq "none" (or .Values.APIKey .User.APIKey) }}'
  - kind: collector_config
    name: honeycombexporter_collector
    format: collector
    meta:
      componentSection: exporters
      signalTypes: [traces, metrics, logs]
      collectorComponentName: otlphttp
    data:
      - key: "{{ .ComponentName }}.headers.x-honeycomb-team"
        value: "{{ .Values.APIKey }}"
`
	var component TemplateComponent
	err := y.Unmarshal([]byte(templateData), &component)
	require.NoError(t, err)

	assert.Equal(t, "TestExporter", component.Kind)
	assert.Equal(t, "v0.1.0", component.Version)
	require.NotNil(t, component.Minimum)
	assert.Equal(t, "v0.120.0", component.Minimum["collector_config"])
	assert.Equal(t, "v2.0.0", component.Minimum["refinery_config"])
	require.NotNil(t, component.Maximum)
	assert.Equal(t, "v0.1.0.0", component.Maximum["collector_config"])
	assert.Equal(t, "v4.0.0", component.Maximum["refinery_config"])
	assert.Equal(t, "test name", component.Name)
	assert.Equal(t, "honeycomb", component.Logo)
	assert.Equal(t, "testing.", component.Summary)
	assert.Equal(t, "testing component config unmarshalling", component.Description)
	assert.Equal(t, "test comment", component.Comment)
	assert.Equal(t, []string{"category:output", "service:refinery"}, component.Tags)
	assert.Equal(t, ComponentTypeBase, component.Type)
	assert.Equal(t, "exporter", component.Style)
	assert.Equal(t, ComponentStatusAlpha, component.Status)
	require.NotNil(t, component.Metadata)
	assert.Equal(t, "bar", component.Metadata["foo"])
	require.Len(t, component.Ports, 1)
	port := component.Ports[0]
	assert.Equal(t, "Events", port.Name)
	assert.Equal(t, "input", port.Direction)
	assert.Equal(t, hpsf.CTYPE_HONEY, port.Type)
	require.Len(t, component.Properties, 1)
	property := component.Properties[0]
	assert.Equal(t, "APIKey", property.Name)
	assert.Equal(t, "The API key to use to authenticate with Honeycomb.", property.Summary)
	assert.Equal(t, "The API key to use to authenticate with Honeycomb.\n", property.Description)
	assert.Equal(t, hpsf.PTYPE_STRING, property.Type)
	assert.Equal(t, []string{"noblanks"}, property.Validations)
	assert.Equal(t, "${HTP_EXPORTER_APIKEY}", property.Default)
	assert.Equal(t, true, property.Advanced)
	require.Len(t, component.Validations, 1)
	assert.Equal(t, "test_validation", component.Validations[0])
	require.Len(t, component.Templates, 2)
	rt := component.Templates[0]
	assert.Equal(t, hpsftypes.RefineryConfig, rt.Kind)
	assert.Equal(t, "HoneycombExporter_RefineryConfig", rt.Name)
	assert.Equal(t, "dotted", rt.Format)
	require.Nil(t, rt.Meta)
	require.Len(t, rt.Data, 1)
	ct := component.Templates[1]
	assert.Equal(t, hpsftypes.CollectorConfig, ct.Kind)
	assert.Equal(t, "honeycombexporter_collector", ct.Name)
	assert.Equal(t, "collector", ct.Format)
	require.NotNil(t, ct.Meta)
	assert.Equal(t, "exporters", ct.Meta["componentSection"])
	assert.Equal(t, []any{"traces", "metrics", "logs"}, ct.Meta["signalTypes"])
	assert.Equal(t, "otlphttp", ct.Meta["collectorComponentName"])
	require.Len(t, ct.Data, 1)
}

// TestConnectionHelperMethods tests the HasTracesConnection, HasLogsConnection, and HasMetricsConnection helper methods
func TestConnectionHelperMethods(t *testing.T) {
	tc := &TemplateComponent{
		Kind: "TestProcessor",
		Name: "Test Processor",
	}

	t.Run("NoConnections", func(t *testing.T) {
		// With no connections, all helper methods should return false
		assert.False(t, tc.HasTracesConnection())
		assert.False(t, tc.HasLogsConnection())
		assert.False(t, tc.HasMetricsConnection())
	})

	t.Run("TracesConnectionAsSource", func(t *testing.T) {
		tc.connections = []*hpsf.Connection{
			{
				Source:      hpsf.ConnectionPort{Component: "comp1", PortName: "Traces", Type: hpsf.CTYPE_TRACES},
				Destination: hpsf.ConnectionPort{Component: "comp2", PortName: "Traces", Type: hpsf.CTYPE_TRACES},
			},
		}
		assert.True(t, tc.HasTracesConnection())
		assert.False(t, tc.HasLogsConnection())
		assert.False(t, tc.HasMetricsConnection())
	})

	t.Run("LogsConnectionAsDestination", func(t *testing.T) {
		tc.connections = []*hpsf.Connection{
			{
				Source:      hpsf.ConnectionPort{Component: "comp1", PortName: "Logs", Type: hpsf.CTYPE_LOGS},
				Destination: hpsf.ConnectionPort{Component: "comp2", PortName: "Logs", Type: hpsf.CTYPE_LOGS},
			},
		}
		assert.False(t, tc.HasTracesConnection())
		assert.True(t, tc.HasLogsConnection())
		assert.False(t, tc.HasMetricsConnection())
	})

	t.Run("MultipleConnectionTypes", func(t *testing.T) {
		tc.connections = []*hpsf.Connection{
			{
				Source:      hpsf.ConnectionPort{Component: "comp1", PortName: "Traces", Type: hpsf.CTYPE_TRACES},
				Destination: hpsf.ConnectionPort{Component: "comp2", PortName: "Traces", Type: hpsf.CTYPE_TRACES},
			},
			{
				Source:      hpsf.ConnectionPort{Component: "comp1", PortName: "Logs", Type: hpsf.CTYPE_LOGS},
				Destination: hpsf.ConnectionPort{Component: "comp2", PortName: "Logs", Type: hpsf.CTYPE_LOGS},
			},
			{
				Source:      hpsf.ConnectionPort{Component: "comp1", PortName: "Metrics", Type: hpsf.CTYPE_METRICS},
				Destination: hpsf.ConnectionPort{Component: "comp2", PortName: "Metrics", Type: hpsf.CTYPE_METRICS},
			},
		}
		assert.True(t, tc.HasTracesConnection())
		assert.True(t, tc.HasLogsConnection())
		assert.True(t, tc.HasMetricsConnection())
	})

	t.Run("OnlyMetrics", func(t *testing.T) {
		tc.connections = []*hpsf.Connection{
			{
				Source:      hpsf.ConnectionPort{Component: "comp1", PortName: "Metrics", Type: hpsf.CTYPE_METRICS},
				Destination: hpsf.ConnectionPort{Component: "comp2", PortName: "Metrics", Type: hpsf.CTYPE_METRICS},
			},
		}
		assert.False(t, tc.HasTracesConnection())
		assert.False(t, tc.HasLogsConnection())
		assert.True(t, tc.HasMetricsConnection())
	})
}
