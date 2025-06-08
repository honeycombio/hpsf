package configtests

import (
	"testing"

	tmpl "github.com/honeycombio/hpsf/pkg/config/tmpl"
	refineryconfig "github.com/honeycombio/hpsf/tests/providers/refinery"
	"github.com/stretchr/testify/assert"
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

	parsedConfig := refineryconfig.GetParsedRulesConfig(t, rulesConfig)

	assert.Equal(t, "test.name1", parsedConfig.Samplers["__default__"].RulesBasedSampler.Rules[0].Name)
	assert.Equal(t, "test.name2", parsedConfig.Samplers["__default__"].RulesBasedSampler.Rules[1].Name)

}
