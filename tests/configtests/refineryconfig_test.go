package configtests

import (
	"testing"

	tmpl "github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/tests/configproviders/refineryconfigprovider"
)

type BasicRefineryConfig struct {
	RulesVersion int            `yaml:"RulesVersion"`
	Samplers     map[string]any `yaml:"Samplers"`
}

func TestRefineryArrayRendering(t *testing.T) {

	rulesConfig := tmpl.NewRulesConfig()
	rulesConfig.Envs = append(rulesConfig.Envs, tmpl.EnvConfig{
		Name: "test",
		ConfigData: tmpl.DottedConfig{
			"Samplers.__default__.RulesBasedSampler.Rules[0].Name": "test.name1",
			"Samplers.__default__.RulesBasedSampler.Rules[1].Name": "test.name2",
		},
	})

	parsedConfig := refineryconfigprovider.GetParsedConfig(t, rulesConfig)

	if parsedConfig.Samplers["__default__"].RulesBasedSampler.Rules[0].Name != "test.name1" {
		t.Errorf("Expected value not found in rendered config")
	}
	if parsedConfig.Samplers["__default__"].RulesBasedSampler.Rules[1].Name != "test.name2" {
		t.Errorf("Expected value not found in rendered config")
	}

}
