package translator

import (
	"fmt"
	"iter"
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

func (t *Translator) MakeConfigComponent(component *hpsf.Component) (config.Component, error) {
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

// OrderedComponentMap is a generic map that maintains the order of insertion.
// It is used to ensure that the order of components and properties is preserved
// when generating the configuration.
type OrderedComponentMap struct {
	// Keys is the list of keys in the order they were added.
	Keys []string
	// Values is the map of keys to values.
	Values map[string]config.Component
}

func NewOrderedComponentMap() *OrderedComponentMap {
	return &OrderedComponentMap{
		Keys:   make([]string, 0),
		Values: make(map[string]config.Component),
	}
}

// Set adds a key-value pair to the ordered map.
func (om *OrderedComponentMap) Set(key string, value config.Component) {
	if _, exists := om.Values[key]; !exists {
		// Only add the key to the Keys slice if it doesn't already exist
		om.Keys = append(om.Keys, key)
	}
	om.Values[key] = value
}

// Get retrieves a value from the ordered map by key.
func (om *OrderedComponentMap) Get(key string) (config.Component, bool) {
	value, exists := om.Values[key]
	return value, exists
}

// Items returns a Go iterable
func (om *OrderedComponentMap) Items() iter.Seq[config.Component] {
	return func(yield func(config.Component) bool) {
		for _, key := range om.Keys {
			if value, exists := om.Values[key]; exists {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func (t *Translator) GenerateConfig(h *hpsf.HPSF, ct config.Type, userdata map[string]any) (tmpl.TemplateConfig, error) {
	// we need to make sure that there is a sampler in the config to produce a valid refinery rules config
	t.maybeAddDefaultSampler(h)

	comps := NewOrderedComponentMap()
	receiverNames := make(map[string]bool)
	// make all the components
	// for _, c := range h.Components {
	visitFunc := func(c *hpsf.Component) error {
		comp, err := t.MakeConfigComponent(c)
		if err != nil {
			return err
		}
		comps.Set(c.GetSafeName(), comp)
		if tc, ok := comp.(*config.TemplateComponent); ok {
			if tc.Style == "receiver" {
				receiverNames[c.GetSafeName()] = true
			}
		}
		return nil
	}

	if err := h.VisitComponents(visitFunc); err != nil {
		return nil, fmt.Errorf("failed to create components: %w", err)
	}

	// now add the connections
	for _, conn := range h.Connections {
		comp, ok := comps.Get(conn.Source.GetSafeName())
		if !ok {
			return nil, fmt.Errorf("unknown source component %s in connection", conn.Source.Component)
		}
		comp.AddConnection(conn)

		comp, ok = comps.Get(conn.Destination.GetSafeName())
		if !ok {
			return nil, fmt.Errorf("unknown target component %s in connection", conn.Destination.Component)
		}
		comp.AddConnection(conn)
	}

	// We need to generate our collection of unique pipelines. A pipeline in
	// this context is the shortest path from a source component to a
	// destination component. We iterate over all starting components (those
	// with no incoming connections) and all ending components (those with no
	// outgoing connections).
	pipelines := h.FindAllPipelines(receiverNames)
	if len(pipelines) == 0 {
		// there were no complete pipelines found, so we construct dummy pipelines with all the components
		// so that all the non-piped components can play
		pipelines = []hpsf.PipelineWithConnectionType{
			{Pipeline: h.Components, ConnType: hpsf.CTYPE_LOGS},
			{Pipeline: h.Components, ConnType: hpsf.CTYPE_METRICS},
			{Pipeline: h.Components, ConnType: hpsf.CTYPE_TRACES},
			{Pipeline: h.Components, ConnType: hpsf.CTYPE_HONEY},
		}
	}

	composites := make([]tmpl.TemplateConfig, 0, len(pipelines))

	// now we can iterate over the pipelines and generate a configuration for each
	for _, pipeline := range pipelines {
		// Start with a base component so we always have a valid config
		dummy := hpsf.Component{Name: "dummy", Kind: "dummy"}
		base := config.GenericBaseComponent{Component: dummy}
		composite, err := base.GenerateConfig(ct, pipeline, userdata)
		if err != nil {
			return nil, err
		}

		for _, comp := range pipeline.Pipeline {
			// look up the component in the ordered map
			c, ok := comps.Get(comp.GetSafeName())
			if !ok {
				return nil, fmt.Errorf("unknown component %s in pipeline", comp.GetSafeName())
			}

			compConfig, err := c.GenerateConfig(ct, pipeline, userdata)
			if err != nil {
				return nil, err
			}
			if compConfig != nil {
				composite.Merge(compConfig)
			}
		}
		composites = append(composites, composite)
	}
	// If we have multiple pipelines, we need to merge them into a single config.
	if len(composites) > 1 {
		// We can use the Merge method to combine all the configurations into one.
		finalConfig := composites[0]
		for _, comp := range composites[1:] {
			finalConfig.Merge(comp)
		}
		return finalConfig, nil
	} else if len(composites) == 1 {
		// If we only have one pipeline, we can return it directly.
		return composites[0], nil
	}
	// If we have no pipelines, we return nil.
	return nil, nil
}

func (t *Translator) maybeAddDefaultSampler(h *hpsf.HPSF) {
	foundDefaultSampler := slices.ContainsFunc(h.Components, func(c *hpsf.Component) bool {
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
		h.Components = append(h.Components, &hpsf.Component{
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
