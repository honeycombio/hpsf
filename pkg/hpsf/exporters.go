package hpsf

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
	// Metadata contains exporter-specific configuration details
	Metadata ExporterMetadata
}

// ExporterMetadata is an interface for exporter-specific metadata
type ExporterMetadata interface {
	// GetType returns the exporter type
	GetType() ExporterType
}

// HoneycombExporterMetadata contains metadata for Honeycomb exporters
type HoneycombExporterMetadata struct {
	// Environment identifies the Honeycomb environment (can be empty)
	Environment string
}

func (m HoneycombExporterMetadata) GetType() ExporterType {
	return ExporterTypeHoneycomb
}

// S3ArchiveExporterMetadata contains metadata for S3 Archive exporters
type S3ArchiveExporterMetadata struct {
	// Region is the AWS region
	Region string
	// Bucket is the S3 bucket name
	Bucket string
	// Prefix is the S3 object prefix
	Prefix string
}

func (m S3ArchiveExporterMetadata) GetType() ExporterType {
	return ExporterTypeAWSS3
}

// EnhanceIndexingS3ExporterMetadata contains metadata for Enhance Indexing S3 exporters
type EnhanceIndexingS3ExporterMetadata struct {
	// Region is the AWS region
	Region string
	// Bucket is the S3 bucket name
	Bucket string
	// Prefix is the S3 object prefix
	Prefix string
}

func (m EnhanceIndexingS3ExporterMetadata) GetType() ExporterType {
	return ExporterTypeEnhanceIndexingS3
}

// OTelGRPCExporterMetadata contains metadata for OTLP gRPC exporters
type OTelGRPCExporterMetadata struct {
}

func (m OTelGRPCExporterMetadata) GetType() ExporterType {
	return ExporterTypeOTelGRPC
}

// OTelHTTPExporterMetadata contains metadata for OTLP HTTP exporters
type OTelHTTPExporterMetadata struct {
}

func (m OTelHTTPExporterMetadata) GetType() ExporterType {
	return ExporterTypeOTelHTTP
}

// DebugExporterMetadata contains metadata for Debug exporters
type DebugExporterMetadata struct {
}

func (m DebugExporterMetadata) GetType() ExporterType {
	return ExporterTypeDebug
}

// NopExporterMetadata contains metadata for Nop (no-op) exporters
type NopExporterMetadata struct {
	// No specific metadata for nop exporters
}

func (m NopExporterMetadata) GetType() ExporterType {
	return ExporterTypeNop
}

// GetExporterInfo extracts all exporter components from the HPSF document.
// It returns a slice of ExporterInfo structs containing the exporter type and metadata.
func (h *HPSF) GetExporterInfo() []ExporterInfo {
	var exporters []ExporterInfo

	// Iterate through all components
	for _, c := range h.Components {
		switch c.Kind {
		case "HoneycombExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeHoneycomb,
				Metadata: extractHoneycombMetadata(c),
			})
		case "S3ArchiveExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeAWSS3,
				Metadata: extractS3ArchiveMetadata(c),
			})
		case "EnhanceIndexingS3Exporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeEnhanceIndexingS3,
				Metadata: extractEnhanceIndexingS3Metadata(c),
			})
		case "OTelGRPCExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeOTelGRPC,
				Metadata: extractOTelGRPCMetadata(c),
			})
		case "OTelHTTPExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeOTelHTTP,
				Metadata: extractOTelHTTPMetadata(c),
			})
		case "DebugExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeDebug,
				Metadata: extractDebugMetadata(c),
			})
		case "NopExporter":
			exporters = append(exporters, ExporterInfo{
				Type:     ExporterTypeNop,
				Metadata: extractNopMetadata(c),
			})
		}
	}

	return exporters
}

// extractHoneycombMetadata extracts Honeycomb exporter metadata
func extractHoneycombMetadata(c *Component) *HoneycombExporterMetadata {
	metadata := &HoneycombExporterMetadata{
		Environment: "", // Can be populated from additional context if available
	}

	return metadata
}

// extractS3ArchiveMetadata extracts S3 Archive exporter metadata
func extractS3ArchiveMetadata(c *Component) *S3ArchiveExporterMetadata {
	metadata := &S3ArchiveExporterMetadata{}

	if prop := c.GetProperty("Region"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata.Region = val
		}
	}

	if prop := c.GetProperty("Bucket"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata.Bucket = val
		}
	}

	if prop := c.GetProperty("Prefix"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata.Prefix = val
		}
	}

	return metadata
}

// extractEnhanceIndexingS3Metadata extracts Enhance Indexing S3 exporter metadata
func extractEnhanceIndexingS3Metadata(c *Component) *EnhanceIndexingS3ExporterMetadata {
	metadata := &EnhanceIndexingS3ExporterMetadata{}

	if prop := c.GetProperty("Region"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata.Region = val
		}
	}

	if prop := c.GetProperty("Bucket"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata.Bucket = val
		}
	}

	if prop := c.GetProperty("Prefix"); prop != nil {
		if val, ok := prop.Value.(string); ok {
			metadata.Prefix = val
		}
	}

	return metadata
}

// extractOTelGRPCMetadata extracts OTLP gRPC exporter metadata
func extractOTelGRPCMetadata(c *Component) *OTelGRPCExporterMetadata {
	return &OTelGRPCExporterMetadata{}
}

// extractOTelHTTPMetadata extracts OTLP HTTP exporter metadata
func extractOTelHTTPMetadata(c *Component) *OTelHTTPExporterMetadata {
	return &OTelHTTPExporterMetadata{}
}

// extractDebugMetadata extracts Debug exporter metadata
func extractDebugMetadata(c *Component) *DebugExporterMetadata {
	return &DebugExporterMetadata{}
}

// extractNopMetadata extracts Nop exporter metadata
func extractNopMetadata(c *Component) *NopExporterMetadata {
	return &NopExporterMetadata{}
}
