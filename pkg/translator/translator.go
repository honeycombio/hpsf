package translator

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/yaml"
)

type Translator struct {
	templateComponents map[string]config.TemplateComponent
}

func NewTranslator() (*Translator, error) {
	tcs, err := config.LoadTemplateComponents()
	if err != nil {
		return nil, err
	}
	return &Translator{templateComponents: tcs}, nil
}

func (t *Translator) MakeConfigComponent(component hpsf.Component) (config.Component, error) {
	// first look in the template components
	tc, ok := t.templateComponents[component.Kind]
	if ok {
		return &tc, nil
	}

	// fall back to the base components
	switch component.Kind {
	case "TraceGRPC", "TraceHTTP", "LogGRPC", "LogHTTP", "RefineryGRPC", "RefineryHTTP":
		return NewInput(component)
	case "DeterministicSampler":
		return config.DeterministicSampler{Component: component}, nil
	default:
		return nil, fmt.Errorf("unknown component kind: %s", component.Kind)
	}
}

func (t *Translator) GenerateConfig(h *hpsf.HPSF, ct config.Type, userdata map[string]any) (yaml.DottedConfig, error) {
	composite := yaml.DottedConfig{}

	// Add base component to the config so we can make a valid config
	// this may be temporary until we have a database of components
	dummy := hpsf.Component{Name: "dummy", Kind: "dummy"}
	var base config.Component
	switch ct {
	case config.RefineryConfigType, config.RefineryRulesType:
		base = config.RefineryBaseComponent{Component: dummy}
	case config.CollectorConfigType:
		base = config.CollectorBaseComponent{Component: dummy}
	}

	cfg, err := base.GenerateConfig(ct, userdata)
	if err != nil {
		return nil, err
	}
	composite.Merge(cfg)

	for _, c := range h.Components {
		comp, err := t.MakeConfigComponent(c)
		if err != nil {
			return nil, err
		}
		refineryConfig, err := comp.GenerateConfig(ct, userdata)
		if err != nil {
			return nil, err
		}
		if refineryConfig != nil {
			composite.Merge(refineryConfig)
		}
	}
	return composite, nil
}
