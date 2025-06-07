package hpsftests

import (
	"os"
	"testing"

	collectorConfigprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/pipeline"
)

func TestMultipleOTLPExporters(t *testing.T) {

	file, err := os.ReadFile("multiple_otlp_exporters.yaml")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	rulesConfig, collectorConfig, errors := hpsfprovider.GetParsedConfigs(t, string(file))
	errors.FailIfError(t)

	if len(collectorConfig.Service.Pipelines[pipeline.NewID(pipeline.SignalTraces)].Exporters) != 2 {
		t.Errorf("Expected 2 exporters, got %d", len(collectorConfig.Service.Pipelines[pipeline.NewID(pipeline.SignalTraces)].Exporters))
	}

	customBackendConfig, findResult := collectorConfigprovider.GetExporterConfig[otlphttpexporter.Config](collectorConfig, "otlphttp/My_Custom_backend")
	if !findResult.Found {
		t.Fatalf("Expected exporter to find \"%v\", found (%v)", findResult.SearchString, findResult.Components)
	}

	if customBackendConfig.ClientConfig.Headers["x-custom-backend"] != "MY_KEY" {
		t.Errorf("Expected custom header to be set to \"MY_KEY\", got %s", customBackendConfig.ClientConfig.Headers["x-custom-backend"])
	}

	if len(rulesConfig.Samplers) != 1 {
		t.Errorf("Expected no samplers, got %d", len(rulesConfig.Samplers))
	}

}
