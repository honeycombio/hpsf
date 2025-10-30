package inspector

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// Inspector provides information about components in HPSF configurations.
// It loads component templates to access default values and metadata.
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

// ComponentInfo represents a component (receiver, processor, or exporter) with its type and metadata
type ComponentInfo struct {
	// Type is the component kind (e.g., "HoneycombExporter", "OTelReceiver", "MemoryLimiterProcessor")
	Type string
	// Metadata contains component-specific configuration details as key-value pairs
	// Users can access values directly without type casting, e.g. metadata["Region"]
	Metadata map[string]any
}

// InspectionResult holds information about all components in an HPSF configuration.
type InspectionResult struct {
	Receivers  []ComponentInfo
	Processors []ComponentInfo
	Exporters  []ComponentInfo
}

// Deprecated type aliases for backward compatibility
type Component = ComponentInfo
type Components = InspectionResult
type ReceiverInfo = ComponentInfo
type ProcessorInfo = ComponentInfo
type ExporterInfo = ComponentInfo

// getPropertyDefault retrieves the default value for a property from the template component.
// Returns empty string if no default is found.
func getPropertyDefault(t config.TemplateComponent, propertyName string) string {
	for _, prop := range t.Properties {
		if prop.Name == propertyName {
			if defaultVal, ok := prop.Default.(string); ok {
				return defaultVal
			}
			return ""
		}
	}

	return ""
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
				Type:     c.Kind,
				Metadata: i.extractComponentMetadata(c, t),
			})
		case "processor":
			result.Processors = append(result.Processors, ComponentInfo{
				Type:     c.Kind,
				Metadata: i.extractComponentMetadata(c, t),
			})
		case "exporter":
			result.Exporters = append(result.Exporters, ComponentInfo{
				Type:     c.Kind,
				Metadata: i.extractExporterMetadata(c, t),
			})
		}
	}

	return result
}

// GetExporterInfo extracts all exporter components from the HPSF document.
// It returns a slice of ExporterInfo structs containing the exporter type and metadata.
// Deprecated: Use GetComponents instead for more comprehensive component extraction.
func (i *Inspector) GetExporterInfo(h hpsf.HPSF) []Component {
	return i.GetComponents(h).Exporters
}

// extractComponentMetadata extracts metadata for receivers and processors (generic components)
func (i *Inspector) extractComponentMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	metadata := make(map[string]any)

	// For generic components, extract all properties with their values
	for _, prop := range c.Properties {
		metadata[prop.Name] = prop.Value
	}

	return metadata
}

// extractExporterMetadata extracts metadata for exporters with special handling
func (i *Inspector) extractExporterMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	// Use specialized extraction for known exporters
	switch c.Kind {
	case "HoneycombExporter":
		return i.extractHoneycombMetadata(c, t)
	case "S3ArchiveExporter":
		return i.extractS3ArchiveMetadata(c, t)
	case "EnhanceIndexingS3Exporter":
		return i.extractEnhanceIndexingS3Metadata(c, t)
	case "OTelGRPCExporter":
		return i.extractOTelGRPCMetadata(c, t)
	case "OTelHTTPExporter":
		return i.extractOTelHTTPMetadata(c, t)
	case "DebugExporter":
		return i.extractDebugMetadata(c, t)
	case "NopExporter":
		return i.extractNopMetadata(c, t)
	default:
		// For unknown exporters, use generic extraction
		return i.extractComponentMetadata(c, t)
	}
}

// extractHoneycombMetadata extracts Honeycomb exporter metadata
func (i *Inspector) extractHoneycombMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	metadata := make(map[string]any)

	// Environment - can be populated from additional context if available
	metadata["Environment"] = ""

	return metadata
}

// extractS3ArchiveMetadata extracts S3 Archive exporter metadata
func (i *Inspector) extractS3ArchiveMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	metadata := make(map[string]any)

	// Get Region - use component value or template default
	if prop := c.GetProperty("Region"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata["Region"] = val
		}
	} else {
		metadata["Region"] = getPropertyDefault(t, "Region")
	}

	// Get Bucket - required property, no default
	if prop := c.GetProperty("Bucket"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata["Bucket"] = val
		}
	}

	// Get Prefix - optional property, no default
	if prop := c.GetProperty("Prefix"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata["Prefix"] = val
		}
	}

	return metadata
}

// extractEnhanceIndexingS3Metadata extracts Enhance Indexing S3 exporter metadata
func (i *Inspector) extractEnhanceIndexingS3Metadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	metadata := make(map[string]any)

	// Get Region - use component value or template default
	if prop := c.GetProperty("Region"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata["Region"] = val
		}
	} else {
		metadata["Region"] = getPropertyDefault(t, "Region")
	}

	// Get Bucket - required property, no default
	if prop := c.GetProperty("Bucket"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata["Bucket"] = val
		}
	}

	// Get Prefix - optional property, no default
	if prop := c.GetProperty("Prefix"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata["Prefix"] = val
		}
	}

	return metadata
}

// extractOTelGRPCMetadata extracts OTLP gRPC exporter metadata
func (i *Inspector) extractOTelGRPCMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	return make(map[string]any)
}

// extractOTelHTTPMetadata extracts OTLP HTTP exporter metadata
func (i *Inspector) extractOTelHTTPMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	return make(map[string]any)
}

// extractDebugMetadata extracts Debug exporter metadata
func (i *Inspector) extractDebugMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	return make(map[string]any)
}

// extractNopMetadata extracts Nop exporter metadata
func (i *Inspector) extractNopMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	return make(map[string]any)
}
