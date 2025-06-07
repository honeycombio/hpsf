package hpsftests

import (
	"testing"

	collectorConfigprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
)

func TestMultipleOTLPExporters(t *testing.T) {

	rulesConfig, collectorConfig, errors := hpsfprovider.GetParsedConfigsFromFile(t, "multiple_otlp_exporters.yaml")
	errors.FailIfError(t)

	_, _, exporters, getResult := collectorConfigprovider.GetPipelineConfig(collectorConfig, "traces")
	if !getResult.Found || len(exporters) != 2 {
		t.Errorf("Expected 2 exporters, got %s", exporters)
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
