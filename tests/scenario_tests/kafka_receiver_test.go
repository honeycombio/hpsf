package hpsftests

import (
	"testing"

	collectorConfigprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKafkaReceiver(t *testing.T) {
	// Test the HPSF parsing and Kafka receiver configuration
	rulesConfig, collectorConfig, errors := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/kafka_receiver.yaml")
	errors.FailIfError(t)

	// Verify that the refinery rules config was generated successfully
	assert.Len(t, rulesConfig.Samplers, 1, "Expected 1 sampler in refinery config")

	// Verify the Kafka receiver is present in all pipelines
	tracesReceivers, _, _, getTracesResult := collectorConfigprovider.GetPipelineConfig(collectorConfig, "traces")
	require.True(t, getTracesResult.Found, "Expected traces pipeline to be found")
	assert.Contains(t, tracesReceivers, "kafka/Kafka_Receiver_1", "Expected Kafka receiver in traces pipeline")

	metricsReceivers, _, _, getMetricsResult := collectorConfigprovider.GetPipelineConfig(collectorConfig, "metrics")
	require.True(t, getMetricsResult.Found, "Expected metrics pipeline to be found")
	assert.Contains(t, metricsReceivers, "kafka/Kafka_Receiver_1", "Expected Kafka receiver in metrics pipeline")

	logsReceivers, _, _, getLogsResult := collectorConfigprovider.GetPipelineConfig(collectorConfig, "logs")
	require.True(t, getLogsResult.Found, "Expected logs pipeline to be found")
	assert.Contains(t, logsReceivers, "kafka/Kafka_Receiver_1", "Expected Kafka receiver in logs pipeline")

	// Get the Kafka receiver configuration using typed access
	kafkaConfig, findResult := collectorConfigprovider.GetReceiverConfig[kafkareceiver.Config](collectorConfig, "kafka/Kafka_Receiver_1")
	require.True(t, findResult.Found, "Expected Kafka receiver to be found: %v", findResult.Components)

	// Test broker configuration
	expectedBrokers := []string{"kafka-broker-1:9092", "kafka-broker-2:9092"}
	assert.Equal(t, expectedBrokers, kafkaConfig.Brokers, "Expected custom brokers configuration")

	// Test protocol version
	assert.Equal(t, "2.1.0", kafkaConfig.ProtocolVersion, "Expected default protocol version")

	// Test consumer group configuration
	assert.Equal(t, "test-consumer-group", kafkaConfig.GroupID, "Expected custom group ID")
	assert.Equal(t, "test-client", kafkaConfig.ClientID, "Expected custom client ID")

	// Test topic configuration
	assert.Equal(t, "custom-traces-topic", kafkaConfig.Traces.Topic, "Expected custom traces topic")
	assert.Equal(t, "custom-metrics-topic", kafkaConfig.Metrics.Topic, "Expected custom metrics topic")
	assert.Equal(t, "custom-logs-topic", kafkaConfig.Logs.Topic, "Expected custom logs topic")

	// Test encoding configuration
	assert.Equal(t, "jaeger_json", kafkaConfig.Traces.Encoding, "Expected Jaeger JSON encoding for traces")
	assert.Equal(t, "otlp_json", kafkaConfig.Metrics.Encoding, "Expected OTLP JSON encoding for metrics")
	assert.Equal(t, "json", kafkaConfig.Logs.Encoding, "Expected JSON encoding for logs")

	// Test initial offset configuration
	assert.Equal(t, "earliest", kafkaConfig.InitialOffset, "Expected earliest initial offset")

	// Test session timeout and heartbeat interval
	assert.Equal(t, "10s", kafkaConfig.SessionTimeout.String(), "Expected default session timeout")
	assert.Equal(t, "3s", kafkaConfig.HeartbeatInterval.String(), "Expected default heartbeat interval")

	// Test SASL authentication configuration
	require.NotNil(t, kafkaConfig.Authentication.SASL, "Expected SASL authentication to be configured")
	assert.Equal(t, "KAFKA_USERNAME", kafkaConfig.Authentication.SASL.Username, "Expected SASL username")
	assert.Equal(t, "KAFKA_PASSWORD", string(kafkaConfig.Authentication.SASL.Password), "Expected SASL password")
	assert.Equal(t, "SCRAM-SHA-256", kafkaConfig.Authentication.SASL.Mechanism, "Expected SCRAM-SHA-256 mechanism")

	// Test TLS configuration
	require.NotNil(t, kafkaConfig.Authentication.TLS, "Expected TLS to be configured")
}

func TestKafkaReceiverDefaults(t *testing.T) {
	// Test with minimal configuration to verify defaults
	minimalConfig := `
name: kafka_receiver_minimal
version: v0.1.0
summary: Minimal Kafka receiver test
description: Test with minimal Kafka receiver configuration

components:
  - name: Kafka Receiver Minimal
    kind: KafkaReceiver
  - name: Debug Exporter
    kind: OTelDebugExporter

connections:
  - source:
      component: Kafka Receiver Minimal
      port: Traces
      type: OTelTraces
    destination:
      component: Debug Exporter
      port: Traces
      type: OTelTraces`

	rulesConfig, collectorConfig, errors := hpsfprovider.GetParsedConfigs(t, minimalConfig)
	errors.FailIfError(t)

	// Verify basic configuration
	assert.Len(t, rulesConfig.Samplers, 1, "Expected 1 sampler in refinery config")

	// Get the Kafka receiver configuration
	kafkaConfig, findResult := collectorConfigprovider.GetReceiverConfig[kafkareceiver.Config](collectorConfig, "kafka/Kafka_Receiver_Minimal")
	require.True(t, findResult.Found, "Expected Kafka receiver to be found")

	// Test default values
	assert.Equal(t, []string{"localhost:9092"}, kafkaConfig.Brokers, "Expected default brokers")
	assert.Equal(t, "otel-collector", kafkaConfig.GroupID, "Expected default group ID")
	assert.Equal(t, "otel-collector", kafkaConfig.ClientID, "Expected default client ID")
	assert.Equal(t, "otlp_spans", kafkaConfig.Traces.Topic, "Expected default traces topic")
	assert.Equal(t, "otlp_metrics", kafkaConfig.Metrics.Topic, "Expected default metrics topic")
	assert.Equal(t, "otlp_logs", kafkaConfig.Logs.Topic, "Expected default logs topic")
	assert.Equal(t, "otlp_proto", kafkaConfig.Traces.Encoding, "Expected default traces encoding")
	assert.Equal(t, "otlp_proto", kafkaConfig.Metrics.Encoding, "Expected default metrics encoding")
	assert.Equal(t, "otlp_proto", kafkaConfig.Logs.Encoding, "Expected default logs encoding")
	assert.Equal(t, "latest", kafkaConfig.InitialOffset, "Expected default initial offset")
}

func TestKafkaReceiverValidation(t *testing.T) {
	// Test configuration with invalid encoding - this should parse successfully
	// but would fail at collector runtime validation
	invalidConfig := `
name: kafka_receiver_invalid
version: v0.1.0
summary: Invalid Kafka receiver test
description: Test with invalid Kafka receiver configuration

components:
  - name: Kafka Receiver Invalid
    kind: KafkaReceiver
    properties:
      - name: TracesEncoding
        value: invalid_encoding
  - name: Debug Exporter
    kind: OTelDebugExporter

connections:
  - source:
      component: Kafka Receiver Invalid
      port: Traces
      type: OTelTraces
    destination:
      component: Debug Exporter
      port: Traces
      type: OTelTraces`

	rulesConfig, collectorConfig, errors := hpsfprovider.GetParsedConfigs(t, invalidConfig)

	// The HPSF parsing should succeed, but the invalid encoding will be in the config
	errors.FailIfError(t)

	// Verify basic configuration
	assert.Len(t, rulesConfig.Samplers, 1, "Expected 1 sampler in refinery config")

	// Get the Kafka receiver configuration and verify the invalid encoding is present
	kafkaConfig, findResult := collectorConfigprovider.GetReceiverConfig[kafkareceiver.Config](collectorConfig, "kafka/Kafka_Receiver_Invalid")
	require.True(t, findResult.Found, "Expected Kafka receiver to be found")

	// The invalid encoding should be present in the configuration
	assert.Equal(t, "invalid_encoding", kafkaConfig.Traces.Encoding, "Expected invalid encoding to be preserved")
}
