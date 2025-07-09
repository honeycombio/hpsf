package tmpl

import (
	"fmt"
	"strconv"

	y "gopkg.in/yaml.v3"
)

type RulesComponentType string

const (
	Output        RulesComponentType = "output"
	StartSampling RulesComponentType = "startsampling"
	Condition     RulesComponentType = "condition"
	Sampler       RulesComponentType = "sampler"
	Dropper       RulesComponentType = "dropper"
)

func RCTFromStyle(style string) RulesComponentType {
	switch style {
	case "startsampling":
		return StartSampling
	case "condition":
		return Condition
	case "sampler":
		return Sampler
	case "dropper":
		return Dropper
	default: // we don't need output because it's not a style
		return "unknown" + "(" + RulesComponentType(style) + ")"
	}
}

// Defines the configuration for a rules-based sampler in the Refinery. This is
// a dual-purpose object: before merge, it holds the private fields so that we
// can defer the rendering of the samplers until we merge the results. This is
// because the final position of sampler configurations will depend on how they
// are wired. After merge, it has been converted to an object that can be
// rendered directly to YAML. The private 'compType' field is used to determine
// the type of component that created the object so that Merge can be done
// correctly; objects ready for rendering will have a compType of Output.
type RulesConfig struct {
	Version  int                         `yaml:"RulesVersion,omitempty"`
	Samplers map[string]*V2SamplerChoice `yaml:"Samplers,omitempty"`
	compType RulesComponentType          `yaml:"-"`
	meta     map[string]string           `yaml:"-"`
	kvs      map[string]any              `yaml:"-"`
}

// keys used to index the metadata map in RulesConfig
const (
	MetaPipelineIndex = "pipeline_index"
	MetaEnv           = "env"
	MetaSampler       = "sampler"
)

func NewRulesConfig(rct RulesComponentType, meta map[string]string, kvs map[string]any) *RulesConfig {
	return &RulesConfig{
		Version:  2,
		Samplers: make(map[string]*V2SamplerChoice),
		compType: rct,
		meta:     meta,
		kvs:      kvs,
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

func (rc *RulesConfig) Merge(other TemplateConfig) error {
	otherRC, ok := other.(*RulesConfig)
	if !ok {
		// if the other TemplateConfig is not a RulesConfig, we can't merge it
		return fmt.Errorf("cannot merge %T with RulesConfig", other)
	}

	// All merges convert the kvs (which is a path-based key and value) to a renderable Refinery configuration,
	// and also flag the output with type "output".
	// Our possible merge types are:
	// - into startsampling from condition (for a condition sampler attached to a startsampling)
	// - into startsampling from sampler (for a sampler attached directly to a startsampling)
	// - into output from condition (for chained conditions)
	// - into output from sampler (for a sampler attached to a condition)
	// - into output from output (when we merge two completed startsampling rules for different environments)
	// We'll do a nested switch here to handle the different cases.

	switch rc.compType {
	case StartSampling:
		rc.compType = Output // we will now treat this as an output
		sampler := &V2SamplerChoice{}

		var keyPrefix string
		switch otherRC.compType {
		case Condition:
			// condition always has a rule-based sampler attached to it
			// so we can write it into the startsampling at index 0 because we know it's first
			keyPrefix = fmt.Sprintf("RulesBasedSampler.Rules.%s.Conditions.0.", rc.meta[MetaPipelineIndex])
		case Sampler:
			// in this case, we are merging a sampler directly into a startsampling
			// so we use the new sampler's type as the key prefix
			keyPrefix = fmt.Sprintf("%s.", otherRC.meta[MetaSampler])
		case Dropper:
			// The refinery syntax for drop is terrible, so we have to handle it specially.
			keyPrefix = fmt.Sprintf("%s.Rules.%s.", otherRC.meta[MetaSampler], rc.meta[MetaPipelineIndex])
		default:
			return fmt.Errorf("cannot merge %T with RulesConfig because it is not valid start merge type", other)
		}

		for key, value := range otherRC.kvs {
			if err := SetMemberValue(keyPrefix+key, sampler, value); err != nil {
				return err
			}
		}
		rc.Samplers[rc.meta[MetaEnv]] = sampler

	case Output:
		switch otherRC.compType {
		case StartSampling:
			// this is what happens at the start of a pipeline
			rc.Version = otherRC.Version
			rc.compType = otherRC.compType
			rc.meta = otherRC.meta
			rc.kvs = otherRC.kvs
		case Condition:
			// We know the pipeline_index (ruleIndex) is in rc.meta.
			// We need to figure out the condition index by looking at the RulesBasedSampler.Rules.Conditions
			// at the correct index, and then we can write the sampler into the output at that index

			// this was put here by Itoa so we don't worry about errors
			ruleIndex, _ := strconv.Atoi(rc.meta[MetaPipelineIndex])
			conditionIndex := len(rc.Samplers[rc.meta[MetaEnv]].RulesBasedSampler.Rules[ruleIndex].Conditions)
			keyPrefix := fmt.Sprintf("RulesBasedSampler.Rules.%d.Conditions.%d.", ruleIndex, conditionIndex)

			sampler := rc.Samplers[rc.meta[MetaEnv]]
			for key, value := range otherRC.kvs {
				if err := SetMemberValue(keyPrefix+key, sampler, value); err != nil {
					return err
				}
			}
			rc.Samplers[rc.meta[MetaEnv]] = sampler
		case Sampler:
			// we need to check if the sampler is connected to a condition or not. If not, we
			// add it directly to the Samplers map, otherwise we add it to the
			// RulesBasedSampler.Rules slice at the correct index.
			ruleIndex, _ := strconv.Atoi(rc.meta[MetaPipelineIndex])
			samplerType := otherRC.meta[MetaSampler]
			sampler := rc.Samplers[rc.meta[MetaEnv]]
			var keyPrefix string
			if sampler.RulesBasedSampler == nil || len(sampler.RulesBasedSampler.Rules) == 0 {
				keyPrefix = fmt.Sprintf("%s.", samplerType)
			} else {
				keyPrefix = fmt.Sprintf("RulesBasedSampler.Rules.%d.", ruleIndex)
			}
			for key, value := range otherRC.kvs {
				if err := SetMemberValue(keyPrefix+key, sampler, value); err != nil {
					return err
				}
			}
			rc.Samplers[rc.meta[MetaEnv]] = sampler
		case Output:
			// if they have the same environment, and both are rules-based, we
			// add to the rules slice. if they have different environments, we
			// add to the Samplers map.
			if rc.meta[MetaEnv] == otherRC.meta[MetaEnv] {
				rc.Samplers[rc.meta[MetaEnv]].RulesBasedSampler.Rules = append(
					rc.Samplers[rc.meta[MetaEnv]].RulesBasedSampler.Rules,
					otherRC.Samplers[otherRC.meta[MetaEnv]].RulesBasedSampler.Rules...)
			} else {
				// we need to add the other environment's sampler to the map
				rc.Samplers[otherRC.meta[MetaEnv]] = otherRC.Samplers[otherRC.meta[MetaEnv]]
			}
		default:
			return fmt.Errorf("cannot merge %T with RulesConfig because it is not valid output merge type", other)
		}
	default:
		return fmt.Errorf("cannot merge into RulesConfig because '%s' is not a valid component type", rc.compType)
	}
	return nil
}
