package translator

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// A Translator is responsible for translating an HPSF document into a
// collection of components, and then further rendering those into configuration
// files.
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
		tc.SetHPSF(component)
		return &tc, nil
	}

	// fall back to the base components
	switch component.Kind {
	case "TraceGRPC", "TraceHTTP", "LogGRPC", "LogHTTP", "RefineryGRPC", "RefineryHTTP":
		return NewInput(component)
	default:
		return nil, fmt.Errorf("unknown component kind: %s", component.Kind)
	}
}

func (t *Translator) GenerateConfig(h *hpsf.HPSF, ct config.Type, userdata map[string]any) (tmpl.TemplateConfig, error) {
	comps := make(map[string]config.Component)
	// make all the components
	for _, c := range h.Components {
		comp, err := t.MakeConfigComponent(c)
		if err != nil {
			return nil, err
		}
		comps[c.Name] = comp
	}

	// now add the connections
	for _, conn := range h.Connections {
		comp, ok := comps[conn.Source.Component]
		if !ok {
			return nil, fmt.Errorf("unknown source component %s in connection", conn.Source.Component)
		}
		comp.AddConnection(conn)

		comp, ok = comps[conn.Destination.Component]
		if !ok {
			return nil, fmt.Errorf("unknown target component %s in connection", conn.Destination.Component)
		}
		comp.AddConnection(conn)
	}

	// Start with a base component so we always have a valid config
	dummy := hpsf.Component{Name: "dummy", Kind: "dummy"}
	base := config.GenericBaseComponent{Component: dummy}
	composite, err := base.GenerateConfig(ct, userdata)
	if err != nil {
		return nil, err
	}

	// merge in the config from each of the components
	for _, comp := range comps {
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
