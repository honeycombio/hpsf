package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareStringEquals(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/string_value_condition_equals.yaml")

	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)

	assert.Len(t, receivers, 1)
	assert.Contains(t, receivers, "otlp/Receive_OTel_1")

	assert.Len(t, processors, 1)
	assert.Contains(t, processors, "usage")

	assert.Len(t, exporters, 1)
	assert.Contains(t, exporters, "otlphttp/Start_Sampling_1")

	assert.Len(t, rulesConfig.Samplers, 1)
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists)

	require.NotNil(t, defaultSampler.RulesBasedSampler)
	assert.Len(t, defaultSampler.RulesBasedSampler.Rules, 1)

	rule := defaultSampler.RulesBasedSampler.Rules[0]
	assert.Equal(t, "Sample_All_1", rule.Name)
	assert.Len(t, rule.Conditions, 1)
	assert.Equal(t, []string{"status"}, rule.Conditions[0].Fields)
	assert.Equal(t, "=", rule.Conditions[0].Operator)
	assert.Equal(t, "success", rule.Conditions[0].Value)
	assert.Equal(t, "string", rule.Conditions[0].Datatype)
}

func TestCompareStringNotEquals(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/string_value_condition_not_equals.yaml")

	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)

	assert.Len(t, receivers, 1)
	assert.Contains(t, receivers, "otlp/Receive_OTel_1")

	assert.Len(t, processors, 1)
	assert.Contains(t, processors, "usage")

	assert.Len(t, exporters, 1)
	assert.Contains(t, exporters, "otlphttp/Start_Sampling_1")

	assert.Len(t, rulesConfig.Samplers, 1)
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists)

	require.NotNil(t, defaultSampler.RulesBasedSampler)
	assert.Len(t, defaultSampler.RulesBasedSampler.Rules, 1)

	rule := defaultSampler.RulesBasedSampler.Rules[0]
	assert.Equal(t, "Sample_All_1", rule.Name)
	assert.Len(t, rule.Conditions, 1)
	assert.Equal(t, []string{"status"}, rule.Conditions[0].Fields)
	assert.Equal(t, "!=", rule.Conditions[0].Operator)
	assert.Equal(t, "error", rule.Conditions[0].Value)
	assert.Equal(t, "string", rule.Conditions[0].Datatype)
}

func TestCompareStringContains(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/string_value_condition_contains.yaml")

	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)

	assert.Len(t, receivers, 1)
	assert.Contains(t, receivers, "otlp/Receive_OTel_1")

	assert.Len(t, processors, 1)
	assert.Contains(t, processors, "usage")

	assert.Len(t, exporters, 1)
	assert.Contains(t, exporters, "otlphttp/Start_Sampling_1")

	assert.Len(t, rulesConfig.Samplers, 1)
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists)

	require.NotNil(t, defaultSampler.RulesBasedSampler)
	assert.Len(t, defaultSampler.RulesBasedSampler.Rules, 1)

	rule := defaultSampler.RulesBasedSampler.Rules[0]
	assert.Equal(t, "Sample_All_1", rule.Name)
	assert.Len(t, rule.Conditions, 1)
	assert.Equal(t, []string{"message"}, rule.Conditions[0].Fields)
	assert.Equal(t, "contains", rule.Conditions[0].Operator)
	assert.Equal(t, "timeout", rule.Conditions[0].Value)
	assert.Equal(t, "string", rule.Conditions[0].Datatype)
}

func TestCompareStringDoesNotContain(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/string_value_condition_does_not_contain.yaml")

	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)

	assert.Len(t, receivers, 1)
	assert.Contains(t, receivers, "otlp/Receive_OTel_1")

	assert.Len(t, processors, 1)
	assert.Contains(t, processors, "usage")

	assert.Len(t, exporters, 1)
	assert.Contains(t, exporters, "otlphttp/Start_Sampling_1")

	assert.Len(t, rulesConfig.Samplers, 1)
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists)

	require.NotNil(t, defaultSampler.RulesBasedSampler)
	assert.Len(t, defaultSampler.RulesBasedSampler.Rules, 1)

	rule := defaultSampler.RulesBasedSampler.Rules[0]
	assert.Equal(t, "Sample_All_1", rule.Name)
	assert.Len(t, rule.Conditions, 1)
	assert.Equal(t, []string{"message"}, rule.Conditions[0].Fields)
	assert.Equal(t, "does-not-contain", rule.Conditions[0].Operator)
	assert.Equal(t, "debug", rule.Conditions[0].Value)
	assert.Equal(t, "string", rule.Conditions[0].Datatype)
}

func TestCompareStringStartsWith(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/string_value_condition_starts_with.yaml")

	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)

	assert.Len(t, receivers, 1)
	assert.Contains(t, receivers, "otlp/Receive_OTel_1")

	assert.Len(t, processors, 1)
	assert.Contains(t, processors, "usage")

	assert.Len(t, exporters, 1)
	assert.Contains(t, exporters, "otlphttp/Start_Sampling_1")

	assert.Len(t, rulesConfig.Samplers, 1)
	defaultSampler, exists := rulesConfig.Samplers["__default__"]
	require.True(t, exists)

	require.NotNil(t, defaultSampler.RulesBasedSampler)
	assert.Len(t, defaultSampler.RulesBasedSampler.Rules, 1)

	rule := defaultSampler.RulesBasedSampler.Rules[0]
	assert.Equal(t, "Sample_All_1", rule.Name)
	assert.Len(t, rule.Conditions, 1)
	assert.Equal(t, []string{"endpoint"}, rule.Conditions[0].Fields)
	assert.Equal(t, "starts-with", rule.Conditions[0].Operator)
	assert.Equal(t, "/api/", rule.Conditions[0].Value)
	assert.Equal(t, "string", rule.Conditions[0].Datatype)
}
