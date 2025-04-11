package translator

import (
	"fmt"
	"slices"

	"maps"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/validator"
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

// GetComponents returns the components installed in the translator.
func (t *Translator) GetComponents() map[string]config.TemplateComponent {
	return t.components
}

// GetTemplates returns the templates installed in the translator.
func (t *Translator) GetTemplates() map[string]hpsf.HPSF {
	return t.templates
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
		// found it, manufacture a new instance of the component
		tc.SetHPSF(component)
		return &tc, nil
	}

	// nothing found so we're done
	return nil, fmt.Errorf("unknown component kind: %s", component.Kind)
}

// ValidateConfig validates the configuration of the HPSF document as it stands with respect to the
// components and templates installed in the translator.
// Note that it returns a validation.Result so that the errors can be collected and reported in a
// structured way. This allows for multiple validation errors to be returned at once, rather than
// stopping at the first error. This is useful for providing feedback to users on multiple issues
// in their configuration.
func (t *Translator) ValidateConfig(h *hpsf.HPSF) error {
	if h == nil {
		return fmt.Errorf("nil HPSF document provided for validation")
	}

	// We assume that the HPSF document has already been validated for syntax and structure since
	// it's already in hpsf format. Our goal here is to make sure that the components and templates
	// can be used to generate a valid configuration. This means checking that all components referenced
	// in the HPSF document are available in the translator's component map and that they can be instantiated
	// correctly, and that all the properties are of the correct type.
	result := validator.NewResult("HPSF document validation failed")
	templateComps := make(map[string]config.TemplateComponent)
	// make all the components
	for _, c := range h.Components {
		err := c.Validate()
		if err != nil {
			// if the component itself is invalid, add the error to the result
			// this means the component itself has some issues
			result.Add(fmt.Errorf("failed to validate component %s: %w", c.Name, err))
			// continue to process other components, since we want to validate all of them
			// before returning an error. This allows us to collect all the errors in one pass.
		}

		if comp, ok := t.components[c.Kind]; ok {
			templateComps[c.GetSafeName()] = comp
		} else {
			result.Add(fmt.Errorf("failed to locate corresponding template component for %s: %w", c.Name, err))
		}
	}
	if !result.IsEmpty() {
		// if we have errors at this point, return early
		// this means we couldn't even instantiate the components
		// so there's no point in continuing to validate the connections
		return result
	}

	// now we have a map of all the components that were successfully instantiated
	// so we can iterate the properties and validate them according to the validations specified in the template components
	for _, c := range h.Components {
		tmpl, ok := templateComps[c.GetSafeName()]
		if !ok {
			// If we don't have a template component for this component, it
			// means we couldn't instantiate it. We caught this earlier, so we
			// should never get here. Just continue.
			continue
		}

		// Get the template properties from the template component.
		tprops := tmpl.Props()
		for _, prop := range c.Properties {
			// validate each property against the template component's basic validation rules
			tp, ok := tprops[prop.Name]
			if !ok {
				// If the property is not found in the template component's
				// properties, something's messed up. This means the property is
				// not defined in the template component.
				return hpsf.NewError("property not found in template component").
					WithComponent(c.Name).
					WithProperty(prop.Name)
			}

			// Now validate the property against the template property's validation rules.
			if err := tp.Validate(prop); err != nil {
				// if the property fails validation, add the error to the result
				// this means the property itself has some issues
				// we want to include the component name and property name in the error message for clarity
				herr := hpsf.NewError("failed to validate property").
					WithCause(err).
					WithComponent(c.Name).
					WithProperty(prop.Name)
				result.Add(herr)
			}
		}
	}

	return result.ErrOrNil()
}

func (t *Translator) GenerateConfig(h *hpsf.HPSF, ct config.Type, userdata map[string]any) (tmpl.TemplateConfig, error) {
	// we need to make sure that there is a sampler in the config to produce a valid refinery rules config
	t.maybeAddDefaultSampler(h)

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
		compConfig, err := comp.GenerateConfig(ct, userdata)
		if err != nil {
			return nil, err
		}
		if compConfig != nil {
			composite.Merge(compConfig)
		}
	}
	return composite, nil
}

func (t *Translator) maybeAddDefaultSampler(h *hpsf.HPSF) {
	foundDefaultSampler := slices.ContainsFunc(h.Components, func(c hpsf.Component) bool {
		if component, ok := t.components[c.Kind]; ok {
			if component.Style != "sampler" {
				return false
			}
			p := c.GetProperty("Environment")
			if p != nil {
				return p.Value == "__default__"
			}
			return slices.ContainsFunc(component.Properties, func(p config.TemplateProperty) bool {
				return p.Name == "Environment" && p.Default == "__default__"
			})
		}
		return false
	})
	if !foundDefaultSampler {
		h.Components = append(h.Components, hpsf.Component{
			Name: "defaultSampler",
			Kind: "DeterministicSampler",
			Properties: []hpsf.Property{
				{
					Name:  "SampleRate",
					Value: 1,
				},
			},
		})
	}
}
