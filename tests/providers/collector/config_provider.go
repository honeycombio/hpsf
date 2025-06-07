package collectorconfigprovider

import (
	"context"
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/opentelemetry-collector-configs/honeycombextension"
	"github.com/honeycombio/opentelemetry-collector-configs/usageprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"gopkg.in/yaml.v2"
)

type ParseError struct {
	IsError    bool
	Config     string
	InnerError error
}

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

func newConfFromString(tb testing.TB, content string) map[string]any {
	var conf map[string]any
	err := yaml.Unmarshal([]byte(content[11:]), &conf)
	require.NoError(tb, err)
	return conf
}

func GetParsedConfig(t *testing.T, cc *tmpl.CollectorConfig) (*otelcol.Config, ParseError) {
	renderedYamlConfig, _ := cc.RenderYAML()
	renderedYamlAsString := string(renderedYamlConfig)

	inmemoryProvider := newFakeProvider("inmemory", func(_ context.Context, uri string, w confmap.WatcherFunc) (*confmap.Retrieved, error) {
		return confmap.NewRetrieved(newConfFromString(t, uri))
	})
	stringProvider := newFakeProvider("string", func(_ context.Context, uri string, w confmap.WatcherFunc) (*confmap.Retrieved, error) {
		return confmap.NewRetrievedFromYAML([]byte(uri[7:]))
	})

	configProvider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs:              []string{"inmemory://" + renderedYamlAsString},
			ProviderFactories: []confmap.ProviderFactory{inmemoryProvider, stringProvider},
			DefaultScheme:     "string",
		},
	})

	if err != nil {
		t.Errorf("Error creating config provider: %v", err)
	}
	componentFactories := defaultComponents()
	parsedConfig, parseError := configProvider.Get(context.Background(), componentFactories)

	// if there's a parseError, return a custom error that includes the rendered config
	if parseError != nil {
		return nil, ParseError{InnerError: parseError, Config: renderedYamlAsString, IsError: true}
	}
	return parsedConfig, ParseError{}
}

type ComponentGetResult struct {
	Found        bool
	SearchString string
	Components   []string
}

func GetProcessorConfig[T any](cfg *otelcol.Config, processorId string) (*T, bool) {
	typeAndName := strings.Split(processorId, "/")
	var processorID component.ID
	if len(typeAndName) == 2 {
		processorID = component.MustNewIDWithName(typeAndName[0], typeAndName[1])
	}
	if len(typeAndName) == 1 {
		processorID = component.MustNewID(typeAndName[0])
	}
	processorCfg, ok := cfg.Processors[processorID]
	if !ok {
		return nil, false
	}
	typedCfg, ok := processorCfg.(*T)
	if ok {
		return typedCfg, true
	}
	// If direct cast fails, try reflection for pointer to struct
	val, ok := any(processorCfg).(T)
	if ok {
		return &val, true
	}
	return nil, false
}

func GetExporterConfig[T any](cfg *otelcol.Config, exporterId string) (*T, ComponentGetResult) {
	typeAndName := strings.Split(exporterId, "/")
	var exporterID component.ID
	if len(typeAndName) == 2 {
		exporterID = component.MustNewIDWithName(typeAndName[0], typeAndName[1])
	}
	if len(typeAndName) == 1 {
		exporterID = component.MustNewID(typeAndName[0])
	}
	exporterCfg, ok := cfg.Exporters[exporterID]
	if !ok {
		return nil, ComponentGetResult{Found: false, SearchString: exporterId, Components: listComponents(cfg.Exporters)}
	}
	typedCfg, ok := exporterCfg.(*T)
	if ok {
		return typedCfg, ComponentGetResult{Found: true, SearchString: exporterId}
	}
	// If direct cast fails, try reflection for pointer to struct
	val, ok := any(exporterCfg).(T)
	if ok {
		return &val, ComponentGetResult{Found: true, SearchString: exporterId}
	}
	return nil, ComponentGetResult{Found: false, SearchString: exporterId, Components: listComponents(cfg.Exporters)}
}

func listComponents(components map[component.ID]component.Config) []string {
	componentList := make([]string, 0)
	for name, _ := range components {
		componentList = append(componentList, name.String())
	}
	return componentList
}
