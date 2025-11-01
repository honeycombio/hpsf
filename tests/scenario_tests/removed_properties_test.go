package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
)

func TestRemovedProperties(t *testing.T) {
	// This test verifies that removed properties (mode in HoneycombExporter,
	// headers in SamplingSequencer) are handled gracefully and still generate
	// valid collector and refinery configs
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/removed_properties.yaml")

	// Verify the traces pipeline exists and has the correct components
	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1, "Expected 1 traces pipeline, got %v", tracesPipelineNames)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found, "Expected traces pipeline to be found")

	// Check pipeline components
	assert.Len(t, receivers, 1, "Expected 1 receiver")
	assert.Contains(t, receivers, "otlp/Receive_OTel_1", "Expected OTel receiver")

	// Sampling components are translated to Refinery rules, not collector processors
	assert.Len(t, processors, 1, "Expected 1 processor (usage)")
	assert.Contains(t, processors, "usage", "Expected usage processor")

	assert.Len(t, exporters, 1, "Expected 1 exporter")
	assert.Contains(t, exporters, "otlphttp/Start_Sampling_1", "Expected SamplingSequencer exporter")

	// Verify the SamplingSequencer exporter has the correct x-honeycomb-team header
	// Since the Headers property is removed/deprecated, it should be ignored and
	// the system should use the default APIKey from the HoneycombExporter
	samplingExporterConfig, findResult := collectorprovider.GetExporterConfig[otlphttpexporter.Config](collectorConfig, "otlphttp/Start_Sampling_1")
	require.True(t, findResult.Found, "Expected SamplingSequencer exporter to be found")

	honeycombTeamHeader, exists := samplingExporterConfig.ClientConfig.Headers["x-honeycomb-team"]
	require.True(t, exists, "Expected x-honeycomb-team header to exist")
	// The removed Headers property should be ignored, and the default APIKey should be used
	assert.Equal(t, "HTP_EXPORTER_APIKEY", string(honeycombTeamHeader), "Expected default APIKey to be used (removed Headers property should be ignored)")

	// Verify Refinery rules configuration
	assert.Len(t, rulesConfig.Samplers, 1, "Expected 1 sampler in refinery config")

	// Check that the __default__ environment has a sampler
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists, "Expected __default__ sampler to exist")

	// Verify the sampler is a DeterministicSampler (KeepAllSampler uses DeterministicSampler with SampleRate=1)
	require.NotNil(t, defaultSampler.DeterministicSampler, "Expected DeterministicSampler configuration")
	assert.Equal(t, 1, defaultSampler.DeterministicSampler.SampleRate, "Expected SampleRate to be 1 for KeepAllSampler")

	// Verify that the config generation succeeded despite the removed properties
	// The presence of valid configs above confirms this implicitly
}
