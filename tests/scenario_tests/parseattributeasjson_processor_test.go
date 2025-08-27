package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttributeJSONParsingProcessorDefaults(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/parseattributeasjson_processor_defaults.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1, "Expected 1 traces pipeline, got %v", tracesPipelineNames)

	_, processors, _, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)
	assert.Contains(t, processors, "transform/json_parser_1")

	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	assert.Len(t, logsPipelineNames, 1, "Expected 1 logs pipeline, got %v", logsPipelineNames)

	_, processors, _, getResult = collectorprovider.GetPipelineConfig(collectorConfig, logsPipelineNames[0].String())
	require.True(t, getResult.Found)
	assert.Contains(t, processors, "transform/json_parser_1")

	transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/json_parser_1")
	require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)

	// Default signal is "log", so should have log statements
	require.Len(t, transformConfig.LogStatements, 1)
	logStatement := transformConfig.LogStatements[0]

	assert.Len(t, logStatement.Conditions, 1)
	assert.Len(t, logStatement.Statements, 3)
}

func TestAttributeJSONParsingProcessorCustom(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/parseattributeasjson_processor_custom.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/json_parser_1")
	require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)

	// Custom signal is "logs", so should have log statements
	require.Len(t, transformConfig.LogStatements, 1)
	logStatement := transformConfig.LogStatements[0]

	assert.Len(t, logStatement.Conditions, 1)
	assert.Len(t, logStatement.Statements, 3)
}

func TestAttributeJSONParsingProcessorSpanSignal(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/parseattributeasjson_processor_span.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/json_parser_1")
	require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)

	// Signal is "span" with custom field, so should have trace statements
	require.Len(t, transformConfig.TraceStatements, 1)
	traceStatement := transformConfig.TraceStatements[0]

	assert.Len(t, traceStatement.Conditions, 1)
	assert.Len(t, traceStatement.Statements, 3)
}
