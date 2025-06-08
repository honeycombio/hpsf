package hpsftests

import (
	"slices"
	"testing"

	collectorConfigprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
)

func TestParseUserAgent(t *testing.T) {
	// Test the HPSF parsing and template generation using typed configuration
	rulesConfig, collectorConfig, errors := hpsfprovider.GetParsedConfigsFromFile(t, "parse_user_agent.yaml")

	// First, verify that the refinery config was generated successfully
	if len(rulesConfig.Samplers) != 1 {
		t.Errorf("Expected 1 sampler in refinery config, got %d", len(rulesConfig.Samplers))
	}

	// Check if there are any errors in parsing - any errors should fail the test
	errors.FailIfError(t)

	// Verify the transform processor is present in the traces pipeline
	_, processors, _, getResult := collectorConfigprovider.GetPipelineConfig(collectorConfig, "traces")
	if !getResult.Found {
		t.Errorf("Expected traces pipeline to be present in collector config, got %s", getResult.Components)
	}

	// We expect 2 processors: usage + transform
	if len(processors) != 2 {
		t.Errorf("Expected 2 processors (usage + transform), got %s", processors)
	}

	// Check that the transform processor is in the pipeline
	expectedTransformProcessor := "transform/User_Agent_Parser"
	if !slices.Contains(processors, expectedTransformProcessor) {
		t.Errorf("Expected processor %s to be in pipeline, got %s", expectedTransformProcessor, processors)
	}

	// Get the typed configuration for transform processor
	_, foundResult := collectorConfigprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, expectedTransformProcessor)
	if !foundResult.Found {
		t.Errorf("Expected transform processor to be present in collector config, got %s", foundResult.Components)
	}
}
