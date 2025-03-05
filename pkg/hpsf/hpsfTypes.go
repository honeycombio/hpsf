package hpsf

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/honeycombio/hpsf/pkg/validator"
	y "gopkg.in/yaml.v3"
)

// DefaultConfiguration is the default HPSF configuration that includes a
// simple Refinery configuration with a deterministic sampler
// and a Collector Nop receiver and exporter.
const DefaultConfiguration = `
components:
  - name: DefaultDeterministicSampler
    kind: DeterministicSampler
    properties:
      - name: SampleRate
        value: 1
        type: int
  - name: DefaultNopReceiver
    kind: NopReceiver
  - name: DefaultNopExporter
    kind: NopExporter
connections:
  - source:
      component: DefaultNopReceiver
      port: Traces
      type: OTelTraces
    destination:
      component: DefaultNopExporter
      port: Traces
      type: OTelTraces
`

type ConnectionType string

const (
	CTYPE_UNKNOWN ConnectionType = "unknown"
	CTYPE_TRACES  ConnectionType = "OTelTraces"
	CTYPE_LOGS    ConnectionType = "OTelLogs"
	CTYPE_METRICS ConnectionType = "OTelMetrics"
	CTYPE_EVENT   ConnectionType = "OTelEvent"
	CTYPE_HONEY   ConnectionType = "Honeycomb"
	CTYPE_NUMBER  ConnectionType = "number"
	CTYPE_STRING  ConnectionType = "string"
	CTYPE_BOOL    ConnectionType = "bool"
)

type PropType string

const (
	PTYPE_INT    PropType = "int"
	PTYPE_FLOAT  PropType = "float"
	PTYPE_STRING PropType = "string"
	PTYPE_BOOL   PropType = "bool"
	PTYPE_ARRSTR PropType = "stringarray"
	PTYPE_MAPSTR PropType = "map" // map[string]any
)

func (p PropType) Validate() error {
	switch p {
	case PTYPE_INT:
	case PTYPE_FLOAT:
	case PTYPE_STRING:
	case PTYPE_BOOL:
	case PTYPE_ARRSTR:
	case PTYPE_MAPSTR:
	default:
		return errors.New("invalid PropType '" + string(p) + "'")
	}
	return nil
}

func (p PropType) ValueConforms(a any) error {
	// null proptype means anything goes
	if p == "" {
		return nil
	}
	switch p {
	case PTYPE_INT:
		if _, ok := a.(int); !ok {
			return errors.New("expected int, got " + fmt.Sprint(a))
		}
	case PTYPE_FLOAT:
		if _, ok := a.(float64); !ok {
			return errors.New("expected float, got " + fmt.Sprint(a))
		}
	case PTYPE_STRING:
		if _, ok := a.(string); !ok {
			return errors.New("expected string, got " + fmt.Sprint(a))
		}
	case PTYPE_BOOL:
		if _, ok := a.(bool); !ok {
			return errors.New("expected bool, got " + fmt.Sprint(a))
		}
	case PTYPE_ARRSTR:
		if _, ok := a.([]string); !ok {
			return errors.New("expected []string, got " + fmt.Sprint(a))
		}
	case PTYPE_MAPSTR:
		if _, ok := a.(map[string]any); !ok {
			return errors.New("expected map[string]any, got " + fmt.Sprint(a))
		}
	default:
		return errors.New("invalid PropType '" + string(p) + "'")
	}
	return nil
}

type Direction string

const (
	DIR_INPUT  Direction = "input"
	DIR_OUTPUT Direction = "output"
)

type Port struct {
	Name      string         `yaml:"name"`
	Direction string         `yaml:"direction"`
	Type      ConnectionType `yaml:"type"`
}

type Property struct {
	Name  string   `yaml:"name"`
	Value any      `yaml:"value"`
	Type  PropType `yaml:"type"`
}

type Component struct {
	Name       string     `yaml:"name"`
	Kind       string     `yaml:"kind"`
	Ports      []Port     `yaml:"ports,omitempty"`
	Properties []Property `yaml:"properties,omitempty"`
}

func (c *Component) Validate() []error {
	results := []error{}
	if c.Name == "" {
		results = append(results, validator.NewError("Component Name must be set"))
	}
	if c.Kind == "" {
		results = append(results, validator.NewErrorf("Component %s Kind must be set", c.Name))
	}
	for _, p := range c.Ports {
		if p.Direction != string(DIR_INPUT) && p.Direction != string(DIR_OUTPUT) {
			results = append(results, validator.NewErrorf(
				"Component %s Port %s Direction must be 'Input' or 'Output'", c.Name, p.Name))
		}
	}
	for _, p := range c.Properties {
		if p.Name == "" {
			results = append(results, validator.NewErrorf("Component %s Property Name must be set", c.Name))
		}
		if p.Value == nil {
			results = append(results, validator.NewErrorf("Component %s Property %s Value must be set", c.Name, p.Name))
		}
		if p.Type != "" {
			if err := p.Type.Validate(); err != nil {
				results = append(results, validator.NewErrorf("Component %s Property %s Type %s", c.Name, p.Name, err))
			}
		}

		switch p.Value.(type) {
		case string:
		case int:
		case float64:
		case bool:
		case map[string]any:
		case []any:
			sa := make([]string, len(p.Value.([]any)))
			for i, v := range p.Value.([]any) {
				if _, ok := v.(string); !ok {
					results = append(results, validator.NewErrorf("Component %s Property %s Value must be a string, number, bool, []any, or map[string]any", c.Name, p.Name))
				}
				sa[i] = v.(string)
			}
			p.Value = sa
		default:
			results = append(results, validator.NewErrorf("Component %s Property %s Value must be a string, number, bool, []any, or map[string]any", c.Name, p.Name))
		}

		err := p.Type.ValueConforms(p.Value)
		if err != nil {
			results = append(results, validator.NewErrorf("Component %s Property %s Value %s", c.Name, p.Name, err))
		}
	}
	return results
}

func safeName(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return re.ReplaceAllString(s, "_")
}

// Returns the safe name of the component (no spaces or special characters)
// This has potential to cause a problem if the resulting name is not unique -- so uniqueness
// should be tested with this name, not the original name.
// we replace any runs of characters not in [a-zA-Z0-9] with an underscore
func (c *Component) GetSafeName() string {
	return safeName(c.Name)
}

// returns the port with the given name, or nil if not found
func (c *Component) GetPort(name string) *Port {
	for _, p := range c.Ports {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// returns the property with the given name, or nil if not found
func (c *Component) GetProperty(name string) *Property {
	for _, p := range c.Properties {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// returns all specified property names as a slice of strings
func (c *Component) GetPropertyNames() []string {
	props := make([]string, len(c.Properties))
	for i, p := range c.Properties {
		props[i] = p.Name
	}
	return props
}

type ConnectionPort struct {
	Component string         `yaml:"component"`
	PortName  string         `yaml:"port"`
	Type      ConnectionType `yaml:"type"`
}

func (cp *ConnectionPort) GetSafeName() string {
	return safeName(cp.Component)
}

func (cp *ConnectionPort) Validate() []error {
	results := []error{}
	if cp.Component == "" {
		results = append(results, validator.NewError("ConnectionPort Component must be set"))
	}
	if cp.PortName == "" {
		results = append(results, validator.NewError("ConnectionPort PortName must be set"))
	}
	if cp.Type == "" {
		results = append(results, validator.NewError("ConnectionPort Type must be set"))
	}
	return results
}

type Connection struct {
	Source      ConnectionPort `yaml:"source"`
	Destination ConnectionPort `yaml:"destination"`
}

func (c *Connection) Validate() []error {
	results := []error{}
	e := c.Source.Validate()
	results = append(results, e...)
	e = c.Destination.Validate()
	results = append(results, e...)
	return results
}

type PublicPort struct {
	Name      string `yaml:"name"`
	Component string `yaml:"component"`
	Port      string `yaml:"port"`
}

func (pp *PublicPort) Validate() []error {
	results := []error{}
	if pp.Name == "" {
		results = append(results, validator.NewError("PublicPort Name must be set"))
	}
	if pp.Component == "" {
		results = append(results, validator.NewError("PublicPort Component must be set"))
	}
	if pp.Port == "" {
		results = append(results, validator.NewError("PublicPort Port must be set"))
	}
	return results
}

type PublicProp struct {
	Name      string `yaml:"name"`
	Component string `yaml:"component"`
	Property  string `yaml:"property"`
}

func (pp *PublicProp) Validate() []error {
	results := []error{}
	if pp.Name == "" {
		results = append(results, validator.NewError("PublicProp Name must be set"))
	}
	if pp.Component == "" {
		results = append(results, validator.NewError("PublicProp Component must be set"))
	}
	if pp.Property == "" {
		results = append(results, validator.NewError("PublicProp Property must be set"))
	}
	return results
}

type Container struct {
	Name       string       `yaml:"name"`
	Components []Component  `yaml:"components,omitempty"`
	Ports      []PublicPort `yaml:"ports,omitempty"`
	Props      []PublicProp `yaml:"props,omitempty"`
}

func (c *Container) Validate() []error {
	results := []error{}
	if c.Name == "" {
		results = append(results, validator.NewError("Container Name must be set"))
	}
	for _, p := range c.Ports {
		e := p.Validate()
		results = append(results, e...)
	}
	for _, p := range c.Props {
		e := p.Validate()
		results = append(results, e...)
	}
	return results
}

// placeholder for where we'll store layout information later
type Layout map[string]any

type HPSF struct {
	Kind        string        `yaml:"kind"`
	Version     string        `yaml:"version"`
	Name        string        `yaml:"name"`
	Summary     string        `yaml:"summary"`
	Description string        `yaml:"description"`
	Components  []Component   `yaml:"components,omitempty"`
	Connections []*Connection `yaml:"connections,omitempty"`
	Containers  []Container   `yaml:"containers,omitempty"`
	Layout      Layout        `yaml:"layout,omitempty"`
}

// use reflect to generate a list of valid yaml tags in a pointer to
// a struct
func getValidKeys(p any) []string {
	keys := []string{}
	v := reflect.ValueOf(p).Elem()
	for i := range v.NumField() {
		f := v.Type().Field(i)
		yamltag := f.Tag.Get("yaml")
		if yamltag != "" {
			// ignore any options like "omitempty"
			if strings.Contains(yamltag, ",") {
				yamltag = strings.Split(yamltag, ",")[0]
			}
			keys = append(keys, yamltag)
		}
	}
	return keys
}

func (h *HPSF) Validate() []error {
	results := []error{}

	if h.Components == nil && h.Connections == nil && h.Containers == nil && h.Layout == nil {
		results = append(results, errors.New("empty and default HPSF structs are considered invalid"))
	}

	for _, c := range h.Components {
		e := c.Validate()
		results = append(results, e...)
	}
	for _, c := range h.Connections {
		e := c.Validate()
		results = append(results, e...)
	}
	for _, c := range h.Containers {
		e := c.Validate()
		results = append(results, e...)
	}
	return results
}

// EnsureHPSFYAML returns an error if the input is not HPSF yaml or invalid HPSF
func EnsureHPSFYAML(input string) error {
	m, err := validator.EnsureYAML([]byte(input))
	if err != nil {
		return err
	}
	// it has to have at least one key
	if len(m) == 0 {
		return errors.New("HPSF yaml is empty")
	}

	// check to see if it has only expected top-level keys
	// (it would be interesting to do this recursively someday, but it's a lot)
	keys := getValidKeys(&HPSF{})
	badkeys := make([]string, 0)
	for k := range m {
		if !slices.Contains(keys, k) {
			badkeys = append(badkeys, k)
		}
	}
	if len(badkeys) > 0 {
		return errors.New("HPSF yaml contains unexpected keys: " + strings.Join(badkeys, ", "))
	}

	var hpsf HPSF
	dec := y.NewDecoder(strings.NewReader(input))
	err = dec.Decode(&hpsf)
	if err != nil {
		return err
	}
	validations := hpsf.Validate()
	if len(validations) != 0 {
		v := make([]string, len(validations))
		for i, e := range validations {
			v[i] = e.Error()
		}
		return errors.New("HPSF validation failed: " + strings.Join(v, ", "))
	}
	return nil
}
