package refineryconfigprovider

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	refineryConfig "github.com/honeycombio/refinery/config"
	"gopkg.in/yaml.v2"
)

func GetParsedConfig(t *testing.T, rulesConfig *tmpl.RulesConfig) *refineryConfig.V2SamplerConfig {
	renderedConfig, err := rulesConfig.RenderYAML()
	if err != nil {
		t.Errorf("Error rendering config: %v", err)
	}

	parsedConfig := refineryConfig.V2SamplerConfig{}
	parsingError := yaml.Unmarshal(renderedConfig, &parsedConfig)
	if parsingError != nil {
		t.Errorf("Error parsing rendered config: %v", parsingError)
	}

	return &parsedConfig
}
