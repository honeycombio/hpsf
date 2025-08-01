package config

import (
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
