package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeepErrors(t *testing.T) {
	// Test the HPSF parsing and KeepErrors sampler configuration
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/keep_errors.yaml")

	// Verify that the refinery rules config was generated successfully
	assert.Equal(t, 2, rulesConfig.RulesVersion)

	// Check that the default environment defaultSampler was created
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists, "Expected '__default__' environment sampler to exist")

	// Verify that it's a RulesBasedSampler
	require.NotNil(t, defaultSampler.RulesBasedSampler, "Expected __default__ sampler to be a RulesBasedSampler")

	// Check that there's exactly one rule
	rules := defaultSampler.RulesBasedSampler.Rules
	assert.Len(t, rules, 1, "Expected 1 rule in __default__ sampler")

	// Verify the rule properties from the KeepErrors template
	rule := rules[0]

	// Test rule name (from templates: Rules[0].Name)
	expectedName := "Keep traces with errors at a sample rate of 5"
	assert.Equal(t, expectedName, rule.Name)

	// Test sample rate (from templates: Rules[0].SampleRate)
	assert.Equal(t, 5, rule.SampleRate)

	// Test rule conditions
	assert.Len(t, rule.Conditions, 1, "Expected 1 condition")

	condition := rule.Conditions[0]

	// Test field name (from FieldName property)
	assert.Equal(t, "error_field", condition.Fields[0])

	// Test operator (from template: o=exists)
	assert.Equal(t, "exists", condition.Operator)

	// verify that the the collectorconfig pipeline includes the exporter to refinery
	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1, "Expected 1 traces pipeline, got %v", tracesPipelineNames)

	_, _, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found, "Expected pipeline to be found")
	assert.Len(t, exporters, 1, "Expected 1 exporter, got %s", exporters)
	assert.Contains(t, exporters, "otlphttp/Start_Sampling_1", "Expected OTel HTTP exporter")
}
