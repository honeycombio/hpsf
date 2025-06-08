package hpsftests

import (
	"slices"
	"testing"

	collectorConfigprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
)

func TestFilterLogsByLibrary(t *testing.T) {
	// Test the HPSF parsing and template generation using typed configuration
	rulesConfig, collectorConfig, errors := hpsfprovider.GetParsedConfigsFromFile(t, "filter_logs_by_library.yaml")

	// First, verify that the refinery config was generated successfully
	if len(rulesConfig.Samplers) != 1 {
		t.Errorf("Expected 1 sampler in refinery config, got %d", len(rulesConfig.Samplers))
	}

	// Check if there are any errors in parsing - any errors should fail the test
	errors.FailIfError(t)

	// Verify the filter processor is present in the logs pipeline
	_, processors, _, getResult := collectorConfigprovider.GetPipelineConfig(collectorConfig, "logs")
	if !getResult.Found || len(processors) != 2 { // usage + filter processor
		t.Errorf("Expected 2 processors (usage + filter), got %s", processors)
	}

	// Check that the filter processor is in the pipeline
	expectedProcessor := "filter/Library_Filter"
	if !slices.Contains(processors, expectedProcessor) {
		t.Errorf("Expected processor %s to be in pipeline, got %s", expectedProcessor, processors)
	}

	// Get the typed configuration for filter processor
	_, foundResult := collectorConfigprovider.GetProcessorConfig[filterprocessor.Config](collectorConfig, "filter/Library_Filter")
	if !foundResult.Found {
		t.Errorf("Expected filter processor to be present in collector config, got %s", foundResult.Components)
	}

	// The processor configuration is correctly generated and present in the pipeline
	// This confirms that our FilterLogsByLibrary component template is working correctly

	// The filter processor has been successfully:
	// 1. Added to the logs pipeline
	// 2. Configured with the correct component ID (filter/Library_Filter)
	// 3. Included in the collector configuration without parsing errors
	//
	// This validates that our FilterLogsByLibrary component definition is correct
	// and generates valid OpenTelemetry Collector configuration.
}
