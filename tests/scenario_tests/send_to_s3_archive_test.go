package hpsftests

import (
	"testing"
	"time"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendToS3Archive(t *testing.T) {
	// Test the HPSF parsing and template generation using typed configuration
	rulesConfig, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/send_to_s3_archive.yaml")

	// First, verify that the refinery config was generated successfully
	assert.Len(t, rulesConfig.Samplers, 1, "Expected 1 sampler in refinery config")

	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1, "Expected 1 traces pipeline, got %v", tracesPipelineNames)

	// Verify the S3 exporter is present in the traces pipeline
	_, _, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found, "Expected pipeline to be found")
	assert.Len(t, exporters, 1, "Expected 1 exporter, got %s", exporters)

	// Check that the exporter is the S3 exporter
	expectedExporter := "awss3/My_S3_Backend"
	if assert.NotEmpty(t, exporters, "Expected at least one exporter") {
		assert.Equal(t, expectedExporter, exporters[0])
	}

	// Get the S3 exporter configuration using typed access
	s3Config, findResult := collectorprovider.GetExporterConfig[awss3exporter.Config](collectorConfig, "awss3/My_S3_Backend")
	require.True(t, findResult.Found, "Expected exporter to find \"%v\", found (%v)", findResult.SearchString, findResult.Components)

	// Test S3 bucket configuration (from templates: s3uploader.s3_bucket)
	assert.Equal(t, "test-bucket", s3Config.S3Uploader.S3Bucket)

	// Test S3 region configuration (from templates: s3uploader.region)
	assert.Equal(t, "us-west-2", s3Config.S3Uploader.Region)

	// Test S3 prefix configuration (from templates: s3uploader.s3_prefix)
	assert.Equal(t, "telemetry-data/", s3Config.S3Uploader.S3Prefix)

	// Test S3 partition format configuration (from templates: s3uploader.s3_partition_format)
	assert.Equal(t, "year=%Y/month=%m/day=%d/hour=%H", s3Config.S3Uploader.S3PartitionFormat)

	// Test marshaler configuration (from templates: marshaler)
	assert.Equal(t, "otlp_proto", string(s3Config.MarshalerName))

	// Test compression configuration (from templates: s3uploader.compression - hardcoded to "gzip")
	assert.Equal(t, "gzip", string(s3Config.S3Uploader.Compression))

	// Test timeout configuration (from templates: timeout)
	expectedTimeout := 10 * time.Second
	assert.Equal(t, expectedTimeout, s3Config.TimeoutSettings.Timeout)

	// Test sending queue configuration (from templates: sending_queue.*)
	assert.Equal(t, int64(500000), s3Config.QueueSettings.QueueSize)

	assert.True(t, s3Config.QueueSettings.Enabled, "Expected sending_queue.enabled to be true")

	// Test batch configuration - note: batch settings are part of the queue configuration
	// The actual batch timeout is configured through the queue settings
}
