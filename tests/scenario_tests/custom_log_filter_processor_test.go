package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomLogFilterProcessor(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/customlogfilterprocessor.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	// Should only have logs pipeline
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	assert.Len(t, logsPipelineNames, 1, "Expected 1 logs pipeline, got %v", logsPipelineNames)

	// Check filter processor is in pipeline
	_, processors, _, getResult := collectorprovider.GetPipelineConfig(collectorConfig, logsPipelineNames[0].String())
	require.True(t, getResult.Found)
	assert.Contains(t, processors, "filter/log_filter_1")

	// Check filter processor configuration exists and is valid
	filterConfig, findResult := collectorprovider.GetProcessorConfig[filterprocessor.Config](collectorConfig, "filter/log_filter_1")
	require.True(t, findResult.Found, "Expected filter processor to be found, found (%v)", findResult.Components)

	// Verify logs config exists
	require.NotNil(t, filterConfig.Logs)
}
