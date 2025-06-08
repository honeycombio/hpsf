package collectorprovider

import (
	"context"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/otelcol"
)

type CollectorConfigParseError struct {
	HasError bool
	Config   string
	Error    error
}

func GetParsedConfig(t *testing.T, cc *tmpl.CollectorConfig) (*otelcol.Config, CollectorConfigParseError) {
	renderedYamlConfig, renderYamlError := cc.RenderYAML()
	require.NoError(t, renderYamlError, "Error during RenderYAML while reading collector Config")
	renderedYamlAsString := string(renderedYamlConfig)

	inmemoryProvider := newFakeConfmapProvider("inmemory", func(_ context.Context, uri string, w confmap.WatcherFunc) (*confmap.Retrieved, error) {
		return confmap.NewRetrievedFromYAML([]byte(uri[9:]))
	})
	stringProvider := newFakeConfmapProvider("string", func(_ context.Context, uri string, w confmap.WatcherFunc) (*confmap.Retrieved, error) {
		return confmap.NewRetrievedFromYAML([]byte(uri[7:]))
	})

	configProvider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs:              []string{"inmemory:" + renderedYamlAsString},
			ProviderFactories: []confmap.ProviderFactory{inmemoryProvider, stringProvider},
			DefaultScheme:     "string",
		},
	})

	require.NoError(t, err, "Error creating collector config provider")
	componentFactories := defaultComponents()
	parsedConfig, parseError := configProvider.Get(context.Background(), componentFactories)

	// if there's a parseError, return a custom error that includes the rendered config
	if parseError != nil {
		return nil, CollectorConfigParseError{Error: parseError, Config: renderedYamlAsString, HasError: true}
	}

	collectorConfigValidationError := parsedConfig.Validate()
	if collectorConfigValidationError != nil {
		return nil, CollectorConfigParseError{Error: collectorConfigValidationError, Config: renderedYamlAsString, HasError: true}
	}

	return parsedConfig, CollectorConfigParseError{}
}
