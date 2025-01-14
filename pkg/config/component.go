package config

import (
	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/hpsf"
)

type Type string

const (
	RefineryConfigType  Type = "refinery_config"
	RefineryRulesType   Type = "refinery_rules"
	CollectorConfigType Type = "collector_config"
)

// The Component interface is implemented by all components.
// If one of these functions returns nil, nil, it means
// that the component has no impact on that particular system.
// Component key names are dotted paths, e.g. "a.b.c", and
// the values are any valid YAML value.
// We will need to convert the dotted paths into real ones later.
type Component interface {
	GenerateConfig(cfgType Type, userdata map[string]any) (tmpl.TemplateConfig, error)
}

type NullComponent struct{}

func (c NullComponent) GenerateConfig(Type, map[string]any) (tmpl.TemplateConfig, error) {
	return nil, nil
}

// This base component is used to make sure that the config will be valid
// even if it stands alone. This is likely to be a temporary solution until we have a
// database of components.
type GenericBaseComponent struct {
	Component hpsf.Component
}

func (c GenericBaseComponent) GenerateConfig(ct Type, userdata map[string]any) (tmpl.TemplateConfig, error) {
	switch ct {
	case RefineryConfigType:
		return tmpl.DottedConfig{
			"General.ConfigurationVersion": 2,
			"General.MinRefineryVersion":   "v2.0",
		}, nil
	case RefineryRulesType:
		return tmpl.DottedConfig{
			"RulesVersion": 2,
		}, nil
	case CollectorConfigType:
		return tmpl.NewCollectorConfig(), nil
	default:
		return nil, nil
	}
}
