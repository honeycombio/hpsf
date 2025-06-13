package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
)

func TestMultiplePipelines(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/multiple_pipelines.yaml")

	//verify that there are 2 logs pipelines
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	assert.Len(t, logsPipelineNames, 2, "Expected 2 logs pipelines, got %v", logsPipelineNames)

	firstLogsPipeline := collectorConfig.Service.Pipelines[logsPipelineNames[0]]
	assert.Len(t, firstLogsPipeline.Exporters, 1, "Expected 1 exporter in pipeline")

	secondLogsPipeline := collectorConfig.Service.Pipelines[logsPipelineNames[1]]
	assert.Len(t, secondLogsPipeline.Exporters, 1, "Expected 1 exporter in pipeline")
	assert.NotEqual(t, secondLogsPipeline.Exporters[0], firstLogsPipeline.Exporters[0], "Expected different exporters in pipelines")

	//  pipelines:
	// 	  logs:
	// 	    receivers: [otlp/OTel_Receiver_1]
	// 	    processors: [usage, filter/Filter_Logs_by_Severity_1]
	// 	    exporters: [otlphttp/Honeycomb_Exporter_1]
	// 	  logs/1:
	// 	    receivers: [otlp/OTel_Receiver_1]
	//      processors: [usage]
	// 	    exporters: [awss3/Send_to_S3_Archive_1]

}

func TestMultiplePipelinesMultipleExporters(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/multiple_pipelines_multiple_exporters.yaml")

	usageProcessor := component.MustNewID("usage")
	filterProcessor := component.MustNewIDWithName("filter", "Filter_Logs_by_Severity_1")
	otelReceiver := component.MustNewIDWithName("otlp", "OTel_Receiver_1")
	honeycombExporter := component.MustNewIDWithName("otlphttp", "Honeycomb_Exporter_1")
	otlpExporter := component.MustNewIDWithName("otlphttp", "Send_to_OTLP")
	s3Exporter := component.MustNewIDWithName("awss3", "Send_to_S3_Archive_1")

	//verify that there are 2 logs pipelines
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	assert.Len(t, logsPipelineNames, 3, "Expected 2 logs pipelines, got %v", logsPipelineNames)

	for _, pipelineName := range logsPipelineNames {
		pipeline := collectorConfig.Service.Pipelines[pipelineName]

		for _, exporter := range pipeline.Exporters {
			if exporter == honeycombExporter || exporter == otlpExporter {
				assert.Len(t, pipeline.Processors, 2, "Expected 2 processors in the pipeline for %s but got %s", exporter.String(), pipeline.Processors)
				assert.Contains(t, pipeline.Processors, usageProcessor, "Expected usage processor")
				assert.Contains(t, pipeline.Processors, filterProcessor, "Expected filter processor")

				assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver  in the pipeline for %s but got %s", exporter.String(), pipeline.Receivers)
				assert.Contains(t, pipeline.Receivers, otelReceiver, "Expected OTel receiver")
			} else if exporter == s3Exporter {
				assert.Len(t, pipeline.Processors, 1, "Expected 1 processor in pipeline got %s", pipeline.Processors)
				assert.Contains(t, pipeline.Processors, usageProcessor, "Expected usage processor")
			} else {
				t.Errorf("Unexpected exporter %s in pipeline %s", exporter.String(), pipelineName.String())
			}
		}
	}

	//  pipelines:
	//    logs:
	//       receivers: [otlp/OTel_Receiver_1]
	//       processors: [usage, filter/Filter_Logs_by_Severity_1]
	//       exporters: [otlphttp/Honeycomb_Exporter_1]
	//    logs/1:
	//      receivers: [otlp/OTel_Receiver_1]
	//      processors: [usage]
	//     e xporters: [awss3/Send_to_S3_Archive_1]
	//    logs/2:
	//      receivers: [otlp/OTel_Receiver_1]
	//      processors: [usage, filter/Filter_Logs_by_Severity_1]
	//     e xporters: [otlphttp/Send_to_OTLP]

}

func TestMultiplePipelinesSubProcessors(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/multiple_pipelines_sub_processors.yaml")

	usageProcessor := component.MustNewID("usage")
	filterProcessor := component.MustNewIDWithName("filter", "Filter_Logs_by_Severity_1")
	transformProcessor := component.MustNewIDWithName("transform", "Parse_Log_Body_As_JSON_1")
	otelReceiver := component.MustNewIDWithName("otlp", "OTel_Receiver_1")
	honeycombExporter := component.MustNewIDWithName("otlphttp", "Honeycomb_Exporter_1")
	otlpExporter := component.MustNewIDWithName("otlphttp", "Send_to_OTLP")
	s3Exporter := component.MustNewIDWithName("awss3", "Send_to_S3_Archive_1")

	//verify that there are 2 logs pipelines
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")

	for _, pipelineName := range logsPipelineNames {
		pipeline := collectorConfig.Service.Pipelines[pipelineName]

		for _, exporter := range pipeline.Exporters {
			if exporter == otlpExporter {
				assert.Len(t, pipeline.Processors, 3, "Expected 3 processors in pipeline got %s", pipeline.Processors)
				assert.Contains(t, pipeline.Processors, usageProcessor, "Expected usage processor")
				assert.Contains(t, pipeline.Processors, filterProcessor, "Expected filter processor")
				assert.Contains(t, pipeline.Processors, transformProcessor, "Expected transform processor")

				assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver in pipeline got %s", pipeline.Receivers)
				assert.Contains(t, pipeline.Receivers, otelReceiver, "Expected OTel receiver")

			} else if exporter == s3Exporter {
				assert.Len(t, pipeline.Processors, 1, "Expected 1 processor in pipeline got %s", pipeline.Processors)
				assert.Contains(t, pipeline.Processors, usageProcessor, "Expected usage processor")

				assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver in pipeline got %s", pipeline.Receivers)
				assert.Contains(t, pipeline.Receivers, otelReceiver, "Expected OTel receiver")

			} else if exporter == honeycombExporter {
				assert.Len(t, pipeline.Processors, 2, "Expected 2 processors in pipeline got %s", pipeline.Processors)
				assert.Contains(t, pipeline.Processors, usageProcessor, "Expected usage processor")
				assert.Contains(t, pipeline.Processors, filterProcessor, "Expected filter processor")

				assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver in pipeline got %s", pipeline.Receivers)
				assert.Contains(t, pipeline.Receivers, otelReceiver, "Expected OTel receiver")
			} else {
				t.Errorf("Unexpected exporter %s in pipeline %s", exporter.String(), pipelineName.String())
			}
		}
	}

	// pipelines:
	//    logs:
	// 	    receivers: [otlp/OTel_Receiver_1]
	// 	    processors: [usage, filter/Filter_Logs_by_Severity_1, transform/Parse_Log_Body_As_JSON_1]
	// 	    exporters: [otlphttp/Send_to_OTLP]
	// 	  logs/1:
	// 	    receivers: [otlp/OTel_Receiver_1]
	// 	    processors: [usage]
	// 	    exporters: [awss3/Send_to_S3_Archive_1]
	// 	  logs/2:
	// 	    receivers: [otlp/OTel_Receiver_1]
	// 	    processors: [usage, filter/Filter_Logs_by_Severity_1]
	// 	   exporters: [otlphttp/Honeycomb_Exporter_1]

}

func TestMultiplePipelinesSingleExporter(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/multiple_pipelines_single_exporter.yaml")

	usageProcessor := component.MustNewID("usage")
	filterProcessor := component.MustNewIDWithName("filter", "Info_Logs_only")
	otelReceiver := component.MustNewIDWithName("otlp", "OTel_Receiver_1")
	honeycombExporter := component.MustNewIDWithName("otlphttp", "Honeycomb_Exporter_1")

	//verify that there are 2 logs pipelines
	logsPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "logs")
	assert.Len(t, logsPipelineNames, 2, "Expected 2 logs pipelines, got %v", logsPipelineNames)

	if len(collectorConfig.Service.Pipelines[logsPipelineNames[0]].Processors) ==
		len(collectorConfig.Service.Pipelines[logsPipelineNames[1]].Processors) {
		t.Errorf("Expected pipelines to have different number of processors")
	}

	for _, pipelineName := range logsPipelineNames {
		pipeline := collectorConfig.Service.Pipelines[pipelineName]
		assert.Len(t, pipeline.Exporters, 1, "Expected 1 exporter in pipeline")
		assert.Equal(t, pipeline.Exporters[0], honeycombExporter, "Expected Honeycomb exporter")

		assert.Len(t, pipeline.Receivers, 1, "Expected 1 receiver in pipeline got %s", pipeline.Receivers)
		assert.Contains(t, pipeline.Receivers, otelReceiver, "Expected OTel receiver")

		if len(pipeline.Processors) == 1 {
			assert.Len(t, pipeline.Processors, 1, "Expected 1 processor in pipeline got %s", pipeline.Processors)
			assert.Contains(t, pipeline.Processors, usageProcessor, "Expected usage processor")
		} else {
			assert.Len(t, pipeline.Processors, 2, "Expected 2 processors in pipeline got %s", pipeline.Processors)
			assert.Contains(t, pipeline.Processors, usageProcessor, "Expected usage processor")
			assert.Contains(t, pipeline.Processors, filterProcessor, "Expected filter processor")
		}
	}

	// pipelines:
	//    logs:
	// 	    receivers: [otlp/OTel_Receiver_1]
	// 	    processors: [usage, filter/Info_Logs_only]
	// 	    exporters: [otlphttp/Honeycomb_Exporter_1]
	// 	  logs/1:
	// 	    receivers: [otlp/OTel_Receiver_1]
	// 	    processors: [usage]
	// 	    exporters: [otlphttp/Honeycomb_Exporter_1]
}
