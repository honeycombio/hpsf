package inspector

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// Inspector provides information about components in HPSF configurations.
// It loads component templates to access default values and properties.
type Inspector struct {
	templates map[string]config.TemplateComponent // kind -> template
}

// NewInspector creates a new Inspector with embedded component templates loaded.
func NewInspector() (*Inspector, error) {
	templates, err := data.LoadEmbeddedComponents()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded components: %w", err)
	}

	return &Inspector{
		templates: templates,
	}, nil
}

// ComponentInfo represents a component (receiver, processor, or exporter) with its name, kind, and properties
type ComponentInfo struct {
	// Name is the user-defined name of the component instance
	Name string
	// Style is the component style (e.g., "receiver", "processor", "exporter")
	Style string
	// Kind is the component kind (e.g., "HoneycombExporter", "OTelReceiver", "MemoryLimiterProcessor")
	Kind string
	// Properties contains component-specific configuration details as key-value pairs
	// Users can access values directly without type casting, e.g. properties["Region"]
	Properties map[string]any
}

// InspectionResult holds information about all components in an HPSF configuration.
type InspectionResult struct {
	Components []ComponentInfo
}

// Exporters returns only the exporter components
func (r InspectionResult) Exporters() []ComponentInfo {
	var exporters []ComponentInfo
	for _, c := range r.Components {
		if c.Style == "exporter" {
			exporters = append(exporters, c)
		}
	}
	return exporters
}

// Receivers returns only the receiver components
func (r InspectionResult) Receivers() []ComponentInfo {
	var receivers []ComponentInfo
	for _, c := range r.Components {
		if c.Style == "receiver" {
			receivers = append(receivers, c)
		}
	}
	return receivers
}

// Processors returns only the processor components
func (r InspectionResult) Processors() []ComponentInfo {
	var processors []ComponentInfo
	for _, c := range r.Components {
		if c.Style == "processor" {
			processors = append(processors, c)
		}
	}
	return processors
}

// Inspect extracts all components from the HPSF document.
// It returns an InspectionResult containing all components.
// Use Exporters(), Receivers(), or Processors() methods to filter by style.
func (i *Inspector) Inspect(h hpsf.HPSF) InspectionResult {
	result := InspectionResult{
		Components: []ComponentInfo{},
	}

	// Iterate through all components
	for _, c := range h.Components {
		// Look up the template for this component
		t, ok := i.templates[c.Kind]
		if !ok {
			continue
		}

		// Add component to result
		result.Components = append(result.Components, ComponentInfo{
			Name:       c.Name,
			Style:      t.Style,
			Kind:       c.Kind,
			Properties: i.getProperties(c, t),
		})
	}

	return result
}

// getProperties extracts all properties from a component, using template defaults as fallback
func (i *Inspector) getProperties(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	properties := make(map[string]any)

	// Start with template defaults for all properties
	for _, templateProp := range t.Properties {
		if templateProp.Default != nil {
			properties[templateProp.Name] = templateProp.Default
		}
	}

	// Override with actual component values
	for _, prop := range c.Properties {
		properties[prop.Name] = prop.Value
	}

	return properties
}
