package tmpl

import (
	"fmt"

	y "gopkg.in/yaml.v3"
)

type RulesComponentType string

const (
	StartSampling RulesComponentType = "startsampling"
	Condition     RulesComponentType = "condition"
	Sampler       RulesComponentType = "sampler"
)

func RCTFromStyle(style string) RulesComponentType {
	switch style {
	case "startsampling":
		return StartSampling
	case "condition":
		return Condition
	case "sampler":
		return Sampler
	default:
		return ""
	}
}

// Defines the configuration for a rules-based sampler in the Refinery.
// The private 'style' field is used to determine the type of component that created the object
// so that Merge can be done correctly.
type RulesConfig struct {
	Version  int
	Samplers map[string]*V2SamplerChoice `yaml:"Samplers,omitempty"`
	compType RulesComponentType          `yaml:"-"`
}

func NewRulesConfig(rct RulesComponentType) *RulesConfig {
	return &RulesConfig{
		Version:  2,
		Samplers: make(map[string]*V2SamplerChoice),
		compType: rct,
	}
}

func (rc *RulesConfig) RenderToMap(m map[string]any) map[string]any {
	// unlike some of the other configs, we don't need to render the samplers
	// to a map, because they are already in the correct format.
	// If we decide we want this, we can have it call RenderYAML and then
	// unmarshal the YAML into a map.

	// if m == nil {
	// 	m = make(map[string]any)
	// }
	// m["RulesVersion"] = rc.Version
	// foundDefault := false
	// for _, env := range rc.Envs {
	// 	if env.Name == "__default__" {
	// 		foundDefault = true
	// 	}
	// 	m = env.ConfigData.RenderToMap(m)
	// }
	// if !foundDefault {
	// 	// if we don't have a default env, we need to add one
	// 	defaultConfig := DottedConfig{
	// 		"Samplers.__default__.DeterministicSampler.SampleRate": 1,
	// 	}
	// 	m = defaultConfig.RenderToMap(m)
	// }
	return m
}

func (rc *RulesConfig) RenderYAML() ([]byte, error) {
	data, err := y.Marshal(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (rc *RulesConfig) Merge(other TemplateConfig) TemplateConfig {
	otherRC, ok := other.(*RulesConfig)
	if !ok {
		// if the other TemplateConfig is not a RulesConfig, we can't merge it
		return rc
	}

	// our merge types are:
	// condition into startsampling (for a condition sampler attached to a startsampling)
	// sampler into startsampling (for a sampler attached to a startsampling)
	// startsampling into startsampling (when we merge two startsampling rules for different environments)
	// so if rc is not a startsampling, we can't merge
	if rc.compType != StartSampling {
		return rc
	}

	for otherEnv, otherSampler := range otherRC.Samplers {
		// if this environment already exists, we have a problem
		if _, ok := rc.Samplers[otherEnv]; ok {
			// we can only scream here, this shouldn't happen and should have been caught
			// in validation
			fmt.Printf("environment %s already exists in RulesConfig, merge will be incorrect", otherEnv)
		}
		// otherwise, we add it to the map
		rc.Samplers[otherEnv] = otherSampler
	}
	return rc
}
