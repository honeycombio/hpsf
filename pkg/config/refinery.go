package config

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/yaml"
)

// This base component is used to make sure that the config will be valid
// even if it stands alone. This is likely to be a temporary solution until we have a
// database of components.
type RefineryBaseComponent struct {
	Component hpsf.Component
}

func (c RefineryBaseComponent) GenerateConfig(ct Type, userdata map[string]any) (yaml.DottedConfig, error) {
	switch ct {
	case RefineryConfigType:
		return yaml.DottedConfig{
			"General.ConfigurationVersion": 2,
			"General.MinRefineryVersion":   "v2.0",
		}, nil
	case RefineryRulesType:
		return yaml.DottedConfig{
			"RulesVersion": 2,
		}, nil
	default:
		return nil, nil
	}
}

type RefineryInputComponent struct {
	Component hpsf.Component
}

// ensure RefineryInputComponent implements Component
var _ Component = RefineryInputComponent{}

func (c RefineryInputComponent) GenerateConfig(ct Type, userdata map[string]any) (yaml.DottedConfig, error) {
	if ct != RefineryConfigType {
		return nil, nil
	}
	if c.Component.Properties == nil {
		return nil, nil
	}

	port := c.Component.GetProperty("Port")
	if port == nil {
		return nil, nil
	}
	pstr := yaml.AsString(port.Value)

	switch c.Component.Kind {
	case "RefineryGRPC":
		return yaml.DottedConfig{
			"GRPCServerParameters.Enabled":    true,
			"GRPCServerParameters.ListenAddr": "0.0.0.0:" + pstr,
		}, nil
	case "RefineryHTTP":
		return yaml.DottedConfig{
			"GRPCServerParameters.Enabled": true,
			"Network.ListenAddr":           "0.0.0.0:" + pstr,
		}, nil
	default:
		return nil, fmt.Errorf("unknown refinery input component: %s", c.Component.Name)
	}
}

type DeterministicSampler struct {
	Component hpsf.Component
}

// ensure DeterministicSampler implements Component
var _ Component = DeterministicSampler{}

func (c DeterministicSampler) GenerateConfig(ct Type, userdata map[string]any) (yaml.DottedConfig, error) {
	if ct != RefineryRulesType {
		return nil, nil
	}

	if c.Component.Properties == nil {
		return nil, nil
	}

	rate := c.Component.GetProperty("SampleRate")
	if rate == nil {
		return nil, nil
	}
	r := yaml.AsInt(rate.Value)

	env := c.Component.GetProperty("Environment")
	if env == nil {
		return nil, nil
	}
	e := yaml.AsString(env.Value)

	return yaml.DottedConfig{
		fmt.Sprintf("Samplers.%s.DeterministicSampler.SampleRate", e): r,
		"Samplers." + e + ".DeterministicSampler.SampleRate":          r,
	}, nil
}
