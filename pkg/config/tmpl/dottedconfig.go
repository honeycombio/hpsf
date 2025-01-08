package tmpl

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	y "gopkg.in/yaml.v3"
)

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

func (dc DottedConfig) RenderYAML() ([]byte, string, error) {
	m := dc.Render()
	data, err := y.Marshal(m)
	if err != nil {
		return nil, "", err
	}
	// we use md5 here because:
	// * this is not a security-centric use case, we just want a hash
	// * it's compatible with command line tools as well as Refinery's existing code
	h := md5.New()
	hash := hex.EncodeToString(h.Sum(data))
	return data, hash, nil
}

func (dc DottedConfig) Merge(other TemplateConfig) TemplateConfig {
	otherDotted, ok := other.(DottedConfig)
	if !ok {
		// if the other TemplateConfig is not a DottedConfig, we can't merge it
		return dc
	}
	for k, v := range otherDotted {
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
