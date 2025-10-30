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
	assert.Equal(t, "HoneycombExporter", exp.Type)

	// Verify metadata is accessible without casting
	assert.Equal(t, "", exp.Metadata["Environment"])
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
	assert.Equal(t, "S3ArchiveExporter", exp.Type)

	// Verify metadata is accessible without casting
	assert.Equal(t, "us-west-2", exp.Metadata["Region"])
	assert.Equal(t, "my-telemetry-bucket", exp.Metadata["Bucket"])
	assert.Equal(t, "telemetry/", exp.Metadata["Prefix"])
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
	assert.Equal(t, "EnhanceIndexingS3Exporter", exp.Type)

	// Verify metadata is accessible without casting
	assert.Equal(t, "eu-west-1", exp.Metadata["Region"])
	assert.Equal(t, "my-indexed-bucket", exp.Metadata["Bucket"])
	assert.Nil(t, exp.Metadata["Prefix"]) // Not set in config
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
	assert.Equal(t, "OTelGRPCExporter", exp.Type)

	// Verify metadata map exists (even if empty)
	assert.NotNil(t, exp.Metadata)
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
	assert.Equal(t, "OTelHTTPExporter", exp.Type)

	// Verify metadata map exists (even if empty)
	assert.NotNil(t, exp.Metadata)
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
	assert.Equal(t, "DebugExporter", exp.Type)

	// Verify metadata map exists (even if empty)
	assert.NotNil(t, exp.Metadata)
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
	assert.Equal(t, "NopExporter", exp.Type)

	// Verify metadata map exists (even if empty)
	assert.NotNil(t, exp.Metadata)
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
	exporterTypes := make(map[string]bool)
	for _, exp := range exporters {
		exporterTypes[exp.Type] = true
	}

	assert.True(t, exporterTypes["HoneycombExporter"])
	assert.True(t, exporterTypes["S3ArchiveExporter"])
	assert.True(t, exporterTypes["DebugExporter"])
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
	assert.Equal(t, "HoneycombExporter", exp.Type)

	// Environment field should have empty string value
	assert.Equal(t, "", exp.Metadata["Environment"])
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
	assert.Equal(t, "S3ArchiveExporter", exp.Type)

	// Region should use default from template
	assert.Equal(t, "us-east-1", exp.Metadata["Region"])
	assert.Equal(t, "my-bucket", exp.Metadata["Bucket"])
	assert.Nil(t, exp.Metadata["Prefix"]) // Not set in config
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
	assert.Equal(t, "EnhanceIndexingS3Exporter", exp.Type)

	// Region should use default from template
	assert.Equal(t, "us-east-1", exp.Metadata["Region"])
	assert.Equal(t, "my-indexed-bucket", exp.Metadata["Bucket"])
	assert.Nil(t, exp.Metadata["Prefix"]) // Not set in config
}

func TestGetComponentInfo_AllComponentTypes(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
    properties:
      - name: GRPCPort
        value: 4317
      - name: HTTPPort
        value: 4318
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
    properties:
      - name: CheckInterval
        value: 1s
      - name: MemoryLimitMiB
        value: 512
  - name: Honeycomb Export
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-key
  - name: S3 Archive
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-bucket
      - name: Region
        value: us-west-2
  - name: Another Receiver
    kind: NopReceiver
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	info := extractor.GetComponents(h)

	// Verify receivers
	require.Len(t, info.Receivers, 2)
	assert.Equal(t, "OTelReceiver", info.Receivers[0].Type)
	assert.Equal(t, 4317, info.Receivers[0].Metadata["GRPCPort"])
	assert.Equal(t, 4318, info.Receivers[0].Metadata["HTTPPort"])
	assert.Equal(t, "NopReceiver", info.Receivers[1].Type)

	// Verify processors
	require.Len(t, info.Processors, 1)
	assert.Equal(t, "MemoryLimiterProcessor", info.Processors[0].Type)
	assert.Equal(t, "1s", info.Processors[0].Metadata["CheckInterval"])
	assert.Equal(t, 512, info.Processors[0].Metadata["MemoryLimitMiB"])

	// Verify exporters
	require.Len(t, info.Exporters, 2)
	assert.Equal(t, "HoneycombExporter", info.Exporters[0].Type)
	assert.Equal(t, "S3ArchiveExporter", info.Exporters[1].Type)
	assert.Equal(t, "us-west-2", info.Exporters[1].Metadata["Region"])
	assert.Equal(t, "my-bucket", info.Exporters[1].Metadata["Bucket"])
}

func TestExporterInfo_MetadataAccessWithoutCasting(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Test S3
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: test-bucket
      - name: Region
        value: eu-central-1
      - name: Prefix
        value: data/
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	extractor, err := NewExtractor()
	require.NoError(t, err)

	exporters := extractor.GetExporterInfo(h)
	require.Len(t, exporters, 1)

	exp := exporters[0]

	// Demonstrate accessing metadata without type casting
	region := exp.Metadata["Region"]
	assert.Equal(t, "eu-central-1", region)

	bucket := exp.Metadata["Bucket"]
	assert.Equal(t, "test-bucket", bucket)

	prefix := exp.Metadata["Prefix"]
	assert.Equal(t, "data/", prefix)

	// Non-existent keys return nil
	assert.Nil(t, exp.Metadata["NonExistentKey"])
}
