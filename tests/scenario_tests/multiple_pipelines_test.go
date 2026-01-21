package hpsftests

import (
	"strings"
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pipeline"
)

// isIngressPipeline checks if a pipeline name indicates an ingress pipeline
func isIngressPipeline(pipelineID pipeline.ID) bool {
	return strings.Contains(pipelineID.String(), "/ingress_")
}

// getDownstreamPipelines returns only non-ingress pipelines
func getDownstreamPipelines(pipelineNames []pipeline.ID) []pipeline.ID {
	var downstream []pipeline.ID
	for _, name := range pipelineNames {
		if !isIngressPipeline(name) {
			downstream = append(downstream, name)
		}
	}
	return downstream
}

func TestMultiplePipelines(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/multiple_pipelines.yaml")

	// Get all logs pipelines (includes ingress + downstream)
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")

	// With shared receiver, we expect: 2 downstream + 1 ingress = 3 total
	assert.Len(t, logsPipelineNames, 3, "Expected 3 logs pipelines (2 downstream + 1 ingress), got %v", logsPipelineNames)

	// Verify only downstream pipelines
	downstreamPipelines := getDownstreamPipelines(logsPipelineNames)
	assert.Len(t, downstreamPipelines, 2, "Expected 2 downstream logs pipelines")

	firstLogsPipeline := collectorConfig.Service.Pipelines[downstreamPipelines[0]]
	assert.Len(t, firstLogsPipeline.Exporters, 1, "Expected 1 exporter in pipeline")

	secondLogsPipeline := collectorConfig.Service.Pipelines[downstreamPipelines[1]]
	assert.Len(t, secondLogsPipeline.Exporters, 1, "Expected 1 exporter in pipeline")
	assert.NotEqual(t, secondLogsPipeline.Exporters[0], firstLogsPipeline.Exporters[0], "Expected different exporters in pipelines")

	// Architecture:
	// logs/ingress_otlp/OTel_Receiver_1:
	//   receivers: [otlp/OTel_Receiver_1]
	//   processors: [memory_limiter, usage]
	//   exporters: [forward/otlp/OTel_Receiver_1]
	// logs/downstream1:
	//   receivers: [forward/otlp/OTel_Receiver_1]
	//   processors: [filter/Filter_Logs_by_Severity_1]
	//   exporters: [otlphttp/Honeycomb_Exporter_1]
	// logs/downstream2:
	//   receivers: [forward/otlp/OTel_Receiver_1]
	//   processors: []
	//   exporters: [awss3/Send_to_S3_Archive_1]
}

func TestMultiplePipelinesMultipleExporters(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/multiple_pipelines_multiple_exporters.yaml")

	usageProcessor := component.MustNewID("usage")
	memoryLimiterProcessor := component.MustNewIDWithName("memory_limiter", "OTel_Receiver_1")
	filterProcessor := component.MustNewIDWithName("filter", "Filter_Logs_by_Severity_1")
	otelReceiver := component.MustNewIDWithName("otlp", "OTel_Receiver_1")
	forwardConnector := component.MustNewIDWithName("forward", "otlp/OTel_Receiver_1")
	honeycombExporter := component.MustNewIDWithName("otlphttp", "Honeycomb_Exporter_1")
	otlpExporter := component.MustNewIDWithName("otlphttp", "Send_to_OTLP")
	s3Exporter := component.MustNewIDWithName("awss3", "Send_to_S3_Archive_1")

	// Get all logs pipelines
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")

	// With shared receiver: 3 downstream + 1 ingress = 4 total
	assert.Len(t, logsPipelineNames, 4, "Expected 4 logs pipelines (3 downstream + 1 ingress), got %v", logsPipelineNames)

	for _, pipelineName := range logsPipelineNames {
		pipeline := collectorConfig.Service.Pipelines[pipelineName]

		if isIngressPipeline(pipelineName) {
			// Ingress pipeline: receiver → memory_limiter → usage → forward connector
			assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver in ingress pipeline")
			assert.Contains(t, pipeline.Receivers, otelReceiver, "Expected OTel receiver in ingress")
			assert.Len(t, pipeline.Processors, 2, "Expected 2 processors in ingress pipeline")
			assert.Contains(t, pipeline.Processors, usageProcessor, "Expected usage processor in ingress")
			assert.Contains(t, pipeline.Processors, memoryLimiterProcessor, "Expected memory_limiter in ingress")
			assert.Len(t, pipeline.Exporters, 1, "Expected 1 exporter (forward connector) in ingress")
			assert.Contains(t, pipeline.Exporters, forwardConnector, "Expected forward connector in ingress")
		} else {
			// Downstream pipeline: forward connector → custom processors → exporter
			assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver (forward connector) in downstream")
			assert.Contains(t, pipeline.Receivers, forwardConnector, "Expected forward connector as receiver")

			for _, exporter := range pipeline.Exporters {
				if exporter == honeycombExporter || exporter == otlpExporter {
					// Pipelines with filter processor
					assert.Len(t, pipeline.Processors, 1, "Expected 1 custom processor in downstream pipeline")
					assert.Contains(t, pipeline.Processors, filterProcessor, "Expected filter processor")
				} else if exporter == s3Exporter {
					// Pipeline with no custom processors
					assert.Len(t, pipeline.Processors, 0, "Expected no custom processors in S3 pipeline")
				} else {
					t.Errorf("Unexpected exporter %s in pipeline %s", exporter.String(), pipelineName.String())
				}
			}
		}
	}

	// Architecture:
	// logs/ingress_otlp/OTel_Receiver_1:
	//   receivers: [otlp/OTel_Receiver_1]
	//   processors: [memory_limiter, usage]
	//   exporters: [forward/otlp/OTel_Receiver_1]
	// logs/downstream1:
	//   receivers: [forward/otlp/OTel_Receiver_1]
	//   processors: [filter/Filter_Logs_by_Severity_1]
	//   exporters: [otlphttp/Honeycomb_Exporter_1]
	// logs/downstream2:
	//   receivers: [forward/otlp/OTel_Receiver_1]
	//   processors: []
	//   exporters: [awss3/Send_to_S3_Archive_1]
	// logs/downstream3:
	//   receivers: [forward/otlp/OTel_Receiver_1]
	//   processors: [filter/Filter_Logs_by_Severity_1]
	//   exporters: [otlphttp/Send_to_OTLP]
}

func TestMultiplePipelinesSubProcessors(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/multiple_pipelines_sub_processors.yaml")

	usageProcessor := component.MustNewID("usage")
	memoryLimiterProcessor := component.MustNewIDWithName("memory_limiter", "OTel_Receiver_1")
	filterProcessor := component.MustNewIDWithName("filter", "Filter_Logs_by_Severity_1")
	transformProcessor := component.MustNewIDWithName("transform", "Parse_Log_Body_As_JSON_1")
	otelReceiver := component.MustNewIDWithName("otlp", "OTel_Receiver_1")
	forwardConnector := component.MustNewIDWithName("forward", "otlp/OTel_Receiver_1")
	honeycombExporter := component.MustNewIDWithName("otlphttp", "Honeycomb_Exporter_1")
	otlpExporter := component.MustNewIDWithName("otlphttp", "Send_to_OTLP")
	s3Exporter := component.MustNewIDWithName("awss3", "Send_to_S3_Archive_1")

	// Get all logs pipelines
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")

	for _, pipelineName := range logsPipelineNames {
		pipeline := collectorConfig.Service.Pipelines[pipelineName]

		if isIngressPipeline(pipelineName) {
			// Ingress pipeline validation
			assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver in ingress pipeline")
			assert.Contains(t, pipeline.Receivers, otelReceiver, "Expected OTel receiver in ingress")
			assert.Len(t, pipeline.Processors, 2, "Expected 2 processors in ingress pipeline")
			assert.Contains(t, pipeline.Processors, usageProcessor, "Expected usage processor in ingress")
			assert.Contains(t, pipeline.Processors, memoryLimiterProcessor, "Expected memory_limiter in ingress")
		} else {
			// Downstream pipeline validation
			assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver (forward connector)")
			assert.Contains(t, pipeline.Receivers, forwardConnector, "Expected forward connector as receiver")

			for _, exporter := range pipeline.Exporters {
				if exporter == otlpExporter {
					assert.Len(t, pipeline.Processors, 2, "Expected 2 custom processors")
					assert.Contains(t, pipeline.Processors, filterProcessor, "Expected filter processor")
					assert.Contains(t, pipeline.Processors, transformProcessor, "Expected transform processor")
				} else if exporter == s3Exporter {
					assert.Len(t, pipeline.Processors, 0, "Expected no custom processors")
				} else if exporter == honeycombExporter {
					assert.Len(t, pipeline.Processors, 1, "Expected 1 custom processor")
					assert.Contains(t, pipeline.Processors, filterProcessor, "Expected filter processor")
				} else {
					t.Errorf("Unexpected exporter %s in pipeline %s", exporter.String(), pipelineName.String())
				}
			}
		}
	}

	// Architecture:
	// logs/ingress_otlp/OTel_Receiver_1:
	//   receivers: [otlp/OTel_Receiver_1]
	//   processors: [memory_limiter, usage]
	//   exporters: [forward/otlp/OTel_Receiver_1]
	// logs/downstream1:
	//   receivers: [forward/otlp/OTel_Receiver_1]
	//   processors: [filter/Filter_Logs_by_Severity_1, transform/Parse_Log_Body_As_JSON_1]
	//   exporters: [otlphttp/Send_to_OTLP]
	// logs/downstream2:
	//   receivers: [forward/otlp/OTel_Receiver_1]
	//   processors: []
	//   exporters: [awss3/Send_to_S3_Archive_1]
	// logs/downstream3:
	//   receivers: [forward/otlp/OTel_Receiver_1]
	//   processors: [filter/Filter_Logs_by_Severity_1]
	//   exporters: [otlphttp/Honeycomb_Exporter_1]
}

func TestMultiplePipelinesSingleExporter(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/multiple_pipelines_single_exporter.yaml")

	usageProcessor := component.MustNewID("usage")
	memoryLimiterProcessor := component.MustNewIDWithName("memory_limiter", "OTel_Receiver_1")
	filterProcessor := component.MustNewIDWithName("filter", "Info_Logs_only")
	otelReceiver := component.MustNewIDWithName("otlp", "OTel_Receiver_1")
	forwardConnector := component.MustNewIDWithName("forward", "otlp/OTel_Receiver_1")
	honeycombExporter := component.MustNewIDWithName("otlphttp", "Honeycomb_Exporter_1")

	// Get all logs pipelines
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")

	// With shared receiver: 2 downstream + 1 ingress = 3 total
	assert.Len(t, logsPipelineNames, 3, "Expected 3 logs pipelines (2 downstream + 1 ingress), got %v", logsPipelineNames)

	// Verify downstream pipelines
	downstreamPipelines := getDownstreamPipelines(logsPipelineNames)
	assert.Len(t, downstreamPipelines, 2, "Expected 2 downstream pipelines")

	// Check that downstream pipelines have different number of processors
	pipeline1Procs := len(collectorConfig.Service.Pipelines[downstreamPipelines[0]].Processors)
	pipeline2Procs := len(collectorConfig.Service.Pipelines[downstreamPipelines[1]].Processors)
	assert.NotEqual(t, pipeline1Procs, pipeline2Procs, "Expected pipelines to have different number of processors")

	for _, pipelineName := range logsPipelineNames {
		pipeline := collectorConfig.Service.Pipelines[pipelineName]

		if isIngressPipeline(pipelineName) {
			// Ingress pipeline validation
			assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver in ingress")
			assert.Contains(t, pipeline.Receivers, otelReceiver, "Expected OTel receiver")
			assert.Len(t, pipeline.Processors, 2, "Expected 2 processors in ingress")
			assert.Contains(t, pipeline.Processors, usageProcessor, "Expected usage processor")
			assert.Contains(t, pipeline.Processors, memoryLimiterProcessor, "Expected memory_limiter")
			assert.Len(t, pipeline.Exporters, 1, "Expected 1 exporter in ingress")
			assert.Contains(t, pipeline.Exporters, forwardConnector, "Expected forward connector")
		} else {
			// Downstream pipeline validation
			assert.Len(t, pipeline.Exporters, 1, "Expected 1 exporter in downstream")
			assert.Equal(t, pipeline.Exporters[0], honeycombExporter, "Expected Honeycomb exporter")

			assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver in downstream")
			assert.Contains(t, pipeline.Receivers, forwardConnector, "Expected forward connector")

			// One pipeline has filter, one doesn't
			if len(pipeline.Processors) == 0 {
				assert.Len(t, pipeline.Processors, 0, "Expected no custom processors")
			} else {
				assert.Len(t, pipeline.Processors, 1, "Expected 1 custom processor")
				assert.Contains(t, pipeline.Processors, filterProcessor, "Expected filter processor")
			}
		}
	}

	// Architecture:
	// logs/ingress_otlp/OTel_Receiver_1:
	//   receivers: [otlp/OTel_Receiver_1]
	//   processors: [memory_limiter, usage]
	//   exporters: [forward/otlp/OTel_Receiver_1]
	// logs/downstream1:
	//   receivers: [forward/otlp/OTel_Receiver_1]
	//   processors: [filter/Info_Logs_only]
	//   exporters: [otlphttp/Honeycomb_Exporter_1]
	// logs/downstream2:
	//   receivers: [forward/otlp/OTel_Receiver_1]
	//   processors: []
	//   exporters: [otlphttp/Honeycomb_Exporter_1]
}
