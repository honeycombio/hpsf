package hpsftests

import (
	"testing"

	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
)

func TestSamplerRuleOverride(t *testing.T) {
	rulesConfig, _, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/sampler_rule_override.yaml")

	assert.Len(t, rulesConfig.Samplers, 1)

	assert.Len(t, rulesConfig.Samplers["__default__"].RulesBasedSampler.Rules, 4)
	rules := rulesConfig.Samplers["__default__"].RulesBasedSampler.Rules
	assert.Contains(t, rules[0].Name, "Keep traces")
	assert.Equal(t, 1, rules[0].SampleRate)
	assert.Equal(t, "exists", rules[0].Conditions[0].Operator)

	assert.Contains(t, rules[1].Name, "500")
	assert.Equal(t, 1, rules[1].SampleRate)
	assert.Equal(t, ">=", rules[1].Conditions[0].Operator)

	assert.Contains(t, rules[2].Name, "400")
	assert.Equal(t, 10, rules[2].SampleRate)
	assert.Equal(t, "400", rules[2].Conditions[0].Value)

	assert.Contains(t, rules[3].Name, "remainder")
	assert.Equal(t, 100, rules[3].SampleRate)
}
