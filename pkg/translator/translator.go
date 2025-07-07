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

// getMatchingTemplateComponents returns the template components that match the components in the HPSF document.
// It validates components before matching them and returns an error if any components are invalid.
func (t *Translator) getMatchingTemplateComponents(h *hpsf.HPSF) (map[string]config.TemplateComponent, validator.Result) {
	result := validator.NewResult("HPSF component fetch failed")
	templateComps := make(map[string]config.TemplateComponent)
	for _, c := range h.Components {
		err := c.Validate()
		if err != nil {
			result.Add(fmt.Errorf("failed to validate component %s: %w", c.Name, err))
			continue
		}
		if comp, ok := t.components[c.Kind]; ok {
			templateComps[c.GetSafeName()] = comp
		} else {
			result.Add(fmt.Errorf("failed to locate corresponding template component for %s: %w", c.Name, err))
		}
	}
	return templateComps, result
}

func (t *Translator) validateProperties(h *hpsf.HPSF, templateComps map[string]config.TemplateComponent) validator.Result {
	result := validator.NewResult("HPSF property validation errors")
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
		templateProperties := tmpl.Props()
		componentProps := make(map[string]hpsf.Property)
		for _, prop := range c.Properties {
			componentProps[prop.Name] = prop
			_, found := templateProperties[prop.Name]
			if !found {
				// If the property is not found in the template component's
				// properties, something's messed up. This means the property is
				// not defined in the template component.
				err := hpsf.NewError("property not found in template component").
					WithComponent(c.Name).
					WithProperty(prop.Name)
				result.Add(err)
			}
		}

		for _, prop := range templateProperties {
			// validate each property against the template component's basic validation rules
			suppliedProperty, propertyFound := componentProps[prop.Name]
			if !propertyFound {
				// If the property is not supplied, use the default value from the template component.
				suppliedProperty.Value = prop.Default
			}

			// Now validate the property against the template property's validation rules.
			if validateError := prop.Validate(suppliedProperty); validateError != nil {
				// if the property fails validation, add the error to the result
				// this means the property itself has some issues
				// we want to include the component name and property name in the error message for clarity
				hspfError := hpsf.NewError("failed to validate property").
					WithCause(validateError).
					WithComponent(c.Name).
					WithProperty(prop.Name)
				result.Add(hspfError)
			}
		}
	}
	return result
}

// this checks that there is exactly one connection on the input and output of each sampler
func (t *Translator) validateSamplerConnections(h *hpsf.HPSF, templateComps map[string]config.TemplateComponent) validator.Result {
	result := validator.NewResult("HPSF sampler connection validation errors")
	// iterate over the components and check for samplers
	for _, c := range h.Components {
		tmpl, ok := templateComps[c.GetSafeName()]
		if !ok {
			// If we don't have a template component for this component, it
			// means we couldn't instantiate it. We caught this earlier, so we
			// should never get here. Just continue.
			continue
		}

		if tmpl.Style == "sampler" {
			// check the connections for the sampler
			inputs := 0
			outputs := 0
			for _, conn := range h.Connections {
				if conn.Destination.GetSafeName() == c.GetSafeName() {
					inputs++
				}
				if conn.Source.GetSafeName() == c.GetSafeName() {
					outputs++
				}
			}
			if inputs != 1 || outputs != 1 {
				err := hpsf.NewError("sampler must have exactly one input and one output connection").
					WithComponent(c.Name)
				result.Add(err)
			}
		}
	}
	return result
}

// validateConnectionPorts checks that all connections have valid ports. The name on the connection
// in hpsf must match the port name on the template component.
func (t *Translator) validateConnectionPorts(h *hpsf.HPSF, templateComps map[string]config.TemplateComponent) validator.Result {
	result := validator.NewResult("HPSF connection port validation errors")
	// iterate over the connections and check that the source and destination components have the
	// specified ports. This is a sanity check to ensure that the connections are valid.
	for _, conn := range h.Connections {
		srcComp, ok := templateComps[conn.Source.GetSafeName()]
		if !ok {
			continue
		}

		if srcComp.GetPort(conn.Source.PortName) == nil {
			err := hpsf.NewErrorf("source component does not have a port called %s", conn.Source.PortName).
				WithComponent(conn.Source.Component)
			result.Add(err)
		}

		dstComp, ok := templateComps[conn.Destination.GetSafeName()]
		if !ok {
			continue
		}

		if dstComp.GetPort(conn.Destination.PortName) == nil {
			err := hpsf.NewErrorf("destination component does not have a port called %s", conn.Destination.PortName).
				WithComponent(conn.Destination.Component)
			result.Add(err)
		}
	}
	return result
}

// validateStartSampling checks that there is at most one component with the "startsampling" style. If it exists,
// there must be at least one sampler in the configuration. If it does not exist, there can be no samplers in the configuration.
// There must also be only one connection from the StartSampling component to a sampler.
func (t *Translator) validateStartSampling(h *hpsf.HPSF, templateComps map[string]config.TemplateComponent) validator.Result {
	result := validator.NewResult("HPSF start sampling validation errors")
	// iterate over the components and check for the StartSampling component
	// startSamplingCount := 0
	// var startSamplingComp string
	// for _, c := range h.Components {
	// 	tmpl, ok := templateComps[c.GetSafeName()]
	// 	if !ok {
	// 		continue
	// 	}

	// 	if tmpl.Style == "startsampling" {
	// 		startSamplingCount++
	// 		startSamplingComp = c.GetSafeName()
	// 		if startSamplingCount > 1 {
	// 			err := hpsf.NewError("only one StartSampling component is allowed").
	// 				WithComponent(c.Name)
	// 			result.Add(err)
	// 		}
	// 	}
	// }
	// if startSamplingCount == 0 {
	// 	// if there is no StartSampling component, we cannot have any samplers in the configuration
	// 	for _, c := range h.Components {
	// 		tmpl, ok := templateComps[c.GetSafeName()]
	// 		if !ok {
	// 			continue
	// 		}

	// 		if tmpl.Style == "sampler" {
	// 			err := hpsf.NewError("if there is no StartSampling component, no samplers are allowed").
	// 				WithComponent(c.Name)
	// 			result.Add(err)
	// 		}
	// 	}
	// } else {
	// 	// if there is a StartSampling component, we must have at least one sampler in the configuration
	// 	hasSampler := false
	// 	for _, c := range h.Components {
	// 		tmpl, ok := templateComps[c.GetSafeName()]
	// 		if !ok {
	// 			continue
	// 		}

	// 		if tmpl.Style == "sampler" {
	// 			hasSampler = true
	// 			break
	// 		}
	// 	}
	// 	if !hasSampler {
	// 		err := hpsf.NewError("if there is a StartSampling component, at least one sampler is required").
	// 			WithComponent("StartSampling")
	// 		result.Add(err)
	// 	}
	// }
	// // now we need to check that there is only one connection from the StartSampling component to a sampler
	// if startSamplingCount == 1 {
	// 	samplerConnections := 0
	// 	for _, conn := range h.Connections {
	// 		if conn.Source.GetSafeName() == startSamplingComp {
	// 			// check if the destination is a sampler
	// 			dstComp, ok := templateComps[conn.Destination.GetSafeName()]
	// 			if !ok {
	// 				continue
	// 			}
	// 			if dstComp.Style == "sampler" {
	// 				samplerConnections++
	// 			}
	// 		}
	// 	}
	// 	if samplerConnections != 1 {
	// 		err := hpsf.NewError("StartSampling component must have exactly one connection to a sampler").
	// 			WithComponent(startSamplingComp)
	// 		result.Add(err)
	// 	}
	// }

	return result
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

	// if we don't pass basic validation, we can't continue
	err := h.Validate()
	if err != nil {
		return err
	}

	// We assume that the HPSF document has already been validated for syntax and structure since
	// it's already in hpsf format. Our goal here is to make sure that the components and templates
	// can be used to generate a valid configuration. This means checking that all components referenced
	// in the HPSF document are available in the translator's component map and that they can be instantiated
	// correctly, and that all the properties are of the correct type.
	templateComps, result := t.getMatchingTemplateComponents(h)
	if !result.IsEmpty() {
		// if we have errors at this point, return early
		// this means we couldn't even instantiate the components
		// so there's no point in continuing to validate the connections
		return result
	}

	result.Add(t.validateProperties(h, templateComps))
	result.Add(t.validateConnectionPorts(h, templateComps))
	result.Add(t.validateStartSampling(h, templateComps))
	result.Add(t.validateSamplerConnections(h, templateComps))

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
		pipelines = []hpsf.PipelineWithConnections{
			{Pipeline: h.Components, ConnType: hpsf.CTYPE_LOGS},
			{Pipeline: h.Components, ConnType: hpsf.CTYPE_METRICS},
			{Pipeline: h.Components, ConnType: hpsf.CTYPE_TRACES},
			{Pipeline: h.Components, ConnType: hpsf.CTYPE_HONEY},
			{Pipeline: h.Components, ConnType: hpsf.CTYPE_SAMPLE},
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

		mergedSomething := false
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
				if err := composite.Merge(compConfig); err != nil {
					return nil, fmt.Errorf("failed to merge component config: %w", err)
				}
				mergedSomething = true
			}
		}
		if mergedSomething {
			composites = append(composites, composite)
		}
	}
	// If we have multiple pipelines, we need to merge them into a single config.
	if len(composites) > 1 {
		// We can use the Merge method to combine all the configurations into one.
		finalConfig := composites[0]
		for _, comp := range composites[1:] {
			if err := finalConfig.Merge(comp); err != nil {
				return nil, fmt.Errorf("failed to merge pipeline configs: %w", err)
			}
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
