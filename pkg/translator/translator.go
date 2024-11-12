package translator

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/yaml"
)

type Translator struct {
}

func NewTranslator() *Translator {
	return &Translator{}
}

func (t *Translator) MakeConfigComponent(component hpsf.Component) (config.Component, error) {
	switch component.Kind {
	case "TraceGRPC", "TraceHTTP", "LogGRPC", "LogHTTP", "RefineryGRPC", "RefineryHTTP":
		return NewInput(component)
	case "EMAThroughputSampler":
		return config.NullComponent{}, nil
	case "DeterministicSampler":
		return config.DeterministicSampler{Component: component}, nil
	case "HoneycombExporter":
		return config.NullComponent{}, nil
	default:
		return nil, fmt.Errorf("unknown component kind: %s", component.Kind)
	}
}

func (t *Translator) GenerateConfig(h *hpsf.HPSF, ct config.Type) (yaml.DottedConfig, error) {
	composite := yaml.DottedConfig{}
	for _, c := range h.Components {
		comp, err := t.MakeConfigComponent(c)
		if err != nil {
			return nil, err
		}
		refineryConfig, err := comp.GenerateConfig(ct)
		if err != nil {
			return nil, err
		}
		if refineryConfig != nil {
			composite.Merge(refineryConfig)
		}
	}
	return composite, nil
}
