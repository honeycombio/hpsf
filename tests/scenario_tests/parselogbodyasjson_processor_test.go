package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogBodyJSONParsingProcessorProcessor(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/parselogbodyasjson_processor_log_body.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	// Should only have logs pipeline since LogBodyJSONParsingProcessor only works with logs
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	assert.Len(t, logsPipelineNames, 1, "Expected 1 logs pipeline, got %v", logsPipelineNames)

	_, processors, _, getResult := collectorprovider.GetPipelineConfig(collectorConfig, logsPipelineNames[0].String())
	require.True(t, getResult.Found)
	assert.Contains(t, processors, "transform/json_parser_1")

	transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/json_parser_1")
	require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)

	// Should have log statements that parse log.body
	require.Len(t, transformConfig.LogStatements, 1)
	logStatement := transformConfig.LogStatements[0]

	assert.Len(t, logStatement.Conditions, 1)
	assert.Contains(t, logStatement.Conditions[0], "log.body")
	assert.Len(t, logStatement.Statements, 3)
	assert.Contains(t, logStatement.Statements[0], "log.body")
}

func TestLogBodyJSONParsingProcessorProcessorStandalone(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/parselogbodyasjson_processor_test.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	// Should only have logs pipeline since LogBodyJSONParsingProcessor only works with logs
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	assert.Len(t, logsPipelineNames, 1, "Expected 1 logs pipeline, got %v", logsPipelineNames)

	_, processors, _, getResult := collectorprovider.GetPipelineConfig(collectorConfig, logsPipelineNames[0].String())
	require.True(t, getResult.Found)
	assert.Contains(t, processors, "transform/parse_log_body_1")

	transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/parse_log_body_1")
	require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)

	// Should have log statements that parse log.body
	require.Len(t, transformConfig.LogStatements, 1)
	logStatement := transformConfig.LogStatements[0]

	assert.Len(t, logStatement.Conditions, 1)
	assert.Contains(t, logStatement.Conditions[0], "log.body")
	assert.Len(t, logStatement.Statements, 3)
	assert.Contains(t, logStatement.Statements[0], "log.body")
}
