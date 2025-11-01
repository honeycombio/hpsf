package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
)

func TestUserDefinedAPIKey(t *testing.T) {
	// This test verifies that a user-defined APIKey in HoneycombExporter
	// is correctly used in the generated collector configuration
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/user_defined_apikey.yaml")

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
	// with the user-defined APIKey
	samplingExporterConfig, findResult := collectorprovider.GetExporterConfig[otlphttpexporter.Config](collectorConfig, "otlphttp/Start_Sampling_1")
	require.True(t, findResult.Found, "Expected SamplingSequencer exporter to be found")

	honeycombTeamHeader, exists := samplingExporterConfig.ClientConfig.Headers["x-honeycomb-team"]
	require.True(t, exists, "Expected x-honeycomb-team header to exist")
	assert.Equal(t, "HONEYCOMB_API_KEY", string(honeycombTeamHeader), "Expected user-defined APIKey to be used in x-honeycomb-team header")

	// Verify Refinery rules configuration
	assert.Len(t, rulesConfig.Samplers, 1, "Expected 1 sampler in refinery config")

	// Check that the __default__ environment has a sampler
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists, "Expected __default__ sampler to exist")

	// Verify the sampler is a DeterministicSampler (KeepAllSampler uses DeterministicSampler with SampleRate=1)
	require.NotNil(t, defaultSampler.DeterministicSampler, "Expected DeterministicSampler configuration")
	assert.Equal(t, 1, defaultSampler.DeterministicSampler.SampleRate, "Expected SampleRate to be 1 for KeepAllSampler")

	// Verify that the user-defined APIKey overrode the default ${HTP_EXPORTER_APIKEY}
	// The APIKey from the HoneycombExporter component is used for the SamplingSequencer's x-honeycomb-team header
	// (verified above), which is the exporter that sends data to Refinery for sampling.
	//
	// Note: The HoneycombExporter itself doesn't appear as a separate exporter in the collector config
	// when sampling is used. Instead, its APIKey configuration is inherited by the SamplingSequencer exporter.
}
