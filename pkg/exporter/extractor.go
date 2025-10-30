package exporter

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// Extractor handles extraction of component information from HPSF configurations.
// It loads component templates to access default values and metadata.
type Extractor struct {
	templates map[string]config.TemplateComponent // kind -> template
}

// NewExtractor creates a new Extractor with embedded component templates loaded.
func NewExtractor() (*Extractor, error) {
	templates, err := data.LoadEmbeddedComponents()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded components: %w", err)
	}

	return &Extractor{
		templates: templates,
	}, nil
}

// Component represents a component (receiver, processor, or exporter) with its type and metadata
type Component struct {
	// Type is the component kind (e.g., "HoneycombExporter", "OTelReceiver", "MemoryLimiterProcessor")
	Type string
	// Metadata contains component-specific configuration details as key-value pairs
	// Users can access values directly without type casting, e.g. metadata["Region"]
	Metadata map[string]any
}

// Components holds information about all components in an HPSF configuration.
type Components struct {
	Receivers  []Component
	Processors []Component
	Exporters  []Component
}

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
// It returns a Components struct containing receivers, processors, and exporters.
func (e *Extractor) GetComponents(h hpsf.HPSF) Components {
	components := Components{
		Receivers:  []Component{},
		Processors: []Component{},
		Exporters:  []Component{},
	}

	// Iterate through all components
	for _, c := range h.Components {
		// Look up the template for this component
		t, ok := e.templates[c.Kind]
		if !ok {
			continue
		}

		// Extract based on component style
		switch t.Style {
		case "receiver":
			components.Receivers = append(components.Receivers, Component{
				Type:     c.Kind,
				Metadata: e.extractComponentMetadata(c, t),
			})
		case "processor":
			components.Processors = append(components.Processors, Component{
				Type:     c.Kind,
				Metadata: e.extractComponentMetadata(c, t),
			})
		case "exporter":
			components.Exporters = append(components.Exporters, Component{
				Type:     c.Kind,
				Metadata: e.extractExporterMetadata(c, t),
			})
		}
	}

	return components
}

// GetExporterInfo extracts all exporter components from the HPSF document.
// It returns a slice of ExporterInfo structs containing the exporter type and metadata.
// Deprecated: Use GetComponents instead for more comprehensive component extraction.
func (e *Extractor) GetExporterInfo(h hpsf.HPSF) []Component {
	return e.GetComponents(h).Exporters
}

// extractComponentMetadata extracts metadata for receivers and processors (generic components)
func (e *Extractor) extractComponentMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	metadata := make(map[string]any)

	// For generic components, extract all properties with their values
	for _, prop := range c.Properties {
		metadata[prop.Name] = prop.Value
	}

	return metadata
}

// extractExporterMetadata extracts metadata for exporters with special handling
func (e *Extractor) extractExporterMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	// Use specialized extraction for known exporters
	switch c.Kind {
	case "HoneycombExporter":
		return e.extractHoneycombMetadata(c, t)
	case "S3ArchiveExporter":
		return e.extractS3ArchiveMetadata(c, t)
	case "EnhanceIndexingS3Exporter":
		return e.extractEnhanceIndexingS3Metadata(c, t)
	case "OTelGRPCExporter":
		return e.extractOTelGRPCMetadata(c, t)
	case "OTelHTTPExporter":
		return e.extractOTelHTTPMetadata(c, t)
	case "DebugExporter":
		return e.extractDebugMetadata(c, t)
	case "NopExporter":
		return e.extractNopMetadata(c, t)
	default:
		// For unknown exporters, use generic extraction
		return e.extractComponentMetadata(c, t)
	}
}

// extractHoneycombMetadata extracts Honeycomb exporter metadata
func (e *Extractor) extractHoneycombMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	metadata := make(map[string]any)

	// Environment - can be populated from additional context if available
	metadata["Environment"] = ""

	return metadata
}

// extractS3ArchiveMetadata extracts S3 Archive exporter metadata
func (e *Extractor) extractS3ArchiveMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
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
func (e *Extractor) extractEnhanceIndexingS3Metadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
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
func (e *Extractor) extractOTelGRPCMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	return make(map[string]any)
}

// extractOTelHTTPMetadata extracts OTLP HTTP exporter metadata
func (e *Extractor) extractOTelHTTPMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	return make(map[string]any)
}

// extractDebugMetadata extracts Debug exporter metadata
func (e *Extractor) extractDebugMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	return make(map[string]any)
}

// extractNopMetadata extracts Nop exporter metadata
func (e *Extractor) extractNopMetadata(c *hpsf.Component, t config.TemplateComponent) map[string]any {
	return make(map[string]any)
}
