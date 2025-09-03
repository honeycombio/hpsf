package tmpl

import (
	"fmt"
	"strconv"
	"strings"

	y "gopkg.in/yaml.v3"
)

type RulesMergeType int

// RulesMergeType represents the types of entities that can be combined to generate rules.
// For components, they come from the "style" inside the components. But we also have the
// output type which is not a component, but is used internally to help with the merging of samplers.
// - StartSampling is the type of component that starts the sampling process.
// - Condition represents a conditional branch in the rules.
// - Sampler represents a sampler in the rules.
// - Dropper is a sampler with no output.
// - Output is not a component, but is used internally to help with the merging of samplers.
const (
	Unknown RulesMergeType = iota
	StartSampling
	Condition
	Sampler
	Dropper
	Output
)

func String(rmt RulesMergeType) string {
	switch rmt {
	case StartSampling:
		return "startsampling"
	case Condition:
		return "condition"
	case Sampler:
		return "sampler"
	case Dropper:
		return "dropper"
	case Output:
		return "output"
	default:
		return "unknown"
	}
}

func RMTFromStyle(style string) (RulesMergeType, error) {
	switch style {
	case "startsampling":
		return StartSampling, nil
	case "condition":
		return Condition, nil
	case "sampler":
		return Sampler, nil
	case "dropper":
		return Dropper, nil
	default: // we don't need output because it's not a style
		return Unknown, fmt.Errorf("unknown RulesComponentType: %s", style)
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
	compType RulesMergeType              `yaml:"-"`
	meta     map[string]string           `yaml:"-"`
	kvs      map[string]any              `yaml:"-"`
}

// keys used to index the metadata map in RulesConfig
const (
	MetaPipelineIndex = "pipeline_index"
	MetaEnv           = "env"
	MetaSampler       = "sampler"
	MetaComponentName = "component_name"
)

func NewRulesConfig(rmt RulesMergeType, meta map[string]string, kvs map[string]any) *RulesConfig {
	return &RulesConfig{
		Version:  2,
		Samplers: make(map[string]*V2SamplerChoice),
		compType: rmt,
		meta:     meta,
		kvs:      kvs,
	}
}

func (rc *RulesConfig) RenderToMap(m map[string]any) map[string]any {
	// unlike some of the other configs, we don't need to render the samplers
	// to a map, because they are already in the correct format.
	// If we decide we want this, we can have it call RenderYAML and then
	// unmarshal the YAML into a map.

	return m
}

// maybePromoteSingleRuleSampler checks if any of the samplers in the RulesConfig should be
// promoted to the top level. This is idiomatic in Refinery rules, where our generation might
// have inserted a rule with no conditions and a single sampler. In this case, we can
// promote the sampler to the top level.
func (rc *RulesConfig) maybePromoteSingleRuleSampler() {
	for env, sampler := range rc.Samplers {
		if sampler != nil && sampler.RulesBasedSampler != nil && len(sampler.RulesBasedSampler.Rules) == 1 {
			rule := sampler.RulesBasedSampler.Rules[0]
			if len(rule.Conditions) == 0 {
				if rule.Sampler != nil {
					// Replace the V2SamplerChoice with the underlying sampler
					if rule.Sampler.DynamicSampler != nil {
						rc.Samplers[env] = &V2SamplerChoice{DynamicSampler: rule.Sampler.DynamicSampler}
					} else if rule.Sampler.EMADynamicSampler != nil {
						rc.Samplers[env] = &V2SamplerChoice{EMADynamicSampler: rule.Sampler.EMADynamicSampler}
					} else if rule.Sampler.EMAThroughputSampler != nil {
						rc.Samplers[env] = &V2SamplerChoice{EMAThroughputSampler: rule.Sampler.EMAThroughputSampler}
					} else if rule.Sampler.WindowedThroughputSampler != nil {
						rc.Samplers[env] = &V2SamplerChoice{WindowedThroughputSampler: rule.Sampler.WindowedThroughputSampler}
					} else if rule.Sampler.TotalThroughputSampler != nil {
						rc.Samplers[env] = &V2SamplerChoice{TotalThroughputSampler: rule.Sampler.TotalThroughputSampler}
					} else if rule.Sampler.DeterministicSampler != nil {
						rc.Samplers[env] = &V2SamplerChoice{DeterministicSampler: rule.Sampler.DeterministicSampler}
					}
				} else if !rule.Drop {
					// The rules sampler had no conditions, no samplers set, and was not dropping.
					// We default to grabbing the 1 rule's SampleRate and making a deterministic sampler.
					rc.Samplers[env] = &V2SamplerChoice{DeterministicSampler: &DeterministicSamplerConfig{SampleRate: rule.SampleRate}}
				}
			}
		}
	}
}

func (rc *RulesConfig) RenderYAML() ([]byte, error) {
	rc.maybePromoteSingleRuleSampler()
	data, err := y.Marshal(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Checks if the sampler type is one of the permitted downstream sampler types.
// We leave deterministic samplers out of this list (even though they're
// technically permitted) because they are handled specially in the rules
// config anyway.
func isDownstreamSamplerType(samplerType string) bool {
	switch samplerType {
	case "EMADynamicSampler",
		"EMAThroughputSampler",
		"WindowedThroughputSampler",
		"DynamicSampler",
		"TotalThroughputSampler":
		return true
	default:
		return false
	}
}

// Add the Name field if we're creating a rule (keyPrefix starts with
// "RulesBasedSampler.Rules." but doesn't contain "Conditions" or ".Sampler.").
func shouldAddNameField(keyPrefix string) bool {
	return strings.HasPrefix(keyPrefix, "RulesBasedSampler.Rules.") &&
		!strings.Contains(keyPrefix, "Conditions") &&
		!strings.Contains(keyPrefix, ".Sampler.")
}

// mergeScopeValues implements the scope merging logic:
// if either value is "span", the result is "span", otherwise "trace"
func mergeScopeValues(scope1, scope2 string) string {
	if scope1 == "span" || scope2 == "span" {
		return "span"
	}
	return "trace"
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
			// so we can write it into the startsampling at the pipeline index
			keyPrefix = fmt.Sprintf("RulesBasedSampler.Rules.%s.Conditions.0.", rc.meta[MetaPipelineIndex])

			for key, value := range otherRC.kvs {
				if err := setMemberValue(keyPrefix+key, sampler, value); err != nil {
					return err
				}
			}

			// Set the Scope field on the rule if the condition has a scope meta value
			// This is ugly because the structure of the rules is ugly.
			if scope, exists := otherRC.meta["scope"]; exists && scope != "" {
				// Merge scope values: if either is "span", result is "span", otherwise "trace"
				currentScope := "trace" // default scope
				ruleIndex, _ := strconv.Atoi(rc.meta[MetaPipelineIndex])
				if sampler.RulesBasedSampler != nil && ruleIndex < len(sampler.RulesBasedSampler.Rules) {
					if existingScope := sampler.RulesBasedSampler.Rules[ruleIndex].Scope; existingScope != "" {
						currentScope = existingScope
					}
				}

				mergedScope := mergeScopeValues(currentScope, scope)
				if err := setMemberValue(fmt.Sprintf("RulesBasedSampler.Rules.%s.Scope", rc.meta[MetaPipelineIndex]), sampler, mergedScope); err != nil {
					return err
				}
			}
		case Sampler:
			// in this case, we are merging a sampler directly into a startsampling
			// so we use the new sampler's type as the key prefix
			// The pipeline index should be propagated from the SamplingSequencer
			// For downstream samplers, we need to create a rule at the correct index
			samplerType := otherRC.meta[MetaSampler]
			if isDownstreamSamplerType(samplerType) {
				// Create a rule-based sampler with the correct index and sampler type
				keyPrefix = fmt.Sprintf("RulesBasedSampler.Rules.%s.Sampler.%s.", rc.meta[MetaPipelineIndex], samplerType)
			} else if samplerType == "DeterministicSampler" {
				// The DeterministicSampler needs a special case because the RulesBasedSampler supports a SampleRate field directly on the rule.
				keyPrefix = fmt.Sprintf("RulesBasedSampler.Rules.%s.", rc.meta[MetaPipelineIndex])
			} else {
				keyPrefix = fmt.Sprintf("%s.", samplerType)
			}
		case Dropper:
			// The refinery syntax for drop is terrible, so we have to handle it specially.
			keyPrefix = fmt.Sprintf("%s.Rules.%s.", otherRC.meta[MetaSampler], rc.meta[MetaPipelineIndex])
		default:
			return fmt.Errorf("cannot merge %T with RulesConfig because it is not valid start merge type", other)
		}

		for key, value := range otherRC.kvs {
			if err := setMemberValue(keyPrefix+key, sampler, value); err != nil {
				return err
			}
		}

		if shouldAddNameField(keyPrefix) {
			if componentName, exists := otherRC.meta[MetaComponentName]; exists {
				if err := setMemberValue(keyPrefix+"Name", sampler, componentName); err != nil {
					return err
				}
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
			sampler := rc.Samplers[rc.meta[MetaEnv]]

			// Ensure the sampler and RulesBasedSampler exist
			if sampler == nil {
				sampler = &V2SamplerChoice{}
				rc.Samplers[rc.meta[MetaEnv]] = sampler
			}
			if sampler.RulesBasedSampler == nil {
				sampler.RulesBasedSampler = &RulesBasedSamplerConfig{}
			}
			if len(sampler.RulesBasedSampler.Rules) <= ruleIndex {
				// Extend the rules slice to accommodate the new rule
				for i := len(sampler.RulesBasedSampler.Rules); i <= ruleIndex; i++ {
					sampler.RulesBasedSampler.Rules = append(sampler.RulesBasedSampler.Rules, &RulesBasedSamplerRule{})
				}
			}

			conditionIndex := len(sampler.RulesBasedSampler.Rules[ruleIndex].Conditions)
			keyPrefix := fmt.Sprintf("RulesBasedSampler.Rules.%d.Conditions.%d.", ruleIndex, conditionIndex)

			for key, value := range otherRC.kvs {
				if err := setMemberValue(keyPrefix+key, sampler, value); err != nil {
					return err
				}
			}

			// Set the Scope field on the rule if the condition has a scope meta value
			// This is ugly because the structure of the rules is ugly.
			if scope, exists := otherRC.meta["scope"]; exists && scope != "" {
				// Merge scope values: if either is "span", result is "span", otherwise "trace"
				currentScope := "trace" // default scope
				if ruleIndex < len(sampler.RulesBasedSampler.Rules) {
					if existingScope := sampler.RulesBasedSampler.Rules[ruleIndex].Scope; existingScope != "" {
						currentScope = existingScope
					}
				}

				mergedScope := mergeScopeValues(currentScope, scope)
				if err := setMemberValue(fmt.Sprintf("RulesBasedSampler.Rules.%d.Scope", ruleIndex), sampler, mergedScope); err != nil {
					return err
				}
			}

			rc.Samplers[rc.meta[MetaEnv]] = sampler
		case Sampler:
			// we need to check if the sampler is connected to a condition or not. If not, we
			// add it directly to the Samplers map, otherwise we add it to the
			// RulesBasedSampler.Rules slice at the correct index.
			// The pipeline index is propagated from the upstream component.
			ruleIndex, _ := strconv.Atoi(otherRC.meta[MetaPipelineIndex])
			samplerType := otherRC.meta[MetaSampler]
			sampler := rc.Samplers[rc.meta[MetaEnv]]
			var keyPrefix string
			if sampler.RulesBasedSampler == nil || len(sampler.RulesBasedSampler.Rules) == 0 {
				keyPrefix = fmt.Sprintf("%s.", samplerType)
			} else {
				if isDownstreamSamplerType(samplerType) {
					keyPrefix = fmt.Sprintf("RulesBasedSampler.Rules.%d.Sampler.%s.", ruleIndex, samplerType)
				} else {
					keyPrefix = fmt.Sprintf("RulesBasedSampler.Rules.%d.", ruleIndex)
				}
			}
			for key, value := range otherRC.kvs {
				if err := setMemberValue(keyPrefix+key, sampler, value); err != nil {
					return err
				}
			}
			// Only set Name if keyPrefix is at the rule level (not inside .Sampler.)
			if shouldAddNameField(keyPrefix) {
				componentName, exists := otherRC.meta[MetaComponentName]
				if !exists {
					// Fallback: try to get component name from current RC's meta
					componentName, exists = rc.meta[MetaComponentName]
				}
				if exists {
					if err := setMemberValue(fmt.Sprintf("RulesBasedSampler.Rules.%d.Name", ruleIndex), sampler, componentName); err != nil {
						return err
					}
				}
			}

			// For downstream samplers, always set the rule's Name field to the sampler's component name
			if isDownstreamSamplerType(samplerType) {
				componentName, exists := otherRC.meta[MetaComponentName]
				if exists {
					if err := setMemberValue(fmt.Sprintf("RulesBasedSampler.Rules.%d.Name", ruleIndex), sampler, componentName); err != nil {
						return err
					}
				}
			}
			rc.Samplers[rc.meta[MetaEnv]] = sampler
		case Output:
			// if they have the same environment, and both are rules-based, we
			// add to the rules slice. if they have different environments, we
			// add to the Samplers map.
			if rc.meta[MetaEnv] == otherRC.meta[MetaEnv] {
				sampler := rc.Samplers[rc.meta[MetaEnv]]
				if sampler == nil {
					sampler = &V2SamplerChoice{}
					rc.Samplers[rc.meta[MetaEnv]] = sampler
				}
				if sampler.RulesBasedSampler == nil {
					sampler.RulesBasedSampler = &RulesBasedSamplerConfig{}
				}
				otherSampler := otherRC.Samplers[otherRC.meta[MetaEnv]]
				if otherSampler != nil && otherSampler.RulesBasedSampler != nil {
					sampler.RulesBasedSampler.Rules = append(
						sampler.RulesBasedSampler.Rules,
						otherSampler.RulesBasedSampler.Rules...)
				}
			} else {
				// we need to add the other environment's sampler to the map
				rc.Samplers[otherRC.meta[MetaEnv]] = otherRC.Samplers[otherRC.meta[MetaEnv]]
			}
		case Dropper:
			ruleIndex, _ := strconv.Atoi(otherRC.meta[MetaPipelineIndex])
			sampler := rc.Samplers[rc.meta[MetaEnv]]
			keyPrefix := fmt.Sprintf("RulesBasedSampler.Rules.%d.", ruleIndex)
			for key, value := range otherRC.kvs {
				if err := setMemberValue(keyPrefix+key, sampler, value); err != nil {
					return err
				}
			}
			componentName, exists := otherRC.meta[MetaComponentName]
			if !exists {
				// Fallback: try to get component name from current RC's meta
				componentName, exists = rc.meta[MetaComponentName]
			}
			if exists {
				if err := setMemberValue(fmt.Sprintf("RulesBasedSampler.Rules.%d.Name", ruleIndex), sampler, componentName); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("cannot merge %T with RulesConfig because it is not valid output merge type", other)
		}
	default:
		return fmt.Errorf("cannot merge into RulesConfig because '%v' is not a valid component type", rc.compType)
	}
	return nil
}
