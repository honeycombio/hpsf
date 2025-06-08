package refineryconfigprovider

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	refineryConfig "github.com/honeycombio/refinery/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func GetParsedRulesConfig(t *testing.T, rulesConfig *tmpl.RulesConfig) *refineryConfig.V2SamplerConfig {
	renderedConfig, err := rulesConfig.RenderYAML()
	require.NoError(t, err, "Error rendering config")

	parsedConfig := refineryConfig.V2SamplerConfig{}
	parsingError := yaml.Unmarshal(renderedConfig, &parsedConfig)
	require.NoError(t, parsingError, "Error parsing rendered config")

	return &parsedConfig
}
