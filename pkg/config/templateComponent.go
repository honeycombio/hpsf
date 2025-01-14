package config

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/hpsf"
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

// A TemplateData describes a template for generating configuration data. It's a
// deliberately simple structure, with a kind (which is the type of
// configuration data it generates), a name (which is used to identify the
// template), a format (which is the format of the data), meta (a map of extra
// component-level info) and the data itself.
type TemplateData struct {
	Kind   Type
	Name   string
	Format string
	Meta   map[string]string
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
	hpsf        *hpsf.Component    // the component from the hpsf document
}

// SetHPSF stores the original component's details and may modify their contents. To
// prevent the original being modified, the argument here should never be changed to a pointer.
func (t *TemplateComponent) SetHPSF(c hpsf.Component) {
	t.hpsf = &c
}

// HProps is a template helper that gets a map of all properties specified in the hpsf document.
func (t *TemplateComponent) HProps() map[string]any {
	props := make(map[string]any)
	if t.hpsf != nil {
		for _, name := range t.hpsf.GetPropertyNames() {
			p := t.hpsf.GetProperty(name)
			props[p.Name] = p.Value
		}
	}
	return props
}

// helper for templates
func (t *TemplateComponent) Props() map[string]TemplateProperty {
	props := make(map[string]TemplateProperty)
	for _, prop := range t.Properties {
		props[prop.Name] = prop
	}
	return props
}

func (t *TemplateComponent) ComponentName() string {
	return t.Name
}

// // ensure that TemplateComponent implements Component
var _ Component = (*TemplateComponent)(nil)

func (t *TemplateComponent) GenerateConfig(cfgType Type, userdata map[string]any) (tmpl.TemplateConfig, error) {
	// we have to find a template with the kind of the config; if it
	// doesn't exist, we return an error
	for _, template := range t.Templates {
		if template.Kind == cfgType {
			switch template.Format {
			case "dotted":
				dct, err := buildDottedConfigTemplate(template.Data)
				if err != nil {
					return nil, fmt.Errorf("error %w building dotted config template named %s", err, t.Name)
				}
				return t.generateDottedConfig(dct, userdata)
			case "collector":
				ct, err := buildCollectorTemplate(template)
				if err != nil {
					return nil, fmt.Errorf("error %w building collector template named %s", err, t.Name)
				}
				return t.generateCollectorConfig(ct, userdata)
			default:
				return nil, fmt.Errorf("unknown template format %s", template.Format)
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
		return "", fmt.Errorf("error executing template: %w", err)
	}
	return b.String(), nil
}

func (t *TemplateComponent) generateCollectorConfig(ct collectorTemplate, userdata map[string]any) (*tmpl.CollectorConfig, error) {
	// we have to fill in the template with the default values
	// and the values from the properties
	config := tmpl.NewCollectorConfig()
	sectionOrder := []string{"receivers", "processors", "exporters", "extensions"}
	for _, section := range sectionOrder {
		svcKey := fmt.Sprintf("pipelines.%s.%s", ct.signalType, section)
		for _, kv := range ct.kvs[section] {
			config.Set("service", svcKey, []string{ct.collectorComponentName})
			key, err := t.applyTemplate(kv.key, userdata)
			if err != nil {
				return nil, err
			}
			key = fmt.Sprintf("%s.%s", section, key)
			value, err := t.applyTemplate(kv.value, userdata)
			if err != nil {
				return nil, err
			}
			config.Set(section, key, value)
		}
	}
	return config, nil
}
