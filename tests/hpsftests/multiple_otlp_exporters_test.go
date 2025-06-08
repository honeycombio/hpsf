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

	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "multiple_otlp_exporters.yaml")

	_, _, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, "traces")
	require.True(t, getResult.Found, "Expected pipeline to be found")
	assert.Len(t, exporters, 2, "Expected 2 exporters, got %s", exporters)

	customBackendConfig, findResult := collectorprovider.GetExporterConfig[otlphttpexporter.Config](collectorConfig, "otlphttp/My_Custom_backend")
	require.True(t, findResult.Found, "Expected exporter to find \"%v\", found (%v)", findResult.SearchString, findResult.Components)

	assert.Equal(t, "MY_KEY", string(customBackendConfig.ClientConfig.Headers["x-custom-backend"]))

	assert.Len(t, rulesConfig.Samplers, 1)

}
