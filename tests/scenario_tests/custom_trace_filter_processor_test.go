package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomTraceFilterProcessor(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/customtracefilterprocessor.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	// Should only have traces pipeline
	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1, "Expected 1 traces pipeline, got %v", tracesPipelineNames)

	// Check filter processor is in pipeline
	_, processors, _, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)
	assert.Contains(t, processors, "filter/trace_filter_1")

	// Check filter processor configuration exists and is valid
	filterConfig, findResult := collectorprovider.GetProcessorConfig[filterprocessor.Config](collectorConfig, "filter/trace_filter_1")
	require.True(t, findResult.Found, "Expected filter processor to be found, found (%v)", findResult.Components)

	// Verify traces config exists
	require.NotNil(t, filterConfig.Traces)
}
