package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
)

func TestMultipleOTLPExporters(t *testing.T) {
	t.Skip("Skipping test for multiple OTLP exporters until we can figure out how to combine exporters from two pipelines")

	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/multiple_otlp_exporters.yaml")

	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1, "Expected 1 traces pipeline, got %v", tracesPipelineNames)

	_, _, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found, "Expected pipeline to be found")
	assert.Len(t, exporters, 2, "Expected 2 exporters, got %s", exporters)

	customBackendConfig, findResult := collectorprovider.GetExporterConfig[otlphttpexporter.Config](collectorConfig, "otlphttp/My_Custom_backend")
	require.True(t, findResult.Found, "Expected exporter to find \"%v\", found (%v)", findResult.SearchString, findResult.Components)

	assert.Equal(t, "MY_KEY", string(customBackendConfig.ClientConfig.Headers["x-custom-backend"]))

	assert.Len(t, rulesConfig.Samplers, 1)

}
