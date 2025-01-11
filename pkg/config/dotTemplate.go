package config

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
)

type dottedConfigTemplateKV struct {
	key   string
	value string
}

// dottedConfigTemplate is the type we use for templates that properly modeled by
// a collection of key-value pairs that may be represented by structure keys (like Refinery config).

type dottedConfigTemplate []dottedConfigTemplateKV

func buildDottedConfigTemplate(data []any) (dottedConfigTemplate, error) {
	var d dottedConfigTemplate
	for _, v := range data {
		m, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected map, got %T", v)
		}
		var sk, sv string
		if mk, ok := m["key"]; !ok {
			return nil, fmt.Errorf("missing key in template data")
		} else {
			if _, ok := mk.(string); !ok {
				return nil, fmt.Errorf("expected string for key, got %T", mk)
			}
			sk = mk.(string)
		}
		if _, ok := m["value"]; !ok {
			return nil, fmt.Errorf("missing value in template data")
		} else {
			if _, ok := m["value"].(string); !ok {
				return nil, fmt.Errorf("expected string for v, got %T", m["value"])
			}
			sv = m["value"].(string)
		}
		d = append(d, dottedConfigTemplateKV{key: sk, value: sv})
	}
	return d, nil
}

func (t *TemplateComponent) generateDottedConfig(dct dottedConfigTemplate, userdata map[string]any) (tmpl.DottedConfig, error) {
	// we have to fill in the template with the default values
	// and the values from the properties
	config := make(tmpl.DottedConfig)
	for _, kv := range dct {
		// do the key
		key, err := t.applyTemplate(kv.key, userdata)
		if err != nil {
			return nil, err
		}
		// and then the value
		value, err := t.applyTemplate(kv.value, userdata)
		if err != nil {
			return nil, err
		}
		config[key] = value
	}
	return config, nil
}
