package config

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	"github.com/honeycombio/hpsf/pkg/yaml"
)

type RefineryInputComponent struct {
	Component   hpsf.Component
	Connections []*hpsf.Connection
}

// ensure RefineryInputComponent implements Component
var _ Component = (*RefineryInputComponent)(nil)

func (c *RefineryInputComponent) GenerateConfig(ct hpsftypes.Type, pipeline hpsf.PathWithConnections, userdata map[string]any) (tmpl.TemplateConfig, error) {
	if ct != RefineryConfigType {
		return nil, nil
	}
	if c.Component.Properties == nil {
		return nil, nil
	}

	port := c.Component.GetProperty("Port")
	if port == nil {
		return nil, nil
	}
	pstr := yaml.AsString(port.Value)

	switch c.Component.Kind {
	case "HoneycombExporter":
		return tmpl.DottedConfig{
			"GRPCServerParameters.Enabled":    true,
			"GRPCServerParameters.ListenAddr": "0.0.0.0:" + pstr,
		}, nil
	case "RefineryHTTP":
		return tmpl.DottedConfig{
			"GRPCServerParameters.Enabled": true,
			"Network.ListenAddr":           "0.0.0.0:" + pstr,
		}, nil
	default:
		return nil, fmt.Errorf("unknown refinery input component: %s", c.Component.Name)
	}
}

func (c *RefineryInputComponent) AddConnection(conn *hpsf.Connection) {
	c.Connections = append(c.Connections, conn)
}

type DeterministicSampler struct {
	Component hpsf.Component
}
