package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	y "gopkg.in/yaml.v3"
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

// Index is an optional field that can be used to indicate that this port
// should be treated as an indexed port (e.g. for use in a pipeline). The
// index values for ports should be sequential within a given component, and
// should start at 1. If the index is not specified, it is assumed to be 0.
type TemplatePort struct {
	Name      string              `yaml:"name"`
	Direction string              `yaml:"direction"`
	Type      hpsf.ConnectionType `yaml:"type"`
	Index     int                 `yaml:"index,omitempty"`
	Note      string              `yaml:"note,omitempty"`
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

type ComponentType string

const (
	ComponentTypeBase     ComponentType = "BASE"
	ComponentTypeMeta     ComponentType = "META"
	ComponentTypeTemplate ComponentType = "TEMPLATE"
)

// we need to be able to unmarshal the component style and status from YAML
// and marshal it back to YAML, and the all-caps nature of the DB constants
// is jarring and doesn't fit with the rest of the yaml styling. So we
// can marshal YAML with some case conversions.
// ensure ComponentStyle implements yaml.Marshaler and yaml.Unmarshaler
var _ y.Marshaler = (*ComponentType)(nil)
var _ y.Unmarshaler = (*ComponentType)(nil)

func (c *ComponentType) UnmarshalYAML(value *y.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	cs := ComponentType(strings.ToUpper(s))
	switch cs {
	case ComponentTypeBase, ComponentTypeMeta, ComponentTypeTemplate:
		*c = cs
		return nil
	default:
		return fmt.Errorf("invalid component style: %s", s)
	}
}

func (c ComponentType) MarshalYAML() (any, error) {
	return strings.ToLower(string(c)), nil
}

type ComponentStatus string

const (
	ComponentStatusAlpha       ComponentStatus = "ALPHA"
	ComponentStatusArchived    ComponentStatus = "ARCHIVED"
	ComponentStatusDeprecated  ComponentStatus = "DEPRECATED"
	ComponentStatusDevelopment ComponentStatus = "DEVELOPMENT"
	ComponentStatusStable      ComponentStatus = "STABLE"
)

var _ y.Marshaler = (*ComponentStatus)(nil)
var _ y.Unmarshaler = (*ComponentStatus)(nil)

func (c *ComponentStatus) UnmarshalYAML(value *y.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	cs := ComponentStatus(strings.ToUpper(s))
	switch cs {
	case ComponentStatusAlpha, ComponentStatusArchived, ComponentStatusDeprecated,
		ComponentStatusDevelopment, ComponentStatusStable:
		*c = cs
		return nil
	default:
		return fmt.Errorf("invalid component status: %s", s)
	}
}

func (c ComponentStatus) MarshalYAML() (any, error) {
	return strings.ToLower(string(c)), nil
}

// A TemplateComponent is a component that can be described with a template.
// We're hoping that most components will be described this way, so that we
// can store most templates in a database and not have to change the code when
// we add new components.
//
// A few notes on the fields:
//   - Kind is the kind of component, e.g. "TraceGRPC", "LogHTTP", etc. The combination of
//     Kind and Version is used to uniquely identify a component.
//   - Version is the version of the component as a semver string.
//   - Name is the name of the component. In a templateComponent, it is used to suggest a name that the
//     end user might want to call the component. It is not used to identify the component in a template,
//     but is used to identify the component in the UI.
//   - Style is currently a string, will be used to help the frontend figure out how to display the component.
//     It will likely become some sort of enum, but for now we don't know what the values will be.
//   - Type is the generalized type of component for broad classification - Base, Meta, or Template.
//   - Status is the development status of the component.
//   - User is only used for templating, but it needs to be exported, so its yaml tag is set to "-"
//   - collName is the name of the OTel collector component that this component is associated with; it may
//     be empty if the component is not associated with a collector. We need to store it in this data type
//     because it's used in the template rendering, but it's not part of the component itself (it's specified
//     in the template metadata).
type TemplateComponent struct {
	Kind        string             `yaml:"kind"`
	Version     string             `yaml:"version"`
	Name        string             `yaml:"name"`
	Logo        string             `yaml:"logo,omitempty"`
	Summary     string             `yaml:"summary,omitempty"`
	Description string             `yaml:"description,omitempty"`
	Comment     string             `yaml:"comment,omitempty"`
	Tags        []string           `yaml:"tags,omitempty"`
	Type        ComponentType      `yaml:"type,omitempty"`
	Style       string             `yaml:"style,omitempty"`
	Status      ComponentStatus    `yaml:"status,omitempty"`
	Metadata    map[string]string  `yaml:"metadata,omitempty"`
	Ports       []TemplatePort     `yaml:"ports,omitempty"`
	Properties  []TemplateProperty `yaml:"properties,omitempty"`
	Templates   []TemplateData     `yaml:"templates,omitempty"`
	User        map[string]any     `yaml:"-"`
	hpsf        *hpsf.Component    // the component from the hpsf document
	connections []*hpsf.Connection
	collName    string
}

// SetHPSF stores the original component's details and may modify their contents. To
// prevent the original being modified, the argument here should never be changed to a pointer.
func (t *TemplateComponent) SetHPSF(c *hpsf.Component) {
	t.hpsf = c
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

// Props is a helper for templates that gets all properties in a template
// component as a map. It's mainly used to look up property defaults.
func (t *TemplateComponent) Props() map[string]TemplateProperty {
	props := make(map[string]TemplateProperty)
	for _, prop := range t.Properties {
		props[prop.Name] = prop
	}
	return props
}

// Values is a helper for templates to get the data values that are available in
// the component including those specified as defaults and the environment. This
// composes HProps, User, and Properties into a single map that can be used in
// templates. You can still use them individually for special cases.
func (t *TemplateComponent) Values() map[string]any {
	result := make(map[string]any)
	for k, v := range t.HProps() {
		if !_isZeroValue(v) {
			// we only want to include non-zero values in the result
			result[k] = v
		}
	}
	for k, v := range t.User {
		if !_isZeroValue(v) {
			// don't overwrite existing values
			if _, exists := result[k]; exists {
				continue
			}
			result[k] = v
		}
	}
	for _, prop := range t.Properties {
		// don't overwrite existing values
		if _, exists := result[prop.Name]; exists {
			continue
		}
		result[prop.Name] = prop.Default
	}
	return result
}

func (t *TemplateComponent) ComponentName() string {
	if t.collName != "" {
		return t.collName + "/" + t.hpsf.GetSafeName()
	}
	return t.Name
}

// Tests if this component has a connection with this signal type
func (t *TemplateComponent) ConnectsUsingAppropriateType(connType hpsf.ConnectionType) bool {
	for _, conn := range t.connections {
		if conn.Source.Type == connType || conn.Destination.Type == connType {
			return true
		}
	}
	return false
}

func (t *TemplateComponent) GetPort(name string) *TemplatePort {
	// returns the port with the given name, or nil if it doesn't exist
	for _, port := range t.Ports {
		if port.Name == name {
			return &port
		}
	}
	return nil
}

func (t *TemplateComponent) GetPortIndex(name string) int {
	// Returns the index for a given port name, or 0 if it's unspecified.
	// This implies that indices start at 1.
	for _, port := range t.Ports {
		if port.Name == name {
			return port.Index
		}
	}
	return 0
}

// // ensure that TemplateComponent implements Component
var _ Component = (*TemplateComponent)(nil)

func (t *TemplateComponent) GenerateConfig(cfgType Type, pipeline hpsf.PipelineWithConnections, userdata map[string]any) (tmpl.TemplateConfig, error) {
	// we have to find a template with the kind of the config; if it
	// doesn't exist, we return an error

	// we might have more than one template for the same config type,
	// so we need to generate all of them and return the merged result.
	generatedTemplates := make([]tmpl.TemplateConfig, 0)

	for _, template := range t.Templates {
		if template.Kind == cfgType {
			switch template.Format {
			case "dotted":
				dct, err := buildDottedConfigTemplate(template.Data)
				if err != nil {
					return nil, fmt.Errorf("error %w building dotted config template for %s",
						err, t.Kind)
				}
				tmpl, err := t.generateDottedConfig(dct, userdata)
				if err != nil {
					return nil, err
				}
				generatedTemplates = append(generatedTemplates, tmpl)
			case "collector":
				ct, err := buildCollectorTemplate(template)
				if err != nil {
					return nil, fmt.Errorf("error %w building collector template for %s",
						err, t.Kind)
				}
				tmpl, err := t.generateCollectorConfig(ct, pipeline, userdata)
				if err != nil {
					return nil, err
				}
				generatedTemplates = append(generatedTemplates, tmpl)
			case "rules":
				// a rules template expects the metadata to include environment
				// information.
				if pipeline.ConnType != hpsf.CTYPE_SAMPLE {
					continue // rules templates are only for sampling pipelines
				}
				rt, err := buildRulesTemplate(template)
				if err != nil {
					return nil, fmt.Errorf("error %w building rules template for %s",
						err, t.Kind)
				}

				// We need to know which pipeline we're generating rules for, so we look up the
				// connection leading to this component.
				conn := pipeline.GetConnectionLeadingTo(t.hpsf.GetSafeName())
				index := 0
				if conn != nil {
					// we have a connection, so we can use its source port name to look up an index
					index = t.GetPortIndex(conn.Source.GetSafeName())
				}

				rmt, err := tmpl.RMTFromStyle(t.Style)
				if err != nil {
					return nil, fmt.Errorf("error %w getting RulesComponentType from style %s for %s",
						err, t.Style, t.Kind)
				}
				tmpl, err := t.generateRulesConfig(rt, rmt, index, userdata)
				if err != nil {
					return nil, err
				}
				generatedTemplates = append(generatedTemplates, tmpl)
			default:
				return nil, fmt.Errorf("unknown template format %s", template.Format)
			}
		}
	}

	if len(generatedTemplates) == 0 {
		return nil, nil // no templates found for this config type
	}

	ret := generatedTemplates[0]
	for _, tmpl := range generatedTemplates[1:] {
		if err := ret.Merge(tmpl); err != nil {
			return nil, fmt.Errorf("failed to merge template: %w", err)
		}
	}

	return ret, nil
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
// These decorations were placed there by the functions in helpers.go.
// Since everything that comes out of a Go template is a string, for things that
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
		// s is encoded as a JSON map, so we need to decode it
		var m map[string]any
		// we ignore the error here because the input string
		// was marshalled by us and we know it's valid JSON,
		// and there's nothing we can do with it anyway.
		json.Unmarshal([]byte(s), &m)
		return m
	}
	return s
}

func (t *TemplateComponent) applyTemplate(tmplVal any, userdata map[string]any) (any, error) {
	switch k := tmplVal.(type) {
	case string:
		if !strings.Contains(k, "{{") {
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
	// right now this is dealing with nop receiver/exporter case
	case map[string]string:
		return k, nil
	case []string:
		for i, v := range k {
			// we need to expand the template for each value in the array
			expanded, err := t.expandTemplateVariable(v, userdata)
			if err != nil {
				return nil, err
			}
			k[i] = expanded
		}
		return k, nil
	case int:
		return k, nil
	case float64:
		return k, nil
	case bool:
		return k, nil
	default:
		return "", fmt.Errorf("invalid templated variable type %T", k)
	}
}

// this is where we do the actual work of generating the config; this thing knows about
// the structure of the collector config and how to fill it in
func (t *TemplateComponent) generateCollectorConfig(ct collectorTemplate, pipeline hpsf.PipelineWithConnections, userdata map[string]any) (*tmpl.CollectorConfig, error) {
	// we have to fill in the template with the default values
	// and the values from the properties
	t.collName = ct.collectorComponentName
	config := tmpl.NewCollectorConfig()
	sectionOrder := []string{"receivers", "processors", "exporters", "extensions"}
	for _, section := range sectionOrder {
		for _, signalType := range hpsf.CollectorSignalTypes {
			if pipeline.ConnType != signalType {
				continue // skip this signal type if it doesn't match the pipeline
			}
			// if this template doesn't have a connection for this signal type, skip it
			if !t.ConnectsUsingAppropriateType(signalType) {
				continue
			}
			svcKey := fmt.Sprintf("pipelines.%s/%s.%s", signalType.AsCollectorSignalType(), pipeline.GetID(), section)
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

func (t *TemplateComponent) AsYAML() (string, error) {
	// this is a mechanism to marshal the template component to YAML
	data, err := y.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("error marshalling template component to YAML: %w", err)
	}
	return string(data), nil
}
