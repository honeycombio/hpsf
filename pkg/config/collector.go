package config

import (
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/yaml"
)

// This base component is used to make sure that the config will be valid
// even if it stands alone. This is likely to be a temporary solution until we have a
// database of components.
type CollectorBaseComponent struct {
	Component hpsf.Component
}

var _ Component = CollectorBaseComponent{}

func (c CollectorBaseComponent) GenerateConfig(ct Type, userdata map[string]any) (yaml.DottedConfig, error) {
	return yaml.DottedConfig{
		"processors": map[string]any{},
		"receivers":  map[string]any{},
		"extensions": map[string]any{},
		"service":    map[string]any{},
	}, nil
}
