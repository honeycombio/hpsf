package hpsf

import (
	"errors"
	"slices"
	"strings"

	"github.com/honeycombio/hpsf/pkg/validator"
	y "gopkg.in/yaml.v3"
)

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
	PTYPE_NUMBER PropType = "number"
	PTYPE_STRING PropType = "string"
	PTYPE_BOOL   PropType = "bool"
)

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
		switch p.Value.(type) {
		case string:
		case int:
		case bool:
		default:
			results = append(results, validator.NewErrorf("Component %s Property %s Value must be a string, number, or bool", c.Name, p.Name))
		}
	}
	return results
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
	Components  []Component   `yaml:"components,omitempty"`
	Connections []*Connection `yaml:"connections,omitempty"`
	Containers  []Container   `yaml:"containers,omitempty"`
	Layout      Layout        `yaml:"layout,omitempty"`
}

func (h *HPSF) Validate() []error {
	results := []error{}

	if h.Components == nil && h.Connections == nil && h.Containers == nil && h.Layout == nil {
		results = append(results, errors.New("default HPSF structs are considered invalid"))
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

// EnsureHPSF returns an error if the input is not HPSF yaml or invalid HPSF
func EnsureHPSF(input string) error {
	m, err := validator.EnsureYAML([]byte(input))
	if err != nil {
		return err
	}
	// it has to have at least one key
	if len(m) == 0 {
		return errors.New("HPSF yaml is empty")
	}

	// check to see if it has only expected keys
	keys := []string{"components", "connections", "containers", "layout"}
	for k := range m {
		if !slices.Contains(keys, k) {
			return errors.New("HPSF yaml contains unexpected keys")
		}
	}

	var hpsf HPSF
	dec := y.NewDecoder(strings.NewReader(input))
	err = dec.Decode(&hpsf)
	if err != nil {
		return err
	}
	if len(hpsf.Validate()) != 0 {
		return errors.New("HPSF validation failed")
	}
	return nil
}
