package yaml

import "strings"

// DottedConfig is a map that allows for keys with dots in them;
// it can convert a regular map into a DottedConfig, and
// when rendered, it will generate nested maps.
// This exists because dotted paths are easier to merge.
type DottedConfig map[string]any

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

func (dc DottedConfig) Render() map[string]any {
	m := make(map[string]any)
	for k, v := range dc {
		dc.renderInto(m, k, v)
	}
	return m
}

func (dc DottedConfig) Merge(other DottedConfig) DottedConfig {
	for k, v := range other {
		dc[k] = v
	}
	return dc
}

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
