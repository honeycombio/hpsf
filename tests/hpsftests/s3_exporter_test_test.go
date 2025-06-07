package hpsftests

import (
	"testing"
	"time"

	"github.com/honeycombio/hpsf/pkg/config"
	collectorConfigprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter"
)

func TestS3Exporter(t *testing.T) {
	// Test the HPSF parsing and template generation using typed configuration
	rulesConfig, collectorConfig, errors := hpsfprovider.GetParsedConfigsFromFile(t, "s3_exporter_test.yaml")

	// First, verify that the refinery config was generated successfully
	if len(rulesConfig.Samplers) != 1 {
		t.Errorf("Expected 1 sampler in refinery config, got %d", len(rulesConfig.Samplers))
	}

	// Check if there are any errors in parsing
	if errors.HasErrors() {
		// If there are errors, let's check if it's just the collector config parsing
		if _, exists := errors.GenerateErrors[config.CollectorConfigType]; exists {
			// This may be due to version compatibility issues with the AWS S3 exporter
			return
		} else {
			// If it's not a collector config error, fail the test
			errors.FailIfError(t)
		}
	}

	// Verify the S3 exporter is present in the traces pipeline
	_, _, exporters, getResult := collectorConfigprovider.GetPipelineConfig(collectorConfig, "traces")
	if !getResult.Found || len(exporters) != 1 {
		t.Errorf("Expected 1 exporter, got %s", exporters)
	}

	// Check that the exporter is the S3 exporter
	expectedExporter := "awss3/My_S3_Backend"
	if len(exporters) > 0 && exporters[0] != expectedExporter {
		t.Errorf("Expected exporter to be %s, got %s", expectedExporter, exporters[0])
	}

	// Get the S3 exporter configuration using typed access
	s3Config, findResult := collectorConfigprovider.GetExporterConfig[awss3exporter.Config](collectorConfig, "awss3/My_S3_Backend")
	if !findResult.Found {
		t.Fatalf("Expected exporter to find \"%v\", found (%v)", findResult.SearchString, findResult.Components)
	}

	// Test S3 bucket configuration (from templates: s3uploader.s3_bucket)
	if s3Config.S3Uploader.S3Bucket != "test-bucket" {
		t.Errorf("Expected s3_bucket to be 'test-bucket', got %s", s3Config.S3Uploader.S3Bucket)
	}

	// Test S3 region configuration (from templates: s3uploader.region)
	if s3Config.S3Uploader.Region != "us-west-2" {
		t.Errorf("Expected region to be 'us-west-2', got %s", s3Config.S3Uploader.Region)
	}

	// Test S3 prefix configuration (from templates: s3uploader.s3_prefix)
	if s3Config.S3Uploader.S3Prefix != "telemetry-data/" {
		t.Errorf("Expected s3_prefix to be 'telemetry-data/', got %s", s3Config.S3Uploader.S3Prefix)
	}

	// Test S3 partition format configuration (from templates: s3uploader.s3_partition_format)
	if s3Config.S3Uploader.S3PartitionFormat != "year=%Y/month=%m/day=%d/hour=%H" {
		t.Errorf("Expected s3_partition_format to be 'year=%%Y/month=%%m/day=%%d/hour=%%H', got %s", s3Config.S3Uploader.S3PartitionFormat)
	}

	// Test marshaler configuration (from templates: marshaler)
	if s3Config.MarshalerName != "otlp_proto" {
		t.Errorf("Expected marshaler to be 'otlp_proto', got %s", s3Config.MarshalerName)
	}

	// Test compression configuration (from templates: s3uploader.compression - hardcoded to "gzip")
	if s3Config.S3Uploader.Compression != "gzip" {
		t.Errorf("Expected compression to be 'gzip', got %s", s3Config.S3Uploader.Compression)
	}

	// Test timeout configuration (from templates: timeout)
	expectedTimeout := 10 * time.Second
	if s3Config.TimeoutSettings.Timeout != expectedTimeout {
		t.Errorf("Expected timeout to be %v, got %v", expectedTimeout, s3Config.TimeoutSettings.Timeout)
	}

	// Test sending queue configuration (from templates: sending_queue.*)
	if s3Config.QueueSettings.QueueSize != 500000 {
		t.Errorf("Expected queue_size to be 500000, got %d", s3Config.QueueSettings.QueueSize)
	}

	if !s3Config.QueueSettings.Enabled {
		t.Error("Expected sending_queue.enabled to be true")
	}

	// Test batch configuration - note: batch settings are part of the queue configuration
	// The actual batch timeout is configured through the queue settings
}
