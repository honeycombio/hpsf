package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/stretchr/testify/require"
)

func TestApplyOTelK8sSemconvProcessor(t *testing.T) {
	_, collectorConfig, err := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/apply_otel_k8s_semconv.yaml")
	require.False(t, err.HasErrors())

	logsPipelines := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	require.Len(t, logsPipelines, 1, "Expected 1 logs pipeline, got %v", logsPipelines)

	_, processors, _, result := collectorprovider.GetPipelineConfig(collectorConfig, logsPipelines[0].String())
	require.True(t, result.Found)
	require.Contains(t, processors, "transform/apply_k8s_semconv")

	transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/apply_k8s_semconv")
	require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)
	require.Equal(t, ottl.IgnoreError, transformConfig.ErrorMode, "Expected ErrorMode to be \"ignore\"")
	// require.Equal(t, "log", transformConfig.LogStatements[0].Context) // not currently possible as the context type is internal
	require.Len(t, transformConfig.LogStatements, 1, "Expected 1 log statement, got %v", len(transformConfig.LogStatements))
	require.Len(t, transformConfig.LogStatements[0].Statements, 6, "Expected 6 statements, got %v", len(transformConfig.LogStatements[0].Statements))
}
