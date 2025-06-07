package collectorprovider

import (
	"github.com/honeycombio/opentelemetry-collector-configs/honeycombextension"
	"github.com/honeycombio/opentelemetry-collector-configs/usageprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func defaultComponents() otelcol.Factories {
	return otelcol.Factories{
		Receivers: map[component.Type]receiver.Factory{
			component.MustNewType("otlp"): otlpreceiver.NewFactory(),
		},
		Exporters: map[component.Type]exporter.Factory{
			component.MustNewType("debug"):    debugexporter.NewFactory(),
			component.MustNewType("otlp"):     otlpexporter.NewFactory(),
			component.MustNewType("otlphttp"): otlphttpexporter.NewFactory(),
		},
		Processors: map[component.Type]processor.Factory{
			component.MustNewType("transform"): transformprocessor.NewFactory(),
			component.MustNewType("usage"):     usageprocessor.NewFactory(),
		},
		Extensions: map[component.Type]extension.Factory{
			component.MustNewType("honeycomb"): honeycombextension.NewFactory(),
		},
	}
}
