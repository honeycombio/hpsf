package inspector

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspect_HoneycombExporter(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "My Honeycomb Exporter", exp.Name)
	assert.Equal(t, "HoneycombExporter", exp.Kind)

	// Verify properties contain actual component values
	assert.NotNil(t, exp.Properties)
}

func TestInspect_S3ArchiveExporter(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "My S3 Archive", exp.Name)
	assert.Equal(t, "S3ArchiveExporter", exp.Kind)

	// Verify properties is accessible without casting
	assert.Equal(t, "us-west-2", exp.Properties["Region"])
	assert.Equal(t, "my-telemetry-bucket", exp.Properties["Bucket"])
	assert.Equal(t, "telemetry/", exp.Properties["Prefix"])
}

func TestInspect_EnhanceIndexingS3Exporter(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "My Enhanced S3", exp.Name)
	assert.Equal(t, "EnhanceIndexingS3Exporter", exp.Kind)

	// Verify properties is accessible without casting
	assert.Equal(t, "eu-west-1", exp.Properties["Region"])
	assert.Equal(t, "my-indexed-bucket", exp.Properties["Bucket"])
	assert.Nil(t, exp.Properties["Prefix"]) // Not set in config
}

func TestInspect_OTelGRPCExporter(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "OTelGRPCExporter", exp.Kind)

	// Verify properties map exists (even if empty)
	assert.NotNil(t, exp.Properties)
}

func TestInspect_OTelHTTPExporter(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "OTelHTTPExporter", exp.Kind)

	// Verify properties map exists (even if empty)
	assert.NotNil(t, exp.Properties)
}

func TestInspect_DebugExporter(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "DebugExporter", exp.Kind)

	// Verify properties map exists (even if empty)
	assert.NotNil(t, exp.Properties)
}

func TestInspect_NopExporter(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: My Nop
    kind: NopExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "NopExporter", exp.Kind)

	// Verify properties map exists (even if empty)
	assert.NotNil(t, exp.Properties)
}

func TestInspect_MultipleExporters(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 3, "should extract only the 3 exporters")

	// Check that we have all the expected exporter types
	exporterTypes := make(map[string]bool)
	exporterNames := make(map[string]bool)
	for _, exp := range exporters {
		exporterTypes[exp.Kind] = true
		exporterNames[exp.Name] = true
	}

	assert.True(t, exporterTypes["HoneycombExporter"])
	assert.True(t, exporterTypes["S3ArchiveExporter"])
	assert.True(t, exporterTypes["DebugExporter"])

	// Verify names are captured
	assert.True(t, exporterNames["Honeycomb Export"])
	assert.True(t, exporterNames["S3 Archive"])
	assert.True(t, exporterNames["Debug Export"])
}

func TestInspect_InvalidYAML(t *testing.T) {
	hpsfConfig := `this is not valid yaml: {[`

	_, err := hpsf.FromYAML(hpsfConfig)
	assert.Error(t, err, "should return error for invalid YAML")
}

func TestInspect_EmptyConfig(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components: []
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	assert.Empty(t, exporters, "should return empty slice for config with no components")
}

func TestInspect_UnknownComponentKind(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	assert.Empty(t, exporters, "should skip unknown component kinds")
}

func TestInspect_MissingOptionalProperties(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "HoneycombExporter", exp.Kind)

	// Properties should be extracted even when minimal config provided
	assert.NotNil(t, exp.Properties)
}

func TestInspect_S3ArchiveExporter_UsesDefaultRegion(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "S3ArchiveExporter", exp.Kind)

	// Region should use default from template
	assert.Equal(t, "us-east-1", exp.Properties["Region"])
	assert.Equal(t, "my-bucket", exp.Properties["Bucket"])
	assert.Nil(t, exp.Properties["Prefix"]) // Not set in config
}

func TestInspect_EnhanceIndexingS3Exporter_UsesDefaultRegion(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "EnhanceIndexingS3Exporter", exp.Kind)

	// Region should use default from template
	assert.Equal(t, "us-east-1", exp.Properties["Region"])
	assert.Equal(t, "my-indexed-bucket", exp.Properties["Bucket"])
	assert.Nil(t, exp.Properties["Prefix"]) // Not set in config
}

func TestInspect_AllComponentTypes(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	result := inspector.Inspect(h)

	// Verify receivers
	receivers := result.Receivers()
	require.Len(t, receivers, 2)
	assert.Equal(t, "OTLP Receiver", receivers[0].Name)
	assert.Equal(t, "OTelReceiver", receivers[0].Kind)
	assert.Equal(t, 4317, receivers[0].Properties["GRPCPort"])
	assert.Equal(t, 4318, receivers[0].Properties["HTTPPort"])
	assert.Equal(t, "Another Receiver", receivers[1].Name)
	assert.Equal(t, "NopReceiver", receivers[1].Kind)

	// Verify processors
	processors := result.Processors()
	require.Len(t, processors, 1)
	assert.Equal(t, "Memory Limiter", processors[0].Name)
	assert.Equal(t, "MemoryLimiterProcessor", processors[0].Kind)
	assert.Equal(t, "1s", processors[0].Properties["CheckInterval"])
	assert.Equal(t, 512, processors[0].Properties["MemoryLimitMiB"])

	// Verify exporters
	exporters := result.Exporters()
	require.Len(t, exporters, 2)
	assert.Equal(t, "Honeycomb Export", exporters[0].Name)
	assert.Equal(t, "HoneycombExporter", exporters[0].Kind)
	assert.Equal(t, "S3 Archive", exporters[1].Name)
	assert.Equal(t, "S3ArchiveExporter", exporters[1].Kind)
	assert.Equal(t, "us-west-2", exporters[1].Properties["Region"])
	assert.Equal(t, "my-bucket", exporters[1].Properties["Bucket"])
}

func TestInspect_PropertiesAccessWithoutCasting(t *testing.T) {
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

	inspector, err := NewInspector()
	require.NoError(t, err)

	exporters := inspector.Inspect(h).Exporters()
	require.Len(t, exporters, 1)

	exp := exporters[0]
	assert.Equal(t, "Test S3", exp.Name)

	// Demonstrate accessing properties without type casting
	region := exp.Properties["Region"]
	assert.Equal(t, "eu-central-1", region)

	bucket := exp.Properties["Bucket"]
	assert.Equal(t, "test-bucket", bucket)

	prefix := exp.Properties["Prefix"]
	assert.Equal(t, "data/", prefix)

	// Non-existent keys return nil
	assert.Nil(t, exp.Properties["NonExistentKey"])
}

func TestInspectionResult_Exporters(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
  - name: Honeycomb Export
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-key
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
  - name: S3 Archive
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-bucket
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	inspector, err := NewInspector()
	require.NoError(t, err)

	result := inspector.Inspect(h)

	// Exporters() should return only exporters
	exporters := result.Exporters()
	require.Len(t, exporters, 2)
	assert.Equal(t, "Honeycomb Export", exporters[0].Name)
	assert.Equal(t, "exporter", exporters[0].Style)
	assert.Equal(t, "HoneycombExporter", exporters[0].Kind)
	assert.Equal(t, "S3 Archive", exporters[1].Name)
	assert.Equal(t, "exporter", exporters[1].Style)
	assert.Equal(t, "S3ArchiveExporter", exporters[1].Kind)
}

func TestInspectionResult_Receivers(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
    properties:
      - name: GRPCPort
        value: 4317
  - name: Honeycomb Export
    kind: HoneycombExporter
    properties:
      - name: APIKey
        value: test-key
  - name: Nop Receiver
    kind: NopReceiver
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	inspector, err := NewInspector()
	require.NoError(t, err)

	result := inspector.Inspect(h)

	// Receivers() should return only receivers
	receivers := result.Receivers()
	require.Len(t, receivers, 2)
	assert.Equal(t, "OTLP Receiver", receivers[0].Name)
	assert.Equal(t, "receiver", receivers[0].Style)
	assert.Equal(t, "OTelReceiver", receivers[0].Kind)
	assert.Equal(t, 4317, receivers[0].Properties["GRPCPort"])
	assert.Equal(t, "Nop Receiver", receivers[1].Name)
	assert.Equal(t, "receiver", receivers[1].Style)
	assert.Equal(t, "NopReceiver", receivers[1].Kind)
}

func TestInspectionResult_Processors(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
    properties:
      - name: CheckInterval
        value: 1s
  - name: Honeycomb Export
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	inspector, err := NewInspector()
	require.NoError(t, err)

	result := inspector.Inspect(h)

	// Processors() should return only processors
	processors := result.Processors()
	require.Len(t, processors, 1)
	assert.Equal(t, "Memory Limiter", processors[0].Name)
	assert.Equal(t, "processor", processors[0].Style)
	assert.Equal(t, "MemoryLimiterProcessor", processors[0].Kind)
	assert.Equal(t, "1s", processors[0].Properties["CheckInterval"])
}

func TestInspectionResult_DirectComponentsAccess(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: OTLP Receiver
    kind: OTelReceiver
  - name: Memory Limiter
    kind: MemoryLimiterProcessor
  - name: Honeycomb Export
    kind: HoneycombExporter
  - name: S3 Archive
    kind: S3ArchiveExporter
    properties:
      - name: Bucket
        value: my-bucket
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	inspector, err := NewInspector()
	require.NoError(t, err)

	result := inspector.Inspect(h)

	// Can access all components directly
	require.Len(t, result.Components, 4)

	// Verify all components are present with correct styles
	assert.Equal(t, "OTLP Receiver", result.Components[0].Name)
	assert.Equal(t, "receiver", result.Components[0].Style)

	assert.Equal(t, "Memory Limiter", result.Components[1].Name)
	assert.Equal(t, "processor", result.Components[1].Style)

	assert.Equal(t, "Honeycomb Export", result.Components[2].Name)
	assert.Equal(t, "exporter", result.Components[2].Style)

	assert.Equal(t, "S3 Archive", result.Components[3].Name)
	assert.Equal(t, "exporter", result.Components[3].Style)

	// Can iterate through all components
	styleCount := make(map[string]int)
	for _, comp := range result.Components {
		styleCount[comp.Style]++
	}
	assert.Equal(t, 1, styleCount["receiver"])
	assert.Equal(t, 1, styleCount["processor"])
	assert.Equal(t, 2, styleCount["exporter"])
}

func TestInspectionResult_EmptyFilters(t *testing.T) {
	hpsfConfig := `
kind: hpsf
version: 1.0
components:
  - name: Honeycomb Export
    kind: HoneycombExporter
`

	h, err := hpsf.FromYAML(hpsfConfig)
	require.NoError(t, err)

	inspector, err := NewInspector()
	require.NoError(t, err)

	result := inspector.Inspect(h)

	// Only has exporters, so receivers and processors should be empty
	assert.Len(t, result.Exporters(), 1)
	assert.Len(t, result.Receivers(), 0)
	assert.Len(t, result.Processors(), 0)
}
