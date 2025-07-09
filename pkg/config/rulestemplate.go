package config

import (
	"fmt"
	"strconv"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
)

type rulesTemplate struct {
	env     string
	sampler string
	kvs     []dottedConfigTemplateKV
}

func buildRulesTemplate(t TemplateData) (*rulesTemplate, error) {
	r := &rulesTemplate{kvs: make([]dottedConfigTemplateKV, 0)}

	for mk, mv := range t.Meta {
		var ok bool
		switch mk {
		case "env":
			r.env, ok = mv.(string)
			if !ok {
				return r, fmt.Errorf("expected string for env, got %T", mv)
			}
		case "sampler":
			r.sampler, ok = mv.(string)
			if !ok {
				return r, fmt.Errorf("expected string for sampler, got %T", mv)
			}
		default:
			// we're going to ignore any other meta keys for now; maybe we can be more strict later
		}
	}

	for _, d := range t.Data {
		kv, ok := getKV(d)
		if !ok {
			return r, fmt.Errorf("expected map for data, got %T", d)
		}

		r.kvs = append(r.kvs, *kv)
	}
	return r, nil
}

// we expand template variables here, but we don't actually apply the kvs yet;
// that's deferred until merge time.
func (t *TemplateComponent) generateRulesConfig(rt *rulesTemplate, compType tmpl.RulesComponentType, pipelineIndex int, userdata map[string]any) (*tmpl.RulesConfig, error) {
	kvs := make(map[string]any)
	meta := make(map[string]string)
	meta[tmpl.MetaPipelineIndex] = strconv.Itoa(pipelineIndex)

	env, err := t.expandTemplateVariable(rt.env, userdata)
	if err != nil {
		return nil, err
	}
	meta[tmpl.MetaEnv] = env

	sampler, err := t.expandTemplateVariable(rt.sampler, userdata)
	if err != nil {
		return nil, err
	}
	meta[tmpl.MetaSampler] = sampler

	for _, kv := range rt.kvs {
		// do the key
		key, err := t.expandTemplateVariable(kv.key, userdata)
		if err != nil {
			return nil, err
		}
		if kv.suppress_if != "" {
			// if the suppress_if condition is met, we skip this key
			condition, err := t.applyTemplate(kv.suppress_if, userdata)
			if err != nil {
				return nil, err
			}
			if condition == "true" {
				continue
			}
		}
		// and then the value
		value, err := t.applyTemplate(kv.value, userdata)
		if err != nil {
			return nil, err
		}
		if kv.value != "" {
			kvs[key] = value
		}
	}
	rc := tmpl.NewRulesConfig(compType, meta, kvs)
	return rc, nil
}
