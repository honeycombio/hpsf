package hpsftests

import (
	"slices"
	"testing"

	collectorConfigprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
)

func TestChangeServiceName(t *testing.T) {
	// Test the HPSF parsing and template generation using typed configuration
	rulesConfig, collectorConfig, errors := hpsfprovider.GetParsedConfigsFromFile(t, "change_service_name.yaml")

	// First, verify that the refinery config was generated successfully
	if len(rulesConfig.Samplers) != 1 {
		t.Errorf("Expected 1 sampler in refinery config, got %d", len(rulesConfig.Samplers))
	}

	// Check if there are any errors in parsing - any errors should fail the test
	errors.FailIfError(t)

	// Verify the processors are present in the traces pipeline
	// This component uses both groupbyattrs and transform processors
	_, processors, _, getResult := collectorConfigprovider.GetPipelineConfig(collectorConfig, "traces")
	if !getResult.Found {
		t.Errorf("Expected traces pipeline to be present in collector config, got %s", getResult.Components)
	}

	// We expect 3 processors: usage + groupbyattrs + transform
	if len(processors) != 3 {
		t.Errorf("Expected 3 processors (usage + groupbyattrs + transform), got %s", processors)
	}

	// Check that the groupbyattrs processor is in the pipeline
	expectedGroupByProcessor := "groupbyattrs/Service_Name_Changer"
	if !slices.Contains(processors, expectedGroupByProcessor) {
		t.Errorf("Expected processor %s to be in pipeline, got %s", expectedGroupByProcessor, processors)
	}

	// Check that the transform processor is in the pipeline
	expectedTransformProcessor := "transform/Service_Name_Changer"
	if !slices.Contains(processors, expectedTransformProcessor) {
		t.Errorf("Expected processor %s to be in pipeline, got %s", expectedTransformProcessor, processors)
	}

	// Get the typed configuration for transform processor
	_, foundResult := collectorConfigprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, expectedTransformProcessor)
	if !foundResult.Found {
		t.Errorf("Expected transform processor to be present in collector config, got %s", foundResult.Components)
	}
}
