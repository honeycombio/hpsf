package config

import "github.com/honeycombio/hpsf/pkg/config/tmpl"

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
	GenerateConfig(Type, map[string]any) (tmpl.TemplateConfig, error)
}

type NullComponent struct{}

func (c NullComponent) GenerateConfig(Type, map[string]any) (tmpl.TemplateConfig, error) {
	return nil, nil
}
