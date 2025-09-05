package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/stretchr/testify/require"
)

func TestExtractLogPropertiesProcessor(t *testing.T) {
	_, collectorConfig, err := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/extract_log_attributes.yaml")
	require.False(t, err.HasErrors())

	logsPipelines := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	require.Len(t, logsPipelines, 1, "Expected 1 logs pipeline, got %v", logsPipelines)

	_, processors, _, result := collectorprovider.GetPipelineConfig(collectorConfig, logsPipelines[0].String())
	require.True(t, result.Found)
	require.Contains(t, processors, "transform/extract_log_attributes")

	transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/extract_log_attributes")
	require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)
	require.Equal(t, ottl.IgnoreError, transformConfig.ErrorMode, "Expected ErrorMode to be \"ignore\"")
	require.Len(t, transformConfig.LogStatements, 1, "Expected 1 log statement, got %v", len(transformConfig.LogStatements))
	require.Len(t, transformConfig.LogStatements[0].Statements, 1, "Expected 1 statement, got %v", len(transformConfig.LogStatements[0].Statements))
}
