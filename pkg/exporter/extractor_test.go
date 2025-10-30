package exporter

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExporterInfo_HoneycombExporter(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My Honeycomb Exporter
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-api-key
      - name: APIEndpoint
        value: api.honeycomb.io
      - name: APIPort
        value: 443
      - name: MetricsDataset
        value: my-metrics
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, ExporterTypeHoneycomb, exp.Type)

	metadata, ok := exp.Metadata.(*HoneycombExporterMetadata)
	require.True(t, ok, "metadata should be HoneycombExporterMetadata")
	assert.Equal(t, "", metadata.Environment)
}

func TestGetExporterInfo_S3ArchiveExporter(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My S3 Archive
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-telemetry-bucket
      - name: Region
        value: us-west-2
      - name: Prefix
        value: telemetry/
      - name: PartitionFormat
        value: year=%Y/month=%m/day=%d
      - name: Marshaler
        value: otlp_json
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, ExporterTypeAWSS3, exp.Type)

	metadata, ok := exp.Metadata.(*S3ArchiveExporterMetadata)
	require.True(t, ok, "metadata should be S3ArchiveExporterMetadata")
	assert.Equal(t, "us-west-2", metadata.Region)
	assert.Equal(t, "my-telemetry-bucket", metadata.Bucket)
	assert.Equal(t, "telemetry/", metadata.Prefix)
}

func TestGetExporterInfo_EnhanceIndexingS3Exporter(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My Enhanced S3
    kind: EnhanceIndexingS3Exporter
    properties:
      - name: Bucket
        value: my-indexed-bucket
      - name: Region
        value: eu-west-1
      - name: APIKey
        value: test-key
      - name: APISecret
        value: test-secret
      - name: APIEndpoint
        value: https://api.honeycomb.io
      - name: IndexedFields
        value:
          - custom.field1
          - custom.field2
      - name: PartitionFormat
        value: year=%Y/month=%m
      - name: Marshaler
        value: otlp_proto
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, ExporterTypeEnhanceIndexingS3, exp.Type)

	metadata, ok := exp.Metadata.(*EnhanceIndexingS3ExporterMetadata)
	require.True(t, ok, "metadata should be EnhanceIndexingS3ExporterMetadata")
	assert.Equal(t, "eu-west-1", metadata.Region)
	assert.Equal(t, "my-indexed-bucket", metadata.Bucket)
	assert.Equal(t, "", metadata.Prefix)
}

func TestGetExporterInfo_OTelGRPCExporter(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My OTLP gRPC
    kind: OTelGRPCExporter
    properties:
      - name: Host
        value: otel-collector.example.com
      - name: Port
        value: 4317
      - name: Insecure
        value: false
      - name: Headers
        value:
          authorization: Bearer token123
          x-custom-header: value
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, ExporterTypeOTelGRPC, exp.Type)

	_, ok := exp.Metadata.(*OTelGRPCExporterMetadata)
	require.True(t, ok, "metadata should be OTelGRPCExporterMetadata")
}

func TestGetExporterInfo_OTelHTTPExporter(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My OTLP HTTP
    kind: OTelHTTPExporter
    properties:
      - name: Host
        value: otel-collector.example.com
      - name: Port
        value: 4318
      - name: Insecure
        value: true
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, ExporterTypeOTelHTTP, exp.Type)

	_, ok := exp.Metadata.(*OTelHTTPExporterMetadata)
	require.True(t, ok, "metadata should be OTelHTTPExporterMetadata")
}

func TestGetExporterInfo_DebugExporter(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My Debug
    kind: DebugExporter
    properties:
      - name: Verbosity
        value: detailed
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, ExporterTypeDebug, exp.Type)

	_, ok := exp.Metadata.(*DebugExporterMetadata)
	require.True(t, ok, "metadata should be DebugExporterMetadata")
}

func TestGetExporterInfo_NopExporter(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My Nop
    kind: NopExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, ExporterTypeNop, exp.Type)

	_, ok := exp.Metadata.(*NopExporterMetadata)
	require.True(t, ok, "metadata should be NopExporterMetadata")
}

func TestGetExporterInfo_MultipleExporters(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Receiver1
    kind: OTelGRPCReceiver
  - name: Honeycomb Export
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-key
      - name: APIEndpoint
        value: api.honeycomb.io
  - name: S3 Archive
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-bucket
      - name: Region
        value: us-east-1
  - name: Processor1
    kind: BatchProcessor
  - name: Debug Export
    kind: DebugExporter
    properties:
      - name: Verbosity
        value: basic
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 3, "should extract only the 3 exporters")

	// Check that we have all the expected exporter types
	exporterTypes := make(map[ExporterType]bool)
	for _, exp := range exporters {
		exporterTypes[exp.Type] = true
	}

	assert.True(t, exporterTypes[ExporterTypeHoneycomb])
	assert.True(t, exporterTypes[ExporterTypeAWSS3])
	assert.True(t, exporterTypes[ExporterTypeDebug])
}

func TestGetExporterInfo_InvalidYAML(t *testing.T) {
	hpsfConfig := `this is not valid yaml: {[`

	_, err := hpsf.FromYAML(hpsfConfig)
	assert.Error(t, err, "should return error for invalid YAML")
}

func TestGetExporterInfo_EmptyConfig(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components: []
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	assert.Empty(t, exporters, "should return empty slice for config with no components")
}

func TestGetExporterInfo_UnknownComponentKind(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Unknown Component
    kind: NonExistentExporter
    properties:
      - name: SomeProp
        value: somevalue
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	assert.Empty(t, exporters, "should skip unknown component kinds")
}

func TestGetExporterInfo_MissingOptionalProperties(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Minimal Honeycomb
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-key
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, ExporterTypeHoneycomb, exp.Type)

	metadata, ok := exp.Metadata.(*HoneycombExporterMetadata)
	require.True(t, ok)
	// Environment field should have zero value
	assert.Empty(t, metadata.Environment)
}

func TestGetExporterInfo_S3ArchiveExporter_UsesDefaultRegion(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Minimal S3
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-bucket
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, ExporterTypeAWSS3, exp.Type)

	metadata, ok := exp.Metadata.(*S3ArchiveExporterMetadata)
	require.True(t, ok)
	// Region should use default from template
	assert.Equal(t, "us-east-1", metadata.Region)
	assert.Equal(t, "my-bucket", metadata.Bucket)
	assert.Empty(t, metadata.Prefix)
}

func TestGetExporterInfo_EnhanceIndexingS3Exporter_UsesDefaultRegion(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Minimal Enhanced S3
    kind: EnhanceIndexingS3Exporter
    properties:
      - name: Bucket
        value: my-indexed-bucket
      - name: APIKey
        value: test-key
      - name: APISecret
        value: test-secret
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, ExporterTypeEnhanceIndexingS3, exp.Type)

	metadata, ok := exp.Metadata.(*EnhanceIndexingS3ExporterMetadata)
	require.True(t, ok)
	// Region should use default from template
	assert.Equal(t, "us-east-1", metadata.Region)
	assert.Equal(t, "my-indexed-bucket", metadata.Bucket)
	assert.Empty(t, metadata.Prefix)
}

func TestExporterMetadata_GetType(t *testing.T) {
	tests := []struct {
		name     string
		metadata ExporterMetadata
		expected ExporterType
	}{
		{
			name:     "HoneycombExporterMetadata",
			metadata: &HoneycombExporterMetadata{},
			expected: ExporterTypeHoneycomb,
		},
		{
			name:     "S3ArchiveExporterMetadata",
			metadata: &S3ArchiveExporterMetadata{},
			expected: ExporterTypeAWSS3,
		},
		{
			name:     "EnhanceIndexingS3ExporterMetadata",
			metadata: &EnhanceIndexingS3ExporterMetadata{},
			expected: ExporterTypeEnhanceIndexingS3,
		},
		{
			name:     "OTelGRPCExporterMetadata",
			metadata: &OTelGRPCExporterMetadata{},
			expected: ExporterTypeOTelGRPC,
		},
		{
			name:     "OTelHTTPExporterMetadata",
			metadata: &OTelHTTPExporterMetadata{},
			expected: ExporterTypeOTelHTTP,
		},
		{
			name:     "DebugExporterMetadata",
			metadata: &DebugExporterMetadata{},
			expected: ExporterTypeDebug,
		},
		{
			name:     "NopExporterMetadata",
			metadata: &NopExporterMetadata{},
			expected: ExporterTypeNop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.metadata.GetType())
		})
	}
}
