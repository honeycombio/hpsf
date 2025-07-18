package hpsftests

import (
	"testing"
	"time"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSampleDefaultDeterministic(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/sample_default_deterministic.yaml")

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

	// Verify Refinery rules configuration
	assert.Len(t, rulesConfig.Samplers, 1, "Expected 1 sampler in refinery config")

	// Check that the __default__ environment has a sampler
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists, "Expected __default__ sampler to exist")

	// Verify the sampler is a RulesBasedSampler with two rules
	require.NotNil(t, defaultSampler.RulesBasedSampler, "Expected RulesBasedSampler configuration")
	assert.Len(t, defaultSampler.RulesBasedSampler.Rules, 2, "Expected 2 rules in the sampler")

	// Check the first rule is a EMAThroughputSampler with a condition
	rule1 := defaultSampler.RulesBasedSampler.Rules[0]
	assert.Equal(t, "Sample_by_Events_per_Second_1", rule1.Name, "Expected rule 1 name to match component name")
	assert.Len(t, rule1.Conditions, 1, "Expected 1 condition in rule 1")
	assert.Equal(t, "duration_ms", rule1.Conditions[0].Field, "Expected duration_ms field in rule 1")
	assert.Equal(t, ">=", rule1.Conditions[0].Operator, "Expected >= operator in rule 1")
	assert.Equal(t, 1000, rule1.Conditions[0].Value, "Expected value 1000 in rule 1")
	assert.Equal(t, "int", rule1.Conditions[0].Datatype, "Expected int datatype in rule 1")
	require.NotNil(t, rule1.Sampler.EMAThroughputSampler, "Expected EMAThroughputSampler in rule 2")
	assert.Equal(t, 200, rule1.Sampler.EMAThroughputSampler.GoalThroughputPerSec, "Expected GoalThroughputPerSec in rule 2")
	assert.Equal(t, time.Duration(60)*time.Second, time.Duration(rule1.Sampler.EMAThroughputSampler.AdjustmentInterval), "Expected AdjustmentInterval in rule 2")
	assert.ElementsMatch(t, []string{"http.method", "http.status_code"}, rule1.Sampler.EMAThroughputSampler.FieldList, "Expected FieldList in rule 2")

	// Check the second rule is a DeterministicSampler without condition
	rule2 := defaultSampler.RulesBasedSampler.Rules[1]
	assert.Equal(t, "Sample_at_a_Fixed_Rate_1", rule2.Name, "Expected rule 2 name to match component name")
	assert.Len(t, rule2.Conditions, 0, "Expected 0 conditions in rule 3")
	assert.Nil(t, rule2.Sampler, "Expected no sampler in rule 2")
	assert.Equal(t, 100, rule2.SampleRate, "Expected SampleRate of 100 in rule 2")
}
