package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListComparisonIn(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/list_comparison_condition_in.yaml")

	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)

	assert.Len(t, receivers, 1)
	assert.Contains(t, receivers, "otlp/Receive_OTel_1")

	assert.Len(t, processors, 2)
	assert.Contains(t, processors, "memory_limiter/Receive_OTel_1")
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
	assert.Equal(t, "trace", rule.Scope)
	assert.Len(t, rule.Conditions, 1)

	condition := rule.Conditions[0]
	assert.Equal(t, []string{"status", "http.status_code"}, condition.Fields)
	assert.Equal(t, "in", condition.Operator)
	assert.Equal(t, []interface{}{"200", "201", "204"}, condition.Value)
	assert.Equal(t, "string", condition.Datatype)
}

func TestListComparisonNotIn(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/list_comparison_condition_not_in.yaml")

	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)

	assert.Len(t, receivers, 1)
	assert.Contains(t, receivers, "otlp/Receive_OTel_1")

	assert.Len(t, processors, 2)
	assert.Contains(t, processors, "memory_limiter/Receive_OTel_1")
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
	assert.Equal(t, "span", rule.Scope)
	assert.Len(t, rule.Conditions, 1)

	condition := rule.Conditions[0]
	assert.Equal(t, []string{"error.type"}, condition.Fields)
	assert.Equal(t, "not-in", condition.Operator)
	assert.Equal(t, []interface{}{"timeout", "connection_refused", "dns_error"}, condition.Value)
	assert.Equal(t, "string", condition.Datatype)
}
