package exporter

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// Extractor handles extraction of exporter information from HPSF configurations.
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

// ExporterType represents the different types of exporters available in HPSF
type ExporterType string

const (
	ExporterTypeHoneycomb         ExporterType = "Honeycomb"
	ExporterTypeAWSS3             ExporterType = "AWSS3"
	ExporterTypeEnhanceIndexingS3 ExporterType = "EnhanceIndexingS3"
	ExporterTypeOTelGRPC          ExporterType = "OTelGRPC"
	ExporterTypeOTelHTTP          ExporterType = "OTelHTTP"
	ExporterTypeDebug             ExporterType = "Debug"
	ExporterTypeNop               ExporterType = "Nop"
)

// ExporterInfo represents an exporter component with its type and metadata
type ExporterInfo struct {
	// Type is the exporter type (e.g., "Honeycomb", "AWSS3")
	Type ExporterType
	// Metadata contains exporter-specific configuration details as key-value pairs
	// Users can access values directly without type casting, e.g. metadata["Region"]
	Metadata map[string]any
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

// GetExporterInfo extracts all exporter components from the HPSF document.
// It returns a slice of ExporterInfo structs containing the exporter type and metadata.
func (e *Extractor) GetExporterInfo(h hpsf.HPSF) []ExporterInfo {
	var exporters []ExporterInfo

	// Iterate through all components
	for _, c := range h.Components {
		// Look up the template for this component
		t, ok := e.templates[c.Kind]
		if !ok {
			continue
		}

		// Check if the component is an exporter
		if t.Style != "exporter" {
			continue
		}

		switch c.Kind {
		case "HoneycombExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeHoneycomb,
				Metadata: e.extractHoneycombMetadata(c, t),
			})
		case "S3ArchiveExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeAWSS3,
				Metadata: e.extractS3ArchiveMetadata(c, t),
			})
		case "EnhanceIndexingS3Exporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeEnhanceIndexingS3,
				Metadata: e.extractEnhanceIndexingS3Metadata(c, t),
			})
		case "OTelGRPCExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeOTelGRPC,
				Metadata: e.extractOTelGRPCMetadata(c, t),
			})
		case "OTelHTTPExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeOTelHTTP,
				Metadata: e.extractOTelHTTPMetadata(c, t),
			})
		case "DebugExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeDebug,
				Metadata: e.extractDebugMetadata(c, t),
			})
		case "NopExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeNop,
				Metadata: e.extractNopMetadata(c, t),
			})
		}
	}

	return exporters
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
