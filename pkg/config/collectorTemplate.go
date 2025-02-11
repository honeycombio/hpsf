package config

import (
	"fmt"
)

// A collectorTemplate implements a template for a collector component. The component
// should be marked with "format: collector" in the template file and it has more
// specific fields than a dottedConfigTemplate.

type collectorTemplate struct {
	componentSection       string
	signalTypes            []string
	collectorComponentName string
	kvs                    map[string][]dottedConfigTemplateKV
}

// getKV is a helper function to extract a key-value pair from a map[string]any.
// It returns the kv and a boolean indicating if the extraction was successful.
func getKV(d any) (*dottedConfigTemplateKV, bool) {
	kv := &dottedConfigTemplateKV{}
	m, ok := d.(map[string]any)
	if !ok {
		return kv, false
	}

	// keys must be strings
	if mk, ok := m["key"]; !ok {
		return kv, false
	} else {
		if _, ok := mk.(string); !ok {
			return kv, false
		}
		kv.key = mk.(string)
	}

	// values can be strings, ints, bools, []string, map[string]string, []any, or map[string]any
	if _, ok := m["value"]; !ok {
		return kv, false
	} else {
		switch val := m["value"].(type) {
		case string:
			kv.value = val
		case int:
			kv.value = val
		case bool:
			kv.value = val
		case []string:
			kv.value = val
		case map[string]string:
			kv.value = val
		case []any:
			sl := make([]string, len(val))
			for _, v := range val {
				if _, ok := v.(string); !ok {
					return kv, false
				}
				sl = append(sl, v.(string))
			}
			kv.value = sl
		case map[string]any:
			mp := make(map[string]string)
			for k, v := range val {
				if _, ok := v.(string); !ok {
					return kv, false
				}
				mp[k] = v.(string)
			}
			kv.value = mp
		default:
			return kv, false
		}
	}

	// suppress_if specifies a condition under which the key-value pair should be suppressed
	// if it evaluates to 'true' then it's suppressed
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
		var ok bool
		switch mk {
		case "componentSection":
			c.componentSection, ok = mv.(string)
			if !ok {
				return c, fmt.Errorf("expected string for componentSection, got %T", mv)
			}
		case "signalType":
			// we can take one or many signalTypes
			st, ok := mv.(string)
			if !ok {
				return c, fmt.Errorf("expected string for signalType, got %T", mv)
			}
			c.signalTypes = []string{st}
		case "signalTypes":
			sts, ok := mv.([]any)
			if !ok {
				return c, fmt.Errorf("expected array for signalTypes, got %T", mv)
			}
			for _, st := range sts {
				if _, ok := st.(string); !ok {
					return c, fmt.Errorf("expected string for signalType, got %T", st)
				}
				c.signalTypes = append(c.signalTypes, st.(string))
			}
		case "collectorComponentName":
			c.collectorComponentName, ok = mv.(string)
			if !ok {
				return c, fmt.Errorf("expected string for collectorComponentName, got %T", mv)
			}
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
