package tmpl

import (
	"fmt"
	"strings"

	y "gopkg.in/yaml.v3"
)

// CollectorConfig is a config specifically focused on creating collector
// YAML configuration files.
type CollectorConfig struct {
	Sections map[string]DottedConfig
}

// ensure CollectorConfig implements TemplateConfig
var _ TemplateConfig = (*CollectorConfig)(nil)

// These types are used to unmarshal the collector config into a struct so that
// we can marshal it back out in a format that's idiomatic for the collector
// (i.e. with the sections in the right order and brackets for the lists --
// that's what the "flow" tag is for).

// signalPipeline is a struct that represents a pipeline in the collector config.
type signalPipeline struct {
	Receivers  []string `yaml:"receivers,flow"`
	Processors []string `yaml:"processors,flow"`
	Exporters  []string `yaml:"exporters,flow"`
}

// collectorConfigService is a struct that represents the service section of the collector config.
type collectorConfigService struct {
	Extensions []string                   `yaml:"extensions,omitempty,flow"`
	Pipelines  map[string]*signalPipeline `yaml:"pipelines"`
}

// collectorConfigFormat is a struct that represents the collector config in a
// format and ordering that's idiomatic for the collector.
type collectorConfigFormat struct {
	Receivers  map[string]any          `yaml:"receivers,omitempty"`
	Processors map[string]any          `yaml:"processors,omitempty"`
	Exporters  map[string]any          `yaml:"exporters,omitempty"`
	Extensions map[string]any          `yaml:"extensions,omitempty"`
	Service    *collectorConfigService `yaml:"service"`
}

// injectHoneycombUsageComponents ensures the collector configuration always has the necessary honeycomb
// components for measuring usage.
func (f *collectorConfigFormat) injectHoneycombUsageComponents() {
	if f.Service == nil {
		f.Service = &collectorConfigService{}
	}

	// ensure the honeycombextension is configured
	if f.Extensions == nil {
		f.Extensions = make(map[string]any)
	}
	f.Extensions["honeycomb"] = map[string]any{}
	if f.Service.Extensions == nil {
		f.Service.Extensions = make([]string, 0, 1)
	}
	f.Service.Extensions = append(f.Service.Extensions, "honeycomb")

	// ensure the usageprocessor is configured for all pipelines
	if f.Processors == nil {
		f.Processors = make(map[string]any)
	}
	f.Processors["usage"] = map[string]any{}

	// now we re-order the processors in each pipeline to:
	// - have memory_limiter first
	// - have usage second
	// - have all others after that
	for _, pipeline := range f.Service.Pipelines {
		// Separate memory_limiter processors from others
		memoryLimiters := []string{}
		others := []string{}

		for _, processor := range pipeline.Processors {
			if strings.HasPrefix(processor, "memory_limiter/") {
				memoryLimiters = append(memoryLimiters, processor)
			} else if processor != "usage" {
				others = append(others, processor)
			}
		}

		// Build ordered list: memory_limiters, usage, others
		orderedProcessors := make([]string, 0, len(memoryLimiters)+1+len(others))
		orderedProcessors = append(orderedProcessors, memoryLimiters...)
		orderedProcessors = append(orderedProcessors, "usage")
		orderedProcessors = append(orderedProcessors, others...)

		pipeline.Processors = dedup(orderedProcessors)
	}
}

func dedup[T comparable](slice []T) []T {
	keys := make(map[T]struct{})
	list := []T{}
	for _, entry := range slice {
		if _, found := keys[entry]; !found {
			keys[entry] = struct{}{}
			list = append(list, entry)
		}
	}
	return list
}

// Set sets a key in the config to a value. If the key already exists, it will
// append the value to the existing value if it's a slice, or overwrite it if
// it's not a slice.
// It will create the section if it doesn't exist.
func (cc *CollectorConfig) Set(section string, key string, value any) {
	if _, ok := cc.Sections[section]; !ok {
		cc.Sections[section] = make(DottedConfig)
	}
	if _, ok := cc.Sections[section][key]; !ok {
		cc.Sections[section][key] = value
	} else {
		switch v := value.(type) {
		case []any:
			// don't add duplicates
			cc.Sections[section][key] = dedup(append(cc.Sections[section][key].([]any), v...))
		case []string:
			cc.Sections[section][key] = dedup(append(cc.Sections[section][key].([]string), v...))
		case []int:
			cc.Sections[section][key] = dedup(append(cc.Sections[section][key].([]int), v...))
		case []float64:
			cc.Sections[section][key] = dedup(append(cc.Sections[section][key].([]float64), v...))
		default:
			cc.Sections[section][key] = v // overwrite if not a slice
		}
	}
}

// renderInto is a helper function that recursively renders a dotted key into a
// map.
func (cc *CollectorConfig) renderInto(m map[string]any, key string, value any) {
	// if the key contains a dot, split it into parts
	if strings.Contains(key, ".") {
		// split the key into parts
		parts := strings.SplitN(key, ".", 2)
		if m[parts[0]] == nil {
			m[parts[0]] = make(map[string]any)
		}
		// recursively call renderInto with the new map
		cc.renderInto(m[parts[0]].(map[string]any), parts[1], value)
	} else {
		// if the key does not contain a dot, assign the value
		m[key] = value
	}
}

// RenderToMap renders the config into a map.
func (cc *CollectorConfig) RenderToMap(m map[string]any) map[string]any {
	if m == nil {
		m = make(map[string]any)
	}
	for section := range cc.Sections {
		cc.Sections[section] = cc.Sections[section].RenderToMap(nil)
		for k, v := range cc.Sections[section] {
			key := section + "." + k
			cc.renderInto(m, key, v)
		}
	}
	return m
}

// RenderYAML renders the config into YAML.
func (cc *CollectorConfig) RenderYAML() ([]byte, error) {
	// we render the config to a map, and then marshal it to yaml
	// but that yaml is not idiomatic for the collector
	m := cc.RenderToMap(nil)
	data, err := y.Marshal(m)
	if err != nil {
		return nil, err
	}

	// now we unmarshal it into a struct so that we can use struct decorators
	// to create a format that's idiomatic for the collector
	var f collectorConfigFormat
	err = y.Unmarshal(data, &f)
	if err != nil {
		return nil, err
	}

	f.injectHoneycombUsageComponents()

	// now marshal from the struct to yaml
	data, err = y.Marshal(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Merge merges another TemplateConfig into this one, but only if it's really
// a CollectorConfig. If it's not, it just returns this config unmodified.
func (cc *CollectorConfig) Merge(other TemplateConfig) error {
	otherCC, ok := other.(*CollectorConfig)
	if !ok {
		// if the other TemplateConfig is not a CollectorConfig, we can't merge it
		return fmt.Errorf("cannot merge %T with CollectorConfig", other)
	}
	for section, items := range otherCC.Sections {
		for k, v := range items {
			cc.Set(section, k, v)
		}
	}
	return nil
}

// NewCollectorConfig creates a new CollectorConfig with an empty map of sections.
func NewCollectorConfig() *CollectorConfig {
	cc := CollectorConfig{Sections: make(map[string]DottedConfig)}
	return &cc
}
