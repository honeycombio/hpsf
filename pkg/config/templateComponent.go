package config

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
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
// value. The advanced flag can be used to indicate that the property should be
// suppressed by default in the UI (only shown if the user selects an "advanced"
// option). We also allow for validations, which can be used to constrain the
// value of the property. The property can also have a summary and a
// description, which are used to document the property.
type TemplateProperty struct {
	Name        string        `yaml:"name"`
	Summary     string        `yaml:"summary,omitempty"`
	Description string        `yaml:"description,omitempty"`
	Type        hpsf.PropType `yaml:"type"`
	Advanced    bool          `yaml:"advanced,omitempty"`
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
	Meta   map[string]any
	Data   []any
}

// A TemplateComponent is a component that can be described with a template.
// We're hoping that most components will be described this way, so that we
// can store most templates in a database and not have to change the code when
// we add new components.
type TemplateComponent struct {
	Name        string             `yaml:"name"`
	CollName    string             `yaml:"collName"`
	Kind        string             `yaml:"kind"`
	Summary     string             `yaml:"summary,omitempty"`
	Description string             `yaml:"description,omitempty"`
	Metadata    map[string]string  `yaml:"metadata,omitempty"`
	Ports       []TemplatePort     `yaml:"ports,omitempty"`
	Properties  []TemplateProperty `yaml:"properties,omitempty"`
	Templates   []TemplateData     `yaml:"templates,omitempty"`
	User        map[string]any     `yaml:"user,omitempty"`
	hpsf        *hpsf.Component    // the component from the hpsf document
	connections []*hpsf.Connection `yaml:"connections,omitempty"`
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
	if t.CollName != "" {
		return t.CollName + "/" + t.hpsf.Name
	}
	return t.Name
}

func (t *TemplateComponent) ConnectsUsingAppropriateType(signalType string) bool {
	typeMapping := map[string]hpsf.ConnectionType{
		"traces":  hpsf.CTYPE_TRACES,
		"logs":    hpsf.CTYPE_LOGS,
		"metrics": hpsf.CTYPE_METRICS,
	}
	connType := typeMapping[signalType]
	for _, conn := range t.connections {
		if conn.Source.Type == connType || conn.Destination.Type == connType {
			return true
		}
	}
	return false
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

func (t *TemplateComponent) AddConnection(conn *hpsf.Connection) {
	t.connections = append(t.connections, conn)
}

func (t *TemplateComponent) expandTemplateVariable(tmplText string, userdata map[string]any) (string, error) {
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

// undecorate removes type decorations from strings and returns the desired type.
// Since everything that comes out of a template is a string, for things that
// needed to not be strings, we flagged them with a decoration indicating the
// desired type. Now we need to do some extra work to make sure that we return
// the indicated type. If it can't be converted to the desired type, we return
// the string as is.
func undecorate(s string) any {
	switch {
	case strings.HasPrefix(s, IntPrefix):
		s := strings.TrimPrefix(s, IntPrefix)
		i, err := strconv.Atoi(s)
		if err == nil {
			return i
		}
	case strings.HasPrefix(s, BoolPrefix):
		s := strings.TrimPrefix(s, BoolPrefix)
		b, err := strconv.ParseBool(s)
		if err == nil {
			return b
		}
	case strings.HasPrefix(s, FloatPrefix):
		s := strings.TrimPrefix(s, FloatPrefix)
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f
		}
	case strings.HasPrefix(s, ArrPrefix):
		s := strings.TrimPrefix(s, ArrPrefix)
		items := strings.Split(s, FieldSeparator)
		// we need to trim the spaces from the items and we don't want blanks
		// in the array
		var arr []string
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item != "" {
				arr = append(arr, item)
			}
		}
		return arr
	case strings.HasPrefix(s, MapPrefix):
		s := strings.TrimPrefix(s, MapPrefix)
		result := make(map[string]string)
		items := strings.Split(s, RecordSeparator)
		// the last item is always blank, so < 2 is what we want
		if len(items) < 2 {
			return result
		}
		items = items[:len(items)-1]
		for _, i := range items {
			sp := strings.Split(i, FieldSeparator)
			result[sp[0]] = sp[1]
		}
		return result
	}
	return s
}

func (t *TemplateComponent) applyTemplate(tmplVal any, userdata map[string]any) (any, error) {
	switch k := tmplVal.(type) {
	case string:
		if tmplVal == "" || !strings.Contains(k, "{{") {
			return k, nil
		}

		// expand the template, but before we return it we need to check if
		// it needs extra handling
		value, err := t.expandTemplateVariable(k, userdata)
		if err != nil {
			return nil, err
		}

		result := undecorate(value)
		return result, nil
	default:
		return "", fmt.Errorf("invalid templated variable type %T", k)
	}
}

// this is where we do the actual work of generating the config; this thing knows about
// the structure of the collector config and how to fill it in
func (t *TemplateComponent) generateCollectorConfig(ct collectorTemplate, userdata map[string]any) (*tmpl.CollectorConfig, error) {
	// we have to fill in the template with the default values
	// and the values from the properties
	t.CollName = ct.collectorComponentName
	config := tmpl.NewCollectorConfig()
	sectionOrder := []string{"receivers", "processors", "exporters", "extensions"}
	for _, section := range sectionOrder {
		for _, signalType := range []string{"traces", "logs", "metrics"} {
			// if the signal type is not in the list of signal types for this collector, skip it
			if !slices.Contains(ct.signalTypes, signalType) {
				continue
			}
			// if this template doesn't have a connection for this signal type, skip it
			if !t.ConnectsUsingAppropriateType(signalType) {
				continue
			}
			svcKey := fmt.Sprintf("pipelines.%s.%s", signalType, section)
			for _, kv := range ct.kvs[section] {
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
				config.Set("service", svcKey, []string{t.ComponentName()})
				key, err := t.expandTemplateVariable(kv.key, userdata)
				if err != nil {
					return nil, err
				}
				value, err := t.applyTemplate(kv.value, userdata)
				if err != nil {
					return nil, err
				}
				config.Set(section, key, value)
			}
		}
	}
	return config, nil
}
