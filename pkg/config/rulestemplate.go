package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
)

type rulesTemplate struct {
	env     string
	sampler string
	kvs     []dottedConfigTemplateKV
}

type rulesCondition struct {
	index    int
	fields   []string
	op       string
	value    string
	datatype string
}

func (r *rulesCondition) Render(prefix string) (map[string]any, error) {
	// render the condition into a map
	dc := make(map[string]any)
	// we need to inject the index if it's not negative
	c := prefix + "Conditions"
	if r.index >= 0 {
		c = fmt.Sprintf("%sConditions[%d]", prefix, r.index)
	}
	dc[c+".Fields"] = r.fields
	dc[c+".Operator"] = r.op
	if r.op == "exists" || r.op == "not_exists" {
		// if the operator is exists or not_exists, we don't need a value or datatype
		return dc, nil
	}
	dc[c+".Value"] = r.value

	switch r.datatype {
	case "string", "s":
		dc[c+".Datatype"] = "string"
	case "int", "i":
		dc[c+".Datatype"] = "int"
	case "float", "f":
		dc[c+".Datatype"] = "float"
	case "bool", "b":
		dc[c+".Datatype"] = "bool"
	case "":
		// if the datatype is empty, don't set it
	default:
		return nil, fmt.Errorf("unknown datatype %q", r.datatype)
	}
	return dc, nil
}

// the format of a condition is multiple key-value pairs
// separated by semicolons, e.g. "k=v;k2=v2;k3=v3"
// the key-value pairs can be either a single value or a list of values
// separated by commas, e.g. "k=v" or "ks=v1,v2,v3"
func splitCondition(condition any) *rulesCondition {
	// if the value is a string, split it by semicolons
	if s, ok := condition.(string); ok {
		parts := strings.Split(s, ";")
		m := make(map[string]any)
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				continue
			}
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			if k == "fs" {
				m[k] = strings.Split(v, ",")
			} else {
				m[k] = v
			}
		}

		// only the fs case can be a []strings, every other v is a string
		cond := &rulesCondition{index: -1}
		for k, v := range m {
			switch k {
			case "ix":
				// parse the index as an int
				if i, err := strconv.Atoi(v.(string)); err == nil {
					cond.index = i
				} else {
					return nil
				}
			case "f":
				cond.fields = append(cond.fields, v.(string))
			case "fs":
				cond.fields = v.([]string)
			case "o":
				cond.op = v.(string)
			case "v":
				cond.value = v.(string)
			case "d":
				cond.datatype = v.(string)
			default:
				// ignore unknown keys
			}
		}
		return cond
	}
	return nil
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
			return r, fmt.Errorf("unknown meta key %q", mk)
		}
	}

	if r.env == "" {
		return r, fmt.Errorf("missing env in meta")
	}
	if r.sampler == "" {
		return r, fmt.Errorf("missing sampler in meta")
	}

	for _, d := range t.Data {
		kv, ok := getKV(d)
		if !ok {
			return r, fmt.Errorf("expected map for data, got %T", d)
		}

		if ix := strings.Index(kv.key, "!condition!"); ix != -1 {
			// this is a special case for the conditions
			// we need to render them into a dottedConfigTemplateKV

			// first, get the prefix
			prefix := kv.key[:ix]
			cond := splitCondition(kv.value)
			if cond == nil {
				return r, fmt.Errorf("expected string for condition, got %T", kv.value)
			}
			m, err := cond.Render(prefix)
			if err != nil {
				return r, err
			}
			for k, v := range m {
				kv := dottedConfigTemplateKV{
					key:   k,
					value: v,
				}
				r.kvs = append(r.kvs, kv)
			}
		} else {
			r.kvs = append(r.kvs, *kv)
		}
	}
	return r, nil
}

func (t *TemplateComponent) generateRulesConfig(rt *rulesTemplate, userdata map[string]any) (*tmpl.RulesConfig, error) {
	dc := tmpl.NewDottedConfig(nil)

	env, err := t.expandTemplateVariable(rt.env, userdata)
	if err != nil {
		return nil, err
	}
	rt.env = env

	sampler, err := t.expandTemplateVariable(rt.sampler, userdata)
	if err != nil {
		return nil, err
	}
	rt.sampler = sampler

	keyPrefix := "Samplers." + rt.env + "." + rt.sampler + "."
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
			dc[keyPrefix+key] = value
		}
	}
	ec := tmpl.EnvConfig{
		Name:       rt.env,
		ConfigData: dc,
	}
	rc := tmpl.NewRulesConfig()
	rc.Envs = append(rc.Envs, ec)
	return rc, nil
}
