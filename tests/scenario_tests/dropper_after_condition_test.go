package hpsftests

import (
	"testing"
	"time"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDropperAfterCondition(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/dropper_after_condition.yaml")

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

	// Check the first rule is a Drop with a condition
	rule1 := defaultSampler.RulesBasedSampler.Rules[0]
	assert.Equal(t, "Drop_1", rule1.Name, "Expected rule 1 name to match component name")
	assert.Len(t, rule1.Conditions, 1, "Expected 1 condition in rule 1")
	assert.Len(t, rule1.Conditions[0].Fields, 1, "Expected 1 Fields in condition 1")
	assert.Equal(t, "error", rule1.Conditions[0].Fields[0], "Expected error field in rule 1")
	assert.Equal(t, "exists", rule1.Conditions[0].Operator, "Expected exists operator in rule 1")
	assert.True(t, rule1.Drop, "Expected this rule to drop")

	// Check the second rule is an EMAThroughputSampler
	rule2 := defaultSampler.RulesBasedSampler.Rules[1]
	assert.Nil(t, rule2.Conditions, "Expected no conditions in rule 2")
	assert.NotNil(t, rule2.Sampler, "Expected sampler configuration in rule 2")
	assert.NotNil(t, rule2.Sampler.EMAThroughputSampler, "Expected EMAThroughputSampler in rule 2")
	assert.Equal(t, 200, rule2.Sampler.EMAThroughputSampler.GoalThroughputPerSec, "Expected GoalThroughputPerSec in rule 2")
	assert.Equal(t, time.Duration(60)*time.Second, time.Duration(rule2.Sampler.EMAThroughputSampler.AdjustmentInterval), "Expected AdjustmentInterval in rule 2")
	assert.ElementsMatch(t, []string{"http.method", "http.status_code"}, rule2.Sampler.EMAThroughputSampler.FieldList, "Expected FieldList in rule 2")
}
