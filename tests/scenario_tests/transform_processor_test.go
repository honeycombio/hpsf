package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformProcessorDefaults(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/transformprocessor_defaults.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	// Check that all three pipeline types are created
	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1, "Expected 1 traces pipeline, got %v", tracesPipelineNames)

	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	assert.Len(t, logsPipelineNames, 1, "Expected 1 logs pipeline, got %v", logsPipelineNames)

	metricsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "metrics")
	assert.Len(t, metricsPipelineNames, 1, "Expected 1 metrics pipeline, got %v", metricsPipelineNames)

	// Check that transform processor is in all pipelines
	_, processors, _, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)
	assert.Contains(t, processors, "transform/transform_1")

	_, processors, _, getResult = collectorprovider.GetPipelineConfig(collectorConfig, logsPipelineNames[0].String())
	require.True(t, getResult.Found)
	assert.Contains(t, processors, "transform/transform_1")

	_, processors, _, getResult = collectorprovider.GetPipelineConfig(collectorConfig, metricsPipelineNames[0].String())
	require.True(t, getResult.Found)
	assert.Contains(t, processors, "transform/transform_1")

	// Check transform processor configuration
	transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/transform_1")
	require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)

	// With default/empty arrays, should have no statements
	assert.Len(t, transformConfig.TraceStatements, 0)
	assert.Len(t, transformConfig.LogStatements, 0)
	assert.Len(t, transformConfig.MetricStatements, 0)

	// Check error mode
	assert.Equal(t, "ignore", string(transformConfig.ErrorMode))
}

func TestTransformProcessorWithStatements(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/transformprocessor_with_statements.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/transform_1")
	require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)

	// Check trace statements - they are wrapped in statement objects
	require.Len(t, transformConfig.TraceStatements, 1)
	traceStatement := transformConfig.TraceStatements[0]
	require.Len(t, traceStatement.Statements, 3)
	assert.Equal(t, `set(span.attributes["processed"], "true")`, traceStatement.Statements[0])
	assert.Equal(t, `set(span.attributes["processor"], "transform")`, traceStatement.Statements[1])
	assert.Equal(t, `delete_key(span.attributes, "temp_field")`, traceStatement.Statements[2])

	// Check log statements - they are wrapped in statement objects
	require.Len(t, transformConfig.LogStatements, 1)
	logStatement := transformConfig.LogStatements[0]
	require.Len(t, logStatement.Statements, 2)
	assert.Equal(t, `set(log.attributes["processed"], "true")`, logStatement.Statements[0])
	assert.Equal(t, `set(log.severity_text, "INFO")`, logStatement.Statements[1])

	// Check metric statements - they are wrapped in statement objects
	require.Len(t, transformConfig.MetricStatements, 1)
	metricStatement := transformConfig.MetricStatements[0]
	require.Len(t, metricStatement.Statements, 2)
	assert.Equal(t, `set(metric.attributes["processed"], "true")`, metricStatement.Statements[0])
	assert.Equal(t, `set(metric.description, Concat([metric.description, " (processed)"], ""))`, metricStatement.Statements[1])

	// Check error mode
	assert.Equal(t, "ignore", string(transformConfig.ErrorMode))
}

func TestTransformProcessorSingleSignal(t *testing.T) {
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/transformprocessor_single_signal.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	// Should only have traces pipeline since only trace statements are defined
	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1, "Expected 1 traces pipeline, got %v", tracesPipelineNames)

	// Should not have logs or metrics pipelines
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	assert.Len(t, logsPipelineNames, 0, "Expected 0 logs pipelines, got %v", logsPipelineNames)

	metricsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "metrics")
	assert.Len(t, metricsPipelineNames, 0, "Expected 0 metrics pipelines, got %v", metricsPipelineNames)

	transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/transform_1")
	require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)

	// Check trace statements - they are wrapped in statement objects
	require.Len(t, transformConfig.TraceStatements, 1)
	traceStatement := transformConfig.TraceStatements[0]
	require.Len(t, traceStatement.Statements, 2)
	assert.Equal(t, `set(span.attributes["trace_only"], "true")`, traceStatement.Statements[0])
	assert.Equal(t, `set(span.name, Concat([span.name, " [transformed]"], ""))`, traceStatement.Statements[1])

	// Should have no log or metric statements
	assert.Len(t, transformConfig.LogStatements, 0)
	assert.Len(t, transformConfig.MetricStatements, 0)

	// Check error mode
	assert.Equal(t, "propagate", string(transformConfig.ErrorMode))
}

func TestTransformProcessorErrorModeValidation(t *testing.T) {
	// Test that all error modes are supported
	errorModes := []string{"ignore", "silent", "propagate"}
	
	for _, errorMode := range errorModes {
		t.Run("ErrorMode_"+errorMode, func(t *testing.T) {
			// Create a temporary test config
			testConfig := `
name: transformprocessor_error_mode_test
version: v0.1.0
summary: Test for TransformProcessor error mode validation

components:
  - name: OTel Receiver 1
    kind: OTelReceiver
  - name: transform_1
    kind: TransformProcessor
    properties:
      - name: ErrorMode
        value: ` + errorMode + `
      - name: TraceStatements
        value:
          - 'set(span.attributes["test"], "true")'
  - name: OTel HTTP Exporter 1
    kind: OTelHTTPExporter

connections:
  - source:
      component: OTel Receiver 1
      port: Traces
      type: OTelTraces
    destination:
      component: transform_1
      port: Traces
      type: OTelTraces
  - source:
      component: transform_1
      port: Traces
      type: OTelTraces
    destination:
      component: OTel HTTP Exporter 1
      port: Traces
      type: OTelTraces`

			rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigs(t, testConfig)
			assert.Len(t, rulesConfig.Samplers, 1)

			transformConfig, findResult := collectorprovider.GetProcessorConfig[transformprocessor.Config](collectorConfig, "transform/transform_1")
			require.True(t, findResult.Found, "Expected transform processor to be found, found (%v)", findResult.Components)

			assert.Equal(t, errorMode, string(transformConfig.ErrorMode))
		})
	}
}