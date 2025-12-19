package hpsftests

import (
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func TestOTelReceiverCustom(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/otelreceiver_custom.yaml")

	// Verify pipeline config
	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)

	// Check pipeline components
	assert.Len(t, receivers, 1)
	assert.Contains(t, receivers, "otlp/otlp_in")

	assert.Len(t, processors, 2)
	assert.Contains(t, processors, "usage")
	assert.Contains(t, processors, "memory_limiter/otlp_in")

	assert.Len(t, exporters, 1)
	assert.Contains(t, exporters, "otlphttp/otlp_out")

	// Verify receiver config with custom values
	receiverConfig, componentGetResult := collectorprovider.GetReceiverConfig[otlpreceiver.Config](collectorConfig, "otlp/otlp_in")
	require.True(t, componentGetResult.Found, "Expected OTLP receiver in config")
	require.True(t, receiverConfig.GRPC.HasValue(), "Expected gRPC config")
	require.True(t, receiverConfig.HTTP.HasValue(), "Expected HTTP config")

	grpcConfig := receiverConfig.GRPC.Get()
	require.NotNil(t, grpcConfig, "Expected non-nil gRPC config")
	assert.Equal(t, "0.0.0.0:5317", grpcConfig.NetAddr.Endpoint)

	httpConfig := receiverConfig.HTTP.Get()
	require.NotNil(t, httpConfig, "Expected non-nil HTTP config")
	assert.Equal(t, "0.0.0.0:5318", httpConfig.ServerConfig.Endpoint)

	// Verify memory limiter processor config with custom values
	memLimiterConfig, componentGetResult := collectorprovider.GetProcessorConfig[memorylimiterprocessor.Config](collectorConfig, "memory_limiter/otlp_in")
	require.True(t, componentGetResult.Found, "Expected memory_limiter processor in config")
	assert.Equal(t, "2s", memLimiterConfig.CheckInterval.String())
	assert.Equal(t, uint32(75), memLimiterConfig.MemoryLimitPercentage)
	assert.Equal(t, uint32(25), memLimiterConfig.MemorySpikePercentage)
}

func TestOTelReceiverDefaults(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/otelreceiver_defaults.yaml")

	// Verify pipeline config
	tracesPipelineNames := collectorprovider.GetPipelinesByType(collectorConfig, "traces")
	assert.Len(t, tracesPipelineNames, 1)

	receivers, processors, exporters, getResult := collectorprovider.GetPipelineConfig(collectorConfig, tracesPipelineNames[0].String())
	require.True(t, getResult.Found)

	// Check pipeline components
	assert.Len(t, receivers, 1)
	assert.Contains(t, receivers, "otlp/otlp_in")

	assert.Len(t, processors, 2)
	assert.Contains(t, processors, "usage")
	assert.Contains(t, processors, "memory_limiter/otlp_in")

	assert.Len(t, exporters, 1)
	assert.Contains(t, exporters, "otlphttp/otlp_out")

	// Verify receiver config with default values
	receiverConfig, componentGetResult := collectorprovider.GetReceiverConfig[otlpreceiver.Config](collectorConfig, "otlp/otlp_in")
	require.True(t, componentGetResult.Found, "Expected OTLP receiver in config")
	require.True(t, receiverConfig.GRPC.HasValue(), "Expected gRPC config")
	require.True(t, receiverConfig.HTTP.HasValue(), "Expected HTTP config")

	grpcConfig := receiverConfig.GRPC.Get()
	require.NotNil(t, grpcConfig, "Expected non-nil gRPC config")
	assert.Contains(t, grpcConfig.NetAddr.Endpoint, "4317", "Expected default gRPC port")

	httpConfig := receiverConfig.HTTP.Get()
	require.NotNil(t, httpConfig, "Expected non-nil HTTP config")
	assert.Contains(t, httpConfig.ServerConfig.Endpoint, "4318", "Expected default HTTP port")

	// Verify memory limiter processor config with default values
	memLimiterConfig, componentGetResult := collectorprovider.GetProcessorConfig[memorylimiterprocessor.Config](collectorConfig, "memory_limiter/otlp_in")
	require.True(t, componentGetResult.Found, "Expected memory_limiter processor in config")
	assert.Equal(t, "1s", memLimiterConfig.CheckInterval.String())
	assert.Equal(t, uint32(80), memLimiterConfig.MemoryLimitPercentage)
	assert.Equal(t, uint32(20), memLimiterConfig.MemorySpikePercentage)
}
