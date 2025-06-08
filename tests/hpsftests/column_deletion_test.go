package hpsftests

import (
	"slices"
	"testing"

	collectorConfigprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
)

func TestColumnDeletion(t *testing.T) {
	// Test the HPSF parsing and template generation using typed configuration
	rulesConfig, collectorConfig, errors := hpsfprovider.GetParsedConfigsFromFile(t, "column_deletion.yaml")

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

	// Check that the transform processor is in the traces pipeline
	expectedTransformProcessor := "transform/Column_Deleter"
	if !slices.Contains(processors, expectedTransformProcessor) {
		t.Errorf("Expected processor %s to be in traces pipeline, got %s", expectedTransformProcessor, processors)
	}

	// Verify the transform processor is present in the metrics pipeline
	_, processors, _, getResult = collectorConfigprovider.GetPipelineConfig(collectorConfig, "metrics")
	if !getResult.Found {
		t.Errorf("Expected metrics pipeline to be present in collector config, got %s", getResult.Components)
	}

	// We expect 2 processors: usage + transform
	if len(processors) != 2 {
		t.Errorf("Expected 2 processors (usage + transform), got %s", processors)
	}

	// Check that the transform processor is in the metrics pipeline
	if !slices.Contains(processors, expectedTransformProcessor) {
		t.Errorf("Expected processor %s to be in metrics pipeline, got %s", expectedTransformProcessor, processors)
	}

	// Verify the transform processor is present in the logs pipeline
	_, processors, _, getResult = collectorConfigprovider.GetPipelineConfig(collectorConfig, "logs")
	if !getResult.Found {
		t.Errorf("Expected logs pipeline to be present in collector config, got %s", getResult.Components)
	}

	// We expect 2 processors: usage + transform
	if len(processors) != 2 {
		t.Errorf("Expected 2 processors (usage + transform), got %s", processors)
	}

	// Check that the transform processor is in the logs pipeline
	if !slices.Contains(processors, expectedTransformProcessor) {
		t.Errorf("Expected processor %s to be in logs pipeline, got %s", expectedTransformProcessor, processors)
	}

	// Get the typed configuration for transform processor
	_, foundResult := collectorConfigprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, expectedTransformProcessor)
	if !foundResult.Found {
		t.Errorf("Expected transform processor to be present in collector config, got %s", foundResult.Components)
	}
}
