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
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"gopkg.in/yaml.v2"
)

type ParseError struct {
	Err    error
	Config string
}

func (e *ParseError) Error() string {
	return e.Err.Error() + "\n" + e.Config
}

func defaultComponents() (otelcol.Factories, error) {
	return otelcol.Factories{
		Receivers: map[component.Type]receiver.Factory{
			component.MustNewType("otlp"): otlpreceiver.NewFactory(),
		},
		Exporters: map[component.Type]exporter.Factory{
			component.MustNewType("debug"): debugexporter.NewFactory(),
		},
		Processors: map[component.Type]processor.Factory{
			component.MustNewType("transform"): transformprocessor.NewFactory(),
			component.MustNewType("usage"):     usageprocessor.NewFactory(),
		},
		Extensions: map[component.Type]extension.Factory{
			component.MustNewType("honeycomb"): honeycombextension.NewFactory(),
		},
	}, nil
}

func newConfFromString(tb testing.TB, content string) map[string]any {
	var conf map[string]any
	err := yaml.Unmarshal([]byte(content[11:]), &conf)
	require.NoError(tb, err)
	return conf
}

func GetParsedConfig(t *testing.T, cc *tmpl.CollectorConfig) (*otelcol.Config, *ParseError) {
	renderedYamlConfig, err := cc.RenderYAML()
	renderedYamlAsString := string(renderedYamlConfig)

	inmemory := newFakeProvider("inmemory", func(_ context.Context, uri string, w confmap.WatcherFunc) (*confmap.Retrieved, error) {
		return confmap.NewRetrieved(newConfFromString(t, uri))
	})
	configProvider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs:              []string{"inmemory://" + renderedYamlAsString},
			ProviderFactories: []confmap.ProviderFactory{inmemory},
			DefaultScheme:     "inmemory",
		},
	})

	if err != nil {
		t.Errorf("Error creating config provider: %v", err)
	}
	componentFactories, err := defaultComponents()
	parsedConfig, parseError := configProvider.Get(context.Background(), componentFactories)

	// if there's a parseError, return a custom error that includes the rendered config
	if parseError != nil {
		return nil, &ParseError{Err: parseError, Config: renderedYamlAsString}
	}
	return parsedConfig, nil
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
