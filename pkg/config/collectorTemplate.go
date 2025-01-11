package config

import (
	"fmt"
)

// A collectorTemplate implements a template for a collector component. The component
// should be marked with "format: collector" in the template file and it has more
// specific fields than a dottedConfigTemplate.

type collectorTemplate struct {
	componentSection       string
	signalType             string
	collectorComponentName string
	kvs                    map[string][]dottedConfigTemplateKV
}

func getKV(d any) (*dottedConfigTemplateKV, bool) {
	kv := &dottedConfigTemplateKV{}
	m, ok := d.(map[string]any)
	if !ok {
		return kv, false
	}
	if mk, ok := m["key"]; !ok {
		return kv, false
	} else {
		if _, ok := mk.(string); !ok {
			return kv, false
		}
		kv.key = mk.(string)
	}
	if _, ok := m["value"]; !ok {
		return kv, false
	} else {
		if _, ok := m["value"].(string); !ok {
			return kv, false
		}
		kv.value = m["value"].(string)
	}
	if _, ok := m["suppress_if"]; ok {
		if _, ok := m["suppress_if"].(string); !ok {
			return kv, false
		}
		kv.suppress_if = m["suppress_if"].(string)
	}
	return kv, true
}

func buildCollectorTemplate(t TemplateData) (collectorTemplate, error) {
	c := collectorTemplate{kvs: make(map[string][]dottedConfigTemplateKV)}

	for mk, mv := range t.Meta {
		switch mk {
		case "componentSection":
			c.componentSection = mv
		case "signalType":
			c.signalType = mv
		case "collectorComponentName":
			c.collectorComponentName = mv
		default:
			return c, fmt.Errorf("unknown meta key %s", mk)
		}
	}
	if c.componentSection == "" {
		return c, fmt.Errorf("missing componentSection in meta")
	}

	for _, d := range t.Data {
		kv, ok := getKV(d)
		if !ok {
			return c, fmt.Errorf("expected map for data, got %T", d)
		}
		c.kvs[c.componentSection] = append(c.kvs[c.componentSection], *kv)
	}
	return c, nil
}
