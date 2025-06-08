package hpsftests

import (
	"testing"

	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
)

func TestKeepErrors(t *testing.T) {
	// Test the HPSF parsing and KeepErrors sampler configuration
	rulesConfig, _, errors := hpsfprovider.GetParsedConfigsFromFile(t, "keep_errors.yaml")
	errors.FailIfError(t)

	// Verify that the refinery rules config was generated successfully
	if rulesConfig.RulesVersion != 2 {
		t.Errorf("Expected RulesVersion to be 2, got %d", rulesConfig.RulesVersion)
	}

	// Check that the production environment sampler was created
	productionSampler, exists := rulesConfig.Samplers["production"]
	if !exists {
		t.Fatal("Expected 'production' environment sampler to exist")
	}

	// Verify that it's a RulesBasedSampler
	if productionSampler.RulesBasedSampler == nil {
		t.Fatal("Expected production sampler to be a RulesBasedSampler")
	}

	// Check that there's exactly one rule
	rules := productionSampler.RulesBasedSampler.Rules
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule in production sampler, got %d", len(rules))
	}

	// Verify the rule properties from the KeepErrors template
	rule := rules[0]

	// Test rule name (from templates: Rules[0].Name)
	expectedName := "Keep traces with errors at a sample rate of 5"
	if rule.Name != expectedName {
		t.Errorf("Expected rule name to be '%s', got '%s'", expectedName, rule.Name)
	}

	// Test sample rate (from templates: Rules[0].SampleRate)
	if rule.SampleRate != 5 {
		t.Errorf("Expected sample rate to be 5, got %d", rule.SampleRate)
	}

	// Test rule conditions (from templates: Rules[0].!condition!)
	if len(rule.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(rule.Conditions))
	}

	condition := rule.Conditions[0]

	// Test field name (from FieldName property)
	if condition.Field != "error_field" {
		t.Errorf("Expected condition field to be 'error_field', got '%s'", condition.Field)
	}

	// Test operator (from template: o=exists)
	if condition.Operator != "exists" {
		t.Errorf("Expected condition operator to be 'exists', got '%s'", condition.Operator)
	}

	// Verify that the default environment also has a sampler (should be DeterministicSampler)
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	if !exists {
		t.Fatal("Expected '__default__' environment sampler to exist")
	}

	// The default should be a DeterministicSampler with rate 1
	if defaultSampler.DeterministicSampler == nil {
		t.Fatal("Expected default sampler to be a DeterministicSampler")
	}

	if defaultSampler.DeterministicSampler.SampleRate != 1 {
		t.Errorf("Expected default sampler rate to be 1, got %d", defaultSampler.DeterministicSampler.SampleRate)
	}
}
