package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterLogsBySeverityDefaults(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/filter_logs_by_severity_defaults.yaml")

	// Verify the logs pipeline exists and has the correct components
	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, "logs")
	require.True(t, getResult.Found, "Expected logs pipeline to be found")

	// Check pipeline components
	assert.Len(t, receivers, 1, "Expected 1 receiver")
	assert.Contains(t, receivers, "otlp/OTel_Receiver_1", "Expected OTel receiver")

	assert.Len(t, processors, 2, "Expected 2 processors (usage + filter)")
	assert.Contains(t, processors, "usage", "Expected usage processor")
	assert.Contains(t, processors, "filter/Log_Severity_Filter", "Expected filter processor")

	assert.Len(t, exporters, 1, "Expected 1 exporter")
	assert.Contains(t, exporters, "otlphttp/OTel_Exporter_1", "Expected OTel HTTP exporter")

	// Verify the filter processor configuration exists
	filterConfig, componentGetResult := collectorprovider.GetProcessorConfig[filterprocessor.Config](collectorConfig, "filter/Log_Severity_Filter")
	require.True(t, componentGetResult.Found, "Expected filter processor to exist in config")
	assert.Equal(t, 1, len(filterConfig.Logs.LogConditions))
}

func TestFilterLogsBySeverityCustomSeverity(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/filter_logs_by_severity_all.yaml")

	// Verify the logs pipeline exists and has the correct components
	_, processors, _, getResult := collectorprovider.GetPipelineConfig(collectorConfig, "logs")
	require.True(t, getResult.Found, "Expected logs pipeline to be found")

	assert.Len(t, processors, 2, "Expected 2 processors (usage + filter)")
	assert.Contains(t, processors, "usage", "Expected usage processor")
	assert.Contains(t, processors, "filter/Error_Log_Filter", "Expected filter processor")

	// Verify the filter processor configuration exists
	filterConfig, componentGetResult := collectorprovider.GetProcessorConfig[filterprocessor.Config](collectorConfig, "filter/Error_Log_Filter")
	require.True(t, componentGetResult.Found, "Expected filter processor to exist in config")
	assert.Equal(t, 1, len(filterConfig.Logs.LogConditions))
}
