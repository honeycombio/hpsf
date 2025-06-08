package hpsftests

import (
	"testing"

	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeepErrors(t *testing.T) {
	// Test the HPSF parsing and KeepErrors sampler configuration
	rulesConfig, _, _ := hpsfprovider.GetParsedConfigsFromFile(t, "keep_errors.yaml")

	// Verify that the refinery rules config was generated successfully
	assert.Equal(t, 2, rulesConfig.RulesVersion)

	// Check that the production environment sampler was created
	productionSampler, exists := rulesConfig.Samplers["production"]
	require.True(t, exists, "Expected 'production' environment sampler to exist")

	// Verify that it's a RulesBasedSampler
	require.NotNil(t, productionSampler.RulesBasedSampler, "Expected production sampler to be a RulesBasedSampler")

	// Check that there's exactly one rule
	rules := productionSampler.RulesBasedSampler.Rules
	assert.Len(t, rules, 1, "Expected 1 rule in production sampler")

	// Verify the rule properties from the KeepErrors template
	rule := rules[0]

	// Test rule name (from templates: Rules[0].Name)
	expectedName := "Keep traces with errors at a sample rate of 5"
	assert.Equal(t, expectedName, rule.Name)

	// Test sample rate (from templates: Rules[0].SampleRate)
	assert.Equal(t, 5, rule.SampleRate)

	// Test rule conditions (from templates: Rules[0].!condition!)
	assert.Len(t, rule.Conditions, 1, "Expected 1 condition")

	condition := rule.Conditions[0]

	// Test field name (from FieldName property)
	assert.Equal(t, "error_field", condition.Field)

	// Test operator (from template: o=exists)
	assert.Equal(t, "exists", condition.Operator)

	// Verify that the default environment also has a sampler (should be DeterministicSampler)
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists, "Expected '__default__' environment sampler to exist")

	// The default should be a DeterministicSampler with rate 1
	require.NotNil(t, defaultSampler.DeterministicSampler, "Expected default sampler to be a DeterministicSampler")

	assert.Equal(t, 1, defaultSampler.DeterministicSampler.SampleRate)
}
