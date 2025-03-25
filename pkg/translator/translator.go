package translator

import (
	"fmt"

	"maps"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// A Translator is responsible for translating an HPSF document into a
// collection of components, and then further rendering those into configuration
// files.
type Translator struct {
	components map[string]config.TemplateComponent
	templates  map[string]hpsf.HPSF
}

// Deprecated: use NewEmptyTranslator and InstallComponents instead
func NewTranslator() (*Translator, error) {
	tr := &Translator{}
	// autoload the template components because we don't want to break existing code
	err := tr.LoadEmbeddedComponents()
	return tr, err
}

// Creates a translator with no components loaded.
func NewEmptyTranslator() *Translator {
	tr := &Translator{
		components: make(map[string]config.TemplateComponent),
		templates:  make(map[string]hpsf.HPSF),
	}
	return tr
}

// InstallComponents installs the given components into the translator.
func (t *Translator) InstallComponents(components map[string]config.TemplateComponent) {
	maps.Copy(t.components, components)
}

// InstallTemplates installs the given templates into the translator.
func (t *Translator) InstallTemplates(components map[string]hpsf.HPSF) {
	maps.Copy(t.templates, components)
}

// Loads the embedded components into the translator.
// Deprecated: use InstallComponents instead
func (t *Translator) LoadEmbeddedComponents() error {
	// load the embedded components
	tcs, err := data.LoadEmbeddedComponents()
	if err != nil {
		return err
	}
	maps.Copy(t.components, tcs)
	return nil
}

func (t *Translator) MakeConfigComponent(component hpsf.Component) (config.Component, error) {
	// first look in the template components
	tc, ok := t.components[component.Kind]
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
		comps[c.GetSafeName()] = comp
	}

	// now add the connections
	for _, conn := range h.Connections {
		comp, ok := comps[conn.Source.GetSafeName()]
		if !ok {
			return nil, fmt.Errorf("unknown source component %s in connection", conn.Source.Component)
		}
		comp.AddConnection(conn)

		comp, ok = comps[conn.Destination.GetSafeName()]
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
