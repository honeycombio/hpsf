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

// ComponentInfo represents a component (receiver, processor, or exporter) with its name, type, and properties
type ComponentInfo struct {
	// Name is the user-defined name of the component instance
	Name string
	// Type is the component kind (e.g., "HoneycombExporter", "OTelReceiver", "MemoryLimiterProcessor")
	Type string
	// Properties contains component-specific configuration details as key-value pairs
	// Users can access values directly without type casting, e.g. properties["Region"]
	Properties map[string]any
}

// InspectionResult holds information about all components in an HPSF configuration.
// TODO: Add support for sampler, startsampling and conditions later.
type InspectionResult struct {
	Receivers  []ComponentInfo
	Processors []ComponentInfo
	Exporters  []ComponentInfo
}

// getPropertyValue retrieves a property value, first checking the component,
// then falling back to the template default if not found.
func getPropertyValue(c *hpsf.Component, t config.TemplateComponent, propertyName string) any {
	// First, try to get from component
	if prop := c.GetProperty(propertyName); prop != nil {
		return prop.Value
	}

	// Fall back to template default
	for _, templateProp := range t.Properties {
		if templateProp.Name == propertyName {
			return templateProp.Default
		}
	}

	return nil
}

// GetComponents extracts all components from the HPSF document.
// It returns an InspectionResult containing receivers, processors, and exporters.
func (i *Inspector) GetComponents(h hpsf.HPSF) InspectionResult {
	result := InspectionResult{
		Receivers:  []ComponentInfo{},
		Processors: []ComponentInfo{},
		Exporters:  []ComponentInfo{},
	}

	// Iterate through all components
	for _, c := range h.Components {
		// Look up the template for this component
		t, ok := i.templates[c.Kind]
		if !ok {
			continue
		}

		// Extract based on component style
		switch t.Style {
		case "receiver":
			result.Receivers = append(result.Receivers, ComponentInfo{
				Name:     c.Name,
				Type:     c.Kind,
				Properties: i.extractComponentProperties(c, t),
			})
		case "processor":
			result.Processors = append(result.Processors, ComponentInfo{
				Name:     c.Name,
				Type:     c.Kind,
				Properties: i.extractComponentProperties(c, t),
			})
		case "exporter":
			result.Exporters = append(result.Exporters, ComponentInfo{
				Name:     c.Name,
				Type:     c.Kind,
				Properties: i.extractExporterProperties(c, t),
			})
		}
	}

	return result
}

// extractComponentProperties extracts properties for receivers and processors (generic components)
func (i *Inspector) extractComponentProperties(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	properties := make(map[string]any)

	// For generic components, extract all properties with their values
	for _, prop := range c.Properties {
		properties[prop.Name] = prop.Value
	}

	return properties
}

// extractExporterProperties extracts properties for exporters with special handling
func (i *Inspector) extractExporterProperties(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	// Use specialized extraction for known exporters
	switch c.Kind {
	case "HoneycombExporter":
		return i.extractHoneycombProperties(c, t)
	case "S3ArchiveExporter":
		return i.extractS3ArchiveProperties(c, t)
	case "EnhanceIndexingS3Exporter":
		return i.extractEnhanceIndexingS3Properties(c, t)
	default:
		// For other exporters, return empty properties
		return make(map[string]any)
	}
}

// extractHoneycombProperties extracts Honeycomb exporter properties
func (i *Inspector) extractHoneycombProperties(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	properties := make(map[string]any)

	// Environment - can be populated from additional context if available
	properties["Environment"] = ""

	return properties
}

// extractS3ArchiveProperties extracts S3 Archive exporter properties
func (i *Inspector) extractS3ArchiveProperties(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	properties := make(map[string]any)

	// Extract selected properties with template defaults as fallback
	if val := getPropertyValue(c, t, "Region"); val != nil {
		properties["Region"] = val
	}
	if val := getPropertyValue(c, t, "Bucket"); val != nil {
		properties["Bucket"] = val
	}
	if val := getPropertyValue(c, t, "Prefix"); val != nil {
		properties["Prefix"] = val
	}

	return properties
}

// extractEnhanceIndexingS3Properties extracts Enhance Indexing S3 exporter properties
func (i *Inspector) extractEnhanceIndexingS3Properties(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	properties := make(map[string]any)

	// Extract selected properties with template defaults as fallback
	if val := getPropertyValue(c, t, "Region"); val != nil {
		properties["Region"] = val
	}
	if val := getPropertyValue(c, t, "Bucket"); val != nil {
		properties["Bucket"] = val
	}
	if val := getPropertyValue(c, t, "Prefix"); val != nil {
		properties["Prefix"] = val
	}

	return properties
}
