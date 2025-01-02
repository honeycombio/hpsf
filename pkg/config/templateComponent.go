package config

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/yaml"
)

// This is the Go support for components read as data.
// Right now we read that data from an embedded filesystem, but
// the hope is to replace the filesystem with a database later.

// The premise here is that for each component, we have a block of YAML
// that describes it. This YAML not only contains the component's public
// list of properties, defaults, and documentation, but also contains
// Go templates that can be used to generate the actual configurations.

// A TemplatePort describes a "port" on a component. A port is a place for
// data to flow in or out of the component. A port can be either an input
// or an output, and has a datatype; at least for now, ports of different
// types cannot be connected to each other.
type TemplatePort struct {
	Name      string              `yaml:"name"`
	Direction string              `yaml:"direction"`
	Type      hpsf.ConnectionType `yaml:"type"`
	Note      string              `yaml:"note,omitempty"`
}

// A TemplateProperty describes a property of a component. A property is a
// user-settable value that can be used to configure the component. Properties
// have a name, a type (which can be used to validate the value), and a default
// value. We also allow for validations, which can be used to constrain the
// value of the property. The property can also have a summary and a
// description, which are used to document the property.
type TemplateProperty struct {
	Name        string        `yaml:"name"`
	Summary     string        `yaml:"summary,omitempty"`
	Description string        `yaml:"description,omitempty"`
	Type        hpsf.PropType `yaml:"type"`
	Validations []string      `yaml:"validation,omitempty"`
	Default     any           `yaml:"default,omitempty"`
}

// A TemplateData describes a template for generating configuration data.
// It's a deliberately simple structure, with a kind (which is the type of
// configuration data it generates), a name (which is used to identify the
// template), a format (which is the format of the data), and the data itself.
type TemplateData struct {
	Kind   Type
	Name   string
	Format string
	Data   []any
}

// A TemplateComponent is a component that can be described with a template.
// We're hoping that most components will be described this way, so that we
// can store most templates in a database and not have to change the code when
// we add new components.
type TemplateComponent struct {
	Name        string             `yaml:"name"`
	Kind        string             `yaml:"kind"`
	Summary     string             `yaml:"summary,omitempty"`
	Description string             `yaml:"description,omitempty"`
	Ports       []TemplatePort     `yaml:"ports,omitempty"`
	Properties  []TemplateProperty `yaml:"properties,omitempty"`
	Templates   []TemplateData     `yaml:"templates,omitempty"`
	User        map[string]any     `yaml:"user,omitempty"`
}

// helper for templates
func (t *TemplateComponent) Props() map[string]TemplateProperty {
	props := make(map[string]TemplateProperty)
	for _, prop := range t.Properties {
		props[prop.Name] = prop
	}
	return props
}

type dottedConfigTemplateKV struct {
	key   string
	value string
}

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

// // ensure that TemplateComponent implements Component
var _ Component = (*TemplateComponent)(nil)

func (t *TemplateComponent) GenerateConfig(cfgType Type, userdata map[string]any) (yaml.DottedConfig, error) {
	// we have find a template with the kind of the config; if it
	// doesn't exist, we return nil, nil
	for _, template := range t.Templates {
		if template.Kind == cfgType {
			switch template.Format {
			case "dottedConfig":
				dct, err := buildDottedConfigTemplate(template.Data)
				if err != nil {
					return nil, fmt.Errorf("error %w building dotted config template named %s", err, t.Name)
				}
				return t.generateDottedConfig(dct, userdata)
			}
		}
	}
	return nil, nil
}

func (t *TemplateComponent) applyTemplate(tmplText string, userdata map[string]any) (string, error) {
	if tmplText == "" || !strings.Contains(tmplText, "{{") {
		return tmplText, nil
	}
	tmpl, err := template.New("template").Funcs(helpers()).Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("error %w parsing template", err)
	}

	t.User = userdata
	var b bytes.Buffer
	err = tmpl.Execute(&b, t)
	if err != nil {
		return "", fmt.Errorf("error %w executing template", err)
	}
	return b.String(), nil
}

func (t *TemplateComponent) generateDottedConfig(dct dottedConfigTemplate, userdata map[string]any) (yaml.DottedConfig, error) {
	// we have to fill in the template with the default values
	// and the values from the properties
	config := make(yaml.DottedConfig)
	for _, kv := range dct {
		key, err := t.applyTemplate(kv.key, userdata)
		if err != nil {
			return nil, err
		}
		value, err := t.applyTemplate(kv.value, userdata)
		if err != nil {
			return nil, err
		}
		config[key] = value
	}
	return config, nil
}
