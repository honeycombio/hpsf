package configtests

import (
	"testing"

	tmpl "github.com/honeycombio/hpsf/pkg/config/tmpl"
	collectorconfigprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func TestValidateCollectorConfig(t *testing.T) {

	cc := tmpl.NewCollectorConfig()
	cc.Set("receivers", "otlp.protocols.http.endpoint", "0.0.0.0:4317")
	cc.Set("service", "pipelines.traces.receivers", []string{"otlp"})
	cc.Set("service", "pipelines.traces.processors", []string{})

	parsedConfig, err := collectorconfigprovider.GetParsedConfig(t, cc)
	if err != nil {
		t.Errorf("Error parsing config: %v\n Rendedered Config: %s\n", err, err.Config)
	}

	if parsedConfig.Receivers[component.MustNewID("otlp")].(*otlpreceiver.Config).HTTP.ServerConfig.Endpoint != "0.0.0.0:4317" {
		t.Errorf("Expected endpoint to be localhost:4317, got %s", parsedConfig.Receivers[component.MustNewID("otlp")].(*otlpreceiver.Config).HTTP.ServerConfig.Endpoint)
	}
}
