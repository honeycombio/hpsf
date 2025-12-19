package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomMetricFilterProcessor(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/custommetricfilterprocessor.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	// Should only have metrics pipeline
	metricsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "metrics")
	assert.Len(t, metricsPipelineNames, 1, "Expected 1 metrics pipeline, got %v", metricsPipelineNames)

	// Check filter processor is in pipeline
	_, processors, _, getResult := collectorprovider.GetPipelineConfig(collectorConfig, metricsPipelineNames[0].String())
	require.True(t, getResult.Found)
	assert.Contains(t, processors, "filter/metric_filter_1")

	// Check filter processor configuration exists and is valid
	filterConfig, findResult := collectorprovider.GetProcessorConfig[filterprocessor.Config](collectorConfig, "filter/metric_filter_1")
	require.True(t, findResult.Found, "Expected filter processor to be found, found (%v)", findResult.Components)

	// Verify metrics config exists
	require.NotNil(t, filterConfig.Metrics)
}
