package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSampleOnlyDeterministic(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/sample_only_deterministic.yaml")

	// Verify the traces pipeline exists and has the correct components
	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1, "Expected 1 traces pipeline, got %v", tracesPipelineNames)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found, "Expected traces pipeline to be found")

	// Check pipeline components
	assert.Len(t, receivers, 1, "Expected 1 receiver")
	assert.Contains(t, receivers, "otlp/Receive_OTel_1", "Expected OTel receiver")

	// Sampling components are translated to Refinery rules, not collector processors
	assert.Len(t, processors, 2, "Expected 2 processors (usage + memory_limiter)")
	assert.Contains(t, processors, "usage", "Expected usage processor")
	assert.Contains(t, processors, "memory_limiter/Receive_OTel_1", "Expected memory_limiter processor")
	assert.Len(t, exporters, 1, "Expected 1 exporter")
	assert.Contains(t, exporters, "otlphttp/Start_Sampling_1", "Expected SamplingSequencer exporter")

	// Verify Refinery rules configuration
	assert.Len(t, rulesConfig.Samplers, 1, "Expected 1 sampler in refinery config")

	// Check that the __default__ environment has a sampler
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists, "Expected __default__ sampler to exist")

	// Verify the sampler is a DeterministicSampler
	require.NotNil(t, defaultSampler.DeterministicSampler, "Expected DeterministicSampler configuration")
	assert.Equal(t, 100, defaultSampler.DeterministicSampler.SampleRate, "Expected SampleRate to be 100")
}
