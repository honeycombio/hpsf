package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirstSamplingPipe(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/firstSamplingPipe.yaml")

	// Verify the traces pipeline exists and has the correct components
	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1, "Expected 1 traces pipeline, got %v", tracesPipelineNames)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found, "Expected traces pipeline to be found")

	// Check pipeline components
	assert.Len(t, receivers, 1, "Expected 1 receiver")
	assert.Contains(t, receivers, "otlp/OTel_Receiver_1", "Expected OTel receiver")

	// Sampling components are translated to Refinery rules, not collector processors
	assert.Len(t, processors, 1, "Expected 1 processor (usage)")
	assert.Contains(t, processors, "usage", "Expected usage processor")

	assert.Len(t, exporters, 1, "Expected 1 exporter")
	assert.Contains(t, exporters, "otlphttp/Start_Sampling_1", "Expected SamplingSequencer exporter")

	// Verify Refinery rules configuration
	assert.Len(t, rulesConfig.Samplers, 1, "Expected 1 sampler in refinery config")

	// Check that the __default__ environment has a sampler
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists, "Expected __default__ sampler to exist")

	// Verify the sampler is a RulesBasedSampler with one rule
	require.NotNil(t, defaultSampler.RulesBasedSampler, "Expected RulesBasedSampler configuration")
	assert.Len(t, defaultSampler.RulesBasedSampler.Rules, 1, "Expected 1 rule in the sampler")

	// Verify the rule has the expected conditions and sampler
	rule := defaultSampler.RulesBasedSampler.Rules[0]
	assert.Len(t, rule.Conditions, 1, "Expected 1 condition in the rule")

	// Check the condition is a LongDurationCondition
	condition := rule.Conditions[0]
	assert.Equal(t, "duration_ms", condition.Field, "Expected duration_ms field")
	assert.Equal(t, ">=", condition.Operator, "Expected >= operator")
	assert.Equal(t, 1000, condition.Value, "Expected duration value of 1000")
	assert.Equal(t, "int", condition.Datatype, "Expected int datatype")

	// Check the rule has SampleRate: 1 (KeepAllSampler is translated to SampleRate: 1)
	assert.Equal(t, 1, rule.SampleRate, "Expected sample rate of 1")
}
