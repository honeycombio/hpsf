package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTwoSamplingPipes(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/twoSamplingPipes.yaml")

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

	require.NotNil(t, defaultSampler.RulesBasedSampler, "Expected RulesBasedSampler configuration")
	assert.Len(t, defaultSampler.RulesBasedSampler.Rules, 2, "Expected 2 rules in the sampler")

	// Rule 1: http.status_code >= 400, SampleRate: 100
	rule1 := defaultSampler.RulesBasedSampler.Rules[0]
	assert.Len(t, rule1.Conditions, 1, "Expected 1 condition in the first rule")
	cond1 := rule1.Conditions[0]
	assert.ElementsMatch(t, []string{"http.status_code", "http.response.status_code"}, cond1.Fields, "Expected http status fields")
	assert.Equal(t, ">=", cond1.Operator, "Expected >= operator")
	assert.Equal(t, 400, cond1.Value, "Expected value 400")
	assert.Equal(t, "int", cond1.Datatype, "Expected int datatype")
	assert.Equal(t, 100, rule1.SampleRate, "Expected sample rate of 100")

	// Rule 2: error exists AND duration_ms >= 1000, SampleRate: 1
	rule2 := defaultSampler.RulesBasedSampler.Rules[1]
	assert.Len(t, rule2.Conditions, 2, "Expected 2 conditions in the second rule")
	cond2a := rule2.Conditions[0]
	assert.ElementsMatch(t, []string{"error"}, cond2a.Fields, "Expected error field")
	assert.Equal(t, "exists", cond2a.Operator, "Expected exists operator")
	assert.Nil(t, cond2a.Value, "Expected value nil for exists operator")
	cond2b := rule2.Conditions[1]
	assert.Equal(t, "duration_ms", cond2b.Field, "Expected duration_ms field")
	assert.Equal(t, ">=", cond2b.Operator, "Expected >= operator")
	assert.Equal(t, 1000, cond2b.Value, "Expected value 1000")
	assert.Equal(t, "int", cond2b.Datatype, "Expected int datatype")
	assert.Equal(t, 1, rule2.SampleRate, "Expected sample rate of 1")
}
