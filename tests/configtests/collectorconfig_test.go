package configtests

import (
	"testing"

	tmpl "github.com/honeycombio/hpsf/pkg/config/tmpl"
	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func TestValidateCollectorConfig(t *testing.T) {

	cc := tmpl.NewCollectorConfig()
	cc.Set("receivers", "otlp.protocols.http.endpoint", "0.0.0.0:4317")
	cc.Set("exporters", "debug", map[string]interface{}{})
	cc.Set("service", "pipelines.traces.receivers", []string{"otlp"})
	cc.Set("service", "pipelines.traces.processors", []string{})
	cc.Set("service", "pipelines.traces.exporters", []string{"debug"})

	parsedConfig, parserError := collectorprovider.GetParsedConfig(t, cc)
	assert.False(t, parserError.HasError, "Error parsing config: %v\n Rendered Config: %s\n", parserError.Error, parserError.Config)
	assert.Equal(t, "0.0.0.0:4317", parsedConfig.Receivers[component.MustNewID("otlp")].(*otlpreceiver.Config).HTTP.ServerConfig.Endpoint)
}
