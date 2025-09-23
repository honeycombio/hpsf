package config

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
)

// The Component interface is implemented by all components.
// If one of these functions returns nil, nil, it means
// that the component has no impact on that particular system.
// Component key names are dotted paths, e.g. "a.b.c", and
// the values are any valid YAML value.
// We will need to convert the dotted paths into real ones later.
// The pipeline identifies which pipeline is being generated.
type Component interface {
	GenerateConfig(cfgType hpsftypes.Type, pipeline hpsf.PathWithConnections, userdata map[string]any) (tmpl.TemplateConfig, error)
	AddConnection(*hpsf.Connection)
}

type NullComponent struct{}

func NewNullComponent() *NullComponent {
	return &NullComponent{}
}

// ensure that NullComponent implements Component
var _ Component = (*NullComponent)(nil)

func (c *NullComponent) GenerateConfig(hpsftypes.Type, hpsf.PathWithConnections, map[string]any) (tmpl.TemplateConfig, error) {
	return nil, nil
}

func (c *NullComponent) AddConnection(*hpsf.Connection) {}

// This base component is used to make sure that the config will be valid
// even if it stands alone. This is likely to be a temporary solution until we have a
// database of components.
type GenericBaseComponent struct {
	Component   hpsf.Component
	Connections []*hpsf.Connection
}

// ensure that GenericBaseComponent implements Component
var _ Component = (*GenericBaseComponent)(nil)

func (c GenericBaseComponent) GenerateConfig(ct hpsftypes.Type, pipeline hpsf.PathWithConnections, userdata map[string]any) (tmpl.TemplateConfig, error) {
	switch ct {
	case hpsftypes.RefineryConfig:
		// DottedConfig is already a map, so we don't need a pointer
		return tmpl.DottedConfig{
			"General.ConfigurationVersion": 2,
			"General.MinRefineryVersion":   "v2.0",
		}, nil
	case hpsftypes.RefineryRules:
		return tmpl.NewRulesConfig(tmpl.Output, nil, nil), nil
	case hpsftypes.CollectorConfig:
		return tmpl.NewCollectorConfig(), nil
	default:
		return nil, nil
	}
}

func (c *GenericBaseComponent) AddConnection(conn *hpsf.Connection) {
	c.Connections = append(c.Connections, conn)
}

// UnconfiguredComponent is used when the user has not added
// any components to the configuration yet. It provides just
// the basic configuration needed to start artifacts.
type UnconfiguredComponent struct {
	Component   hpsf.Component
	Connections []*hpsf.Connection
}

// ensure that UnconfiguredRefineryComponent implements Component
var _ Component = (*UnconfiguredComponent)(nil)

func (c UnconfiguredComponent) GenerateConfig(ct hpsftypes.Type, pipeline hpsf.PathWithConnections, userdata map[string]any) (tmpl.TemplateConfig, error) {
	switch ct {
	case hpsftypes.RefineryConfig:
		// DottedConfig is already a map, so we don't need a pointer
		return tmpl.DottedConfig{
			"General.ConfigurationVersion": 2,
			"General.MinRefineryVersion":   "v2.0",
		}, nil
	case hpsftypes.RefineryRules:
		rules := tmpl.NewRulesConfig(tmpl.Output, nil, nil)
		rules.Samplers["__default__"] = &tmpl.V2SamplerChoice{
			DeterministicSampler: &tmpl.DeterministicSamplerConfig{
				SampleRate: 1,
			},
		}
		return rules, nil
	case hpsftypes.CollectorConfig:
		return tmpl.NewCollectorConfig(), nil
	default:
		return nil, nil
	}
}

func (c *UnconfiguredComponent) AddConnection(conn *hpsf.Connection) {
	c.Connections = append(c.Connections, conn)
}

// MetaComponent represents a logical container that groups related components together.
// It can contain any combination of child components and manages their internal connections
// while presenting a unified interface to external systems.
type MetaComponent struct {
	Component     hpsf.Component
	Children      []Component           // Child components within this meta component
	InternalConns []*hpsf.Connection   // Connections between child components
	ExternalConns []*hpsf.Connection   // Connections to/from external components
}

// ensure that MetaComponent implements Component
var _ Component = (*MetaComponent)(nil)

// NewMetaComponent creates a new MetaComponent with the given component definition
func NewMetaComponent(component hpsf.Component) *MetaComponent {
	return &MetaComponent{
		Component:     component,
		Children:      make([]Component, 0),
		InternalConns: make([]*hpsf.Connection, 0),
		ExternalConns: make([]*hpsf.Connection, 0),
	}
}

// AddChild adds a child component to this meta component
func (m *MetaComponent) AddChild(child Component) {
	m.Children = append(m.Children, child)
}

// AddInternalConnection adds a connection between child components within this meta component
func (m *MetaComponent) AddInternalConnection(conn *hpsf.Connection) {
	m.InternalConns = append(m.InternalConns, conn)
}

// AddConnection adds an external connection to/from this meta component
func (m *MetaComponent) AddConnection(conn *hpsf.Connection) {
	m.ExternalConns = append(m.ExternalConns, conn)
}

// GenerateConfig generates configuration for this meta component by delegating to its children
// and merging their configurations based on the target configuration type
func (m *MetaComponent) GenerateConfig(cfgType hpsftypes.Type, pipeline hpsf.PathWithConnections, userdata map[string]any) (tmpl.TemplateConfig, error) {
	switch cfgType {
	case hpsftypes.RefineryConfig:
		return m.generateRefineryConfig(pipeline, userdata)
	case hpsftypes.RefineryRules:
		return m.generateRefineryRulesConfig(pipeline, userdata)
	case hpsftypes.CollectorConfig:
		return m.generateCollectorConfig(pipeline, userdata)
	default:
		return nil, nil
	}
}

// generateRefineryConfig generates refinery configuration by merging child component configs
func (m *MetaComponent) generateRefineryConfig(pipeline hpsf.PathWithConnections, userdata map[string]any) (tmpl.TemplateConfig, error) {
	// Start with base configuration
	baseConfig := tmpl.DottedConfig{
		"General.ConfigurationVersion": 2,
		"General.MinRefineryVersion":   "v2.0",
	}

	// Merge configurations from all child components
	for _, child := range m.Children {
		childConfig, err := child.GenerateConfig(hpsftypes.RefineryConfig, pipeline, userdata)
		if err != nil {
			return nil, err
		}
		if childConfig != nil {
			err = baseConfig.Merge(childConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	return baseConfig, nil
}

// generateRefineryRulesConfig generates refinery rules by composing child components into rules
func (m *MetaComponent) generateRefineryRulesConfig(pipeline hpsf.PathWithConnections, userdata map[string]any) (tmpl.TemplateConfig, error) {
	// For refinery rules, we need to compose conditions and samplers from child components
	// This is a simplified initial implementation - we'll enhance this based on specific needs

	rulesConfig := tmpl.NewRulesConfig(tmpl.Output, nil, nil)

	// Collect configurations from all child components and merge them
	for _, child := range m.Children {
		childConfig, err := child.GenerateConfig(hpsftypes.RefineryRules, pipeline, userdata)
		if err != nil {
			return nil, err
		}
		if childConfig != nil {
			err = rulesConfig.Merge(childConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	return rulesConfig, nil
}

// generateCollectorConfig generates collector configuration by merging child component configs
func (m *MetaComponent) generateCollectorConfig(pipeline hpsf.PathWithConnections, userdata map[string]any) (tmpl.TemplateConfig, error) {
	// Start with base collector configuration
	collectorConfig := tmpl.NewCollectorConfig()

	// Merge configurations from all child components
	for _, child := range m.Children {
		childConfig, err := child.GenerateConfig(hpsftypes.CollectorConfig, pipeline, userdata)
		if err != nil {
			return nil, err
		}
		if childConfig != nil {
			err = collectorConfig.Merge(childConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	return collectorConfig, nil
}

// Validate validates the MetaComponent and its children
func (m *MetaComponent) Validate() error {
	// Ensure meta component has at least one child
	if len(m.Children) == 0 {
		return fmt.Errorf("meta component %s must have at least one child component", m.Component.Name)
	}

	// Validate each child component if it has validation support
	for i, child := range m.Children {
		// Try to validate child if it supports validation
		if validator, ok := child.(interface{ Validate() error }); ok {
			if err := validator.Validate(); err != nil {
				return fmt.Errorf("child component %d validation failed in meta component %s: %w", i, m.Component.Name, err)
			}
		}
	}

	return nil
}
