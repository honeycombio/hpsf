package tmpl

import (
	"strings"

	y "gopkg.in/yaml.v3"
)

// DottedConfig is a map that allows for keys with dots in them;
// it can convert a regular map into a DottedConfig, and
// when rendered, it will generate nested maps.
// This exists because dotted paths are easier to merge.
type DottedConfig map[string]any

// renderInto is a helper function that recursively renders a DottedConfig into a map.
func (dc DottedConfig) renderInto(m map[string]any, key string, value any) {
	// if the key contains a dot, split it into parts
	if strings.Contains(key, ".") {
		// split the key into parts
		parts := strings.SplitN(key, ".", 2)
		if m[parts[0]] == nil {
			m[parts[0]] = make(map[string]any)
		}
		// recursively call renderInto with the new map
		dc.renderInto(m[parts[0]].(map[string]any), parts[1], value)
	} else {
		// if the key does not contain a dot, assign the value
		m[key] = value
	}
}

// RenderToMap renders the config into a map.
func (dc DottedConfig) RenderToMap() map[string]any {
	m := make(map[string]any)
	for k, v := range dc {
		dc.renderInto(m, k, v)
	}
	return m
}

// RenderYAML renders the config into YAML and returns a hash of it.
func (dc DottedConfig) RenderYAML() ([]byte, error) {
	m := dc.RenderToMap()
	data, err := y.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Merge combines two `DottedConfig` structs together; the values from the
// `DottedConfig` passed in will override any values that are not slices.
func (dc DottedConfig) Merge(other TemplateConfig) TemplateConfig {
	otherDotted, ok := other.(DottedConfig)
	if !ok {
		// if the other TemplateConfig is not a DottedConfig, we can't merge it
		return dc
	}
	for k, v := range otherDotted {
		if _, ok := dc[k]; !ok {
			dc[k] = v
		} else {
			switch v := v.(type) {
			case []any:
				dc[k] = append(dc[k].([]any), v...)
			case []string:
				dc[k] = append(dc[k].([]string), v...)
			case []int:
				dc[k] = append(dc[k].([]int), v...)
			case []float64:
				dc[k] = append(dc[k].([]float64), v...)
			default:
				dc[k] = v // overwrite if not a slice
			}
		}
	}
	return dc
}

// NewDottedConfig recursively converts a map into a DottedConfig.
func NewDottedConfig(m map[string]any) DottedConfig {
	dc := DottedConfig{}
	for k, v := range m {
		switch v := v.(type) {
		case map[string]any:
			for kk, vv := range NewDottedConfig(v) {
				dc[k+"."+kk] = vv
			}
		default:
			dc[k] = v
		}
	}
	return dc
}
